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

package driver

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	goexec "os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/cernops/cvmfs-csi/internal/cvmfs/controller"
	"github.com/cernops/cvmfs-csi/internal/cvmfs/identity"
	"github.com/cernops/cvmfs-csi/internal/cvmfs/node"
	"github.com/cernops/cvmfs-csi/internal/exec"
	"github.com/cernops/cvmfs-csi/internal/log"
	"github.com/cernops/cvmfs-csi/internal/version"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"k8s.io/apimachinery/pkg/util/validation"
)

type (
	// Service role name.
	ServiceRole string

	// Opts holds init-time driver configuration.
	Opts struct {
		// DriverName is the name of this CSI driver that's then
		// advertised via NodeGetPluginInfo RPC.
		DriverName string

		// CSIEndpoint is URL path to the UNIX socket where the driver
		// will serve requests.
		CSIEndpoint string

		// NodeID is unique identifier of the node on which this
		// CVMFS CSI node plugin pod is running.
		NodeID string

		// HasAlienCache determines whether we're using alien cache.
		// If so, we need to prepare the alien cache volume first (e.g.
		// make sure it has correct permissions).
		HasAlienCache bool

		// StartAutomountDaemon determines whether CVMFS CSI nodeplugin Pod
		// should run automount daemon. This is required for automounts to work.
		// If however worker nodes are already running automount daemon (e.g.
		// as a systemd service), we can use that one (since we should be running
		// in host's PID namespace anyway).
		StartAutomountDaemon bool

		// Role under which will the driver operate.
		Roles map[ServiceRole]bool

		// How many seconds to wait for automount daemon to start up,
		// and /cvmfs mountpoint to become available.
		AutomountDaemonStartupTimeoutSeconds int

		// After how many seconds of inactivity of an autofs-managed mount
		// should the volume be unmounted, i.e. automount --timeout <value>.
		// 0: Never unmount.
		// -1: Leave automount's default timeout.
		AutomountDaemonUnmountAfterIdleSeconds int
	}

	// Driver holds CVMFS-CSI driver runtime state.
	Driver struct {
		*Opts
	}
)

const (
	IdentityServiceRole   = "identity"   // Enable identity service role.
	NodeServiceRole       = "node"       // Enable node service role.
	ControllerServiceRole = "controller" // Enable controller service role.
)

const (
	// CVMFS-CSI driver name.
	DefaultName = "cvmfs.csi.cern.ch"

	// Maximum driver name length as per CSI spec.
	maxDriverNameLength = 63
)

func (o *Opts) validate() error {
	required := func(name, value string) error {
		if value == "" {
			return fmt.Errorf("%s is a required parameter", name)
		}

		return nil
	}

	if err := required("drivername", o.DriverName); err != nil {
		return err
	}

	if len(o.DriverName) > maxDriverNameLength {
		return fmt.Errorf("driver name too long: is %d characters, maximum is %d",
			len(o.DriverName), maxDriverNameLength)
	}

	// As per CSI spec, driver name must follow DNS format.
	if errMsgs := validation.IsDNS1123Subdomain(strings.ToLower(o.DriverName)); len(errMsgs) > 0 {
		return fmt.Errorf("driver name is invalid: %v", errMsgs)
	}

	if err := required("endpoint", o.CSIEndpoint); err != nil {
		return err
	}

	if err := required("nodeid", o.NodeID); err != nil {
		return err
	}

	return nil
}

// New creates a new instance of Driver.
func New(opts *Opts) (*Driver, error) {
	if err := opts.validate(); err != nil {
		return nil, fmt.Errorf("invalid driver options: %v", err)
	}

	return &Driver{
		Opts: opts,
	}, nil
}

// Run starts CSI services and blocks.
func (d *Driver) Run() error {
	log.Infof("Driver: %s", d.DriverName)

	log.Infof(
		"Version: %s (commit: %s; build time: %s; metadata: %s)",
		version.Version(),
		version.Commit(),
		version.BuildTime(),
		version.Metadata(),
	)

	s, err := newGRPCServer(d.CSIEndpoint)
	if err != nil {
		return fmt.Errorf("failed to create GRPC server: %v", err)
	}

	if d.Opts.Roles[IdentityServiceRole] {
		log.Debugf("Registering Identity server")
		csi.RegisterIdentityServer(
			s.server,
			identity.New(
				d.DriverName,
				d.Opts.Roles[ControllerServiceRole],
			),
		)
	}

	if d.Opts.Roles[NodeServiceRole] {
		ver, err := cvmfsVersion()
		if err != nil {
			return err
		}
		log.Infof("%s", ver)

		err = setupCvmfs(d.Opts)
		if err != nil {
			return err
		}

		ns := node.New(d.NodeID)

		caps, err := ns.NodeGetCapabilities(
			context.TODO(),
			&csi.NodeGetCapabilitiesRequest{},
		)
		if err != nil {
			return fmt.Errorf("failed to get Node server capabilities: %v", err)
		}

		log.Debugf("Registering Node server with capabilities %+v", caps.GetCapabilities())
		csi.RegisterNodeServer(s.server, node.New(d.NodeID))
	}

	if d.Opts.Roles[ControllerServiceRole] {
		cs := controller.New()

		caps, err := cs.ControllerGetCapabilities(
			context.TODO(),
			&csi.ControllerGetCapabilitiesRequest{},
		)
		if err != nil {
			return fmt.Errorf("failed to get Controller server capabilities: %v", err)
		}

		log.Debugf("Registering Controller server with capabilities %+v", caps.GetCapabilities())
		csi.RegisterControllerServer(s.server, controller.New())
	}

	return s.serve()
}

