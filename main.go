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

package main

import (
	"log"
	"time"

	arg "github.com/alexflint/go-arg"

	"github.com/TheCacophonyProject/event-reporter/api"
	"github.com/TheCacophonyProject/event-reporter/eventstore"
)

var version = "No version provided"

type argSpec struct {
	ConfigFile string `arg:"-c,--config" help:"path to thermal-uploader configuration file"`
	DBPath     string `arg:"-d,--db" help:"path to state database"`
}

func (argSpec) Version() string {
	return version
}

func procArgs() argSpec {
	args := argSpec{
		ConfigFile: "/etc/thermal-uploader.yaml",
		DBPath:     "/var/run/event-reporter.db",
	}
	arg.MustParse(&args)
	return args
}

func main() {
	err := runMain()
	if err != nil {
		log.Fatal(err.Error())
	}
}

func runMain() error {
	args := procArgs()
	log.SetFlags(0) // Removes default timestamp flag
	log.Printf("running version: %s", version)

	store, err := eventstore.Open(args.DBPath)
	if err != nil {
		return err
	}

	api, err := api.Open(args.ConfigFile)
	if err != nil {
		return err
	}

	err = StartService(store)
	if err != nil {
		return err
	}

	for {
		// XXX randomise the time somewhat
		time.Sleep(10 * time.Second) // XXX make this a command line option

		events, err := store.All()
		if err != nil {
			return err
		}

		// XXX log events to send
		// XXX bail if too many failures
		for _, event := range events {
			err := api.ReportEvent(event.Details, event.Timestamps)
			// XXX distinguish between 4xxs and everything else
			if err == nil || permanent {
				// XXX log permanent error
				store.Discard(event)
			} else {
				log.Printf("event report failed: %v", err)
			}
		}

		// XXX log counts
	}
}
