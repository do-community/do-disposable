package main

import "crypto/rsa"

type configStructure struct {
	Token string
	DefaultRegion string
	DefaultSize string
	PrivateKey *rsa.PrivateKey
	KeyID int
}

var config *configStructure
