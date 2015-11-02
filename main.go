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

package main

/*#include <ldap.h>
#include <stdlib.h>
#include <sys/time.h>
#include <stdio.h>
#include <lber.h>
typedef struct ldapmod_str {
	int	 mod_op;
	char	  *mod_type;
	char    **mod_vals;
} LDAPModStr;
int _ldap_add(LDAP *ld, char* dn, LDAPModStr **attrs){
	return ldap_add_ext_s(ld, dn, (LDAPMod **)attrs, NULL, NULL);
}
*/
// #cgo CFLAGS: -DLDAP_DEPRECATED=1
// #cgo linux CFLAGS: -DLINUX=1
// #cgo LDFLAGS: -lldap -llber
import "C"

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/ldap.v2"
	"log"
	"os/exec"
	"strconv"
	"unicode"
	"unicode/utf16"
	"unsafe"

	"github.com/dullgiulio/pingo"

	nan "nanocloud.com/plugins/ldap/libnan"
)

const (
	LDAP_OPT_SUCCESS          = 0
	LDAP_OPT_ERROR            = -1
	LDAP_VERSION3             = 3
	LDAP_OPT_PROTOCOL_VERSION = 0x0011
	LDAP_SUCCESS              = 0x00
	LDAP_NO_LIMIT             = 0
	LDAP_OPT_REFERRALS        = 0x0008
	LDAP_MOD_REPLACE          = 0x0002
)

const (
	LDAP_SCOPE_BASE        = 0x0000
	LDAP_SCOPE_ONELEVEL    = 0x0001
	LDAP_SCOPE_SUBTREE     = 0x0002
	LDAP_SCOPE_SUBORDINATE = 0x0003 // OpenLDAP extension
	LDAP_SCOPE_DEFAULT     = -1     // OpenLDAP extension
)

type LDAPConfig struct {
	ScriptsDir string
	Username   string
	Password   string
	ServerURL  string
}

type AccountParams struct {
	UserEmail string
	Password  string
}

type ChangePasswordParams struct {
	SamAccountName string
	NewPassword    string
}

type Ldap struct{}

type ldap_conf struct {
	ldapConnection *C.LDAP
	host           string
	login          string
	passwd         string
	ou             string
}

var (
	g_LDAPConfig LDAPConfig
)

func SetOptions(ldapConnection *C.LDAP, pOutMsg *string) error {
	// Setting LDAP version and referrals
	var version C.int
	var opt C.int
	version = LDAP_VERSION3
	opt = 0
	err := C.ldap_set_option(ldapConnection, LDAP_OPT_PROTOCOL_VERSION, unsafe.Pointer(&version))
	if err != LDAP_SUCCESS {
		return answerWithError(pOutMsg, "Options settings error: "+C.GoString(C.ldap_err2string(err)), nil)
	}
	err = C.ldap_set_option(ldapConnection, LDAP_OPT_REFERRALS, unsafe.Pointer(&opt))
	if err != LDAP_SUCCESS {
		return answerWithError(pOutMsg, "Options settings error: "+C.GoString(C.ldap_err2string(err)), nil)
	}
	return nil

}

func answerWithError(pOutString *string, msg string, e error) error {

	// TODO return JSON answer
	// r := nan.NewExitCode(0, "ERROR: failed to  : "+err.Error())
	// log.Printf(r.Message) // for on-screen debug output
	// *pOutMsg = r.ToJson() // return codes for IPC should use JSON as much as possible
	answer := "plugin Ldap: " + msg
	if e != nil {
		answer += e.Error()
	}

	log.Printf(answer)
	if pOutString != nil {
		*pOutString = answer
	}
	return errors.New(answer)
}

func (p *Ldap) Configure(jsonConfig string, pOutMsg *string) error {
	// Initialization of values needed for LDAP connection
	var ldapConfig map[string]string

	if e := json.Unmarshal([]byte(jsonConfig), &ldapConfig); e != nil {
		return answerWithError(pOutMsg, "Configure() failed to unmarshal LDAP plugin configuration : ", e)
	}

	g_LDAPConfig.ServerURL = ldapConfig["serverUrl"]
	g_LDAPConfig.Username = ldapConfig["username"]
	g_LDAPConfig.Password = ldapConfig["password"]
	g_LDAPConfig.ScriptsDir = ldapConfig["scriptsDir"]

	return nil
}

