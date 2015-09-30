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
