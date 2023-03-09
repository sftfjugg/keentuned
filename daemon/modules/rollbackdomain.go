package modules

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"keentune/daemon/common/config"
	"sync"
)

const dumpDomainFile = "rollbackDomain.json"
const (
	targetPrefix = "target-group"
)

func dumpRollbackDomainsFile() string {
	return config.GetTempPath(dumpDomainFile)
}

// initRB init rollback
func (tuner *Tuner) initRB() {
	var target = new(Group)
	for _, group := range config.KeenTune.Group {
		target.IPs = group.IPs
		target.Port = group.Port
		target.GroupName = group.GroupName
		target.GroupNo = group.GroupNo
		tuner.Group = append(tuner.Group, *target)
	}

	tuner.loadRBDomains()
}

// dumpForRBDomains dump for rollback domains based on tuning or setting job
func (tuner *Tuner) dumpForRBDomains() {
	var mutex = &sync.RWMutex{}
	mutex.Lock()
	var domainDict = make(map[string][]string)
	for _, target := range tuner.Group {
		key := fmt.Sprintf("%v%v", targetPrefix, target.GroupNo)
		domainDict[key] = target.initDomains
	}

	info, _ := json.Marshal(domainDict)
	ioutil.WriteFile(dumpRollbackDomainsFile(), info, 0644)
	mutex.Unlock()
}

func (tuner *Tuner) StoreRBDomain(i int, domains *sync.Map) {
	gp := tuner.Group[0]
	key := fmt.Sprintf("%v%v", targetPrefix, gp.GroupNo)
	domains.Store(key, gp.initDomains)
}

// loadRBDomains load rollback domains from dumped file or new job domains
func (tuner *Tuner) loadRBDomains() {
	rollbackFile := config.GetTempPath(dumpDomainFile)
	info, _ := ioutil.ReadFile(rollbackFile)

	var domainDict = make(map[string][]string)

	json.Unmarshal(info, &domainDict)

	for idx := range tuner.Group {
		key := fmt.Sprintf("%v%v", targetPrefix, tuner.Group[idx].GroupNo)
		domains, find := domainDict[key]
		if !find {
			tuner.Group[idx].rollbackDomains = tuner.Group[idx].initDomains
			continue
		}

		tuner.Group[idx].rollbackDomains = domains
	}
}

// updateRBDomains update rollback domains used for unexpected job failure rollback
func (gp *Group) updateRBDomains() {
	gp.rollbackDomains = gp.initDomains
}

// cleanRBDomains clean rollback domains
func cleanRBDomains() {
	var mutex = &sync.RWMutex{}
	mutex.Lock()
	ioutil.WriteFile(dumpRollbackDomainsFile(), []byte{}, 0644)
	mutex.Unlock()
}

// DumpDefaultRBDomains ...
func DumpDefaultRBDomains(domains *sync.Map) {
	var mutex = &sync.RWMutex{}
	mutex.Lock()
	defer mutex.Unlock()
	var dumpDomain = make(map[string]interface{})
	domains.Range(func(key, value interface{}) bool {
		groupName, ok := key.(string)
		if ok {
			dumpDomain[groupName] = value
		}
		return true
	})

	if len(dumpDomain) == 0 {
		return
	}

	info, err := json.Marshal(dumpDomain)
	if err != nil {
		return
	}

	ioutil.WriteFile(dumpRollbackDomainsFile(), info, 0644)
}


