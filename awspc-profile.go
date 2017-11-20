package main

import (
	"fmt"
	"github.com/go-ini/ini"
	"github.com/shiena/ansicolor"
	"os"
	"os/exec"
	"sort"
	"strings"
)

/**
 * checks the profile against aws
 * returns {error} eroor
 * @param {string} profile (named profile)
**/
func check_profile(profile string) error {

	toOut := ansicolor.NewAnsiColorWriter(os.Stdout)

	cmd := exec.Command("aws", "sts", "get-caller-identity", "--profile", profile)
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(toOut, "%sfailed%s\n", cred, coff)
		return err
	} else {
		fmt.Fprintf(toOut, "%sOK%s\n", cgreen, coff)
	}
	return err
}

/**
* list all profilenames
* returns (int) number of profiles
* @param {*ini.Files} profiles
**/
func listProfiles(profiles *ini.File) int {
	i := 0
	var color string

	profilenames := profiles.SectionStrings()
	sort.Strings(profilenames)
	for _, name := range profilenames {

		if name[0:7] == "profile" {
			profile := name[8:]
			hasMFA := profiles.Section(name).HasKey("mfa_serial")
			i++
			if hasMFA == true {
				color = cred
			} else {
				color = coff
			}
			fmt.Printf("%d\t%s%s%s\n", i, color, profile, coff)
		}

	}
	return i
}

/**
 * directly add an piped account from json
**/
func pipedAccount2Profile(res AWS_ACCOUNT, profiles *ini.File) {

	for _, a := range res.Accounts {

		name := strings.Replace(strings.ToLower(a.Name), " ", "-", -1)
		// fmt.Print("Enter Profilename (default: "+ name +"): ")
		// iname := getUserInput()
		// if (iname != name) && (iname != "")  {name = iname}

		role_arn := "arn:aws:iam::" + a.ID + ":role/role-moia-" + name + "-account"
		// fmt.Print("Enter role_arn (default: "+ role_arn +"): ")
		// irole_arn := getUserInput()
		// if (irole_arn != role_arn) && (irole_arn != "")  {role_arn = irole_arn}

		source_profile := "default"
		// fmt.Print("Enter source_profile (default: default): ")
		// isource_profile:= getUserInput()
		// if (isource_profile != source_profile) && (isource_profile != "")  {source_profile = isource_profile}

		mfa_serial := getmyMFA()
		// fmt.Print("enter any key to add your mfa_serail or leave empty if not necessary: ")
		// imfa_serial := getUserInput()
		// if (imfa_serial == "")  {mfa_serial = imfa_serial}

		fmt.Printf("[profile %s]\n", name)
		fmt.Printf("role_arn = %s\n", role_arn)
		fmt.Printf("source_profile = %s\nregion = %s\n", source_profile, "eu-central-1")
		if mfa_serial != "" {
			fmt.Println("mfa_serial = " + mfa_serial)
		}
		res := addProfile(name, role_arn, source_profile, mfa_serial, profiles)
		if res == true {
			fmt.Printf("%sprofile '"+name+"' successfully added%s\n", cgreen, coff)
		}
		fmt.Println()
	}
}

/**
* automaticly add profile (without intaction)
* returns {bool} true on success
* @param {string} profile_name
* @param {string} role_arn
* @param {string} source_profile
* @param {string} mfa_serial
* @param {*ini.File} config file
**/
func addProfile(profile_name, role_arn, source_profile, mfa_serial string, profiles *ini.File) bool {
	_, err := profiles.GetSection("profile " + profile_name)
	if err == nil {
		fmt.Printf("%sProfile with name '"+profile_name+"' already exists!%s\n", cred, coff)
		return false
	}
	_, err = profiles.NewSection("profile " + profile_name)
	if err != nil {
		fmt.Printf("sonething went wrong!\n ( error: %s)", err)
		return false
	} else {
		if role_arn != "" {
			_, err := profiles.Section("profile "+profile_name).NewKey("role_arn", strings.Replace(role_arn, "\n", "", -1))
			if err != nil {
				return false
			}
		}
		if source_profile != "" {
			_, err := profiles.Section("profile "+profile_name).NewKey("source_profile", strings.Replace(source_profile, "\n", "", -1))
			if err != nil {
				return false
			}
		}
		if mfa_serial != "" {
			_, err := profiles.Section("profile "+profile_name).NewKey("mfa_serial", strings.Replace(mfa_serial, "\n", "", -1))
			if err != nil {
				return false
			}
		}
		err = profiles.SaveTo(ConfigFile)
		if err == nil {
			return true
		}
	}
	return false
}

/**
* add a new profile
* returns boolean true = success
* @param {string} profilename from flag create (-c)
* @param {string} configfile usually ~/.aws/config
* @param {inifile} the current opened innifile
**/
func inputProfile(profilename string, profiles *ini.File) bool {
	profilename = strings.Replace(profilename, " ", "-", -1)
	_, err := profiles.GetSection("profile" + profilename)
	if err != nil {
		_, err := profiles.NewSection("profile " + profilename)
		if err != nil {
			fmt.Printf("sonething went wrong!\n ( error: %s)", err)
			return false
		} else {
			var role_arn, source_profile, mfa_serial string

			fmt.Print("type the role_arn: ")
			role_arn = getUserInput()

			fmt.Print("type the source_profile (default: default): ")
			source_profile = getUserInput()

			fmt.Print("type the mfa_serial (leave empty if not needed): ")
			mfa_serial = getUserInput()

			return addProfile(profilename, role_arn, source_profile, mfa_serial, profiles)
		}
	} else {
		fmt.Println("a Profile with the name '" + profilename + "' already exists!\ncreate aborted!")
	}
	return false

}
