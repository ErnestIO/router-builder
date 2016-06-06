/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"encoding/json"
	"log"

	"github.com/nats-io/nats"
	"gopkg.in/redis.v3"
)

func provisionRouter(n *nats.Conn, r router, s string, t string) {
	event := routerEvent{}
	event.load(r, t, s)

	n.Publish(t, []byte(event.toJSON()))
}

func processRequest(n *nats.Conn, r *redis.Client, subject string, resSubject string) {
	n.Subscribe(subject, func(m *nats.Msg) {
		event := RoutersCreate{}
		json.Unmarshal(m.Data, &event)
		persistEvent(r, &event)

		if len(event.Routers) == 0 || event.Status == "completed" {
			event.Status = "completed"
			event.ErrorCode = ""
			event.ErrorMessage = ""
			n.Publish(subject+".done", []byte(event.toJSON()))
			return
		}
		for _, router := range event.Routers {
			if ok, msg := router.Valid(); ok == false {
				event.Status = "error"
				event.ErrorCode = "0001"
				event.ErrorMessage = msg
				n.Publish(subject+".error", []byte(event.toJSON()))
				return
			}
		}
		sw := false
		for i, router := range event.Routers {
			if event.Routers[i].completed() == false {
				sw = true
				event.Routers[i].processing()
				provisionRouter(n, router, event.Service, resSubject)
				if true == event.SequentialProcessing {
					break
				}
			}
		}
		if sw == false {
			event.Status = "completed"
			event.ErrorCode = ""
			event.ErrorMessage = ""
			n.Publish(subject+".done", []byte(event.toJSON()))
			return
		}
		persistEvent(r, &event)
	})
}

func processResponse(n *nats.Conn, r *redis.Client, s string, res string, p string, t string) {
	n.Subscribe(s, func(m *nats.Msg) {
		stored, completed := processNext(n, r, s, p, m.Data, t)

		if completed {
			complete(n, stored, res)
		}
	})
}

func complete(n *nats.Conn, stored *RoutersCreate, subject string) {
	if isErrored(stored) == true {
		stored.Status = "error"
		stored.ErrorCode = "0002"
		stored.ErrorMessage = "Some routers could not be successfully processed"
		n.Publish(subject+"error", []byte(stored.toJSON()))
	} else {
		stored.Status = "completed"
		n.Publish(subject+"done", []byte(stored.toJSON()))
	}
}

func isErrored(stored *RoutersCreate) bool {
	for _, v := range stored.Routers {
		if v.isErrored() {
			return true
		}
	}
	return false
}

func processNext(n *nats.Conn, r *redis.Client, subject string, procSubject string, body []byte, status string) (*RoutersCreate, bool) {
	event := &routerCreatedEvent{}
	json.Unmarshal(body, event)

	message, err := r.Get(event.cacheKey()).Result()
	if err != nil {
		log.Println(err)
	}
	stored := &RoutersCreate{}
	json.Unmarshal([]byte(message), stored)
	completed := true
	scheduled := false
	for i := range stored.Routers {
		if stored.Routers[i].Name == event.RouterName {
			stored.Routers[i].Status = status
			stored.Routers[i].IP = event.RouterIP
			stored.Routers[i].ErrorCode = string(event.Error.Code)
			stored.Routers[i].ErrorMessage = event.Error.Message
		}
		if stored.Routers[i].completed() == false && stored.Routers[i].errored() == false {
			completed = false
		}
		if stored.Routers[i].toBeProcessed() && scheduled == false {
			scheduled = true
			completed = false
			stored.Routers[i].processing()
			provisionRouter(n, stored.Routers[i], event.Service, procSubject)
		}
	}
	persistEvent(r, stored)

	return stored, completed
}
