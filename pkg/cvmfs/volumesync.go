package cvmfs

import "sync"

type volumeSync struct {
	vols map[string]byte
	m    sync.Mutex
}

func newVolumeSync() *volumeSync {
	return &volumeSync{vols: make(map[string]byte)}
}

// If volId is unmarked, mark it and return true. Otherwise return false
func (vs *volumeSync) markOrFail(volId string) bool {
	vs.m.Lock()
	defer vs.m.Unlock()

	if _, found := vs.vols[volId]; found {
		return false
	}

	vs.vols[volId] = 0

	return true
}

func (vs *volumeSync) unmark(volId string) {
	vs.m.Lock()
	delete(vs.vols, volId)
	vs.m.Unlock()
}
