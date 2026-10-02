// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package scan

import (
	"archive/tar"
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"gotest.tools/v3/assert"
)

func TestGetContextFromReaderUntarsAbsoluteLinks(t *testing.T) {
	t.Run("symlink", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Windows does not support absolute Unix symlinks")
		}

		archiveData := createTarArchive(t, []tarEntry{
			{
				header: tar.Header{
					Name:     "var/run",
					Typeflag: tar.TypeSymlink,
					Linkname: "/run",
					Mode:     0777,
					Uid:      -1,
					Gid:      -1,
				},
			},
			{
				header: tar.Header{
					Name: "var/run/existing/non-existing/file",
					Mode: 0644,
					Uid:  -1,
					Gid:  -1,
				},
				content: "content",
			},
		})

		destination := t.TempDir()
		scanner := &Scanner{destinationFolder: destination}
		assert.NilError(t, scanner.getContextFromReader(bytes.NewReader(archiveData)))

		content, err := os.ReadFile(filepath.Join(destination, "run", "existing", "non-existing", "file"))
		assert.NilError(t, err)
		assert.Equal(t, string(content), "content")

		linkTarget, err := os.Readlink(filepath.Join(destination, "var", "run"))
		assert.NilError(t, err)
		assert.Equal(t, linkTarget, "/run")
	})

	t.Run("hardlink", func(t *testing.T) {
		archiveData := createTarArchive(t, []tarEntry{
			{
				header: tar.Header{
					Name: "usr/bin/perlbug",
					Mode: 0755,
					Uid:  -1,
					Gid:  -1,
				},
				content: "hello",
			},
			{
				header: tar.Header{
					Name:     "usr/bin/perlthanks",
					Typeflag: tar.TypeLink,
					Linkname: "/usr/bin/perlbug",
					Mode:     0755,
					Uid:      -1,
					Gid:      -1,
				},
			},
		})

		destination := t.TempDir()
		scanner := &Scanner{destinationFolder: destination}
		assert.NilError(t, scanner.getContextFromReader(bytes.NewReader(archiveData)))

		targetInfo, err := os.Stat(filepath.Join(destination, "usr", "bin", "perlbug"))
		assert.NilError(t, err)
		linkInfo, err := os.Stat(filepath.Join(destination, "usr", "bin", "perlthanks"))
		assert.NilError(t, err)
		assert.Assert(t, os.SameFile(targetInfo, linkInfo))
	})
}

type tarEntry struct {
	header  tar.Header
	content string
}

func createTarArchive(t *testing.T, entries []tarEntry) []byte {
	t.Helper()

	var archive bytes.Buffer
	writer := tar.NewWriter(&archive)
	for _, entry := range entries {
		entry.header.Size = int64(len(entry.content))
		assert.NilError(t, writer.WriteHeader(&entry.header))
		if entry.content != "" {
			_, err := writer.Write([]byte(entry.content))
			assert.NilError(t, err)
		}
	}
	assert.NilError(t, writer.Close())
	return archive.Bytes()
}
