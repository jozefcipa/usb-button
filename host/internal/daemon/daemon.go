package daemon

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

func getPIDFilePath() string {
	home, err := os.UserHomeDir()
	if err == nil {
		dir := filepath.Join(home, ".cache")
		_ = os.MkdirAll(dir, 0755)
		return filepath.Join(dir, "hid_listener.pid")
	}
	return filepath.Join(os.TempDir(), fmt.Sprintf("hid_listener_%d.pid", os.Getuid()))
}

func Start() {
	exe, err := os.Executable()
	if err != nil {
		exe = os.Args[0]
	}
	cmd := exec.Command(exe)

	// Detach from terminal: daemon has no stdin/stdout/stderr (same pattern as many Unix daemons)
	devNull, err := os.Open(os.DevNull)
	if err != nil {
		log.Fatalf("failed to open %s: %v", os.DevNull, err)
	}
	defer devNull.Close()
	cmd.Stdin = devNull
	cmd.Stdout = devNull
	cmd.Stderr = devNull
	cmd.Dir = "/"

	if err := cmd.Start(); err != nil {
		log.Fatalf("failed to start daemon: %v", err)
	}
	if err := cmd.Wait(); err != nil {
		log.Fatalf("daemon process exited with error: %v", err)
	}

	pidPath := getPIDFilePath()
	_ = os.WriteFile(pidPath, []byte(strconv.Itoa(cmd.Process.Pid)+"\n"), 0644)
	fmt.Fprintf(os.Stderr, "Started daemon PID %d (PID file: %s). Use 'hid_listener stop' to stop.\n", cmd.Process.Pid, pidPath)
}

func Stop() {
	path := getPIDFilePath()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "Daemon not running (no PID file).")
			return
		}
		log.Fatalf("reading PID file: %v", err)
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid PID file %s: %v\n", path, err)
		_ = os.Remove(path)
		return
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not find process %d: %v\n", pid, err)
		_ = os.Remove(path)
		return
	}

	if err := proc.Signal(syscall.SIGTERM); err != nil {
		if err.Error() == "os: process already finished" || strings.Contains(err.Error(), "already finished") {
			fmt.Fprintln(os.Stderr, "Daemon was not running (stale PID file removed).")
		} else {
			fmt.Fprintf(os.Stderr, "Failed to stop daemon (PID %d): %v\n", pid, err)
			return
		}
	} else {
		fmt.Fprintf(os.Stderr, "Sent SIGTERM to daemon (PID %d). It will exit and remove the PID file.\n", pid)
	}

	_ = os.Remove(path)
}
