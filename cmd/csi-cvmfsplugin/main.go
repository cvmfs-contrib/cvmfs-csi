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

package main

import (
	"flag"
	"os"
	"path"

	"github.com/cernops/cvmfs-csi/pkg/cvmfs"
	"github.com/golang/glog"
)

func init() {
	_ = flag.Set("logtostderr", "true")
}

var (
	endpoint   = flag.String("endpoint", "unix://tmp/csi.sock", "CSI endpoint")
	driverName = flag.String("drivername", "csi-cvmfs", "name of the driver") //nolint
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