func (p *Ldap) ListUser(jsonParams string, pOutMsg *string) error {
	if pOutMsg == nil {
		return answerWithError(pOutMsg, "PoutMsg is nil", nil)
	}
	ldapConnection, err := ldap.DialTLS("tcp", g_LDAPConfig.ServerURL[8:]+":636",
		&tls.Config{
			InsecureSkipVerify: true,
		})
	err = ldapConnection.Bind(g_LDAPConfig.Username, g_LDAPConfig.Password)
	if err != nil {
		return answerWithError(pOutMsg, "Binding error: ", err)
	}

	defer ldapConnection.Close()
	searchRequest := ldap.NewSearchRequest(
		"OU=NanocloudUsers,DC=intra,DC=localdomain,DC=com",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectCategory=person)(objectGUID=*))",
		[]string{"dn", "cn", "mail", "sAMAccountName", "userAccountControl"},
		nil,
	)
	sr, err := ldapConnection.Search(searchRequest)
	if err != nil {
		return answerWithError(pOutMsg, "Search error: ", err)
	}
	// Struct needed for JSON encoding
	var res struct {
		Count int
		Users []map[string]string
	}
	res.Count = len(sr.Entries)
	res.Users = make([]map[string]string, res.Count)
	i := 0

	for _, entry := range sr.Entries {
		res.Users[i] = make(map[string]string, 6)
		res.Users[i]["dn"] = entry.DN
		res.Users[i]["cn"] = entry.GetAttributeValue("cn")
		res.Users[i]["mail"] = entry.GetAttributeValue("mail")
		res.Users[i]["samaccountname"] = entry.GetAttributeValue("sAMAccountName")
		res.Users[i]["useraccountcontrol"] = entry.GetAttributeValue("userAccountControl")
		h, _ := strconv.Atoi(res.Users[i]["useraccountcontrol"])
		if h&0x0002 == 0 { // 0x0002 activated means user is disabled
			res.Users[i]["status"] = "Enabled"
		} else {
			res.Users[i]["status"] = "Disabled"

		}

		i++
	}
	g, _ := json.Marshal(res)
	*pOutMsg = string(g)
	return nil
}

func test_password(pass string) bool {
	// Windows AD password needs at leat 7 characters password,  and must contain characters from three of the following five categories:
	// uppercase character
	// lowercase character
	// digit character
	// nonalphanumeric characters
	// any Unicode character that is categorized as an alphabetic character but is not uppercase or lowercase
	if len(pass) < 7 {
		return false
	}
	d := 0
	l := 0
	u := 0
	p := 0
	o := 0
	for _, c := range pass {
		if unicode.IsDigit(c) { // check digit character
			d = 1
		} else if unicode.IsLower(c) { // check lowercase character
			l = 1
		} else if unicode.IsUpper(c) { // check uppercase character
			u = 1
		} else if unicode.IsPunct(c) { // check nonalphanumeric character
			p = 1
		} else { // other unicode character
			o = 1
		}
	}
	if d+l+u+p+o < 3 {
		return false
	}
	return true
}

func (p *Ldap) DeleteUsers(mails []string, pOutMsg *string) error {
	var conf ldap_conf

	conf.host = g_LDAPConfig.ServerURL
	conf.login = g_LDAPConfig.Username
	conf.passwd = g_LDAPConfig.Password
	var version C.int
	var v C.int
	version = LDAP_VERSION3
	v = 0
	err := C.ldap_set_option(conf.ldapConnection, LDAP_OPT_PROTOCOL_VERSION, unsafe.Pointer(&version))
	if err != LDAP_SUCCESS {
		return answerWithError(pOutMsg, "Options settings error: "+C.GoString(C.ldap_err2string(err)), nil)
	}

	err = C.ldap_set_option(conf.ldapConnection, LDAP_OPT_REFERRALS, unsafe.Pointer(&v))
	if err != LDAP_SUCCESS {
		return answerWithError(pOutMsg, "Deletion error: "+C.GoString(C.ldap_err2string(err)), nil)
	}

	rc := C.ldap_initialize(&conf.ldapConnection, C.CString(conf.host+":636"))
	if conf.ldapConnection == nil {
		return answerWithError(pOutMsg, "Initialization error: ", nil)
	}
	rc = C.ldap_simple_bind_s(conf.ldapConnection, C.CString(conf.login), C.CString(conf.passwd))
	if rc != LDAP_SUCCESS {
		return answerWithError(pOutMsg, "Binding error: "+C.GoString(C.ldap_err2string(rc)), nil)
	}
	c := 0
	for c < len(mails) {
		rc := C.ldap_delete_s(conf.ldapConnection, C.CString(mails[c]))
		if rc != 0 {
			return answerWithError(pOutMsg, "Deletion error: "+C.GoString(C.ldap_err2string(rc)), nil)
		}
		c++
	}
	return nil

}

