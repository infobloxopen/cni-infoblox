// Copyright 2016 Infoblox Inc.
// All Rights Reserved.
//
//    Licensed under the Apache License, Version 2.0 (the "License"); you may
//    not use this file except in compliance with the License. You may obtain
//    a copy of the License at
//
//         http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
//    WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
//    License for the specific language governing permissions and limitations
//    under the License.

package ibcni

import (
	"log"
	"os"
)

const defaultSocketDir = "/run/cni"

type DriverSocket struct {
	SocketDir  string
	DriverName string

	SocketFile string
}

func dirExists(dirname string) (bool, error) {
	fileInfo, err := os.Stat(dirname)
	if err == nil {
		if fileInfo.IsDir() {
			return true, nil
		} else {
			return false, nil
		}
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func createDir(dirname string) error {
	return os.MkdirAll(dirname, 0700)
}

func fileExists(filePath string) (bool, error) {
	_, err := os.Stat(filePath)

	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	}

	return true, err
}

func deleteFile(filePath string) error {
	return os.Remove(filePath)
}

func GetDefaultSocketDir() string {
	return defaultSocketDir
}

func (s *DriverSocket) SetupSocket() string {
	exists, err := dirExists(s.SocketDir)
	if err != nil {
		log.Panicf("Stat Socket Directory error '%s'", err)
		os.Exit(1)
	}
	if !exists {
		err = createDir(s.SocketDir)
		if err != nil {
			log.Panicf("Create Socket Directory error: '%s'", err)
			os.Exit(1)
		}
		log.Printf("Created Socket Directory: '%s'", s.SocketDir)
	}

	log.Printf("SocketFile: '%s'", s.SocketFile)
	exists, err = fileExists(s.SocketFile)
	if err != nil {
		log.Panicf("Stat Socket File error: '%s'", err)
		os.Exit(1)
	}
	if exists {
		err = deleteFile(s.SocketFile)
		if err != nil {
			log.Panicf("Delete Socket File error: '%s'", err)
			os.Exit(1)
		}
		log.Printf("Deleted Old Socket File: '%s'", s.SocketFile)
	}

	return s.SocketFile
}

func (s *DriverSocket) GetSocketFile() string {
	return s.SocketFile
}

func NewDriverSocket(socketDir string, driverName string) *DriverSocket {
	if socketDir == "" {
		socketDir = GetDefaultSocketDir()
	}
	return &DriverSocket{
		SocketDir:  socketDir,
		DriverName: driverName,
		SocketFile: socketDir + "/" + driverName + ".sock"}
}
