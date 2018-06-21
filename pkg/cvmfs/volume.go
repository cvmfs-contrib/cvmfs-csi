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
	cvmfsCacheRoot   = "/var/cache/cvmfs"
	volumeRootPrefix = PluginFolder + "/controller/volumes/vol-"
)

var (
	cvmfsUid = -1
)

func init() {
	u, err := user.Lookup("cvmfs")
	if err != nil {
		panic(err)
	}

	cvmfsUid, _ = strconv.Atoi(u.Uid)
}

func getVolumeCachePath(volUuid string) string {
	return path.Join(cvmfsCacheRoot, "csi-"+volUuid)
}

func getVolumeSharedCachePath(volUuid string) string {
	return path.Join(getVolumeCachePath(volUuid), "shared")
}

func getVolumeRootPath(volUuid string) string {
	return volumeRootPrefix + volUuid
}

func createVolumeCache(volUuid string) error {
	cachePath := getVolumeCachePath(volUuid)

	if err := os.MkdirAll(cachePath, 0755); err != nil {
		return err
	}

	if err := os.Chown(cachePath, cvmfsUid, 0); err != nil {
		return err
	}

	return nil
}

func purgeVolumeCache(volUuid string) error {
	return os.RemoveAll(getVolumeCachePath(volUuid))
}

func mountCvmfs(volOptions *volumeOptions, volUuid string) error {
	return execCommandAndValidate("mount",
		"-t", "cvmfs",
		volOptions.Repository, getVolumeRootPath(volUuid),
		"-o", "config="+getConfigFilePath(volUuid),
	)
}

func mountVolume(mountPoint string, volOptions *volumeOptions, volUuid string) error {
	volRoot := getVolumeRootPath(volUuid)

	if err := createMountPoint(volRoot); err != nil {
		return err
	}

	if err := mountCvmfs(volOptions, volUuid); err != nil {
		return err
	}

	return bindMount(volRoot, mountPoint)
}

func bindMount(from, to string) error {
	if err := execCommandAndValidate("mount", "--bind", from, to); err != nil {
		return fmt.Errorf("failed bind-mount of %s to %s: %v", from, to, err)
	}

	return execCommandAndValidate("mount", "-o", "remount,ro,bind", to)
}

func unmountVolume(mountPoint, volUuid string) error {
	if err := execCommandAndValidate("umount", mountPoint); err != nil {
		return err
	}

	return execCommandAndValidate("umount", getVolumeRootPath(volUuid))
}

func createMountPoint(p string) error {
	return os.MkdirAll(p, 0755)
}

func isMountPoint(p string) (bool, error) {
	notMnt, err := mount.New("").IsLikelyNotMountPoint(p)
	if err != nil {
		return false, err
	}

	return !notMnt, nil
}
