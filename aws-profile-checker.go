package main

import (
	"flag"
	"fmt"
	"github.com/go-ini/ini"
	"os"
	"os/user"
	"sort"
	"strconv"
)

/**
* global var for the config
**/
var ConfigFile string

/**
 * reads all profilenames from $HomeDir/.aws/config
 * output/action on the commandline-optios:
 * -h         => shows all options
 * -a         => checks all profilenames against AWS
 * -mfa       => checks only those how have MFA configured
 * -nomfa     => checks only those how don't have MFA configured
 * -l         => lists only the profilenames (red: MFA, white: no MFA)
 * -n <name>  => checks only this named profile
 * -c <name>  => create a new named profile
 * -e         => edit a existing profile
**/
func main() {

	usr, err := user.Current()
	ConfigFile = usr.HomeDir + "/.aws/config"

	executable := os.Args[0]

	var all, mfa, nomfa, listonly, pedit bool
	var specific, pcreate, color string

	flag.BoolVar(&all, "a", false, "check all profiles")
	flag.BoolVar(&mfa, "mfa", false, "check only profiles that require MFA")
	flag.BoolVar(&nomfa, "nomfa", false, "check only profiles that don't require MFA")
	flag.BoolVar(&listonly, "l", false, "only list the profiles")
	flag.BoolVar(&pedit, "e", false, "edit a profile")
	flag.StringVar(&pcreate, "c", "", "create a new named profile")
	flag.StringVar(&specific, "n", "", "check one named profile")
	flag.Parse()

	profiles, err := ini.LooseLoad(ConfigFile)

	checkError(err)

	profilenames := profiles.SectionStrings()
	sort.Strings(profilenames)

	fmt.Println("using config-file: " + cyellow + ConfigFile + coff + "\n")

	i := 0

	PipedInput, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}
	if PipedInput.Mode()&os.ModeNamedPipe != 0 {
		i = 1
		// myAccounts()
		accoounts := newAccounts()
		pipedAccount2Profile(accoounts, profiles)
	}

	if listonly == true {
		listProfiles(profiles)
		os.Exit(0)
	}
	if pedit == true {
		pn := listProfiles(profiles)
		fmt.Print("Enter the number of the listed profile you want to edit: ")
		en := getUserInput()
		ei, _ := strconv.ParseInt(en, 10, 0)
		if (ei <= int64(pn)) && (ei > 0) {
			ep_name := profilenames[ei+1]
			fmt.Printf("selectd profile: %s%s%s\n", cgreen, ep_name[8:], coff)
			keys := profiles.Section(ep_name).Keys()
			values := profiles.Section(ep_name).KeyStrings()
			hasMFA := false
			mfa := getmyMFA()
			for i, k := range keys {
				if values[i] == "mfa_serial" {

					if k.String() != mfa {
						fmt.Printf("%smfa_serial differs from yours!%s\nyours should be: %s\n >", cred, coff, mfa)
					}
				}
				fmt.Printf("new value for %s (leave empty if no change is needed)\ncurrent value: %s\n > ", values[i], k)
				nKey := getUserInput()
				if nKey == "" {
					nKey = k.String()
				}
				_, err := profiles.Section(ep_name).NewKey(values[i], nKey)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

			}
			if hasMFA == false {
				fmt.Printf("no mfa_serial found! add yours? (leave empty if not)\nany key to add this: %s > ", mfa)
				repl_mfa := getUserInput()
				if repl_mfa != "" {
					_, err := profiles.Section(ep_name).NewKey("mfa_serial", mfa)
					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}
				}
			}
			err = profiles.SaveTo(ConfigFile)
			if err == nil {
				fmt.Println("Profile updated.")
			}
		} else {
			fmt.Printf("Error: %d is not betwwen 1 and %d\n", ei, pn)
			os.Exit(0)
		}
	}
	if pcreate != "" {
		res := inputProfile(pcreate, profiles)
		if res == true {
			fmt.Println("prifile " + pcreate + " has been created.")
		}
		return
	} else {
		for _, name := range profilenames {

			if name[0:7] == "profile" {
				i++
				hasMFA := profiles.Section(name).HasKey("mfa_serial")
				profile := name[8:]
				if (all == true) || ((hasMFA == true) && (mfa == true)) || ((hasMFA == false) && (nomfa == true)) || (profile == specific) {
					if hasMFA == true {
						color = cred
					} else {
						color = coff
					}
					fmt.Printf("%d\t%s%s%s", i,color, padRight(profile, " ", 25), coff)
					err := check_profile(profile)
					if err != nil {
						//checkError(err)
					}
				}
			}
		}
	}
	if ((i == 0) && (listonly == false)) || ((i == 0) && (specific != "")) {
		fmt.Printf("\n%sno profile matches to the selected options%s\ncheck possible optins with the '-h' flag\ne.g.: %s -h\n", cyellow, coff, executable)
	}
	fmt.Println()
}
