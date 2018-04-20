package cvmfs

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/golang/glog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/kubernetes-csi/drivers/pkg/csi-common"
)

type nodeServer struct {
	*csicommon.DefaultNodeServer
}

var (
	pendingVols = newVolumeSync()
)

func validateNodePublishVolumeRequest(req *csi.NodePublishVolumeRequest) error {
	if req.GetVolumeCapability() == nil {
		return errors.New("Volume capability missing in request")
	}

	if req.GetVolumeId() == "" {
		return errors.New("Volume ID missing in request")
	}

	if req.GetTargetPath() == "" {
		return errors.New("Target path missing in request")
	}

	return nil
}

func validateNodeUnpublishVolumeRequest(req *csi.NodeUnpublishVolumeRequest) error {
	if req.GetVolumeId() == "" {
		return errors.New("Volume ID missing in request")
	}

	if req.GetTargetPath() == "" {
		return errors.New("Target path missing in request")
	}

	return nil
}

func (ns *nodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	if err := validateNodePublishVolumeRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Configuration

	targetPath := req.GetTargetPath()
	volId := req.GetVolumeId()
	volUuid := uuidFromVolumeId(volId)

	if !pendingVols.markOrFail(volId) {
		msg := fmt.Sprintf("cvmfs: NodePublishVolume for volume %s is pending", volId)
		glog.Infoln(msg)
		return nil, status.Error(codes.Aborted, msg)
	}

	defer pendingVols.unmark(volId)

	volOptions, err := newVolumeOptions(req.GetVolumeAttributes())
	if err != nil {
		glog.Errorf("error reading volume options: %v", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err = createMountPoint(targetPath); err != nil {
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
		glog.V(4).Infof("cephfs: volume %s is already mounted to %s", volId, targetPath)
		return &csi.NodePublishVolumeResponse{}, nil
	}

	// It's not, mount now

	if err = mountVolume(targetPath, volOptions, volUuid); err != nil {
		glog.Errorf("failed to mount volume %s: %v", volId, err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	glog.V(4).Infof("cvmfs: volume %s successfuly mounted to %s", volId, targetPath)

	return &csi.NodePublishVolumeResponse{}, nil
}

func (ns *nodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	if err := validateNodeUnpublishVolumeRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	targetPath := req.GetTargetPath()
	volId := req.GetVolumeId()
	volUuid := uuidFromVolumeId(volId)

	if err := unmountVolume(targetPath, volUuid); err != nil {
		glog.Errorf("failed to unmount volume %s: %v")
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := os.Remove(getVolumeRootPath(volUuid)); err != nil {
		glog.Error("failed to remove root for volume %s: %v", volId, err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	glog.V(4).Infof("cvmfs: volume %s successfuly unmounted from %s", req.GetVolumeId(), targetPath)

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (ns *nodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (ns *nodeServer) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
