package command

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func RunCommand(cmdName string, cmdArgs []string, timeout int) (int, string) {

	// the command we're going to run
	cmd := exec.Command(cmdName, cmdArgs...)

	// Copy calling ENV to the Cmd
	cmd.Env = os.Environ()

	// assign vars for output and stderr
	var output bytes.Buffer
	var stderr bytes.Buffer

	var combined string

	// get the stdout and stderr and assign to pointers
	cmd.Stderr = &stderr
	cmd.Stdout = &output

	// Start the command
	if err := cmd.Start(); err != nil {
		log.Fatalf("Command not found: %s", cmdName)
	}

	timer := time.AfterFunc(time.Second*time.Duration(timeout), func() {
		// if timeout is set, kill the process
		if timeout > 0 {
			err := cmd.Process.Kill()
			if err != nil {
				panic(err)
			}
		}
	})

	// Here's the good stuff
	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// Command ! exit 0, capture it
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				combined = stderr.String() + output.String()
				return status.ExitStatus(), combined
			}
		} else {
			log.Fatalf("cmd.Wait: %v", err)
		}
		timer.Stop()
	}
	// We didn't get captured, continue!
	return 0, output.String()

}
