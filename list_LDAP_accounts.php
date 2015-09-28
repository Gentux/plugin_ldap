<?php

include './connection.php';

$ldap_connection = connect_AD();

// Our DN
$ldap_base_dn = 'OU=NanocloudUsers,DC=intra,DC=nanocloud,DC=com';

// This filter will get all the users
$search_filter = '(&(objectCategory=person)(samaccountname=*))';

// Query the LDAP server
$ldap_result = ldap_search($ldap_connection, $ldap_base_dn, $search_filter);


$result = array();
$result['count'] = ldap_count_entries($ldap_connection, $ldap_result);

$result['users'] = array();

$info = ldap_get_entries($ldap_connection, $ldap_result);

for ($i=0; $i<$info["count"]; $i++) {
  $user = array(
    "dn" => $info[$i]["dn"],
    "cn" => $info[$i]["cn"][0],
    "mail" => $info[$i]["mail"][0],
    "samaccountname" => $info[$i]["samaccountname"][0],
    "useraccountcontrol" => $info[$i]["useraccountcontrol"][0]
  );

  $account_control = $info[$i]["useraccountcontrol"][0];
  if (($account_control & 2) == 2) {
    $user["status"] = "Disabled";
  } else {
    $user["status"] = "Enabled";
  }

  array_push($result['users'], $user);
}

echo json_encode($result);

disconnect_AD($ldap_connection);
?>
