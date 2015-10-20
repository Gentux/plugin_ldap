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

import (
	"testing"

	"encoding/json"
	. "github.com/smartystreets/goconvey/convey"
)

func TestConfigure(t *testing.T) {
	Convey("Should  fill the LDAPConfig struct with the data in the main JSON file", t, func() {
		var l Ldap
		err := l.Configure(` `, nil)
		So(err, ShouldNotEqual, nil)
		err = l.Configure("qzodkqdmojqdodj", nil)
		So(err, ShouldNotEqual, nil)
		err = l.Configure(`{"password":"Nanocloud123+","scriptsDir":"/home/antoine/community/backend/src/nanocloud.com/plugins/ldap/","serverUrl":"ldaps://10.20.12.20","username":"CN=Administrator,CN=Users,DC=intra,DC=localdomain,DC=com"}`, nil)
		So(err, ShouldEqual, nil)
		So(g_LDAPConfig.ServerURL, ShouldEqual, "ldaps://10.20.12.20")
		So(g_LDAPConfig.Username, ShouldEqual, "CN=Administrator,CN=Users,DC=intra,DC=localdomain,DC=com")
		So(g_LDAPConfig.Password, ShouldEqual, "Nanocloud123+")
		So(g_LDAPConfig.ScriptsDir, ShouldEqual, "/home/antoine/community/backend/src/nanocloud.com/plugins/ldap/")

	})
}

func TestLdapAddUser(t *testing.T) {
	var l Ldap

	l.Configure(`{"password":"Nanocloud123+","scriptsDir":"/home/antoine/community/backend/src/nanocloud.com/plugins/ldap/","serverUrl":"ldaps://10.20.12.20","username":"CN=Administrator,CN=Users,DC=intra,DC=localdomain,DC=com"}`, nil)
	Convey("Should create users", t, func() {
		sam := ""
		err := l.AddUser("", nil)
		So(err, ShouldNotEqual, nil)

		Convey("Should create new users, with new SAMACCOUNTS", func() {

			var tests = []struct {
				params   string
				expected bool
			}{
				{`{ "UserEmail" : "lalala1@mail.com", "password" : "aaa" }`, false},
				{`{ "UserEmail" : "lalala2@mail.com", "password" : "$B" }`, false},
				{`{ "UserEmail" : "lalala3@mail.com", "password" : "bonjour" }`, false},
				{`{ "UserEmail" : "lalala4@mail.com", "password" : "bonjourBNSOIR" }`, false},
				{`{ "UserEmail" : "lalala5@mail.com", "password" : "1233221+++++++++" }`, false},
				{`{ "UserEmail" : "lalala6@mail.com", "password" : "BBBBBBBBB++++++" }`, false},
				{`{ "UserEmail" : "lalala7@mail.com", "password" : "bbbbbbbbbbb+++++" }`, false},
				{`{ "UserEmail" : "lalala8@mail.com", "password" : "QOZDJpodjqz2324+--*-" }`, true},
				{`{ "UserEmail" : "lal", "password" : "Ab3++++++++++++" }`, true},
			}

			for _, test := range tests {

				e := l.AddUser(test.params, &sam)
				So((e == nil && !test.expected) || (e != nil && test.expected), ShouldBeFalse)
			}
			err := l.ListUser(`{ "UserEmail" : "lalala6@mail.com", "password" : "PP1923902_"à&é"_é&àè'" }`, &sam)
			So(err, ShouldEqual, nil)
			var res struct {
				Count int
				Users []map[string]string
			}
			err = json.Unmarshal([]byte(sam), &res)
			So(err, ShouldBeNil)
			So(res.Count, ShouldEqual, 2)
			So(res.Users[0]["cn"], ShouldEqual, "1")
			So(res.Users[0]["dn"], ShouldEqual, "CN=1,OU=NanocloudUsers,DC=intra,DC=localdomain,DC=com")
			So(res.Users[0]["mail"], ShouldEqual, "lalala8@mail.com")
			So(res.Users[0]["status"], ShouldEqual, "Enabled")
			So(res.Users[0]["useraccountcontrol"], ShouldEqual, "512")
			So(res.Users[1]["cn"], ShouldEqual, "2")
			So(res.Users[1]["dn"], ShouldEqual, "CN=2,OU=NanocloudUsers,DC=intra,DC=localdomain,DC=com")
			So(res.Users[1]["mail"], ShouldEqual, "lal")
			So(res.Users[1]["status"], ShouldEqual, "Enabled")
			So(res.Users[1]["useraccountcontrol"], ShouldEqual, "512")
		})

		Convey("Should add new users, but by reusing disabled SAMACCOUNTS", func() {
			sam := ""
			err := l.ForceDisableAccount(`{ "UserEmail" : "lalala8@mail.com" }`, &sam)

			So(err, ShouldEqual, nil)
			err = l.ForceDisableAccount(`{ "UserEmail" : "lal" }`, &sam)

			So(err, ShouldEqual, nil)
			err = l.AddUser(`{ "UserEmail" : "lalala10@mail.com", "password" : "FFFggg123+++" }`, &sam)
			So(err, ShouldBeNil)
			err = l.AddUser(`{ "UserEmail" : "lalala11@mail.com", "password" : "CCCfff5667-é&77" }`, &sam)
			So(err, ShouldBeNil)
			err = l.ListUser(`{ "UserEmail" : "lalala6@mail.com", "password" : "PP1923902_"à&é"_é&àè'" }`, &sam)
			So(err, ShouldBeNil)
			var res struct {
				Count int
				Users []map[string]string
			}
			err = json.Unmarshal([]byte(sam), &res)
			So(err, ShouldBeNil)
			So(res.Count, ShouldEqual, 2)

		})

	})
}

