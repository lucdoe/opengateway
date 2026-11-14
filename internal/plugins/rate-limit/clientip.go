// Copyright 2024 lucdoe
// SPDX-License-Identifier: Apache-2.0

package ratelimit

import (
	"net"
	"net/http"
)

// clientIP returns the address of the peer that opened the connection,
// without the port. The port changes with every connection, so it must not be
// part of anything that identifies a client.
//
// Forwarding headers such as X-Forwarded-For are ignored on purpose: a client
// can set them to any value.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
