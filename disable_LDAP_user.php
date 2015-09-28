<?php

include './connection.php';

$ldap_connection = connect_AD();

// Our DN
$ldap_base_dn = 'OU=NanocloudUsers,DC=intra,DC=nanocloud,DC=com';

// Command line parameters
$sam = $argv[1];

// This filter will get the user
$search_filter = '(&(objectCategory=person)(samaccountname=' . $sam . '))';

$result = ldap_search($ldap_connection, $ldap_base_dn, $search_filter);

$count_accounts = ldap_count_entries($ldap_connection, $result);

if ($count_accounts == 1) {

  $account = ldap_get_entries($ldap_connection, $result);
  $dn=$account[0]["dn"];
  $cn=$account[0]["cn"][0];

  $ldaprecord["userPrincipalName"] = $cn . "@demo.com";

  $ldaprecord["objectClass"] = "User";
  $ldaprecord["UserAccountControl"] = "514";

  // Update account
  $r = ldap_modify($ldap_connection, $dn, $ldaprecord);

  if ($r == FALSE) {
    fwrite(STDERR, "An error occurred during LDAP modification\n");
    exit(1);
  }
} else {
  fwrite(STDERR, "An error occurred. SAM account not available\n");
  exit(1);
}

disconnect_AD($ldap_connection);
?>
