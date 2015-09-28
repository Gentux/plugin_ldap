package main

import (
	"fmt"

	nan "nanocloud.com/lib/libnan"

	"os/exec"
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