func TestListUsers(t *testing.T) {
	var l Ldap
	sam := ""
	Convey("Should list the Users and encode the result in JSON format", t, func() {
		err := l.ListUser("", nil)
		So(err, ShouldNotEqual, nil)
		err = l.ListUser(`{ "UserEmail" : "lalala6@mail.com", "password" : "PP1923902_"à&é"_é&àè'" }`, &sam)
		So(err, ShouldBeNil)
		var res struct {
			Count int
			Users []map[string]string
		}
		err = json.Unmarshal([]byte(sam), &res)
		So(err, ShouldBeNil)
		So(res.Count, ShouldEqual, 2)
		So(res.Users[0]["cn"], ShouldEqual, "1")
		So(res.Users[0]["dn"], ShouldEqual, "CN=1,OU=NanocloudUsers,DC=intra,DC=localdomain,DC=com")
		So(res.Users[0]["mail"], ShouldEqual, "lalala10@mail.com")
		So(res.Users[0]["status"], ShouldEqual, "Enabled")
		So(res.Users[0]["useraccountcontrol"], ShouldEqual, "512")
		So(res.Users[1]["cn"], ShouldEqual, "2")
		So(res.Users[1]["dn"], ShouldEqual, "CN=2,OU=NanocloudUsers,DC=intra,DC=localdomain,DC=com")
		So(res.Users[1]["mail"], ShouldEqual, "lalala11@mail.com")
		So(res.Users[1]["status"], ShouldEqual, "Enabled")
		So(res.Users[1]["useraccountcontrol"], ShouldEqual, "512")

	})

}

func TestForceDisableAccount(t *testing.T) {
	Convey("Should Disable an Account by passing the email of the account", t, func() {
		var l Ldap
		sam := ""
		err := l.ForceDisableAccount("", nil)
		So(err, ShouldNotEqual, nil)
		err = l.DeleteUsers([]string{"CN=1,OU=NanocloudUsers,DC=intra,DC=localdomain,DC=com", "CN=2,OU=NanocloudUsers,DC=intra,DC=localdomain,DC=com"}, &sam)
		So(err, ShouldBeNil)
		err = l.Configure(`{"password":"Nanocloud123+","scriptsDir":"/home/antoine/community/backend/src/nanocloud.com/plugins/ldap/","serverUrl":"ldaps://10.20.12.20","username":"CN=Administrator,CN=Users,DC=intra,DC=localdomain,DC=com"}`, nil)

		So(err, ShouldBeNil)
		err = l.AddUser(`{ "UserEmail" : "lalala12@mail.com", "password" : "Bonjour123+" }`, &sam)
		So(err, ShouldBeNil)
		err = l.ForceDisableAccount(`{ "UserEmail" : "lalala12@mail.com" }`, &sam)
		So(err, ShouldBeNil)
		err = l.ListUser(`{ "UserEmail" : "lalala6@mail.com", "password" : "PP1923902_"à&é"_é&àè'" }`, &sam)
		So(err, ShouldBeNil)
		var res struct {
			Count int
			Users []map[string]string
		}
		err = json.Unmarshal([]byte(sam), &res)
		So(err, ShouldBeNil)
		So(res.Users[0]["status"], ShouldEqual, "Disabled")
		So(res.Users[0]["useraccountcontrol"], ShouldEqual, "514")
		defer l.DeleteUsers([]string{"CN=1,OU=NanocloudUsers,DC=intra,DC=localdomain,DC=com"}, &sam)

	})

}

