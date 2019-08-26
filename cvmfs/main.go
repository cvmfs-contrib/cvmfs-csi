package main

import (
	"flag"
	"os"
	"path"

	"github.com/golang/glog"
	"gitlab.cern.ch/cloud-infrastructure/cvmfs-csi/pkg/cvmfs"
)

func init() {
	flag.Set("logtostderr", "true")
}

var (
	endpoint   = flag.String("endpoint", "unix://tmp/csi.sock", "CSI endpoint")
	driverName = flag.String("drivername", "csi-cvmfs", "name of the driver")
	nodeId     = flag.String("nodeid", "", "node id")
)

func main() {
	flag.Parse()

	if err := os.MkdirAll(path.Join(cvmfs.PluginFolder, "controller"), 0755); err != nil {
		glog.Errorf("failed to create persistent storage for controller: %v", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(path.Join(cvmfs.PluginFolder, "node"), 0755); err != nil {
		glog.Errorf("failed to create persistent storage for node: %v", err)
		os.Exit(1)
	}

	driver := cvmfs.NewDriver(*nodeId, *endpoint)
	driver.Run()

	os.Exit(0)
}
