package cvmfs

import (
	"context"
	"fmt"
	"os"

	"github.com/golang/glog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/kubernetes-csi/drivers/pkg/csi-common"
)

type controllerServer struct {
	*csicommon.DefaultControllerServer
}

func (cs *controllerServer) validateCreateVolumeRequest(req *csi.CreateVolumeRequest) error {
	if err := cs.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		return fmt.Errorf("Invalid CreateVolumeRequest: %v", err)
	}

	if req.GetName() == "" {
		return fmt.Errorf("Volume Name cannot be empty")
	}

	if req.GetVolumeCapabilities() == nil {
		return fmt.Errorf("Volume Capabilities cannot be empty")
	}

	return nil
}

func (cs *controllerServer) validateDeleteVolumeRequest(req *csi.DeleteVolumeRequest) error {
	if err := cs.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		return fmt.Errorf("Invalid DeleteVolumeRequest: %v", err)
	}

	return nil
}

func (cs *controllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	if err := cs.validateCreateVolumeRequest(req); err != nil {
		glog.Errorf("CreateVolumeRequest validation failed: %v", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	volOptions, err := newVolumeOptions(req.GetParameters())
	if err != nil {
		glog.Errorf("failed to validate volume options: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	volId := newVolumeIdentifier(volOptions, req)

	confData := cvmfsConfigData{
		VolUuid: volId.uuid,
		Tag:     volOptions.Tag,
		Hash:    volOptions.Hash,
	}

	if err := confData.writeToFile(); err != nil {
		glog.Errorf("failed to write cvmfs config for volume %s: %v", volId.id, err)
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	if err := createVolumeCache(volId.uuid); err != nil {
		glog.Errorf("failed to create cache for volume %s: %v", volId.id, err)
	}

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			Id:         volId.id,
			Attributes: req.GetParameters(),
		},
	}, nil
}

func (cs *controllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	if err := cs.validateDeleteVolumeRequest(req); err != nil {
		glog.Errorf("DeleteVolumeRequest validation failed: %v", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	volId := req.GetVolumeId()
	volUuid := uuidFromVolumeId(volId)

	if err := os.Remove(getConfigFilePath(volUuid)); err != nil {
		glog.Warningf("cannot remove config for volume %s: %v", volId, err)
	}

	if err := purgeVolumeCache(volUuid); err != nil {
		glog.Errorf("failed to delete cache for volume %s: %v", volId, err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &csi.DeleteVolumeResponse{}, nil
}

func (cs *controllerServer) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	for _, c := range req.GetVolumeCapabilities() {
		if c.AccessMode.GetMode() != csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY {
			return &csi.ValidateVolumeCapabilitiesResponse{Supported: false}, nil
		}
	}

	return &csi.ValidateVolumeCapabilitiesResponse{Supported: true}, nil
}