func Initialize(conf *ldap_conf, pOutMsg *string) error {

	if SetOptions(nil, pOutMsg) != nil {
		return answerWithError(pOutMsg, "Options error", nil)
	}
	rc := C.ldap_initialize(&conf.ldapConnection, C.CString(conf.host+":636"))
	if conf.ldapConnection == nil {
		return answerWithError(pOutMsg, "Initialization error: "+C.GoString(C.ldap_err2string(rc)), nil)
	}
	rc = C.ldap_simple_bind_s(conf.ldapConnection, C.CString(conf.login), C.CString(conf.passwd))
	if rc != LDAP_SUCCESS {
		return answerWithError(pOutMsg, "Binding error: "+C.GoString(C.ldap_err2string(rc)), nil)
	}
	return nil

}

func CheckSamAvailability(ldapConnection *ldap.Conn, pOutMsg *string) (error, string, int) {
	searchRequest := ldap.NewSearchRequest(
		"OU=NanocloudUsers,DC=intra,DC=localdomain,DC=com",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectCategory=person)(objectGUID=*))",
		[]string{"dn", "cn", "mail", "sAMAccountName", "userAccountControl"},
		nil,
	)

	sr, err := ldapConnection.Search(searchRequest)
	if err != nil {
		return answerWithError(pOutMsg, "Search error: ", err), "", 0
	}
	count := len(sr.Entries)
	cn := ""
	for _, entry := range sr.Entries {
		h, err := strconv.Atoi(entry.GetAttributeValue("userAccountControl"))
		if err != nil {
			return answerWithError(pOutMsg, "Atoi conversion error: ", err), "", 0
		}
		if h&0x0002 == 0 { //0x0002 means disabled account
		} else {
			cn = entry.GetAttributeValue("cn")
			break
		}
	}
	return nil, cn, count
}

func CreateNewUser(conf ldap_conf, pOutMsg *string, params AccountParams, count int, mods [3]*C.LDAPModStr, ldapConnection *ldap.Conn) error {
	if pOutMsg == nil {
		return answerWithError(pOutMsg, "PoutMsg is nil", nil)
	}
	if !test_password(params.Password) {

		return answerWithError(pOutMsg, "Password does not meet minimum requirements", nil)

	}
	dn := "cn=" + fmt.Sprintf("%d", count+1) + "," + conf.ou

	rc := C._ldap_add(conf.ldapConnection, C.CString(dn), &mods[0])

	if rc != LDAP_SUCCESS {
		return answerWithError(pOutMsg, "Adding error: "+C.GoString(C.ldap_err2string(rc)), nil)
	}
	pwd := EncodePassword(params.Password)
	modify := ldap.NewModifyRequest("cn=" + fmt.Sprintf("%d", count+1) + ",OU=NanocloudUsers,DC=intra,DC=localdomain,DC=com")
	modify.Replace("unicodePwd", []string{string(pwd)}) // field where the windows password is stored
	modify.Replace("userAccountControl", []string{"512"})
	err := ldapConnection.Modify(modify)
	if err != nil {
		return answerWithError(pOutMsg, "Modify error: ", err)
	}
	return nil

}

func EncodePassword(pass string) []byte {
	s := pass
	// Windows AD needs a UTF16-LE encoded password, with double quotes at the beginning and at the end
	enc := utf16.Encode([]rune(s))
	pwd := make([]byte, len(enc)*2+4)

	pwd[0] = '"'
	i := 2
	for _, n := range enc {
		pwd[i] = byte(n)
		pwd[i+1] = byte(n >> 8)
		i += 2
	}
	pwd[i] = '"'
	return pwd
}

