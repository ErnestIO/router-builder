/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/nats-io/nats"
	"gopkg.in/redis.v3"
)

var setup = false
var nc *nats.Conn
var rc *redis.Client

func wait(ch chan bool) error {
	return waitTime(ch, 500*time.Millisecond)
}

func waitTime(ch chan bool, timeout time.Duration) error {
	select {
	case <-ch:
		return nil
	case <-time.After(timeout):
	}
	return errors.New("timeout")
}

func testSetup() {
	if setup == false {
		os.Setenv("NATS_URI", "nats://localhost:4222")

		nc = natsClient()
		nc.Subscribe("config.get.redis", func(msg *nats.Msg) {
			nc.Publish(msg.Reply, []byte(`{"DB":0,"addr":"localhost:6379","password":""}`))
		})
		rc = redisClient()
		setup = true
	}
}

func TestProvisionAllRoutersBasic(t *testing.T) {
	testSetup()

	processRequest(nc, rc, "routers.create", "provision-router")

	ch := make(chan bool)

	nc.Subscribe("provision-router", func(msg *nats.Msg) {
		event := &routerEvent{}
		json.Unmarshal(msg.Data, event)
		if event.Type == "provision-router" &&
			event.RouterName == "supu" &&
			event.DatacenterName == "name" &&
			event.DatacenterPassword == "password" &&
			event.DatacenterRegion == "region" &&
			event.DatacenterType == "type" {
			log.Println("Message Received")
			var key bytes.Buffer
			key.WriteString("GPBRouters_")
			key.WriteString(event.Service)
			message, _ := rc.Get(key.String()).Result()
			stored := &RoutersCreate{}
			json.Unmarshal([]byte(message), stored)
			if stored.Service != event.Service {
				t.Fatal("Event is not persisted correctly")
			}
			ch <- true
		} else {
			t.Fatal("Message received from nats does not match")
		}
	})

	message := []byte("{\"service\":\"service\", \"routers\":[{\"name\":\"supu\",\"client\":\"supu\",\"datacenter_name\":\"test\",\"datacenter_name\":\"name\",\"datacenter_password\":\"password\",\"datacenter_region\":\"region\",\"datacenter_type\":\"type\"}]}")
	nc.Publish("routers.create", message)
	time.Sleep(500 * time.Millisecond)

	if e := wait(ch); e != nil {
		t.Fatal("Message not received from nats for subscription")
	}
}

func TestProvisionAllRoutersSendingTwoRouters(t *testing.T) {
	testSetup()

	processRequest(nc, rc, "routers.create", "provision-router")

	ch := make(chan bool)
	ch2 := make(chan bool)

	nc.Subscribe("provision-router", func(msg *nats.Msg) {
		event := &routerEvent{}
		json.Unmarshal(msg.Data, event)
		if event.Type == "provision-router" &&
			event.RouterName == "supu" &&
			event.Service == "service" {
			ch <- true
		} else {
			if event.Type == "provision-router" &&
				event.RouterName == "tupu" &&
				event.Service == "service" {
				ch2 <- true
			} else {
				t.Fatal("Message received from nats does not match")
			}
		}
	})

	message := []byte("{\"service\":\"service\",\"routers\":[{\"name\":\"supu\",\"client\":\"supu\",\"datacenter_name\":\"supu\"}, {\"name\":\"tupu\",\"client\":\"tupu\",\"datacenter_name\":\"tupu\"}]}")
	nc.Publish("routers.create", message)
	time.Sleep(500 * time.Millisecond)

	if e := wait(ch); e != nil {
		t.Fatal("Message not received from nats for subscription")
	}
	if e := wait(ch2); e != nil {
		t.Fatal("Message not received from nats for subscription")
	}
}

func TestProvisionAllRoutersWithInvalidMessage(t *testing.T) {
	testSetup()

	processRequest(nc, rc, "routers.create", "provision-router")

	ch := make(chan bool)
	ch2 := make(chan bool)

	nc.Subscribe("provision-router", func(msg *nats.Msg) {
		ch <- true
	})

	nc.Subscribe("routers.create.error", func(msg *nats.Msg) {
		ch2 <- true
	})

	message := []byte("{\"service\": \"service\", \"routers\": [{\"name\":\"supu\",\"client_id\":\"supu\"}]}")
	nc.Publish("routers.create.error", message)

	if e := wait(ch); e == nil {
		t.Fatal("Produced a provision-router message when I shouldn't")
	}
	if e := wait(ch2); e != nil {
		t.Fatal("Should produce a routers.create.error message on nats")
	}
}

