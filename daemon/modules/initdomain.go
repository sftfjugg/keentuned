package modules

import (
	"encoding/json"
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/utils/http"
	"strings"
	"sync"
)

const (
	errFlag  = "ERROR"
	warnFlag = "WARNING"
)

func initDomain(host string, req interface{}) (string, error) {
	url := fmt.Sprintf("%v/init", host)
	resp, err := http.RemoteCall("POST", url, req)
	if err != nil {
		return "", err
	}

	var respDict map[string]struct {
		Success bool        `json:"suc"`
		Msg     interface{} `json:"msg"`
	}

	err = json.Unmarshal(resp, &respDict)
	if err != nil {
		return "", err
	}

	total := len(respDict)
	var failedInfo string
	var failureNum int

	for domain, val := range respDict {
		if val.Success {
			continue
		}
		failureNum++
		failedInfo += fmt.Sprintf("'%v' init failed, %v;", domain, val.Msg)
	}

	failedInfo = strings.TrimSuffix(failedInfo, ";")

	if total == failureNum {
		return failedInfo, fmt.Errorf("init all domain failed")
	}

	return failedInfo, nil
}

func (tuner *Tuner) initDomain() ([]string, error) {
	wg := sync.WaitGroup{}
	var initResults = make([]string, len(config.KeenTune.IPMap))
	for _, group := range tuner.Group {
		for _, ip := range group.IPs {
			wg.Add(1)
			req := request{
				ip:      ip,
				ipIndex: config.KeenTune.IPMap[ip] - 1,
				body: map[string]interface{}{
					"domain_list": group.Domains,
					"backup_all":  config.KeenTune.BackupAll,
				},
			}

			host := fmt.Sprintf("%v:%v", ip, group.Port)

			go doInitDomain(initResults, host, req, &wg)
		}
	}

	wg.Wait()

	var err error

	for idx := range initResults {
		if strings.Contains(initResults[idx], errFlag) && err == nil {
			initResults[idx] = strings.TrimPrefix(initResults[idx], errFlag)
			err = fmt.Errorf("init failure occurs")
		}

		initResults[idx] = strings.TrimPrefix(initResults[idx], warnFlag)
	}

	return initResults, err
}

func doInitDomain(results []string, host string, req request, w *sync.WaitGroup) {
	defer w.Done()
	result, err := initDomain(host, req.body)
	if err != nil {
		results[req.ipIndex] = fmt.Sprintf("%v %v %v", errFlag, req.ip, result)
		return
	}

	if result != "" {
		results[req.ipIndex] = fmt.Sprintf("%v %v %v", warnFlag, req.ip, result)
	}
}


