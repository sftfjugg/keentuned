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

func initDomain(host string, req interface{}, domains *sync.Map) (string, error) {
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

		domains.Store(domain, true)
		failureNum++
		failedInfo += fmt.Sprintf("'%v' init failed, %v%v", domain, val.Msg, multiSeparator)
	}

	failedInfo = strings.TrimSuffix(failedInfo, multiSeparator)

	if total == failureNum {
		return failedInfo, fmt.Errorf("init all domain failed")
	}

	return failedInfo, nil
}

func (gp *Group) initDomain() ([]string, error) {
	var (
		snDomains   = &sync.Map{}
		wg          = sync.WaitGroup{}
		initResults = make([]string, len(config.KeenTune.IPMap))
	)

	for _, ip := range gp.IPs {
		wg.Add(1)
		host := fmt.Sprintf("%v:%v", ip, gp.Port)
		req := request{
			ip:      ip,
			host:    host,
			ipIndex: config.KeenTune.IPMap[ip] - 1,
			body: map[string]interface{}{
				"domain_list": gp.Domains,
				"backup_all":  config.KeenTune.BackupAll,
			},
		}

		go doInitDomain(initResults, req, &wg, snDomains)
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

	gp.updateDomains(snDomains)

	return initResults, err
}

func doInitDomain(results []string, req request, w *sync.WaitGroup, domains *sync.Map) {
	defer w.Done()
	result, err := initDomain(req.host, req.body, domains)
	if err != nil {
		results[req.ipIndex] = fmt.Sprintf("%v %v %v", errFlag, req.ip, result)
		return
	}

	if result != "" {
		results[req.ipIndex] = fmt.Sprintf("%v %v %v", warnFlag, req.ip, result)
	}
}