func TestProvisionAllRoutersWithDifferentMessageType(t *testing.T) {
	testSetup()
	processRequest(nc, rc, "routers.create", "provision-router")

	ch := make(chan bool)

	nc.Subscribe("provision-router", func(msg *nats.Msg) {
		ch <- true
	})

	message := []byte("{\"service\":\"service\",\"routers\":[{\"name\":\"supu\",\"client\":\"supu\",\"datacenter_name\":\"supu\"}, {\"name\":\"tupu\",\"client\":\"tupu\",\"datacenter_name\":\"tupu\"}]}")
	nc.Publish("non-routers-create", message)

	if e := wait(ch); e == nil {
		t.Fatal("Produced a provision-router message when I shouldn't")
	}
}

func TestRouterCreatedForAMultiRequest(t *testing.T) {
	testSetup()

	processResponse(nc, rc, "router-created", "routers.create.", "provision-router", "completed")

	ch := make(chan bool)
	service := "sss"

	nc.Subscribe("routers.create.done", func(msg *nats.Msg) {
		log.Printf("DATA RECEIVED: %s\n", string(msg.Data))
		t.Fatal("Message received from nats does not match")
	})

	original := "{\"service\": \"sss\", \"routers\": [{\"name\":\"supu\",\"client_name\":\"supu\",\"datacenter_name\":\"supu\"}, {\"name\":\"tupu\",\"client_name\":\"supu\",\"datacenter_name\":\"supu\"}]}"

	if err := rc.Set("GPBRouters_sss", original, 0).Err(); err != nil {
		log.Println(err)
		t.Fatal("Can't write on redis")
	}
	message := fmt.Sprintf("{\"type\":\"router-created\",\"service_id\":\"%s\",\"router_name\":\"supu\"}", service)

	nc.Publish("router-created", []byte(message))

	if e := wait(ch); e != nil {
		return
	}
}

func TestRouterCreated(t *testing.T) {
	testSetup()

	processResponse(nc, rc, "router-created", "routers.create.", "provision-router", "completed")

	ch := make(chan bool)
	service := "sss"

	nc.Subscribe("routers.create.done", func(msg *nats.Msg) {
		event := &RoutersCreate{}
		json.Unmarshal(msg.Data, event)
		if service == event.Service && event.Status == "completed" && len(event.Routers) == 1 {
			ch <- true
		} else {
			t.Fatal("Message received from nats does not match")
		}
	})

	original := "{\"service\": \"sss\", \"routers\": [{\"name\":\"supu\",\"client_id\":\"supu\"}]}"

	if err := rc.Set("GPBRouters_sss", original, 0).Err(); err != nil {
		log.Println(err)
		t.Fatal("Can't write on redis")
	}
	message := fmt.Sprintf("{\"type\":\"router-created\",\"service_id\":\"%s\",\"router_name\":\"supu\"}", service)

	nc.Publish("router-created", []byte(message))

	if e := wait(ch); e != nil {
		t.Fatal("Message not received from nats for subscription")
	}
}

func TestProvisionRouterError(t *testing.T) {
	testSetup()

	processResponse(nc, rc, "provision-router-error", "routers.create.", "provision-router", "errored")

	ch := make(chan bool)
	service := "sss"

	nc.Subscribe("routers.create.error", func(msg *nats.Msg) {
		event := &RoutersCreate{}
		json.Unmarshal(msg.Data, event)
		if service == event.Service && event.Status == "error" {
			ch <- true
		} else {
			t.Fatal("Message received from nats does not match")
		}
	})

	original := "{\"service\": \"sss\", \"routers\": [{\"name\":\"sss\",\"client_id\":\"supu\"}]}"

	if err := rc.Set("GPBRouters_sss", original, 0).Err(); err != nil {
		log.Println(err)
		t.Fatal("Can't write on redis")
	}
	message := fmt.Sprintf("{\"type\":\"provision-router-error\",\"service_id\":\"%s\",\"router_name\":\"sss\"}", service)

	nc.Publish("provision-router-error", []byte(message))

	if e := wait(ch); e != nil {
		t.Fatal("Message not received from nats for subscription")
	}
}
