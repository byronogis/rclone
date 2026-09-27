package bisync

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
