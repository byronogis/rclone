// Test Webdav filesystem interface
package webdav

import (
	"testing"
	"time"

	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fstest"
	"github.com/rclone/rclone/fstest/fstests"
)

// TestIntegration runs integration tests against the remote
func TestIntegration(t *testing.T) {
	fstests.Run(t, &fstests.Opt{
		RemoteName: "TestWebdavNextcloud:",
		NilObject:  (*Object)(nil),
		ChunkedUpload: fstests.ChunkedUploadConfig{
			MinChunkSize: 1 * fs.Mebi,
		},
	})
}

// TestIntegration runs integration tests against the remote
func TestIntegration2(t *testing.T) {
	if *fstest.RemoteName != "" {
		t.Skip("skipping as -remote is set")
	}
	fstests.Run(t, &fstests.Opt{
		RemoteName: "TestWebdavOwncloud:",
		NilObject:  (*Object)(nil),
		ChunkedUpload: fstests.ChunkedUploadConfig{
			Skip: true,
		},
	})
}

// TestIntegration runs integration tests against the remote
func TestIntegration3(t *testing.T) {
	if *fstest.RemoteName != "" {
		t.Skip("skipping as -remote is set")
	}
	fstests.Run(t, &fstests.Opt{
		RemoteName: "TestWebdavRclone:",
		NilObject:  (*Object)(nil),
		ChunkedUpload: fstests.ChunkedUploadConfig{
			Skip: true,
		},
	})
}

// TestIntegration runs integration tests against the remote
func TestIntegration4(t *testing.T) {
	if *fstest.RemoteName != "" {
		t.Skip("skipping as -remote is set")
	}
	fstests.Run(t, &fstests.Opt{
		RemoteName: "TestWebdavNTLM:",
		NilObject:  (*Object)(nil),
	})
}

func (f *Fs) SetUploadChunkSize(cs fs.SizeSuffix) (fs.SizeSuffix, error) {
	return f.setUploadChunkSize(cs)
}

func TestTrustServerModTimeDisabledByDefault(t *testing.T) {
	f := &Fs{
		features:  &fs.Features{},
		precision: fs.ModTimeNotSupported,
	}

	err := f.setQuirks(t.Context(), "other")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := f.Precision(), fs.ModTimeNotSupported; got != want {
		t.Fatalf("Precision() = %v, want %v", got, want)
	}
}

func TestTrustServerModTime(t *testing.T) {
	f := &Fs{
		opt: Options{
			TrustServerModTime: true,
		},
		features:  &fs.Features{},
		precision: fs.ModTimeNotSupported,
	}

	err := f.setQuirks(t.Context(), "other")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := f.Precision(), time.Second; got != want {
		t.Fatalf("Precision() = %v, want %v", got, want)
	}
	if !f.useStandardProps {
		t.Fatal("trust_server_modtime must enable standard WebDAV properties")
	}
}
