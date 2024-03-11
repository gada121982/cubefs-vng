package datanode

import (
	"time"

	"github.com/cubefs/cubefs/util/log"
)

const (
	UpdateVolInfoTicket = 30 * time.Second
	BatchUpdateSize     = 1000
)

var (
	volInfoStopC = make(chan struct{}, 0)
)

func (m *DataNode) startUpdateVolInfo() {
	ticker := time.NewTicker(UpdateVolInfoTicket)
	defer ticker.Stop()

	for {
		select {
		case <-volInfoStopC:
			log.LogInfo("datanode volume info go routine stopped")
			return
		case <-ticker.C:
			m.updateVolInfo()
		}
	}
}

func (m *DataNode) updateVolInfo() {
	volNames := m.getVolNames()
	if len(volNames) == 0 {
		return
	}
	for len(volNames) != 0 {
		batch := []string{}
		if len(volNames) < BatchUpdateSize {
			batch = volNames
			volNames = []string{}
		} else {
			batch = volNames[0:BatchUpdateSize]
			volNames = volNames[BatchUpdateSize:]
		}
		go m.queryAndSetVolInfo(batch)
	}
}

func (m *DataNode) queryAndSetVolInfo(volNamesInBatch []string) {
	volsSpaceInfo, err := MasterClient.AdminAPI().ListVolsByNames(volNamesInBatch)
	if err != nil {
		log.LogErrorf("[updateVolInfo] %s", err.Error())
		return
	}
	for _, vol := range volsSpaceInfo {
		m.volInfo.Store(vol.Name, vol)
	}
}

// volumeID actually is volume name
func (m *DataNode) getVolNames() (volIds []string) {
	volumeHandled := map[string]bool{}
	m.space.RangePartitions(func(part *DataPartition) bool {
		if !volumeHandled[part.volumeID] {
			volIds = append(volIds, part.volumeID)
			volumeHandled[part.volumeID] = true
		}
		return true
	})
	return
}
