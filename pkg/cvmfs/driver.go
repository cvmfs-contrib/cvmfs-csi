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
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
)

const (
	PluginFolder = "/var/lib/kubelet/plugins/cvmfs.csi.cern.ch"
	version      = "1.0.1"
)

type cvmfsDriver struct {
	driver   *csicommon.CSIDriver
	endpoint string

	is *identityServer   //nolint
	ns *nodeServer       //nolint
	cs *controllerServer //nolint

	caps   []*csi.VolumeCapability_AccessMode //nolint
	cscaps []*csi.ControllerServiceCapability //nolint
}

func NewDriver(nodeID, endpoint, driverName string) *cvmfsDriver {
	glog.Infof("Driver: %v version: %v", driverName, version)

	d := &cvmfsDriver{}

	d.endpoint = endpoint

	csiDriver := csicommon.NewCSIDriver(driverName, version, nodeID)
	csiDriver.AddVolumeCapabilityAccessModes([]csi.VolumeCapability_AccessMode_Mode{csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY})
	csiDriver.AddControllerServiceCapabilities([]csi.ControllerServiceCapability_RPC_Type{csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME})

	// Create gRPC servers

	d.driver = csiDriver

	return d
}

func NewNodeServer(d *cvmfsDriver) *nodeServer {
	return &nodeServer{
		DefaultNodeServer: csicommon.NewDefaultNodeServer(d.driver),
	}
}

func NewControllerServer(d *cvmfsDriver) *controllerServer {
	return &controllerServer{
		DefaultControllerServer: csicommon.NewDefaultControllerServer(d.driver),
	}
}

func NewIdentityServer(d *cvmfsDriver) *identityServer {
	return &identityServer{
		DefaultIdentityServer: csicommon.NewDefaultIdentityServer(d.driver),
	}
}

func (d *cvmfsDriver) Run() {

	csicommon.RunControllerandNodePublishServer(d.endpoint, d.driver, NewControllerServer(d), NewNodeServer(d))
}
