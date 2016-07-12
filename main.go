/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"os"
	"runtime"

	l "github.com/ernestio/builder-library"
)

var s l.Scheduler

func main() {
	s.Setup(os.Getenv("NATS_URI"))

	s.ProcessRequest("routers.create", "router.create")
	s.ProcessRequest("routers.delete", "router.delete")

	s.ProcessSuccessResponse("router.create.done", "router.create", "routers.create.done")
	s.ProcessSuccessResponse("router.delete.done", "router.delete", "routers.delete.done")

	s.ProcessFailedResponse("router.create.error", "routers.create.error")
	s.ProcessFailedResponse("router.delete.error", "routers.delete.error")

	runtime.Goexit()
}
