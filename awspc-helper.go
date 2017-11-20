package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

/**
* some colors for the console
**/
const cred  	= "\x1b[31m"
const cgreen 	= "\x1b[32m"
const cyellow  	= "\x1b[33m"
const coff  	= "\x1b[0m"

/**
* reads piped accounts into structs
**/
func newAccounts() AWS_ACCOUNT {
	input, _ := ioutil.ReadAll(os.Stdin)
	outStr := string(input)
	res := AWS_ACCOUNT{}
	json.Unmarshal([]byte(outStr), &res)
	return res
}

/**
 * get the mfa-arn from aws
 * returns {string} mfa_device arn
**/
func getmyMFA() string {

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
* for receiving user input
* returns {string} with no newline (|n)
**/
func getUserInput() string {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan() // use `for scanner.Scan()` to keep reading
	line := scanner.Text()
	return strings.Replace(string(line), "\n", "", -1)
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
