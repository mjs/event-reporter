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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// NewAPI creates a CacophonyAPI instance and obtains a fresh JSON Web
// Token. If no password is given then the device is registered.
func NewAPI(serverURL, group, deviceName, password string) (*CacophonyAPI, error) {
	api := &CacophonyAPI{
		serverURL:  serverURL,
		group:      group,
		deviceName: deviceName,
		password:   password,
	}
	if password == "" {
		err := api.register()
		if err != nil {
			return nil, err
		}
		api.justRegistered = true
	} else {
		err := api.newToken()
		if err != nil {
			return nil, err
		}
	}
	return api, nil
}

type CacophonyAPI struct {
	serverURL      string
	group          string
	deviceName     string
	password       string
	token          string
	justRegistered bool
}

func (api *CacophonyAPI) Password() string {
	return api.password
}

func (api *CacophonyAPI) JustRegistered() bool {
	return api.justRegistered
}

func (api *CacophonyAPI) register() error {
	if api.password != "" {
		return errors.New("already registered")
	}

	password := randString(20)
	payload, err := json.Marshal(map[string]string{
		"group":      api.group,
		"devicename": api.deviceName,
		"password":   password,
	})
	if err != nil {
		return err
	}
	postResp, err := http.Post(
		api.serverURL+"/api/v1/devices",
		"application/json",
		bytes.NewReader(payload),
	)
	if err != nil {
		return err
	}
	defer postResp.Body.Close()

	var respData tokenResponse
	d := json.NewDecoder(postResp.Body)
	if err := d.Decode(&respData); err != nil {
		return fmt.Errorf("decode: %v", err)
	}
	if !respData.Success {
		return fmt.Errorf("registration failed: %v", respData.message())
	}

	api.password = password
	api.token = respData.Token
	return nil
}

func (api *CacophonyAPI) newToken() error {
	if api.password == "" {
		return errors.New("no password set")
	}
	payload, err := json.Marshal(map[string]string{
		"devicename": api.deviceName,
		"password":   api.password,
	})
	if err != nil {
		return err
	}
	postResp, err := http.Post(
		api.serverURL+"/authenticate_device",
		"application/json",
		bytes.NewReader(payload),
	)
	if err != nil {
		return err
	}
	defer postResp.Body.Close()

	var resp tokenResponse
	d := json.NewDecoder(postResp.Body)
	if err := d.Decode(&resp); err != nil {
		return fmt.Errorf("decode: %v", err)
	}
	if !resp.Success {
		return fmt.Errorf("registration failed: %v", resp.message())
	}
	api.token = resp.Token
	return nil
}

type tokenResponse struct {
	Success  bool
	Messages []string
	Token    string
}

func (r *tokenResponse) message() string {
	if len(r.Messages) > 0 {
		return r.Messages[0]
	}
	return "unknown"
}

func (api *CacophonyAPI) ReportEvent(jsonDetails []byte, times []time.Time) error {
	// XXX unmarshal into map[string]interface{}
	// XXX add in `dateTimes` key
	// XXX marshal back to JSON

	jsonBytes, err := json.Marshal(&event)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", api.serverURL+"/api/v1/events", bytes.NewReader(jsonBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", api.token)

	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// XXX check status

	fmt.Println(resp.Status)
	fmt.Println(resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	fmt.Println(string(body))

	return nil
}

func formatTimestamp(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

// XXX
type Event struct {
	Description EventDescription `json:"description"`
	Timestamps  []string         `json:"dateTimes"`
}

type EventDescription struct {
	Type    string      `json:"type"`
	Details interface{} `json:"details"`
}
