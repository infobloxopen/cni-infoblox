package main

import (
	"fmt"
	"strings"

	ibclient "github.com/infobloxopen/infoblox-go-client"
	"github.com/sirupsen/logrus"
)

//Checks for cloud license in nios
func CheckForCloudLicense(objMgr *ibclient.ObjectManager) {
	err := checkLicense(objMgr, "cloud")
	if err != nil {
		logrus.Fatal(err)
	}
}

func checkLicense(objMgr *ibclient.ObjectManager, licenseType string) (err error) {
	license, err := objMgr.GetLicense()

	if err != nil {
		return
	}
	for _, v := range license {
		if strings.ToLower(v.Licensetype) == licenseType {
			if v.ExpirationStatus == "DELETED" || v.ExpirationStatus == "EXPIRED" {
				err = fmt.Errorf("%s license is not applied/deleted for the grid. Apply the license and try again", licenseType)
				return
			}
		}
	}
	return
}
