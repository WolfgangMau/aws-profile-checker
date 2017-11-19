# aws-profile-checker

this little program checks all listed profiles which are
located in $HOME/.aws/config against aws and proofs if they are
valid configured.
there is no really need for this kind of tool, but I just needed an reason for my first golang-coding

[![Build Status](https://travis-ci.org/WolfgangMau/aws-profile-checker.svg?branch=master)](https://travis-ci.org/WolfgangMau/aws-profile-checker)

I provides following copmmandline-options:
- __-h__ / __-help__ showns the help
- __-l__ list all profiles wthiout checking them profiles, which needs MFA are colored in red, others are not colored
- __-n <name>__ checks only this named profile
- __-mfa__ checks all profiles which need MFA
- __-nomfa__ checks all profiles which don't need MFA
- __-a__ checks all profiles
- __-e__ edit a existing profile
- __-c <name>__ helps you to create a new named profile

## dependencies
* __aws-cli__ (https://github.com/aws/aws-cli)
* __go__ (https://golang.org/)
* external go-packages:
  - github.com/go-ini/ini
  - github.com/shiena/ansicolor

## build
- go get github.com/go-ini/ini
- go get github.com/shiena/ansicolor
- __go build__

## Usage examples:

show help
```bash
aws-profile-checker -h
Usage of aws-profile-checker:

  -a	check all profiles
  -c	create  a new profile
  -l	only list the profiles
  -mfa
    	check only profiles that require MFA
  -n string
    	check one named profile
  -nomfa
    	check only profiles that don´t require MFA
```

list all profiles
```bash
aws-profile-checker -l
using config-file: /Users/Moe/.aws/config

Bart
Homer
Lisa
Maggie
Marge
```

check all profiles
```bash
aws-profile-checker -a
using config-file: /Users/Moe/.aws/config

1	Bart                     OK
2	Homer                    OK
3	Lisa                     OK
4	Maggie                   OK
5	Marge                    OK
```


check a named profile
```bash
aws-profile-checker -n Bart
using config-file: /Users/Moe/.aws/config

1	Bart                     OK
```

edit profile
```bash
aws-profile-checker -e
```

bulk add profiles
```bash
aws organizations list-accounts | aws-profile-checker
OR
cat sample_accounts.json | aws-profile-checker
```
