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

package node

import (
	"os"
)

func ensureCVMFSClientConfigFile(contents, filepath string) error {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0444)
	if err != nil {
		if os.IsExist(err) {
			// File already exists, exit early.
			return nil
		}

		return err
	}

	_, err = f.WriteString(contents)

	errClose := f.Close()
	if errClose != nil && err == nil {
		return errClose
	}

	return err
}
