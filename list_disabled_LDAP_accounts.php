<?php

include './connection.php';

$ldap_connection = connect_AD();

// This filter will get all the users with disabled account
$search_filter = '(&(objectClass=User)(userAccountControl:1.2.840.113556.1.4.803:=2))';

// Enabled accounts
// '(&(objectClass=User)(!userAccountControl:1.2.840.113556.1.4.803:=2))'

// Query the LDAP server
$result = ldap_search($ldap_connection, $ldap_base_dn, $search_filter);

echo "Number of entries returned is " . ldap_count_entries($ldap_connection, $result) . "\n";

echo "Getting entries ...\n";
$info = ldap_get_entries($ldap_connection, $result);
echo "Data for " . $info["count"] . " items returned:\n";

for ($i=0; $i<$info["count"]; $i++) {
  echo "--------------------------------------------\n";
  echo "dn is: " . $info[$i]["dn"] . "\n";
  echo "first cn entry is: " . $info[$i]["cn"][0] . "\n";
  echo "Samaccountname is: " . $info[$i]["samaccountname"][0] . "\n";
  echo "UserAccountControl is: " . $info[$i]["useraccountcontrol"][0] . "\n";
  $ac = $info[$i]["useraccountcontrol"][0];
  if (($ac & 2)==2) $status="Disabled"; else $status="Enabled";
  echo "User is " . $status . "\n";
}

disconnect_AD($ldap_connection);
?>
