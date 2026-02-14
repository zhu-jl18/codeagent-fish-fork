//go:build unix || darwin || linux
// +build unix darwin linux

package main

import (
	"errors"
	"os/exec"
	"syscall"
)

func processGroupLeaderPID(proc processHandle) int {
	rp, ok := proc.(*realProcess)
	if !ok || rp == nil || rp.proc == nil {
		return 0
	}
	if rp.proc.Pid <= 0 {
		return 0
	}
	return rp.proc.Pid
}

func prepareCommandForSignals(cmd *exec.Cmd) {
	if cmd == nil {
		return
	}
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setpgid = true
}

// sendTermSignal sends SIGTERM to the command process group first, then falls back to the process.
func sendTermSignal(proc processHandle) error {
	if proc == nil {
		return nil
	}
	if pid := processGroupLeaderPID(proc); pid > 0 {
		if err := syscall.Kill(-pid, syscall.SIGTERM); err == nil {
			return nil
		} else if !errors.Is(err, syscall.ESRCH) {
			return err
		}
	}
	return proc.Signal(syscall.SIGTERM)
}

// sendKillSignal sends SIGKILL to the command process group first, then falls back to direct process kill.
func sendKillSignal(proc processHandle) error {
	if proc == nil {
		return nil
	}
	if pid := processGroupLeaderPID(proc); pid > 0 {
		if err := syscall.Kill(-pid, syscall.SIGKILL); err == nil {
			return nil
		} else if !errors.Is(err, syscall.ESRCH) {
			return err
		}
	}
	return proc.Kill()
}
