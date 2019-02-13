package cvmfs

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	"github.com/kubernetes-csi/drivers/pkg/csi-common"
)

const (
	PluginFolder = "/var/lib/kubelet/plugins/csi-cvmfsplugin"
	driverName = "cvmfsDriver"
	version = "1.0.1"
)

var (
	driver *cvmfsDriver
)

type cvmfsDriver struct {
	driver *csicommon.CSIDriver
	endpoint  string

	is *identityServer
	ns  *nodeServer
	cs  *controllerServer

	caps   []*csi.VolumeCapability_AccessMode
	cscaps []*csi.ControllerServiceCapability
}

func NewDriver(nodeID, endpoint string) *cvmfsDriver {
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
