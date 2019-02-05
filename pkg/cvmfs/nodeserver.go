package cvmfs

import (
	"context"
	"os"

	"github.com/golang/glog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/drivers/pkg/csi-common"
)

type nodeServer struct {
	*csicommon.DefaultNodeServer
}

func (ns *nodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	if err := validateNodeStageVolumeRequest(req); err != nil {
		glog.Errorf("failed to validate NodeStageVolumeRequest: %v", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Configuration

	stagingTargetPath := req.GetStagingTargetPath()
	volId := volumeID(req.GetVolumeId())

	volOptions, err := newVolumeOptions(req.GetVolumeContext())
	if err != nil {
		glog.Errorf("invalid volume attributes for volume %s: %v", volId, err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err = createMountPoint(stagingTargetPath); err != nil {
		glog.Errorf("failed to create staging mount point at %s for volume %s: %v", stagingTargetPath, volId, err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	confData := cvmfsConfigData{
		VolumeId: volId,
		Tag:      volOptions.Tag,
		Hash:     volOptions.Hash,
		Proxy:    volOptions.Proxy,
	}

	if err := confData.writeToFile(); err != nil {
		glog.Errorf("failed to write cvmfs config for volume %s: %v", volId, err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := createVolumeCache(volId); err != nil {
		glog.Errorf("failed to create cache for volume %s: %v", volId, err)
	}

	// Check if the volume is already mounted

	isMnt, err := isMountPoint(stagingTargetPath)

	if err != nil {
		glog.Errorf("stat failed: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	if isMnt {
		glog.Infof("cvmfs: volume %s is already mounted to %s, skipping", volId, stagingTargetPath)
		return &csi.NodeStageVolumeResponse{}, nil
	}

	// It's not, mount now

	if err = mountCvmfs(volOptions, volId, stagingTargetPath); err != nil {
		glog.Errorf("failed to mount volume %s to %s: %v", volId, stagingTargetPath, err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	glog.Infof("cvmfs: successfuly mounted volume %s to %s", volId, stagingTargetPath)

	return &csi.NodeStageVolumeResponse{}, nil
}

func (ns *nodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	if err := validateNodePublishVolumeRequest(req); err != nil {
		glog.Errorf("failed to validate NodePublishVolumeRequest: %v", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Configuration

	targetPath := req.GetTargetPath()
	volId := volumeID(req.GetVolumeId())

	if err := createMountPoint(targetPath); err != nil {
		glog.Errorf("failed to create mount point for volume %s at %s: %v", volId, targetPath, err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Check if the volume is already mounted

	isMnt, err := isMountPoint(targetPath)

	if err != nil {
		glog.Errorf("stat failed: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	if isMnt {
		glog.Infof("cvmfs: volume %s is already bind-mounted to %s", volId, targetPath)
		return &csi.NodePublishVolumeResponse{}, nil
	}

	// It's not, bind-mount now

	if err = bindMount(req.GetStagingTargetPath(), targetPath); err != nil {
		glog.Errorf("failed to bind-mount volume %s to %s: %v", volId, targetPath, err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	glog.Infof("cvmfs: successfuly bind-mounted volume %s to %s", volId, targetPath)

	return &csi.NodePublishVolumeResponse{}, nil
}

func (ns *nodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	if err := validateNodeUnpublishVolumeRequest(req); err != nil {
		glog.Errorf("failed to validate NodeUnpublishVolumeRequest: %v", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	targetPath := req.GetTargetPath()
	volId := volumeID(req.GetVolumeId())

	// Unbind the volume

	if err := unmountVolume(targetPath); err != nil {
		glog.Errorf("failed to unbind volume %s: %v", targetPath, err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Clean up

	if err := os.Remove(targetPath); err != nil {
		glog.Errorf("cvmfs: cannot delete target path %s for volume %s: %v", targetPath, volId, err)
	}

	glog.Infof("cvmfs: successfuly unbinded volume %s from %s", volId, targetPath)

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (ns *nodeServer) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	if err := validateNodeUnstageVolumeRequest(req); err != nil {
		glog.Errorf("failed to validate NodeUnstageVolumeRequest: %v", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	stagingTargetPath := req.GetStagingTargetPath()
	volId := volumeID(req.GetVolumeId())

	// Unmount the volume

	if err := unmountVolume(stagingTargetPath); err != nil {
		glog.Errorf("failed unmount volume %s: %v", volId, err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Clean up

	if err := os.Remove(getConfigFilePath(volId)); err != nil {
		glog.Errorf("cvmfs: cannot remove config for volume %s: %v", volId, err)
	}

	if err := purgeVolumeCache(volId); err != nil {
		glog.Errorf("cvmfs: cannot delete cache for volume %s: %v", volId, err)
	}

	if err := os.Remove(stagingTargetPath); err != nil {
		glog.Errorf("cvmfs: cannot delete staging target path %s for volume %s: %v", stagingTargetPath, volId, err)
	}

	glog.Infof("cvmfs: successfuly unmounted volume %s from %s", volId, stagingTargetPath)

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
