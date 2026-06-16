package socketscan

import (
	"cmp"
	"strconv"
	"strings"

	"github.com/prometheus/procfs"
)

// ProcessInfo holds information about a process associated with a socket
type ProcessInfo struct {
	PID     int
	FD      int
	Comm    string
	Cmdline string
	Exe     string
}

// IsEq -
func (p ProcessInfo) IsEq(other ProcessInfo) bool {
	return p == other
}

// Cmp compares two ProcessInfo instances
func (p ProcessInfo) Cmp(o ProcessInfo) int {
	return cmp.Or(
		cmp.Compare(p.PID, o.PID),
		cmp.Compare(p.FD, o.FD),
		cmp.Compare(p.Comm, o.Comm),
		cmp.Compare(p.Cmdline, o.Cmdline),
		cmp.Compare(p.Exe, o.Exe),
	)
}

func makeInodeToProc() (map[uint64][]ProcessInfo, error) {
	fs, err := procfs.NewFS("/proc")
	if err != nil {
		return nil, err
	}

	var procs procfs.Procs
	if procs, err = fs.AllProcs(); err != nil {
		return nil, err
	}

	result := make(map[uint64][]ProcessInfo)

	for _, p := range procs {
		fds, e := p.FileDescriptors()
		if e != nil {
			continue
		}

		targets, e := p.FileDescriptorTargets()
		if e != nil {
			continue
		}

		comm, e := p.Comm()
		if e != nil {
			continue
		}
		cmdlineParts, e := p.CmdLine()
		if e != nil {
			continue
		}
		exe, e := p.Executable()
		if e != nil {
			continue
		}

		cmdline := strings.Join(cmdlineParts, " ")

		for i, target := range targets {
			inode, ok := extractSocketInode(target)
			if !ok {
				continue
			}

			fd := -1
			if i < len(fds) {
				fd = int(fds[i]) //nolint:gosec
			}

			result[inode] = append(result[inode], ProcessInfo{
				PID:     p.PID,
				FD:      fd,
				Comm:    comm,
				Cmdline: cmdline,
				Exe:     exe,
			})
		}
	}

	return result, nil
}

func extractSocketInode(s string) (uint64, bool) {
	if !strings.HasPrefix(s, "socket:[") || !strings.HasSuffix(s, "]") {
		return 0, false
	}

	raw := strings.TrimSuffix(strings.TrimPrefix(s, "socket:["), "]")
	inode, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, false
	}

	return inode, true
}
