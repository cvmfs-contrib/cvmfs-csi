// Copyright CERN.
//
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package node

import (
	"bytes"
	goexec "os/exec"

	"github.com/cernops/cvmfs-csi/internal/exec"

	mount "k8s.io/mount-utils"
)

type (
	mountState int
)

const (
	msUnknown mountState = iota
	msNotMounted
	msMounted
	msCorrupted
)

var (
	dummyMounter = mount.New("")
)

func (ms mountState) String() string {
	return [...]string{
		"UNKNOWN",
		"NOT_MOUNTED",
		"MOUNTED",
		"CORRUPTED",
	}[int(ms)]
}

func getMountState(p string) (mountState, error) {
	isNotMnt, err := mount.IsNotMountPoint(dummyMounter, p)
	if err != nil {
		if mount.IsCorruptedMnt(err) {
			return msCorrupted, nil
		}

		return msUnknown, err
	}

	if !isNotMnt {
		return msMounted, nil
	}

	return msNotMounted, nil
}

func bindMount(from, to string) error {
	_, err := exec.CombinedOutput(goexec.Command("mount", "--bind", from, to))
	return err
}

func slaveRecursiveBind(from, to string) error {
	_, err := exec.CombinedOutput(goexec.Command(
		"mount",
		from,
		to,

		// We bindmount recursively in order to retain any
		// existing CVMFS mounts inside of the autofs root.
		"--rbind",

		// We expect the autofs root in /cvmfs to be already marked
		// as shared, making it possible to send and receive mount
		// and unmount events between bindmounts. We need to make event
		// propagation one-way only (from autofs root to bindmounts)
		// however, because, when unmounting, we do so recursively, and
		// this would then mean attempting to unmount autofs-CVMFS mounts
		// in the rest of the bindmounts (used by other Pods on the node
		// that also use CVMFS), which is not desirable of course.
		"--make-slave",
	))

	return err
}

func unmount(mountpoint string, extraArgs ...string) error {
	out, err := exec.CombinedOutput(goexec.Command("umount", append(extraArgs, mountpoint)...))
	if err != nil {
		// There are no well-defined exit codes for cases of "not mounted"
		// and "doesn't exist". We need to check the output.
		if bytes.HasSuffix(out, []byte(": not mounted")) ||
			bytes.Contains(out, []byte("No such file or directory")) {
			return nil
		}
	}

	return err
}

func recursiveUnmount(mountpoint string) error {
	// We need recursive unmount because there are live mounts inside the bindmount.
	// Unmounting only the upper autofs mount would result in EBUSY.
	return unmount(mountpoint, "--recursive")
}