func RecycleSam(params AccountParams, ldapConnection *ldap.Conn, pOutMsg *string, cn string) error {
	if pOutMsg == nil {
		return answerWithError(pOutMsg, "PoutMsg is nil", nil)
	}
	pwd := EncodePassword(params.Password)
	modify := ldap.NewModifyRequest("cn=" + cn + ",OU=NanocloudUsers,DC=intra,DC=localdomain,DC=com")
	modify.Replace("unicodePwd", []string{string(pwd)})
	modify.Replace("userAccountControl", []string{"512"})
	modify.Replace("mail", []string{params.UserEmail})
	err := ldapConnection.Modify(modify)
	if err != nil {
		return answerWithError(pOutMsg, "Modify error: ", err)
	}
	return nil
}

func (p *Ldap) ModifyPassword(jsonParams string, pOutMsg *string) error {
	if pOutMsg == nil {
		return answerWithError(pOutMsg, "PoutMsg is nil", nil)
	}
	*pOutMsg = "0" // return code meaning failure of operation

	var params AccountParams

	if e := json.Unmarshal([]byte(jsonParams), &params); e != nil {
		return answerWithError(pOutMsg, "AddUser() failed: ", e)
	}
	bindusername := g_LDAPConfig.Username
	bindpassword := g_LDAPConfig.Password
	c := 0
	for i, val := range g_LDAPConfig.ServerURL { //Passing letters/symbols before IP adress ( ex : ldaps:// )
		if unicode.IsDigit(val) {
			c = i
			break
		}
	}
	ldapConnection, err := ldap.DialTLS("tcp", g_LDAPConfig.ServerURL[c:]+":636",
		&tls.Config{
			InsecureSkipVerify: true,
		})

	err = ldapConnection.Bind(bindusername, bindpassword)
	if err != nil {
		return answerWithError(pOutMsg, "Binding error: ", err)
	}

	defer ldapConnection.Close()

	searchRequest := ldap.NewSearchRequest(
		"OU=NanocloudUsers,DC=intra,DC=localdomain,DC=com",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectCategory=person)(sAMAccountName="+params.UserEmail+"))",
		[]string{"dn", "cn", "mail", "sAMAccountName", "userAccountControl"},
		nil,
	)

	sr, err := ldapConnection.Search(searchRequest)
	if err != nil {
		return answerWithError(pOutMsg, "Search error: ", err)
	}

	var cn string
	if len(sr.Entries) != 1 {
		return answerWithError(pOutMsg, "invalid sAMAccountName", nil)
	}
	for _, entry := range sr.Entries {
		cn = entry.GetAttributeValue("cn")

	}
	pwd := EncodePassword(params.Password)

	modify := ldap.NewModifyRequest("cn=" + cn + ",OU=NanocloudUsers,DC=intra,DC=localdomain,DC=com")
	modify.Replace("unicodePwd", []string{string(pwd)})
	err = ldapConnection.Modify(modify)
	if err != nil {
		return answerWithError(pOutMsg, "Password modification failed: ", err)
	}
	return nil

}

