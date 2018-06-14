package cvmfs

import (
	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/pborman/uuid"
)

const (
	volumePrefix    = "csi-cvmfs-"
	volumePrefixLen = len(volumePrefix)
)

type volumeIdentifier struct {
	name, uuid, id string
}

func newVolumeIdentifier(req *csi.CreateVolumeRequest) *volumeIdentifier {
	volId := volumeIdentifier{
		name: req.GetName(),
		uuid: uuid.NewUUID().String(),
	}

	volId.id = volumePrefix + volId.uuid

	return &volId
}

func uuidFromVolumeId(volId string) string {
	return volId[volumePrefixLen:]
}
