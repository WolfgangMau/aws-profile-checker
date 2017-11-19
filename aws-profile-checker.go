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
  "io/ioutil"
  "strconv"
)

/**
* Structure of the result from 'aws iam list-mfa-devices'
**/
type AWS_MFA struct {
  MFADevices []struct {
    UserName     string    `json:"UserName"`
    SerialNumber string    `json:"SerialNumber"`
    EnableDate   string    `json:"EnableDate"`
  } `json:"MFADevices"`
}

/**
* Structure oa an Account
**/
type AWS_ACCOUNT struct {
	Accounts []struct {
		Status          string  `json:"Status"`
		Name            string  `json:"Name"`
		Email           string  `json:"Email"`
		JoinedMethod    string  `json:"JoinedMethod"`
		JoinedTimestamp float64 `json:"JoinedTimestamp"`
		ID              string  `json:"Id"`
		Arn             string  `json:"Arn"`
	} `json:"Accounts"`
}

/**
* some colors for the console
**/
const cred    string   = "\x1b[31m"
const cgreen  string   = "\x1b[32m"
const cyellow string   = "\x1b[33m"
const coff    string   = "\x1b[0m"

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
**/
func main() {
  executable := os.Args[0]
  var all, mfa, nomfa, listonly, pedit bool
  var specific, pcreate string
  usr, err := user.Current()

  flag.BoolVar(&all, "a", false, "check all profiles")
  flag.BoolVar(&mfa, "mfa", false, "check only profiles that require MFA")
  flag.BoolVar(&nomfa, "nomfa", false, "check only profiles that don't require MFA")
  flag.BoolVar(&listonly, "l", false, "only list the profiles")
  flag.BoolVar(&pedit, "e", false, "edit a profile")
  flag.StringVar(&pcreate, "c", "", "create a new named profile")
  flag.StringVar(&specific, "n", "", "check one named profile")
  flag.Parse()

  ConfigFile = usr.HomeDir + "/.aws/config"
  profiles, err := ini.LooseLoad(ConfigFile)
  checkError(err)

  profilenames := profiles.SectionStrings()
  sort.Strings(profilenames)

  fmt.Println("using config-file: "+ cyellow + ConfigFile + coff +"\n")

  i :=0;

  PipedInput, err := os.Stdin.Stat()
  if err != nil {
    panic(err)
  }
  if PipedInput.Mode() & os.ModeNamedPipe != 0 {
    i = 1
    // myAccounts()
    accoounts := newAccounts()
    pipedaccount(accoounts, profiles)
  }

  if listonly == true {
    listProfiles(profiles)
    os.Exit(0)
  }
  if pedit == true {
    pn := listProfiles(profiles)
    fmt.Print("Enter the number of the listed profile you want to edit: ")
    en := getUserInput()
    ei,_ := strconv.ParseInt(en, 10, 0)
    if (ei <= int64(pn)) && (ei > 0) {
      ep_name := profilenames[ei+1]
      fmt.Printf("selectd profile: %s%s%s\n",cgreen,ep_name[8:len(ep_name)],coff)
      keys    := profiles.Section(ep_name).Keys()
      values  := profiles.Section(ep_name).KeyStrings()
      hasMFA:=false
      mfa:=mymfa()
      for i,k := range keys {
        if values[i] == "mfa_serial" {
          hasMFA=true
          if (k.String() != mfa) {
            fmt.Printf("%smfa_serial differs from yours!%s\nyours should be: %s\n >",cred,coff,mfa)
          }
        }
        fmt.Printf("new value for %s (leave empty if no change is needed)\ncurrent value: %s\n > ",values[i],k)
        nKey:=getUserInput()
        if (nKey == "") {nKey=k.String()}
        _, err := profiles.Section(ep_name).NewKey(values[i], nKey)
        if err != nil {
         fmt.Println(err)
         os.Exit(1)
        }

      }
      if hasMFA == false {
        fmt.Printf("no mfa_serial found! add yours? (leave empty if not)\nany key to add this: %s > ",mfa)
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
      fmt.Printf("Error: %d is not betwwen 1 and %d\n",ei, pn)
      os.Exit(0)
    }
  }
  if pcreate != "" {
    res := inputProfile(pcreate, profiles)
    if res == true {
      fmt.Println("prifile "+ pcreate +" has been created.")
    }
    return
  } else {
    for _,name := range profilenames {

      if name[0:7] == "profile" {
        i++
        hasMFA := profiles.Section(name).HasKey("mfa_serial")
        profile := name[8:len(name)]
        if (all == true) || ( (hasMFA == true)  && (mfa == true) ) || ( (hasMFA == false) && (nomfa == true) ) || (profile == specific)  {
          fmt.Printf("%d\t%s", i, padRight(profile," ", 25))
      	   err := check_profile(profile)
           if err != nil {
             //checkError(err)
           }
        }
      }
    }
  }
  if ( (i==0) && (listonly == false) ) || ( (i==0) && (specific != "") ) {
    fmt.Printf("\n%sno profile matches to the selected options%s\ncheck possible optins with the '-h' flag\ne.g.: %s -h\n", cyellow, coff, executable)
  }
  fmt.Println()
}

/**
* list all profilenames
* returns (int) number of profiles
* @param {*ini.Files} profiles
**/
func listProfiles(profiles *ini.File) (int) {
  i:=0
  var color string

  profilenames := profiles.SectionStrings()
  sort.Strings(profilenames)
  for _,name := range profilenames {

    if name[0:7] == "profile" {
      profile := name[8:len(name)]
      hasMFA := profiles.Section(name).HasKey("mfa_serial")
      i++
      if hasMFA == true {
        color = cred
      } else {
        color = coff
       }
      fmt.Printf("%d\t%s%s%s\n",i , color, profile, coff)
    }

  }
  return i
}

/**
 * checks the profile against aws
 * returns {error} eroor
 * @param {string} profile (named profile)
**/
func check_profile(profile string) (error) {

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

	cmd := exec.Command("aws", "iam", "list-mfa-devices")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

  err := cmd.Run()
	checkError(err)

	outStr, _ := string(stdout.Bytes()), string(stderr.Bytes())
  res := AWS_MFA{}
    json.Unmarshal([]byte(outStr), &res)
    return res.MFADevices[0].SerialNumber
}

/**
* reads piped accounts into structs
**/
func newAccounts() (AWS_ACCOUNT) {
  input, _ := ioutil.ReadAll(os.Stdin)
  outStr := string(input)
  res := AWS_ACCOUNT{}
  json.Unmarshal([]byte(outStr), &res)
  return res
}
/**
 * directly add an piped account from json
**/
func pipedaccount(res AWS_ACCOUNT, profiles *ini.File) {

	for _,a  := range res.Accounts {

		name := strings.Replace(strings.ToLower(a.Name), " ", "-", -1)
    // fmt.Print("Enter Profilename (default: "+ name +"): ")
    // iname := getUserInput()
    // if (iname != name) && (iname != "")  {name = iname}

    role_arn := "arn:aws:iam::"+ a.ID +":role/role-moia-"+ name +"-account"
    // fmt.Print("Enter role_arn (default: "+ role_arn +"): ")
    // irole_arn := getUserInput()
    // if (irole_arn != role_arn) && (irole_arn != "")  {role_arn = irole_arn}

    source_profile := "default"
    // fmt.Print("Enter source_profile (default: default): ")
    // isource_profile:= getUserInput()
    // if (isource_profile != source_profile) && (isource_profile != "")  {source_profile = isource_profile}

    mfa_serial := mymfa()
    // fmt.Print("enter any key to add your mfa_serail or leave empty if not necessary: ")
    // imfa_serial := getUserInput()
    // if (imfa_serial == "")  {mfa_serial = imfa_serial}

		fmt.Printf("[profile %s]\n", name)
		fmt.Printf("role_arn = %s\n", role_arn)
		fmt.Printf("source_profile = %s\nregion = %s\n",source_profile, "eu-central-1")
    if mfa_serial != "" {
      fmt.Println("mfa_serial = "+ mfa_serial)
    }
    res := addProfile(name, role_arn, source_profile, mfa_serial, profiles)
    if res == true {
      fmt.Printf("%sprofile '"+ name +"' successfully added%s\n",cgreen, coff)
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
func addProfile(profile_name, role_arn, source_profile, mfa_serial string, profiles *ini.File) (bool) {
  _, err := profiles.GetSection("profile "+ profile_name)
  if err == nil {
    fmt.Printf("%sProfile with name '"+ profile_name +"' already exists!%s\n",cred,coff)
    return false
  }
  _,err = profiles.NewSection("profile "+ profile_name)
  if err != nil {
    fmt.Printf("sonething went wrong!\n ( error: %s)", err)
    return false
  } else {
    if role_arn != "" {
      _,err := profiles.Section("profile "+ profile_name).NewKey("role_arn", strings.Replace(role_arn, "\n", "",  -1))
      if err != nil {
        return false
      }
    }
    if source_profile != "" {
      _,err := profiles.Section("profile "+ profile_name).NewKey("source_profile", strings.Replace(source_profile, "\n", "",  -1))
      if err != nil {
        return false
      }
    }
    if  mfa_serial != "" {
      _,err := profiles.Section("profile "+ profile_name).NewKey("mfa_serial", strings.Replace(mfa_serial, "\n", "",  -1))
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
func inputProfile(profilename string, profiles *ini.File) (bool) {
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

      fmt.Print("type the source_profile (default: default): ")
      source_profile = getUserInput();

      fmt.Print("type the mfa_serial (leave empty if not needed): ")
      mfa_serial = getUserInput();

      return addProfile(profilename, role_arn, source_profile, mfa_serial , profiles)
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
  return strings.Replace(string(line),"\n","", -1)
}

/**
* global error-handler*
* @param {error} err
**/
func checkError(err error) {
  if err != nil {
    fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
    os.Exit(1)
  }
}
