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

	"github.com/golang/glog"
	"github.com/pborman/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/container-storage-interface/spec/lib/go/csi"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
)

type controllerServer struct {
	*csicommon.DefaultControllerServer
	Name string
}

func NewControllerServer(d *cvmfsDriver) *controllerServer {
	return &controllerServer{
		DefaultControllerServer: csicommon.NewDefaultControllerServer(d.driver),
		Name:                    d.Name,
	}
}

func (cs *controllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	var err error
	if e := cs.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); e != nil {
		err = fmt.Errorf("invalid CreateVolumeRequest: %v", e)
	}

	name := req.GetName()

	if name == "" {
		err = fmt.Errorf("volume name cannot be empty")
	}

	if err != nil {
		glog.Errorf("failed to validate CreateVolumeRequest: %v", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	volId := "csi-cvmfs-" + name + "-" + uuid.NewUUID().String()

	glog.Infof("cvmfs: Assigned new volume ID (%s) to volume %s", volId, name)

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      volId,
			VolumeContext: req.GetParameters(),
			CapacityBytes: req.GetCapacityRange().GetRequiredBytes(),
		},
	}, nil
}

func (cs *controllerServer) ControllerGetVolume(ctx context.Context, request *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	// TODO this is largely stubbed and needs more refinement
	return &csi.ControllerGetVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      request.VolumeId,
			CapacityBytes: 0,
		},
	}, nil
}

func (cs *controllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	if err := cs.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		glog.Errorf("failed to validate DeleteVolumeRequest: %v", err)
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("invalid DeleteVolumeRequest: %v", err).Error())
	}
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

	return &csi.ValidateVolumeCapabilitiesResponse{
		Confirmed: &csi.ValidateVolumeCapabilitiesResponse_Confirmed{
			VolumeCapabilities: []*csi.VolumeCapability{
				{
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY,
					},
				},
			},
		},
	}, nil
}

func (cs *controllerServer) ControllerExpandVolume(ctx context.Context, request *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
