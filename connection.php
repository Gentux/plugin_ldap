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

include("configuration.php");

function connect_AD()
{
  global $ldap_server;
  global $ldap_user;
  global $ldap_pass;

  $ldap_connection = ldap_connect($ldap_server) ;

  // We have to set this option for the version of Active Directory we are using.
  ldap_set_option($ldap_connection, LDAP_OPT_PROTOCOL_VERSION, 3) or die('Unable to set LDAP protocol version');
  ldap_set_option($ldap_connection, LDAP_OPT_REFERRALS, 0) or die('Unable to set LDAP referrals');

  $bound = ldap_bind($ldap_connection, $ldap_user, $ldap_pass) ;

  return $ldap_connection ;
}

function disconnect_AD($ldap_connection)
{
  ldap_unbind($ldap_connection) or die('Unable to close LDAP connection');
}
?>
