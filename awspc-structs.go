package main


/**
* Structure of the result from 'aws iam list-mfa-devices'
**/
type AWS_MFA struct {
	MFADevices []struct {
		UserName     string `json:"UserName"`
		SerialNumber string `json:"SerialNumber"`
		EnableDate   string `json:"EnableDate"`
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
* Structure of an Credential
**/
type AWS_CREDENTIALS struct {
	Credentials struct {
		SecretAccessKey string    `json:"SecretAccessKey"`
		SessionToken    string    `json:"SessionToken"`
		Expiration      string 	  `json:"Expiration"`
		AccessKeyID     string    `json:"AccessKeyId"`
	} `json:"Credentials"`
}