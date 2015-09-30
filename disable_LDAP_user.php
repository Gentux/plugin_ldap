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
