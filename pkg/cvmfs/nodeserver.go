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
package cvmfs

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path"
	"strconv"
	"time"

	"github.com/golang/glog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/container-storage-interface/spec/lib/go/csi"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
	mount "k8s.io/mount-utils"
)

const (
	cvmfsConfigRoot = "/etc/cvmfs"
	unmountTimeout  = 6 * time.Second
)

var cvmfsUid int

func init() {
	u, err := user.Lookup("cvmfs")
	if err != nil {
		panic(err)
	}

	cvmfsUid, _ = strconv.Atoi(u.Uid)
}

type nodeServer struct {
	*csicommon.DefaultNodeServer
	Name           string
	mounter        mount.MounterForceUnmounter
	cvmfsCacheRoot string
}

func NewNodeServer(d *cvmfsDriver) *nodeServer {
	return &nodeServer{
		DefaultNodeServer: csicommon.NewDefaultNodeServer(d.driver),
		Name:              d.Name,
		mounter:           &mount.Mounter{},
		cvmfsCacheRoot:    d.cvmfsCacheRoot,
	}
}

func getConfigFilePath(volId string) string {
	return path.Join(cvmfsConfigRoot, "config-"+volId)
}

func (ns *nodeServer) stage(stagingTargetPath, volId string, options map[string]string) error {
	var repository string
	var ok bool
	var err error
	if repository, ok = options["repository"]; !ok {
		msg := "missing required field 'repository'"
		glog.Errorf("invalid volume attributes for volume %s: %s", volId, msg)
		return status.Error(codes.InvalidArgument, msg)
	}

	cachePath := getVolumeCachePath(ns.cvmfsCacheRoot, volId)
	confData := cvmfsConfigData{
		VolumeId:  volId,
		Tag:       options["tag"],
		Hash:      options["hash"],
		Proxy:     options["proxy"],
		CachePath: cachePath,
	}

	if confData.Hash == "" && confData.Tag == "" {
		confData.Tag = "trunk"
	}

	if confData.Hash != "" && confData.Tag != "" {
		msg := "specifying both hash and tag is not allowed"
		glog.Errorf("invalid volume attributes for volume %s: %s", volId, msg)
		return status.Error(codes.InvalidArgument, msg)
	}

	if err = os.MkdirAll(stagingTargetPath, 0755); err != nil {
		glog.Errorf("failed to create staging path at %s for volume %s: %v", stagingTargetPath, volId, err)
		return status.Error(codes.Internal, err.Error())
	}

	if notMount, e := ns.mounter.IsLikelyNotMountPoint(stagingTargetPath); notMount && e == nil {
		// Write config for specific mount
		configPath := getConfigFilePath(volId)
		if err = confData.writeToFile(configPath); err != nil {
			glog.Errorf("failed to write volume config: %v", err)
			return status.Error(codes.Internal, err.Error())
		}

		// Each mount requires its own cache path
		err = os.MkdirAll(cachePath, 0755)
		if err == nil {
			err = os.Chown(cachePath, cvmfsUid, 0)
		}
		if err != nil {
			glog.Errorf("failed to create cache for volume %s: %v", volId, err)
			return status.Error(codes.Internal, err.Error())
		}

		// Mount the cvmfs repo directly to the stage with the custom config
		if err = ns.mounter.Mount(repository, stagingTargetPath, "cvmfs", []string{"config=" + configPath}); err != nil {
			glog.Errorf("failed to mount volume: %v", err)
			return status.Error(codes.Internal, err.Error())
		}
	} else if e != nil {
		glog.Errorf("failed to determine if %s is already mounted to %s: %v", volId, stagingTargetPath, err)
		return status.Error(codes.Internal, err.Error())
	}

	glog.V(5).Infof("cvmfs: successfuly mounted volume %s to %s", volId, stagingTargetPath)
	return nil
}

