// worker_limit_unix.go
//go:build linux || darwin

package git

import "syscall"

// fileDescriptorLimit reports the process's current soft RLIMIT_NOFILE.
// Used to size the branch-evaluation pool so it cannot outrun the file
// descriptors available for subprocess pipes.
func fileDescriptorLimit() (soft uint64, ok bool) {
	var rlimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlimit); err != nil {
		return 0, false
	}
	return uint64(rlimit.Cur), true
}