func TestDisableAccount(t *testing.T) {
	Convey("Should Disable an Account by passing the SAMACCOUNT of the account", t, func() {
		var l Ldap
		err := l.DisableAccount("", nil)
		So(err, ShouldNotEqual, nil)
		l.Configure(`{"password":"Nanocloud123+","scriptsDir":"/home/antoine/community/backend/src/nanocloud.com/plugins/ldap/","serverUrl":"ldaps://10.20.12.20","username":"CN=Administrator,CN=Users,DC=intra,DC=localdomain,DC=com"}`, nil)
		sam := ""
		l.AddUser(`{ "UserEmail" : "lalala13@mail.com", "password" : "Bonjour123+" }`, &sam)
		l.ListUser(`{ "UserEmail" : "lalala6@mail.com", "password" : "PP1923902_"à&é"_é&àè'" }`, &sam)
		var res struct {
			Count int
			Users []map[string]string
		}
		err = json.Unmarshal([]byte(sam), &res)
		So(err, ShouldBeNil)
		err = l.DisableAccount(`{ "UserEmail" : "`+res.Users[0]["samaccountname"]+`", "password" : "PP1923902++++rf" }`, &sam)
		So(err, ShouldBeNil)
		l.ListUser(`{ "UserEmail" : "lalala6@mail.com", "password" : "PP1923902_"à&é"_é&àè'" }`, &sam)
		err = json.Unmarshal([]byte(sam), &res)
		So(err, ShouldBeNil)
		So(res.Users[0]["useraccountcontrol"], ShouldEqual, "514")
		defer l.DeleteUsers([]string{"CN=1,OU=NanocloudUsers,DC=intra,DC=localdomain,DC=com"}, &sam)
	})
}

func TestModifyPassword(t *testing.T) {
	Convey("Should Modify the account's password", t, func() {
		var l Ldap
		err := l.Configure(`{"password":"Nanocloud123+","scriptsDir":"/home/antoine/community/backend/src/nanocloud.com/plugins/ldap/","serverUrl":"ldaps://10.20.12.20","username":"CN=Administrator,CN=Users,DC=intra,DC=localdomain,DC=com"}`, nil)
		So(err, ShouldBeNil)
		sam := ""
		err = l.AddUser(`{ "UserEmail" : "lalala15@mail.com", "password" : "Bonjour123+" }`, &sam)
		So(err, ShouldBeNil)
		l.ListUser(`{ "UserEmail" : "lalala6@mail.com", "password" : "PP1923902_"à&é"_é&àè'" }`, &sam)
		var res struct {
			Count int
			Users []map[string]string
		}
		err = json.Unmarshal([]byte(sam), &res)
		So(err, ShouldBeNil)
		err = l.ModifyPassword("", nil)
		So(err, ShouldNotBeNil)
		err = l.ModifyPassword(`{ "UserEmail" : "`+res.Users[0]["samaccountname"]+`", "password" : "PP1923902++++rf" }`, &sam)
		So(err, ShouldBeNil)
		err = l.DeleteUsers([]string{"CN=1,OU=NanocloudUsers,DC=intra,DC=localdomain,DC=com"}, &sam)
		So(err, ShouldBeNil)

	})
}

func TestTestPassword(t *testing.T) {
	Convey("Should test if a password meets the requirements policy of Windows AD", t, func() {
		var res = []struct {
			pass     string
			expected bool
		}{
			{"aaa", false},
			{"aaaBBBBBB", false},
			{"aaa+++++++", false},
			{"aaaBBBBB9999", true},
			{"Ba7+", false},
			{"", false},
			{"  ", false},
			{"Bb33111", true},
		}

		for _, test := range res {
			So(test_password(test.pass), ShouldEqual, test.expected)
		}
	})
}
