package agent

import (
	"os"
	"time"
)

// fileInfoAdapter exposes just the bits the runtime test needs from os.FileInfo.
// Lives in a non-_test file so it is reachable from runtime_test.go without
// importing os there (keeps the test file's import set focused on the public
// surface being verified).
type fileInfoAdapter struct {
	fi os.FileInfo
}

func (a fileInfoAdapter) size() int64        { return a.fi.Size() }
func (a fileInfoAdapter) modTime() time.Time { return a.fi.ModTime() }

func osStat(path string) (fileInfoAdapter, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return fileInfoAdapter{}, err
	}
	return fileInfoAdapter{fi: fi}, nil
}
