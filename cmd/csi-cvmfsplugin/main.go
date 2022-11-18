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
	"fmt"
	"os"
	"strings"

	"github.com/cernops/cvmfs-csi/internal/cvmfs/driver"
	"github.com/cernops/cvmfs-csi/internal/log"
	cvmfsversion "github.com/cernops/cvmfs-csi/internal/version"

	"k8s.io/klog/v2"
)

type rolesFlag []driver.ServiceRole

func (rf rolesFlag) String() string {
	return fmt.Sprintf("%v", []driver.ServiceRole(rf))
}

var (
	knownServiceRoles = map[driver.ServiceRole]struct{}{
		driver.IdentityServiceRole:   {},
		driver.NodeServiceRole:       {},
		driver.ControllerServiceRole: {},
	}
)

func (rf *rolesFlag) Set(newRoleFlag string) error {
	for _, part := range strings.Split(newRoleFlag, ",") {
		if _, ok := knownServiceRoles[driver.ServiceRole(part)]; !ok {
			return fmt.Errorf("unknown role %s", part)
		}

		*rf = append(*rf, driver.ServiceRole(part))
	}

	return nil
}

var (
	endpoint   = flag.String("endpoint", fmt.Sprintf("unix:///var/lib/kubelet/plugins/%s/csi.sock", driver.DefaultName), "CSI endpoint.")
	driverName = flag.String("drivername", driver.DefaultName, "Name of the driver.") //nolint
	nodeId     = flag.String("nodeid", "", "Node id.")
	version    = flag.Bool("version", false, "Print driver version and exit.")
	roles      rolesFlag

	hasAlienCache        = flag.Bool("has-alien-cache", false, "CVMFS client is using alien cache volume")
	startAutomountDaemon = flag.Bool("start-automount-daemon", true, "start automount daemon when initializing CVMFS CSI driver")

	automountDaemonStartupTimeoutSeconds   = flag.Int("automount-startup-timeout", 5, "number of seconds to wait for automount daemon to start up before exiting")
	automountDaemonUnmountAfterIdleSeconds = flag.Int("automount-unmount-timeout", -1, "number of seconds of idle time after which an autofs-managed CVMFS mount will be unmounted. '0' means never unmount, '-1' leaves automount default option.")
)

func printVersion() {
	fmt.Printf(
		"CVMFS CSI plugin version %s (commit: %s; build time: %s; metadata: %s)\n",
		cvmfsversion.Version(), cvmfsversion.Commit(), cvmfsversion.BuildTime(), cvmfsversion.Metadata(),
	)
}

func main() {
	// Handle flags and initialize logging.

	flag.Var(&roles, "role", "Enable driver service role (comma-separated list or repeated --role flags). Allowed values are: 'identity', 'node', 'controller'.")

	klog.InitFlags(nil)
	if err := flag.Set("logtostderr", "true"); err != nil {
		klog.Exitf("failed to set logtostderr flag: %v", err)
	}
	flag.Parse()

	if *version {
		printVersion()
		os.Exit(0)
	}

	log.Infof("Running CVMFS CSI plugin with %v", os.Args)

	// Initialize and run the driver.

	driverRoles := make(map[driver.ServiceRole]bool, len(roles))
	for _, role := range roles {
		driverRoles[role] = true
	}

	driver, err := driver.New(&driver.Opts{
		DriverName:  *driverName,
		CSIEndpoint: *endpoint,
		NodeID:      *nodeId,
		Roles:       driverRoles,

		StartAutomountDaemon: *startAutomountDaemon,
		HasAlienCache:        *hasAlienCache,

		AutomountDaemonStartupTimeoutSeconds:   *automountDaemonStartupTimeoutSeconds,
		AutomountDaemonUnmountAfterIdleSeconds: *automountDaemonUnmountAfterIdleSeconds,
	})

	if err != nil {
		log.Fatalf("Failed to initialize the driver: %v", err)
	}

	err = driver.Run()
	if err != nil {
		log.Fatalf("Failed to run the driver: %v", err)
	}

	os.Exit(0)
}