func cvmfsVersion() (string, error) {
	out, err := exec.CombinedOutput(goexec.Command("cvmfs2", "--version"))
	if err != nil {
		return "", fmt.Errorf("failed to get CVMFS version: %v", err)
	}

	return string(bytes.TrimSpace(out)), nil
}

func setupCvmfs(o *Opts) error {
	if o.HasAlienCache {
		// Make sure the volume is writable by CVMFS processes.
		if err := os.Chmod("/cvmfs-aliencache", 0777); err != nil {
			return fmt.Errorf("failed to chmod /cvmfs-aliencache: %v", err)
		}
	}

	// Set up configuration required for autofs with CVMFS to work properly.
	if _, err := exec.CombinedOutput(goexec.Command("cvmfs_config", "setup")); err != nil {
		return fmt.Errorf("failed to setup CVMFS config: %v", err)
	}

	if o.StartAutomountDaemon {
		// Start the automount daemon.
		if err := automountRunner(o); err != nil {
			return fmt.Errorf("failed to start automount daemon: %v", err)
		}
	}

	// The autofs root must be made to be shared, so that mount/unmount events
	// can be propagated to bindmounts we'll be making for consumer Pods.
	if _, err := exec.CombinedOutput(goexec.Command("mount", "--make-shared", "/cvmfs")); err != nil {
		return fmt.Errorf("failed to share /cvmfs: %v", err)
	}

	return nil
}

func automountRunner(o *Opts) error {
	var (
		confBuffer bytes.Buffer
		args       = []string{
			"--foreground",
		}
	)

	// Build automount config file and cmd args.

	confBuffer.WriteString("USE_MISC_DEVICE=\"yes\"\n")

	if log.LevelEnabled(log.LevelDebug) {
		// Enable automount verbose logging.
		args = append(args, "--verbose")
	}

	if log.LevelEnabled(log.LevelTrace) {
		// automount passes -O options to the underlying fs mounts.
		// Enable CVMFS debug logging.
		args = append(args, "-O", "debug")
	}

	if o.AutomountDaemonUnmountAfterIdleSeconds != -1 {
		// automount in the image ignores --timeout flag,
		// and only reads configuration from /etc/sysconfig/autofs.
		confBuffer.WriteString(fmt.Sprintf("TIMEOUT=%d\n", o.AutomountDaemonUnmountAfterIdleSeconds))
	}

	if err := os.WriteFile("/etc/sysconfig/autofs", confBuffer.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write autofs configuration to /etc/sysconfig/autofs: %v", err)
	}

	cmd := goexec.Command("automount", args...)

	// Set-up piping output for stdout and stderr to driver's logging.

	outp, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = cmd.Stdout

	// Run automount.

	log.Infof("Starting automount daemon prog=%s args=%v", cmd.Path, cmd.Args)
	if err := cmd.Start(); err != nil {
		return err
	}
	log.Infof("Started automount daemon PID %d", cmd.Process.Pid)

	scanner := bufio.NewScanner(outp)
	scanner.Split(bufio.ScanLines)

	go func() {
		// Log and wait.

		for scanner.Scan() {
			log.Infof("automount[%d]: %s", cmd.Process.Pid, scanner.Text())
		}

		cmd.Wait()

		if cmd.ProcessState.ExitCode() != 0 {
			panic(fmt.Sprintf("automount[%d] has exited unexpectedly: %s", cmd.Process.Pid, cmd.ProcessState))
		}

		log.Infof("automount[%d] has exited: %s", cmd.Process.Pid, cmd.ProcessState)
	}()

	// Wait until autofs is mounted under /cvmfs.

	retryFor := func(attempts int, f func() (bool, error)) error {
		trial := 0
		for trial < attempts {
			res, err := f()

			if err != nil {
				return err
			}

			if res {
				return nil
			}

			trial++
			time.Sleep(1 * time.Second)
		}

		return fmt.Errorf("timed-out while waiting for autofs to be mounted")
	}

	const autofsStatfsType = 0x187

	err = retryFor(o.AutomountDaemonStartupTimeoutSeconds, func() (bool, error) {
		statfs := syscall.Statfs_t{}
		err = syscall.Statfs("/cvmfs", &statfs)
		if err != nil {
			return false, err
		}

		return statfs.Type == autofsStatfsType, nil
	})

	return nil
}
