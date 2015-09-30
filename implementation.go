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
	"fmt"

	"os/exec"

	//todo vendor this dependency
	nan "nanocloud.com/plugins/ldap/libnan"
)

var ()

// Improve naming/documentation of this procedure which is in fact a function
func ImpLdapForceDisableAccount(_Email string) string {
	sPhpScript := fmt.Sprintf("force_disable_LDAP_user.php")
	cmd := exec.Command("/usr/bin/php", "-f", sPhpScript, "--", _Email)

	ans := nan.Err{}

	out, err := cmd.Output()

	if err != nil {
		ans.Code = 0
		ans.Details = "Failed to run script force_disable_LDAP_user.php, error: " + err.Error()
	} else {
		ans.Code = 1
		ans.Details = fmt.Sprint("LDAP Check... %s account(s) disabled", string(out))
	}

	return ans.ToJson()
}

func ImpLdapDisableAccount(_Sam string) string {
	sPhpScript := "disable_LDAP_user.php"
	cmd := exec.Command("/usr/bin/php", "-f", sPhpScript, "--", _Sam)

	ans := nan.Err{}

	_, err := cmd.Output()
	if err != nil {
		ans.Code = 0
		ans.Details = "Error returned by script disable_LDAP_user.php, error: " + err.Error()
	} else {
		ans.Code = 1
	}

	return ans.ToJson()
}
