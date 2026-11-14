// Copyright 2024 lucdoe
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logger

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAppendsJSONLinesToTheFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "gateway.log")
	require.NoError(t, os.WriteFile(path, []byte("earlier line\n"), 0o644))

	l, err := New(Config{FilePath: path})
	require.NoError(t, err)
	l.Request(context.Background(), slog.Int("status", 204))
	require.NoError(t, l.Close())

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	require.Len(t, lines, 2)
	assert.Equal(t, "earlier line", lines[0], "the file is appended to")

	var entry map[string]any
	require.NoError(t, json.Unmarshal([]byte(lines[1]), &entry))
	assert.Equal(t, "INFO", entry["level"])
	assert.Equal(t, "request", entry["msg"])
	assert.Equal(t, float64(204), entry["status"])
	assert.NotEmpty(t, entry["time"])
}

func TestNewWritesToStdoutWithoutAFilePath(t *testing.T) {
	read, write, err := os.Pipe()
	require.NoError(t, err)
	stdout := os.Stdout
	os.Stdout = write
	t.Cleanup(func() { os.Stdout = stdout })

	l, err := New(Config{})
	require.NoError(t, err)
	l.Request(context.Background(), slog.Int("status", 200))
	require.NoError(t, l.Close())
	write.Close()

	out, err := io.ReadAll(read)
	require.NoError(t, err)
	assert.Contains(t, string(out), `"msg":"request","status":200`)
	assert.Nil(t, l.file, "stdout is not closed")
}

func TestNewFailsForAFileItCannotOpen(t *testing.T) {
	l, err := New(Config{FilePath: filepath.Join(t.TempDir(), "missing", "gateway.log")})

	assert.Nil(t, l)
	assert.ErrorContains(t, err, "opening log file")
}
