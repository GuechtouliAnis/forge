// worker_limit_other.go
//go:build !linux && !darwin

package git

// FileDescriptorLimit is unavailable on platforms without a POSIX rlimit
// model (Windows). The pool falls back to the CPU-based signal alone.
func FileDescriptorLimit() (soft uint64, ok bool) {
	return 0, false
}
