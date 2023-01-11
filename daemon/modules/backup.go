package modules

import (
	"encoding/json"
	"fmt"
	"keentune/daemon/common/utils/http"
)

// application
const (
	myConfApp = "MySQL"
)

// avoid domain
const (
	selinuxDomain = "selinux"
)

const (
	myConfBackupFile = "/etc/my.cnf"
)

const backupAllErr = "All of the domain backup failed"

var (
	backupENVNotMetFmt = "Can't find %v, please check %v is installed"
)

var unavailableDomainReason = map[string]string{
	myConfDomain:  fmt.Sprintf(backupENVNotMetFmt, myConfBackupFile, myConfApp),
	selinuxDomain: fmt.Sprintf("unvaliable domain '%v': %v", selinuxDomain, notSupportRecommend),
}

func (tuner *Tuner) backup() error {
	err := tuner.concurrent("backup")
	if tuner.Flag == JobProfile {
		return err
	}

	if err != nil {
		return err
	}

	if tuner.backupWarning != "" {
		tuner.deleteUnAVLParams()
	}

	return nil
}

func callBackup(method, url string, request interface{}) (map[string]map[string]string, string, int) {
	var response map[string]interface{}
	resp, err := http.RemoteCall(method, url, request)
	if err != nil {
		return nil, err.Error(), FAILED
	}

	if err = json.Unmarshal(resp, &response); err != nil {
		return nil, err.Error(), FAILED
	}

	req, ok := request.(map[string]interface{})
	if !ok {
		return nil, "assert request type to map error", FAILED
	}

	var unAVLParam = make(map[string]map[string]string)

	for domain, param := range req {
		reason, match := response[domain].(string)
		if match {
			// whole domain is not available
			defReason, ok := unavailableDomainReason[domain]
			if ok {
				unAVLParam[domain] = map[string]string{
					domain: defReason,
				}
			} else {
				unAVLParam[domain] = map[string]string{
					domain: reason,
				}
			}

			continue
		}

		domainParam, _ := response[domain].(map[string]interface{})
		parameter := param.(map[string]interface{})
		for name, _ := range parameter {
			_, exists := domainParam[name]
			if !exists {
				_, notExist := unAVLParam[domain]
				if !notExist {
					unAVLParam[domain] = make(map[string]string)
				}
				unAVLParam[domain][name] = fmt.Sprintf("'%v' can not backup", name)
			}
		}
	}

	return unAVLParam, "", SUCCESS
}


