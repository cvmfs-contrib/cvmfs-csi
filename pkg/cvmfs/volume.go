package cvmfs

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"strconv"

	"k8s.io/kubernetes/pkg/util/mount"
)

const (
	cvmfsCacheRoot = "/var/cache/cvmfs"
)

var (
	cvmfsUid     = -1
	dummyMounter = mount.New("") // Used in isMountPoint()
)

func init() {
	u, err := user.Lookup("cvmfs")
	if err != nil {
		panic(err)
	}

	cvmfsUid, _ = strconv.Atoi(u.Uid)
}

func getVolumeCachePath(volId volumeID) string {
	return path.Join(cvmfsCacheRoot, "csi-"+string(volId))
}

func getVolumeSharedCachePath(volId volumeID) string {
	return path.Join(getVolumeCachePath(volId), "shared")
}

func createVolumeCache(volId volumeID) error {
	cachePath := getVolumeCachePath(volId)

	if err := os.MkdirAll(cachePath, 0755); err != nil {
		return err
	}

	if err := os.Chown(cachePath, cvmfsUid, 0); err != nil {
		return err
	}

	return nil
}

func purgeVolumeCache(volId volumeID) error {
	return os.RemoveAll(getVolumeCachePath(volId))
}

func mountCvmfs(volOptions *volumeOptions, volId volumeID, mountPoint string) error {
	return execCommandAndValidate("mount",
		"-t", "cvmfs",
		volOptions.Repository, mountPoint,
	)
}

func bindMount(from, to string) error {
	if err := execCommandAndValidate("mount", "--bind", from, to); err != nil {
		return fmt.Errorf("failed bind-mount of %s to %s: %v", from, to, err)
	}

	return execCommandAndValidate("mount", "-o", "remount,ro,bind", to)
}

func unmountVolume(mountPoint string) error {
	return execCommandAndValidate("umount", mountPoint)
}

func createMountPoint(p string) error {
	return os.MkdirAll(p, 0755)
}

func isMountPoint(p string) (bool, error) {
	notMnt, err := dummyMounter.IsLikelyNotMountPoint(p)
	if err != nil {
		return false, err
	}

	return !notMnt, nil
}
