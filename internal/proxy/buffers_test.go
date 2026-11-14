// Copyright 2024 lucdoe
// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBufferPoolHandsOutFullSizeBuffers(t *testing.T) {
	b := buffers.Get()
	assert.Len(t, b, copyBufferSize)

	buffers.Put(b[:10]) // a resliced buffer comes back at full length
	assert.Len(t, buffers.Get(), copyBufferSize)
}

func TestBufferPoolDropsBuffersThatAreTooSmall(t *testing.T) {
	buffers.Put(make([]byte, 1024))

	for i := 0; i < 10; i++ {
		assert.Len(t, buffers.Get(), copyBufferSize)
	}
}
