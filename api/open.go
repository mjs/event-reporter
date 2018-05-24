/*
event-reporter - report events to the Cacophony Project API.
Copyright (C) 2018, The Cacophony Project

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package api

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
)

func Open(configFile string) (*CacophonyAPI, error) {
	// TODO(mjs) - much of this is copied straight from
	// thermal-uploader and should be extracted.
	conf, err := ParseConfigFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("configuration error: %v", err)
	}
	privConfigFilename := genPrivConfigFilename(configFile)
	log.Println("private settings file:", privConfigFilename)
	password, err := ReadPassword(privConfigFilename)
	if err != nil {
		return nil, err
	}

	// XXX there's a race here if both thermal-uploader and
	// event-reporter register at about the same time - do some simple locking?
	api, err := NewAPI(conf.ServerURL, conf.Group, conf.DeviceName, password)
	if err != nil {
		return nil, err
	}

	if api.JustRegistered() {
		log.Println("first time registration - saving password")
		err := WritePassword(privConfigFilename, api.Password())
		if err != nil {
			return nil, err
		}
	}

	return api, nil
}

func genPrivConfigFilename(configFile string) string {
	dirname, filename := filepath.Split(configFile)
	bareFilename := strings.TrimSuffix(filename, ".yaml")
	return filepath.Join(dirname, bareFilename+"-priv.yaml")
}
