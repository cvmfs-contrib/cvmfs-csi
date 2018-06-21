package cvmfs

import (
	"fmt"
	"os/exec"
)

func execCommand(program string, args ...string) ([]byte, error) {
	cmd := exec.Command(program, args[:]...)
	return cmd.CombinedOutput()
}

func execCommandAndValidate(program string, args ...string) error {
	if out, err := execCommand(program, args[:]...); err != nil {
		return fmt.Errorf("cvmfs: %s failed with following error: %v\ncvmfs: %s output: %s", program, err, program, out)
	}

	return nil
}
