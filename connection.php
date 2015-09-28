<?php

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
