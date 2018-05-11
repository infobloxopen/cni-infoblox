package main

import (
	"fmt"
	"strings"

	ibclient "github.com/infobloxopen/infoblox-go-client"
	"github.com/sirupsen/logrus"
)

type licenseName string

const (
	cloud licenseName = "Cloud Network Automation"
)

//Checks for cloud license in nios
func CheckForCloudLicense(objMgr *ibclient.ObjectManager, user string) {
	err := CheckLicense(objMgr, "cloud", user)
	if err != nil {
		logrus.Fatal("Error while checking for cloud license: ", err)
	}
}

func CheckLicense(objMgr *ibclient.ObjectManager, licenseType string, user string) (err error) {
	license, err := objMgr.GetLicense()
	if err != nil {
		return
	}
	for _, v := range license {
		if strings.ToLower(v.Licensetype) == licenseType {
			if v.ExpirationStatus != "DELETED" && v.ExpirationStatus != "EXPIRED" {
				return
			}
		}
	}
	err = fmt.Errorf("%s License not available or User: %s not having sufficient permissions. ", GetLicenseName(licenseType), user)
	return
}

func GetLicenseName(licenseType string) (licenseName licenseName) {
	switch licenseType {
	case "cloud":
		licenseName = cloud
	}
	return
}
