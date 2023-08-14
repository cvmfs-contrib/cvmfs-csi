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

package mountreconcile

import (
	"bytes"
	"fmt"
	"os"
	goexec "os/exec"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/cvmfs-contrib/cvmfs-csi/internal/exec"
	"github.com/cvmfs-contrib/cvmfs-csi/internal/log"
	"github.com/cvmfs-contrib/cvmfs-csi/internal/mountutils"

	"github.com/moby/sys/mountinfo"
)

const mountPathPrefix = "/cvmfs/"

type Opts struct {
	Period time.Duration
}

func RunBlocking(o *Opts) error {
	t := time.NewTicker(o.Period)

	for {
		select {
		case <-t.C:
			log.Tracef("Reconciling /cvmfs")
			if err := reconcile(); err != nil {
				log.Errorf("Failed to reconcile /cvmfs: %v", err)
			}
		}
	}
}

// List CVMFS mounts in /cvmfs that the kernel knows about.
// We do that by listing mounts in /proc/self/mountinfo and filtering
// those where the device is "fuse" and the mountpoint is rooted in /cvmfs.
func getMountedRepositories() ([]string, error) {
	cvmfsMountInfos, err := mountinfo.GetMounts(func(info *mountinfo.Info) (skip, stop bool) {
		return info.FSType != "fuse" || !strings.HasPrefix(info.Mountpoint, mountPathPrefix),
			false
	})
	if err != nil {
		return nil, err
	}

	repositories := make([]string, len(cvmfsMountInfos))

	for i := range cvmfsMountInfos {
		repositories[i] = cvmfsMountInfos[i].Mountpoint[len(mountPathPrefix):]
	}

	return repositories, nil
}

func doCvmfsTalk(repo, command string) ([]byte, error) {
	return exec.CombinedOutput(
		goexec.Command(
			"cvmfs_talk",
			"-i", repo,
			command,
		),
	)
}

// repoNeedsUnmount checks if a /cvmfs/<repo> mountpoint is healthy.
// Because mounts under /cvmfs are managed by autofs, we cannot check
// them directly (with a stat() for example), as this would trigger
// autofs's unmount timeout reset. Instead, we use cvmfs_talk to probe
// for CVMFS client, and only if this fails with "Connection refused",
// we use stat("/cvmfs/<repo>") to check the mount.
func repoNeedsUnmount(repo string) (bool, error) {
	out, err := doCvmfsTalk(repo, "mountpoint")
	if err == nil {
		if bytes.HasPrefix(out, []byte(mountPathPrefix)) {
			return false, nil
		}

		// The mountpoint is outside of /cvmfs?
		// Normally this shouldn't happen, report an error.
		return false, fmt.Errorf(
			"repository is mounted at an unexpected location \"%s\", expected /cvmfs", out)
	}

	// The CVMFS client exited unexpectedly, and the watchdog
	// didn't remount it automatically.
	const cvmfsErrConnRefused = "(111 - Connection refused)\x0A"
	const cvmfsErrClientNotRunning = "Seems like CernVM-FS is not running"

	if bytes.HasSuffix(out, []byte(cvmfsErrConnRefused)) ||
		bytes.HasPrefix(out, []byte(cvmfsErrClientNotRunning)) {
		// It seems that the CVMFS client exited.
		// Use stat syscall to check for ENOTCONN, i.e. the mount is corrupted,
		// confirming what cvmfs_talk returned.

		_, err := os.Stat(path.Join(mountPathPrefix, repo))
		if err != nil {
			if err.(*os.PathError).Err == syscall.ENOTCONN {
				return true, nil
			}

			// It's something else.
			return false, fmt.Errorf("unexpected error from stat: %v", err)
		}

		// stat should have failed! Fall through and fail.
	}

	// If we got here, the error reported by cvmfs_talk
	// is something else and we should fail too.
	return false, fmt.Errorf("failed to talk to CVMFS client (%v): %s", err, out)
}

func reconcile() error {
	// List mounted CVMFS repositories in /cvmfs.

	mountedRepos, err := getMountedRepositories()
	if err != nil {
		return err
	}

	log.Tracef("CVMFS mounts in /cvmfs: %v", mountedRepos)

	// Check each mountpoint we found above. In case it's corrupted,
	// we unmount it. autofs will then take care of automatically remounting
	// it when the path is accessed.

	for _, repo := range mountedRepos {
		needsUnmount, err := repoNeedsUnmount(repo)
		mountpoint := path.Join(mountPathPrefix, repo)

		if err != nil {
			log.Errorf("Failed to reconcile %s: %v", mountpoint, err)
			continue
		}

		if needsUnmount {
			log.Infof("%s is corrupted, unmounting", mountpoint)

			if err := mountutils.Unmount(mountpoint); err != nil {
				log.Errorf("Failed to unmount %s during mount reconciliation: %v", mountpoint, err)
				continue
			}
		}
	}

	return nil
}
