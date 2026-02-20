//go:build linux

package system

import (
	"os"
	"time"

	"golang.org/x/sys/unix"
)

func fileCreatedAt(path string, _ os.FileInfo) (time.Time, bool) {
	// FileInfo does not include the original path, and Linux creation time
	// requires statx(2) by path.
	var stx unix.Statx_t
	if err := unix.Statx(unix.AT_FDCWD, path, unix.AT_STATX_SYNC_AS_STAT, unix.STATX_BTIME, &stx); err != nil {
		return time.Time{}, false
	}
	if stx.Mask&unix.STATX_BTIME == 0 || (stx.Btime.Sec == 0 && stx.Btime.Nsec == 0) {
		return time.Time{}, false
	}

	return time.Unix(int64(stx.Btime.Sec), int64(stx.Btime.Nsec)), true
}
