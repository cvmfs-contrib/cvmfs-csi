package cvmfs

import (
	"errors"
	"fmt"
)

type volumeOptions struct {
	Repository string
	Tag        string
	Hash       string
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

	extractOption(&opts.Tag, "tag", volOptions)
	extractOption(&opts.Hash, "hash", volOptions)

	if err = opts.validate(); err != nil {
		return nil, err
	}

	return &opts, nil
}
