package main

import (
	"fmt"
	"strings"

	ibclient "github.com/infobloxopen/infoblox-go-client"
	"github.com/sirupsen/logrus"
)

type license string

const (
	cloud license = "Cloud Network Automation"
)

//Checks for cloud license in nios
func CheckForCloudLicense(objMgr *ibclient.ObjectManager) {
	err := CheckLicense(objMgr, "cloud")
	if err != nil {
		logrus.Fatal("Error while checking for cloud license: ", err)
	}
}

func CheckLicense(objMgr *ibclient.ObjectManager, licenseType string) (err error) {
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
	err = fmt.Errorf("%s License not available/applied. Apply the license for the grid and try again", GetLicenseType(licenseType))
	return
}

func GetLicenseType(p_licenseType string) (r_licenseType license) {
	switch p_licenseType {
	case "cloud":
		r_licenseType = cloud
	}
	return
}
