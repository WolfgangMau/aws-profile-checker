package main

import (
  "os"
  "os/user"
  "os/exec"
  "fmt"
  "github.com/go-ini/ini"
  "github.com/shiena/ansicolor"
  "bytes"
  "encoding/json"
  "strings"
  "sort"
  "flag"
  "bufio"
)

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
**/
func main() {
  ex, _ := os.Executable()
  var all, mfa, nomfa, listonly bool
  var specific, color, create string
  cyellow := "\x1b[33m"
  cred := "\x1b[31m"
  coff := "\x1b[0m"
  usr, err := user.Current()

  flag.BoolVar(&all, "a", false, "check all profiles")
  flag.BoolVar(&mfa, "mfa", false, "check only profiles that require MFA")
  flag.BoolVar(&nomfa, "nomfa", false, "check only profiles that don't require MFA")
  flag.BoolVar(&listonly, "l", false, "only list the profiles")
  flag.StringVar(&create, "c", "", "create a new named profile")
  flag.StringVar(&specific, "n", "", "check one named profile")
  flag.Parse()

  profiles, err := ini.LooseLoad(usr.HomeDir +"/.aws/config")
  if err != nil {
    fmt.Println(err)
  }
  profilenames := profiles.SectionStrings()
  sort.Strings(profilenames)

  fmt.Println("using config-file: "+ cyellow + usr.HomeDir +"/.aws/config"+ coff +"\n")



  i :=0;
  if create != "" {
    res := addprofile(create, usr.HomeDir+"/.aws/config", profiles)
    if res == true {
      fmt.Println("prifile "+ create +" has been created.")
    }
    return
  } else {
    for _,name := range profilenames {

      if name[0:7] == "profile" {
        hasMFA := profiles.Section(name).HasKey("mfa_serial")
        profile := name[8:len(name)]
        if (all == true) || ( (hasMFA == true)  && (mfa == true) ) || ( (hasMFA == false) && (nomfa == true) ) || (profile == specific)  {
          i++;
          fmt.Printf("%d\t%s", i, padRight(profile," ", 25))
      	   err := check_profile(profile)
           if err != nil {
             //
           }
         } else {
           if listonly == true {
             if hasMFA == true {
               color = cred
             } else {
               color = coff
              }
             fmt.Printf("%s%s%s\n",color, profile, coff)
           }
         }
      }
    }
  }
  if ( (i==0) && (listonly == false) ) || ( (i==0) && (specific != "") ) {
    fmt.Printf("\n%sno profile matches to the selected options%s\ncheck possible optins with the '-h' flag\ne.g.: %s -h\n", cyellow, coff, ex)
  }
  fmt.Println()
}


/**
 * checks the profile against aws
 * returns {error} eroor
 * @param {string} profile (named profile)
**/
func check_profile(profile string) (error) {

  cred := "\x1b[31m"
  cgreen := "\x1b[32m"
  coff := "\x1b[0m"
  toOut := ansicolor.NewAnsiColorWriter(os.Stdout)

  cmd := exec.Command("aws", "sts", "get-caller-identity", "--profile", profile)
  _, err := cmd.CombinedOutput()
  if err != nil {
    fmt.Fprintf(toOut,"%sfailed%s\n", cred, coff)
    return err
  } else {
    fmt.Fprintf(toOut,"%sOK%s\n", cgreen, coff)
  }
  return err
}

/**
* returns haystack with replaced needle
* @param {string} haystack String in which we search
* @param {string} needle String wich should be replaced
* @param {string} repl String replacement for the needle
**/
func replaceStr(haystack, needle, repl string) (string) {

  return strings.Replace(haystack, needle, repl, -1)
}

/**
* makes paddings to the right side of a text - only for nicer display
* returns a String with a fixed lenght
* @param {string} str String that should be padded
* @param {string} pad String that is being used for paddings
* @param {integer} lenght Number og chars that the Striung should provide at the end
**/
func padRight(str, pad string, lenght int) string {

    for {
        str += pad
        if len(str) > lenght {
            return str[0:lenght]
        }
    }
}

/**
 * get the mfa-arn from aws
 * returns {string} mfa_device arn
**/
func mymfa() (string) {

  type AWS_MFA struct {
  	MFADevices []struct {
  		UserName     string    `json:"UserName"`
  		SerialNumber string    `json:"SerialNumber"`
  		EnableDate   string    `json:"EnableDate"`
  	} `json:"MFADevices"`
  }

	cmd := exec.Command("aws", "iam", "list-mfa-devices")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("cmd.Run() failed with %s\n", err)
	}
	outStr, _ := string(stdout.Bytes()), string(stderr.Bytes())
  res := AWS_MFA{}
    json.Unmarshal([]byte(outStr), &res)
    return res.MFADevices[0].SerialNumber
}

/**
* add a new profile
* returns boolean true = success
* @param {string} profilename from flag create (-c)
* @param {string} configfile usually ~/.aws/config
* @param {inifile} the current opened innifile
**/
func addprofile(profilename, inifilename string, profiles *ini.File) (bool) {
  profilename = strings.Replace(profilename," ","-", -1)
  _, err := profiles.GetSection("profile"+ profilename)
  if err != nil {
    _,err := profiles.NewSection("profile "+ profilename)
    if err != nil {
      fmt.Printf("sonething went wrong!\n ( error: %s)", err)
      return false
    }else {
      var role_arn, source_profile, mfa_serial string

      fmt.Print("type the role_arn: ")
      role_arn = getUserInput();
      if role_arn != "" {
        _,err := profiles.Section("profile "+ profilename).NewKey("role_arn", strings.Replace(role_arn, "\n", "",  -1))
        if err != nil {
          return false
        }
      }

      fmt.Print("type the source_profile (default: default): ")
      source_profile = getUserInput();
      if source_profile == "" {
        source_profile = "default"
      }
      _,err := profiles.Section("profile "+ profilename).NewKey("source_profile", strings.Replace(source_profile, "\n", "",  -1))
      if err != nil {
        return false
      }

      fmt.Print("type the mfa_serial (leave empty if not needed): ")
      mfa_serial = getUserInput();
      if  mfa_serial != "" {
        _,err := profiles.Section("profile "+ profilename).NewKey("mfa_serial", strings.Replace(mfa_serial, "\n", "",  -1))
        if err != nil {
          return false
        }
      }
      err = profiles.SaveTo(inifilename)
      if err == nil {
        return true
      }
    }
  }else {
    fmt.Println("a Profile with the name '"+profilename+"' already exists!\ncreate aborted!")
  }
  return false
}

/**
* for receiving user input
* returns {string} with no newline (|n)
**/
func getUserInput() (string) {
  scanner := bufio.NewScanner(os.Stdin)
  scanner.Scan() // use `for scanner.Scan()` to keep reading
  line := scanner.Text()
  return strings.Replace(line,"\n","", -1)
}
