<?php
/*
 * Nanocloud Community, a comprehensive platform to turn any application
 * into a cloud solution.
 *
 * Copyright (C) 2015 Nanocloud Software
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program. If not, see <http://www.gnu.org/licenses/>.
 */

include './connection.php';

$ldap_connection = connect_AD();

// Our DN
$ldap_base_dn = 'OU=NanocloudUsers,DC=intra,DC=nanocloud,DC=com';


// Command line parameters
$user_email = $argv[1];
$password = $argv[2];

$ldaprecord = array(
  "mail" => $user_email,
  "givenName" => $user_email,
  "userPrincipalName" => $user_email,
  "objectClass" => "User",
  "unicodePwd" => mb_convert_encoding('"' . $password . '"', 'utf-16le'),
  "UserAccountControl" => "512",
);

// This filter will get all the users with disabled account
$search_filter = '(&(objectClass=User)(userAccountControl:1.2.840.113556.1.4.803:=2))';
$result = ldap_search($ldap_connection, $ldap_base_dn, $search_filter);
$count_disabled_account = ldap_count_entries($ldap_connection, $result);

if ($count_disabled_account) {

  $disabled_accounts = ldap_get_entries($ldap_connection, $result);
  $dn = $disabled_accounts[0]["dn"];
  $sam_account_name = $disabled_accounts[0]["samaccountname"][0];

  // Update account
  $r = ldap_modify($ldap_connection, $dn, $ldaprecord);
  if ($r == FALSE) {
    fwrite(STDERR, "An error occurred during LDAP account update.\n");
    exit(1);
  }
} else {
  // This filter will get all the users
  $search_filter = '(&(objectCategory=person)(samaccountname=*))';
  $result = ldap_search($ldap_connection, $ldap_base_dn, $search_filter);

  $count_users = ldap_count_entries($ldap_connection, $result);
  $cn = "demo" . sprintf('%04d', ++$count_users);
  $dn = "CN=$cn,OU=NanocloudUsers,DC=intra,DC=nanocloud,DC=com";

  $ldaprecord["CN"] = $cn;
  $ldaprecord["givenName"] = $cn;
  $ldaprecord["userPrincipalName"] = $cn;

  // Insert new account
  $r = ldap_add($ldap_connection, $dn, $ldaprecord);
  if ($r == FALSE) {
    fwrite(STDERR, "An error occurred.\n");
    exit(1);
  }

  $sr = ldap_search($ldap_connection,"OU=NanocloudUsers,DC=intra,DC=nanocloud,DC=com","cn=$cn");
  $info = ldap_get_entries($ldap_connection,$sr);
  $sam_account_name =  $info[0]["samaccountname"][0];
}

echo $sam_account_name;

disconnect_AD($ldap_connection);
?>
