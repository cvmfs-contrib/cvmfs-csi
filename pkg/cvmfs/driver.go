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
	version      = "1.2.0"
)

type cvmfsDriver struct {
	driver         *csicommon.CSIDriver
	endpoint       string
	cvmfsCacheRoot string
	Name           string

	ns *nodeServer
	cs *controllerServer

	caps   []*csi.VolumeCapability_AccessMode
	cscaps []*csi.ControllerServiceCapability
}

func NewDriver(nodeID, endpoint, driverName, cvmfsCacheRoot string) *cvmfsDriver {
	glog.Infof("Driver: %v version: %v", driverName, version)

	csiDriver := csicommon.NewCSIDriver(driverName, version, nodeID)
	csiDriver.AddVolumeCapabilityAccessModes([]csi.VolumeCapability_AccessMode_Mode{csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY})
	csiDriver.AddControllerServiceCapabilities([]csi.ControllerServiceCapability_RPC_Type{csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME})

	// Create gRPC servers
	return &cvmfsDriver{
		Name:           driverName,
		endpoint:       endpoint,
		cvmfsCacheRoot: cvmfsCacheRoot,
		driver:         csiDriver,
	}
}

func (d *cvmfsDriver) Run() {
	csicommon.RunControllerandNodePublishServer(d.endpoint, d.driver, NewControllerServer(d), NewNodeServer(d))
}
