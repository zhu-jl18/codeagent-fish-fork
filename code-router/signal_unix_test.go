//go:build unix || darwin || linux
// +build unix darwin linux

package main

import (
	"os"
	"os/exec"
	"syscall"
	"testing"
)

type fallbackProcess struct {
	pid      int
	signaled os.Signal
	killed   bool
}

func (p *fallbackProcess) Pid() int { return p.pid }
func (p *fallbackProcess) Kill() error {
	p.killed = true
	return nil
}
func (p *fallbackProcess) Signal(sig os.Signal) error {
	p.signaled = sig
	return nil
}

func TestPrepareCommandForSignalsUnix(t *testing.T) {
	cmd := exec.Command("sleep", "0")
	prepareCommandForSignals(cmd)
	if cmd.SysProcAttr == nil {
		t.Fatalf("expected SysProcAttr to be initialized")
	}
	if !cmd.SysProcAttr.Setpgid {
		t.Fatalf("expected Setpgid=true for child process isolation")
	}
}

func TestProcessGroupLeaderPIDOnlyForRealProcess(t *testing.T) {
	if got := processGroupLeaderPID(&fallbackProcess{pid: 123}); got != 0 {
		t.Fatalf("expected non-real process to return pgid 0, got %d", got)
	}

	rp := &realProcess{proc: &os.Process{Pid: 456}}
	if got := processGroupLeaderPID(rp); got != 456 {
		t.Fatalf("expected real process pid 456, got %d", got)
	}
}

func TestUnixSignalFallbackForNonRealProcess(t *testing.T) {
	proc := &fallbackProcess{pid: 999}
	if err := sendTermSignal(proc); err != nil {
		t.Fatalf("sendTermSignal returned error: %v", err)
	}
	if proc.signaled != syscall.SIGTERM {
		t.Fatalf("expected SIGTERM fallback signal, got %v", proc.signaled)
	}

	if err := sendKillSignal(proc); err != nil {
		t.Fatalf("sendKillSignal returned error: %v", err)
	}
	if !proc.killed {
		t.Fatalf("expected fallback process Kill to be called")
	}
}

