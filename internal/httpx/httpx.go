// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

// Package httpx holds small HTTP helpers several plugins share.
package httpx

import (
	"math"
	"net"
	"net/http"
	"strconv"
	"time"
)

// ClientIP returns the address of the peer that opened the connection,
// without the port. The port changes with every connection, so it must not be
// part of anything that identifies a client.
//
// Forwarding headers such as X-Forwarded-For are ignored on purpose: a client
// can set them to any value.
func ClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// Seconds rounds d up to whole seconds, as Retry-After wants them, so a
// client that waits that long is past the moment d ends.
func Seconds(d time.Duration) string {
	return strconv.Itoa(int(math.Ceil(d.Seconds())))
}
