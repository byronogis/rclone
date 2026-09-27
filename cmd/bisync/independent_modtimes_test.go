package bisync

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rclone/rclone/fs/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileInfoEqualIndependentModtimes(t *testing.T) {
	t1 := time.Date(2026, 9, 27, 10, 0, 0, 0, time.UTC)
	t2 := t1.Add(5 * time.Minute)

	path1 := newFileList()
	path2 := newFileList()
	path1.put("same.txt", 4, t1, "", "-", "-")
	path2.put("same.txt", 4, t2, "", "-", "-")

	b := &bisyncRun{
		opt: &Options{
			Compare: CompareOpt{
				Size:    true,
				Modtime: true,
			},
			IndependentModtimes: true,
		},
	}

	assert.True(t, b.fileInfoEqual("same.txt", "same.txt", path1, path2),
		"cross-path modtime differences must be allowed in independent modtime mode")

	path2.get("same.txt").size = 5
	assert.False(t, b.fileInfoEqual("same.txt", "same.txt", path1, path2),
		"independent modtime mode must still validate other enabled comparison fields")
}

func TestRefreshTransferredMetadata(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	srcRoot := filepath.Join(root, "src")
	dstRoot := filepath.Join(root, "dst")
	require.NoError(t, os.MkdirAll(srcRoot, 0o755))
	require.NoError(t, os.MkdirAll(dstRoot, 0o755))

	const remote = "file.txt"
	srcPath := filepath.Join(srcRoot, remote)
	dstPath := filepath.Join(dstRoot, remote)
	require.NoError(t, os.WriteFile(srcPath, []byte("source"), 0o600))
	require.NoError(t, os.WriteFile(dstPath, []byte("destination"), 0o600))

	srcTime := time.Date(2026, 9, 27, 10, 0, 1, 0, time.Local)
	dstTime := time.Date(2026, 9, 27, 10, 0, 9, 0, time.Local)
	require.NoError(t, os.Chtimes(srcPath, srcTime, srcTime))
	require.NoError(t, os.Chtimes(dstPath, dstTime, dstTime))

	src, err := cache.Get(ctx, srcRoot)
	require.NoError(t, err)
	dst, err := cache.Get(ctx, dstRoot)
	require.NoError(t, err)

	stale := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	srcList := newFileList()
	dstList := newFileList()
	srcList.put(remote, 1, stale, "src-hash", "-", "-")
	dstList.put(remote, 1, stale, "dst-hash", "-", "-")

	b := &bisyncRun{
		opt: &Options{
			Compare: CompareOpt{Modtime: true},
		},
	}

	files := map[string]struct{}{remote: {}}
	require.NoError(t, b.refreshTransferredMetadata(ctx, src, dst, srcList, dstList, files))

	srcObj, err := src.NewObject(ctx, remote)
	require.NoError(t, err)
	dstObj, err := dst.NewObject(ctx, remote)
	require.NoError(t, err)

	assert.Equal(t, srcObj.Size(), srcList.getSize(remote))
	assert.Equal(t, dstObj.Size(), dstList.getSize(remote))
	assert.True(t, srcObj.ModTime(ctx).Equal(srcList.getTime(remote)))
	assert.True(t, dstObj.ModTime(ctx).Equal(dstList.getTime(remote)))
	assert.Equal(t, "src-hash", srcList.getHash(remote), "refresh must preserve existing hash metadata")
	assert.Equal(t, "dst-hash", dstList.getHash(remote), "refresh must preserve existing hash metadata")
}

func TestRefreshTransferredMetadataFailure(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	srcRoot := filepath.Join(root, "src")
	dstRoot := filepath.Join(root, "dst")
	require.NoError(t, os.MkdirAll(srcRoot, 0o755))
	require.NoError(t, os.MkdirAll(dstRoot, 0o755))

	const remote = "file.txt"
	require.NoError(t, os.WriteFile(filepath.Join(srcRoot, remote), []byte("source"), 0o600))
	// Keep the destination in the listing but deliberately omit the actual object.
	// This simulates a successful transfer followed by a metadata refresh failure.

	src, err := cache.Get(ctx, srcRoot)
	require.NoError(t, err)
	dst, err := cache.Get(ctx, dstRoot)
	require.NoError(t, err)

	srcList := newFileList()
	dstList := newFileList()
	srcList.put(remote, 1, time.Time{}, "", "-", "-")
	dstList.put(remote, 1, time.Time{}, "", "-", "-")

	b := &bisyncRun{
		opt: &Options{
			Compare: CompareOpt{Modtime: true},
		},
	}

	files := map[string]struct{}{remote: {}}
	err = b.refreshTransferredMetadata(ctx, src, dst, srcList, dstList, files)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "refresh metadata")
}
