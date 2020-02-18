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

	"github.com/golang/glog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/container-storage-interface/spec/lib/go/csi"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
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
			VolumeId:      string(volId),
			VolumeContext: req.GetParameters(),
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
