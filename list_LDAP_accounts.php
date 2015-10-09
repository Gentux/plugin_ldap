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
$ldap_base_dn = 'OU=NanocloudUsers,DC=intra,DC=localdomain,DC=com';

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