func (p *Ldap) AddUser(jsonParams string, pOutMsg *string) error {
	if pOutMsg == nil {
		return answerWithError(pOutMsg, "PoutMsg is nil", nil)
	}
	*pOutMsg = "0" // return code meaning failure of operation

	var params AccountParams

	if e := json.Unmarshal([]byte(jsonParams), &params); e != nil {
		return answerWithError(pOutMsg, "AddUser() failed: ", e)
	}
	// OpenLDAP and CGO needed here to add a new user
	var conf ldap_conf
	conf.host = g_LDAPConfig.ServerURL
	conf.login = g_LDAPConfig.Username
	conf.passwd = g_LDAPConfig.Password
	conf.ou = "OU=NanocloudUsers,DC=intra,DC=localdomain,DC=com"
	err := Initialize(&conf, pOutMsg)
	if err != nil {
		return err
	}
	var mods [3]*C.LDAPModStr
	var modClass, modCN C.LDAPModStr
	var vclass [5]*C.char
	var vcn [4]*C.char
	modClass.mod_op = 0
	modClass.mod_type = C.CString("objectclass")
	vclass[0] = C.CString("top")
	vclass[1] = C.CString("person")
	vclass[2] = C.CString("organizationalPerson")
	vclass[3] = C.CString("User")
	vclass[4] = nil
	modClass.mod_vals = &vclass[0]

	modCN.mod_op = 0
	modCN.mod_type = C.CString("mail")
	vcn[0] = C.CString(params.UserEmail)
	vcn[1] = nil
	modCN.mod_vals = &vcn[0]

	mods[0] = &modClass
	mods[1] = &modCN
	mods[2] = nil

	bindusername := g_LDAPConfig.Username
	bindpassword := g_LDAPConfig.Password
	// Return to ldap go API to set the password
	c := 0
	for i, val := range g_LDAPConfig.ServerURL { //Passing letters/symbols before IP adress ( ex : ldaps:// )
		if unicode.IsDigit(val) {
			c = i
			break
		}
	}
	ldapConnection, err := ldap.DialTLS("tcp", g_LDAPConfig.ServerURL[c:]+":636",
		&tls.Config{
			InsecureSkipVerify: true,
		})
	if err != nil {
		return answerWithError(pOutMsg, "DialTLS failed: ", err)
	}
	err = ldapConnection.Bind(bindusername, bindpassword)
	if err != nil {
		return answerWithError(pOutMsg, "Binding error: ", err)
	}

	defer ldapConnection.Close()

	err, cn, count := CheckSamAvailability(ldapConnection, pOutMsg) // If an account is disabled, this function will look for his CN
	if err != nil {
		return err
	}

	// If no disabled accounts were found, real new user created
	if cn == "" {
		err = CreateNewUser(conf, pOutMsg, params, count, mods, ldapConnection)
		if err != nil {
			return err
		}
		// Freeing various structures needed for adding entry with OpenLDAP
		C.free(unsafe.Pointer(vclass[0]))
		C.free(unsafe.Pointer(vclass[1]))
		C.free(unsafe.Pointer(vclass[2]))
		C.free(unsafe.Pointer(vclass[3]))
		C.free(unsafe.Pointer(vcn[0]))
		C.free(unsafe.Pointer(modCN.mod_type))
		//C._ldap_mods_free(&mods[0], 1)   Should work but doesnt...
	} else {
		// If a disabled account is found, modifying this account instead of creating a new one
		err = RecycleSam(params, ldapConnection, pOutMsg, cn)
		if err != nil {
			return err
		}

	}

	*pOutMsg = params.UserEmail + " added"

	return nil
}

func (p *Ldap) ForceDisableAccount(jsonParams string, pOutMsg *string) error {
	if pOutMsg == nil {
		return answerWithError(pOutMsg, "PoutMsg is nil", nil)
	}
	*pOutMsg = "0" // return code meaning failure of operation

	var params AccountParams

	if e := json.Unmarshal([]byte(jsonParams), &params); e != nil {
		return answerWithError(pOutMsg, "ForceDisableAccount() failed ", e)
	}

	bindusername := g_LDAPConfig.Username
	bindpassword := g_LDAPConfig.Password
	c := 0
	for i, val := range g_LDAPConfig.ServerURL { //Passing letters/symbols before IP adress ( ex : ldaps:// )
		if unicode.IsDigit(val) {
			c = i
			break
		}
	}
	ldapConnection, err := ldap.DialTLS("tcp", g_LDAPConfig.ServerURL[c:]+":636",
		&tls.Config{
			InsecureSkipVerify: true,
		})

	if err != nil {
		return answerWithError(pOutMsg, "DialTLS error: ", err)
	}
	err = ldapConnection.Bind(bindusername, bindpassword)
	if err != nil {
		return answerWithError(pOutMsg, "Binding error: ", err)
	}
	defer ldapConnection.Close()
	searchRequest := ldap.NewSearchRequest(
		"OU=NanocloudUsers,DC=intra,DC=localdomain,DC=com",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectCategory=person)(mail="+params.UserEmail+"))",
		[]string{"userAccountControl", "cn"},
		nil,
	)

	sr, err := ldapConnection.Search(searchRequest)
	if err != nil {
		return answerWithError(pOutMsg, "Searching error: ", err)
	}

	if len(sr.Entries) != 1 {
		// Means entered mail was not valid, or several user have the same mail ?
		return answerWithError(pOutMsg, "Email does not match any user, or several users have the same mail adress", nil)
	} else {
		var cn string
		for _, entry := range sr.Entries {
			cn = entry.GetAttributeValue("cn")
		}
		modify := ldap.NewModifyRequest("cn=" + cn + ",OU=NanocloudUsers,DC=intra,DC=localdomain,DC=com")
		modify.Replace("userAccountControl", []string{"514"}) // 512 is a normal account, 514 is disabled ( 512 + 0x0002 )
		err = ldapConnection.Modify(modify)
		if err != nil {
			return answerWithError(pOutMsg, "Modify error: ", err)
		}
		*pOutMsg = "1" //success

	}
	return nil
}