// NodeStageVolume stages the CVMFS repo mount with a custom configuration
func (ns *nodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	stagingTargetPath := req.GetStagingTargetPath()
	volId := req.GetVolumeId()
	options := req.GetVolumeContext()

	var err error
	if req.GetVolumeCapability() == nil {
		err = fmt.Errorf("volume capability missing in request")
	}

	if volId == "" {
		err = fmt.Errorf("volume ID missing in request")
	}

	if stagingTargetPath == "" {
		err = fmt.Errorf("staging target path missing in request")
	}
	if err != nil {
		glog.Errorf("failed to validate NodeStageVolumeRequest: %v", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = ns.stage(stagingTargetPath, volId, options)
	if err != nil {
		return nil, err
	}

	return &csi.NodeStageVolumeResponse{}, nil
}

// NodePublishVolume bind mounts the stage to the location where the requesting container access the repo
func (ns *nodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	var err error
	targetPath := req.GetTargetPath()
	stagingTargetPath := req.GetStagingTargetPath()
	volId := req.GetVolumeId()

	if req.GetVolumeCapability() == nil {
		err = fmt.Errorf("volume capability missing in request")
	}

	if volId == "" {
		err = fmt.Errorf("volume ID missing in request")
	}

	if targetPath == "" {
		err = fmt.Errorf("target path missing in request")
	}

	if stagingTargetPath == "" {
		err = fmt.Errorf("staging target path missing in request")
	}

	if err != nil {
		glog.Errorf("failed to validate NodePublishVolumeRequest: %v", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err = os.MkdirAll(targetPath, 0755); err != nil {
		glog.Errorf("failed to create bind mount target path at %s for volume %s: %v", targetPath, volId, err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	if notMount, _ := ns.mounter.IsLikelyNotMountPoint(stagingTargetPath); notMount {
		glog.Errorf("Attempting to publish unstaged volume %s at %s. Staging first..", volId, stagingTargetPath)
		err = ns.stage(stagingTargetPath, volId, req.GetVolumeContext())
		if err != nil {
			return nil, err
		}
	}

	if notMount, e := ns.mounter.IsLikelyNotMountPoint(targetPath); notMount && e == nil {
		if err = ns.mounter.Mount(stagingTargetPath, targetPath, "", []string{"bind", "ro"}); err != nil {
			glog.Errorf("failed to bind mount volume %s: %v", volId, err)
			return nil, status.Error(codes.Internal, err.Error())
		}
	} else if e != nil {
		glog.Errorf("failed to determine if %s is already mounted to %s: %v", volId, stagingTargetPath, err)
		return nil, status.Error(codes.Internal, err.Error())
	} else if !notMount {
		glog.Infof("cvmfs: volume %s is already bind-mounted to %s", volId, targetPath)
		return &csi.NodePublishVolumeResponse{}, nil
	}

	glog.V(5).Infof("cvmfs: successfuly bind-mounted volume %s to %s", volId, targetPath)

	return &csi.NodePublishVolumeResponse{}, nil
}

// NodeUnpublishVolume unmounts the stage from location where the requesting container access the repo, cleaning up any created directories
func (ns *nodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	var err error
	targetPath := req.GetTargetPath()
	volId := req.GetVolumeId()

	if volId == "" {
		err = fmt.Errorf("volume ID missing in request")
	}

	if targetPath == "" {
		err = fmt.Errorf("target path missing in request")
	}

	if err != nil {
		glog.Errorf("failed to validate NodeUnpublishVolumeRequest: %v", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Unbind the volume
	if _, err := os.Stat(targetPath); os.IsExist(err) {
		if err := ns.mounter.UnmountWithForce(targetPath, unmountTimeout); err != nil {
			glog.Errorf("failed to unbind volume %s: %v", targetPath, err)
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	// Clean up
	if err := os.Remove(targetPath); err != nil {
		glog.Errorf("cvmfs: cannot delete target path %s for volume %s: %v", targetPath, volId, err)
	}

	glog.V(5).Infof("cvmfs: successfuly unbinded volume %s from %s", volId, targetPath)

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

// NodeUnstageVolume unmounts CVMFS repo mount and removes the custom configuration file
func (ns *nodeServer) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	var err error
	stagingTargetPath := req.GetStagingTargetPath()
	volId := req.GetVolumeId()
	cachePath := getVolumeCachePath(ns.cvmfsCacheRoot, volId)

	if volId == "" {
		err = fmt.Errorf("volume ID missing in request")
	}

	if stagingTargetPath == "" {
		err = fmt.Errorf("staging target path missing in request")
	}

	if err != nil {
		glog.Errorf("failed to validate NodeUnstageVolumeRequest: %v", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Unmount the volume
	if _, err := os.Stat(stagingTargetPath); os.IsExist(err) {
		if err := ns.mounter.UnmountWithForce(stagingTargetPath, unmountTimeout); err != nil {
			glog.Errorf("failed unmount volume %s: %v", volId, err)
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	// Clean up
	if err := os.Remove(cachePath); err != nil {
		glog.Errorf("cvmfs: cannot remove config for volume %s: %v", volId, err)
	}

	if err := os.RemoveAll(cachePath); err != nil {
		glog.Errorf("cvmfs: cannot delete cache for volume %s: %v", volId, err)
	}

	if err := os.Remove(stagingTargetPath); err != nil {
		glog.Errorf("cvmfs: cannot delete staging target path %s for volume %s: %v", stagingTargetPath, volId, err)
	}

	glog.V(5).Infof("cvmfs: successfuly unmounted volume %s from %s", volId, stagingTargetPath)

	return &csi.NodeUnstageVolumeResponse{}, nil
}

func (ns *nodeServer) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: []*csi.NodeServiceCapability{
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
					},
				},
			},
		},
	}, nil
}

func (ns *nodeServer) NodeExpandVolume(ctx context.Context, request *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
