package git

import (
	"fmt"
	"runtime"
)

// BranchInfo holds metadata about a branch collected during the clean pass.
type BranchInfo struct {
	Name        string
	DaysOld     int
	Behind      int
	Merged      bool
	Current     bool
	Protected   bool
	Interrupted bool // true if evaluation was cut short by cancellation (e.g. Ctrl+C)
}

// behindResult is produced by a worker goroutine and consumed exclusively by
// the CleanGit goroutine, so staleMap is never written from more than one
// goroutine at a time.
type behindResult struct {
	name        string
	behind      int
	interrupted bool
}

const (
	// fdPerWorker is a conservative estimate of descriptors held by one
	// in-flight git subprocess call: the pipe pair backing captured stdout,
	// plus headroom for stdin/stderr and transient fds git itself opens
	// while reading refs and objects.
	fdPerWorker = 8

	// fdReserve is set aside for Forge's own stdio and any open config/log
	// handles, so the worker budget never claims the entire limit.
	fdReserve = 32

	// absoluteWorkerCeiling is a hard backstop independent of detected
	// limits, so a misreported or unusually high rlimit cannot spawn an
	// unreasonable number of concurrent git processes.
	absoluteWorkerCeiling = 32
)

// reportInterrupted prints a summary when evaluation is cut short by
// cancellation and returns a non-nil error so callers can distinguish this
// from a clean run.
func reportInterrupted(evaluated, total int) error {
	fmt.Println()
	if total > 0 {
		fmt.Printf("Interrupted — %d/%d branches evaluated before cancellation.\n", evaluated, total)
	} else {
		fmt.Println("Interrupted before branch evaluation began.")
	}
	fmt.Println("No branches were deleted.")
	return fmt.Errorf("[git clean]: clean cancelled")
}

// computeWorkerCount derives the branch-evaluation pool size from runtime
// signals rather than a fixed constant:
//
//   - runtime.GOMAXPROCS(0) as the CPU-availability signal. On Go 1.25+ and
//     Linux this is already container-aware — it reflects a cgroup CPU
//     quota, not just host core count — and self-updates if the quota
//     changes, so no external library is needed for this to behave
//     correctly under Docker/Kubernetes.
//   - the process's soft RLIMIT_NOFILE where the platform exposes one
//     (Linux, macOS), since each worker holds one subprocess's pipes open
//     at a time.
//   - the number of branches actually awaiting evaluation, so idle workers
//     are never created.
//
// cfgOverride, when > 0, comes from forge.toml (max_workers) and takes
// precedence over both detected signals.
func computeWorkerCount(jobCount int, cfgOverride int) int {
	if jobCount == 0 {
		return 0
	}
	if cfgOverride > 0 {
		return clamp(cfgOverride, 1, jobCount)
	}

	n := runtime.GOMAXPROCS(0) * 2

	if limit, ok := fileDescriptorLimit(); ok {
		budget := int((limit - fdReserve) / fdPerWorker)
		if budget < n {
			n = budget
		}
	}

	n = clamp(n, 1, absoluteWorkerCeiling)
	if n > jobCount {
		n = jobCount
	}
	return n
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