func (p *Ldap) DisableAccount(jsonParams string, pOutMsg *string) error {
	if pOutMsg == nil {
		return answerWithError(pOutMsg, "PoutMsg is nil", nil)
	}
	*pOutMsg = "0" // return code meaning failure of operation

	var params AccountParams

	if err := json.Unmarshal([]byte(jsonParams), &params); err != nil {
		r := nan.NewExitCode(0, "ERROR: failed to unmarshal Ldap.AccountParams : "+err.Error())
		log.Printf(r.Message)
		*pOutMsg = r.ToJson() // return codes for IPC should use JSON as much as possible
		return nil
	}

	bindusername := g_LDAPConfig.Username
	bindpassword := g_LDAPConfig.Password
	c := 0
	for i, val := range g_LDAPConfig.ServerURL { //Passing letters/symbols before IP adress ( ex : ldaps:// )
		if unicode.IsDigit(val) {
			c = i
			break
		}
	}
	ldapConnection, err := ldap.DialTLS("tcp", g_LDAPConfig.ServerURL[c:]+":636",
		&tls.Config{
			InsecureSkipVerify: true,
		})
	if err != nil {
		return answerWithError(pOutMsg, "DialTLS error: ", err)
	}
	err = ldapConnection.Bind(bindusername, bindpassword)
	if err != nil {
		return answerWithError(pOutMsg, "Binding error: ", err)
	}

	defer ldapConnection.Close()
	searchRequest := ldap.NewSearchRequest(
		"OU=NanocloudUsers,DC=intra,DC=localdomain,DC=com",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectCategory=person)(samaccountname="+params.UserEmail+"))",
		[]string{"userAccountControl", "cn"},
		nil,
	)
	sr, err := ldapConnection.Search(searchRequest)
	if err != nil {
		return answerWithError(pOutMsg, "Search error: ", err)
	}

	if len(sr.Entries) != 1 { //wrong samaccount
		return answerWithError(pOutMsg, "SAMACCOUNT does not match any user", nil)
	} else {
		var cn string
		for _, entry := range sr.Entries {
			cn = entry.GetAttributeValue("cn")
		}

		modify := ldap.NewModifyRequest("cn=" + cn + ",OU=NanocloudUsers,DC=intra,DC=localdomain,DC=com")
		modify.Replace("userAccountControl", []string{"514"})
		err = ldapConnection.Modify(modify)
		if err != nil {
			return answerWithError(pOutMsg, "Modify error: ", err)
		}

	}
	*pOutMsg = "1" //success
	return nil
}

func (p *Ldap) ChangePassword(jsonParams string, pOutMsg *string) error {
	var params ChangePasswordParams
	*pOutMsg = "0"

	if e := json.Unmarshal([]byte(jsonParams), &params); e != nil {
		return answerWithError(pOutMsg, "ChangePassword() failed ", e)
	}

	cmd := exec.Command("/usr/bin/php", "changepw_LDAP_user.php", params.SamAccountName, params.NewPassword)
	cmd.Dir = g_LDAPConfig.ScriptsDir

	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("Failed to run script changepw_LDAP_user.php for sam <%s>, error: %s, output: %s\n", params.SamAccountName, err, string(out))
		return err
	}

	*pOutMsg = "1"
	return nil
}

func main() {
	plugin := &Ldap{}

	pingo.Register(plugin)

	pingo.Run()
}
