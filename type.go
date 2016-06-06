/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"bytes"
	"encoding/json"
	"log"

	"gopkg.in/redis.v3"
)

// RoutersCreate : Represents a routers.create message
type RoutersCreate struct {
	Service              string   `json:"service"`
	Status               string   `json:"status"`
	ErrorCode            string   `json:"error_code"`
	ErrorMessage         string   `json:"error_message"`
	Routers              []router `json:"routers"`
	SequentialProcessing bool     `json:"sequential_processing"`
}

func (e *RoutersCreate) toJSON() string {
	message, _ := json.Marshal(e)
	return string(message)
}

func (e *RoutersCreate) cacheKey() string {
	return composeCacheKey(e.Service)
}

func composeCacheKey(service string) string {
	var key bytes.Buffer
	key.WriteString("GPBRouters_")
	key.WriteString(service)

	return key.String()
}

type router struct {
	Name               string `json:"name"`
	Type               string `json:"type"`
	ClientName         string `json:"client_name"`
	DatacenterName     string `json:"datacenter_name"`
	DatacenterPassword string `json:"datacenter_password"`
	DatacenterRegion   string `json:"datacenter_region"`
	DatacenterType     string `json:"datacenter_type"`
	DatacenterUsername string `json:"datacenter_username"`
	ExternalNetwork    string `json:"external_network"`
	VCloudURL          string `json:"vcloud_url"`
	VseURL             string `json:"vse_url"`
	IP                 string `json:"ip"`
	Created            bool   `json:"created"`
	Status             string `json:"status"`
	ErrorCode          string `json:"error_code"`
	ErrorMessage       string `json:"error_message"`
}

func (r *router) fail() {
	r.Status = "errored"
}

func (r *router) complete() {
	r.Status = "completed"
}

func (r *router) processing() {
	r.Status = "processed"
}

func (r *router) errored() bool {
	return r.Status == "errored"
}

func (r *router) completed() bool {
	println(r.Status)
	return r.Status == "completed"
}

func (r *router) isProcessed() bool {
	return r.Status == "processed"
}

func (r *router) isErrored() bool {
	return r.Status == "errored"
}

func (r *router) toBeProcessed() bool {
	return r.Status != "processed" && r.Status != "completed" && r.Status != "errored"
}

func (r *router) Valid() (bool, string) {
	if r.Name == "" {
		return false, "Router name can not be empty"
	}
	if r.DatacenterName == "" {
		return false, "Specifying a datacenter is necessary when creating a router"
	}

	return true, ""
}

type routers struct {
	Collection []router
}

type routerEvent struct {
	Service            string `json:"service_id"`
	Type               string `json:"type"`
	RouterName         string `json:"router_name"`
	RouterType         string `json:"router_type"`
	ClientName         string `json:"client_name"`
	DatacenterName     string `json:"datacenter_name"`
	DatacenterUsername string `json:"datacenter_username"`
	DatacenterPassword string `json:"datacenter_password"`
	DatacenterRegion   string `json:"datacenter_region"`
	DatacenterType     string `json:"datacenter_type"`
	ExternalNetwork    string `json:"external_network"`
	VCloudURL          string `json:"vcloud_url"`
	VseURL             string `json:"vse_url"`
	Status             string `json:"status"`
}

func (e *routerEvent) load(rt router, t string, s string) {
	e.Service = s
	e.Type = t
	e.RouterType = rt.Type
	e.RouterName = rt.Name
	e.ClientName = rt.ClientName
	e.DatacenterName = rt.DatacenterName
	e.DatacenterUsername = rt.DatacenterUsername
	e.DatacenterPassword = rt.DatacenterPassword
	e.DatacenterRegion = rt.DatacenterRegion
	e.DatacenterType = rt.DatacenterType
	e.ExternalNetwork = rt.ExternalNetwork
	e.VCloudURL = rt.VCloudURL
	e.VseURL = rt.VseURL
	e.Status = rt.Status
}

func (e *routerEvent) toJSON() string {
	message, _ := json.Marshal(e)
	return string(message)
}

// Error : partial error codes
type Error struct {
	Code    json.Number `json:"code,Number"`
	Message string      `json:"message"`
}

type routerCreatedEvent struct {
	Type       string `json:"type"`
	Service    string `json:"service_id"`
	RouterID   string `json:"router_id"`
	RouterName string `json:"router_name"`
	RouterIP   string `json:"router_ip"`
	Error      Error  `json:"error"`
}

func (e *routerCreatedEvent) cacheKey() string {
	return composeCacheKey(e.Service)
}

func persistEvent(redisClient *redis.Client, event *RoutersCreate) {
	if event.Service == "" {
		panic("Service is null!")
	}
	if err := redisClient.Set(event.cacheKey(), event.toJSON(), 0).Err(); err != nil {
		log.Println(err)
	}
}
