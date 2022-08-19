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

package exec

import (
	"fmt"
	"os/exec"
	"sync/atomic"

	"github.com/cernops/cvmfs-csi/internal/log"
)

// This file only provides wrappers around "os/exec" and logs the executed commands.

var (
	// Counter value used for pairing pre- and post-exec log messages.
	execCounter uint64
)

func fmtLogMsg(execID uint64, msg string) string {
	return fmt.Sprintf("Exec-ID %d: %s", execID, msg)
}

func Run(cmd *exec.Cmd) error {
	c := atomic.AddUint64(&execCounter, 1)
	log.InfofDepth(2, fmtLogMsg(c, "Running command env=%v prog=%s cmd=%v"), cmd.Env, cmd.Path, cmd.Args)

	err := cmd.Run()
	log.InfofDepth(2, fmtLogMsg(c, "Process exited: %s"), cmd.ProcessState)

	if err != nil {
		log.ErrorfDepth(2, fmtLogMsg(c, "Error: %v"), err)
	}

	return err
}

func Output(cmd *exec.Cmd) ([]byte, error) {
	c := atomic.AddUint64(&execCounter, 1)
	log.InfofDepth(2, fmtLogMsg(c, "Running command env=%v prog=%s args=%v"), cmd.Env, cmd.Path, cmd.Args)

	out, err := cmd.Output()
	log.InfofDepth(2, fmtLogMsg(c, "Process exited: %s"), cmd.ProcessState)

	if err != nil {
		log.ErrorfDepth(2, fmtLogMsg(c, "Error: %v"), err)
	}

	return out, err
}

func CombinedOutput(cmd *exec.Cmd) ([]byte, error) {
	c := atomic.AddUint64(&execCounter, 1)
	log.InfofDepth(2, fmtLogMsg(c, "Running command env=%v prog=%s args=%v"), cmd.Env, cmd.Path, cmd.Args)

	out, err := cmd.CombinedOutput()
	log.InfofDepth(2, fmtLogMsg(c, "Process exited: %s"), cmd.ProcessState)

	if err != nil {
		log.ErrorfDepth(2, fmtLogMsg(c, "Error: %v; Output: %s"), err, out)
	}

	return out, err
}
