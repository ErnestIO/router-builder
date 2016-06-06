/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import "runtime"

func main() {
	n := natsClient()
	r := redisClient()

	// Process requests
	processRequest(n, r, "routers.create", "router.create")
	processRequest(n, r, "routers.delete", "router.delete")

	// Process resulting success
	processResponse(n, r, "router.create.done", "routers.create.", "router.create", "completed")
	processResponse(n, r, "router.delete.done", "routers.delete.", "router.delete", "completed")

	// Process resulting errors
	processResponse(n, r, "router.create.error", "routers.create.", "router.create", "errored")
	processResponse(n, r, "router.delete.error", "routers.delete.", "router.delete", "errored")

	runtime.Goexit()
}
