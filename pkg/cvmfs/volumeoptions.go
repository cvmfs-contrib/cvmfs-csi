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
	"errors"
	"fmt"
)

type volumeOptions struct {
	Repository string
	Tag        string
	Hash       string
	Proxy      string
}

func validateNonEmptyField(field, fieldName string) error {
	if field == "" {
		return fmt.Errorf("parameter '%s' cannot be empty", fieldName)
	}

	return nil
}

func (o *volumeOptions) validate() error {
	if err := validateNonEmptyField(o.Repository, "repository"); err != nil {
		return err
	}

	if o.Hash == "" && o.Tag == "" {
		o.Tag = "trunk"
	}

	if o.Hash != "" && o.Tag != "" {
		return errors.New("specifying both hash and tag is not allowed")
	}

	return nil
}

func extractOption(dest *string, optionLabel string, options map[string]string) error {
	if opt, ok := options[optionLabel]; !ok {
		return errors.New("missing required field " + optionLabel)
	} else {
		*dest = opt
		return nil
	}
}

func newVolumeOptions(volOptions map[string]string) (*volumeOptions, error) {
	var (
		opts volumeOptions
		err  error
	)

	if err = extractOption(&opts.Repository, "repository", volOptions); err != nil {
		return nil, err
	}
	if err = extractOption(&opts.Tag, "tag", volOptions); err != nil {
		return nil, err
	}
	if err = extractOption(&opts.Hash, "hash", volOptions); err != nil {
		return nil, err
	}
	if err = extractOption(&opts.Proxy, "proxy", volOptions); err != nil {
		return nil, err
	}
	if err = opts.validate(); err != nil {
		return nil, err
	}

	return &opts, nil
}
