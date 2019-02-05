package cvmfs

import (
	"context"

	"github.com/golang/glog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/drivers/pkg/csi-common"
)

type controllerServer struct {
	*csicommon.DefaultControllerServer
}

func (cs *controllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	if err := cs.validateCreateVolumeRequest(req); err != nil {
		glog.Errorf("failed to validate CreateVolumeRequest: %v", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	volId := newVolumeID()

	glog.Infof("cvmfs: successfuly created volume %s", volId)

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:            string(volId),
			VolumeContext:    req.GetParameters(),
			CapacityBytes: req.GetCapacityRange().GetRequiredBytes(),
		},
	}, nil
}

func (cs *controllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	if err := cs.validateDeleteVolumeRequest(req); err != nil {
		glog.Errorf("failed to validate DeleteVolumeRequest: %v", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	glog.Infof("cvmfs: successfuly deleted volume %s", req.GetVolumeId())

	return &csi.DeleteVolumeResponse{}, nil
}




func (cs *controllerServer) ValidateVolumeCapabilities(
	ctx context.Context,
	req *csi.ValidateVolumeCapabilitiesRequest,
) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	for _, c := range req.GetVolumeCapabilities() {
		if c.AccessMode.GetMode() != csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY || c.GetBlock() != nil {
			return nil, status.Error(codes.Unimplemented, "")
		}
	}

	supportedAccessMode := &csi.VolumeCapability_AccessMode{
		Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY,
	}

	return &csi.ValidateVolumeCapabilitiesResponse{
		Confirmed: &csi.ValidateVolumeCapabilitiesResponse_Confirmed{
			VolumeCapabilities: []*csi.VolumeCapability{
				{
					AccessMode: supportedAccessMode,
				},
			},
		},
	}, nil
}
