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
	"errors"
	"time"

	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"

	"github.com/TheCacophonyProject/event-reporter/eventstore"
)

const dbusName = "org.cacophony.Events"
const dbusPath = "/org.cacophony/Events"

// XXX
func StartService(store *eventstore.EventStore) error {
	conn, err := dbus.SystemBus()
	if err != nil {
		return err
	}
	reply, err := conn.RequestName(dbusName, dbus.NameFlagDoNotQueue)
	if err != nil {
		return err
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		return errors.New("name already taken")
	}

	svc := &service{
		store: store,
	}
	conn.Export(svc, dbusPath, dbusName)
	conn.Export(genIntrospectable(svc), dbusPath, "org.freedesktop.DBus.Introspectable")
	return nil
}

func genIntrospectable(v interface{}) introspect.Introspectable {
	node := &introspect.Node{
		Interfaces: []introspect.Interface{{
			Name:    dbusName,
			Methods: introspect.Methods(v),
		}},
	}
	return introspect.NewIntrospectable(node)
}

type service struct {
	store *eventstore.EventStore
}

func (svc *service) Queue(details []byte, nanos int64) *dbus.Error {
	err := svc.store.Queue(details, time.Unix(0, nanos))
	if err != nil {
		return &dbus.Error{
			Name: dbusName + ".Errors.QueueFailed",
			Body: []interface{}{err.Error()},
		}
	}
	return nil
}
