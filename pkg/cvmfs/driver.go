package cvmfs

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	"github.com/kubernetes-csi/drivers/pkg/csi-common"
)

const (
	PluginFolder = "/var/lib/kubelet/plugins/csi-cvmfsplugin"
	Version      = "1.0.1"
)

type cvmfsDriver struct {
	driver *csicommon.CSIDriver
	endpoint  string

	ids *csicommon.DefaultIdentityServer
	ns  *nodeServer
	cs  *controllerServer

	caps   []*csi.VolumeCapability_AccessMode
	cscaps []*csi.ControllerServiceCapability
}

var (
	driver *cvmfsDriver
)

const (
	driverName = "cvmfsDriver"
)

var (
	version = "1.0.0-rc2"
)

func NewDriver(nodeID, endpoint string) *cvmfsDriver {
	glog.Infof("Driver: %v version: %v", driverName, version)

	d := &cvmfsDriver{}

	d.endpoint = endpoint

	csiDriver := csicommon.NewCSIDriver(driverName, version, nodeID)
	csiDriver.AddVolumeCapabilityAccessModes([]csi.VolumeCapability_AccessMode_Mode{csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER})

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

func (d *cvmfsDriver) Run() {
	csicommon.RunNodePublishServer(d.endpoint, d.driver, NewNodeServer(d))
}
