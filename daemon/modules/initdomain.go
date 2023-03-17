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
	initDomainFailed = "keentune-target domain init failed."
)

// InitResult init domain result
type InitResult struct {
	Warn string
	Err  string
}

func initDomain(host string, req interface{}, domains *sync.Map) (string, error) {
	if len(host) == 0 {
		return "", fmt.Errorf("empty host")
	}

	ip := strings.Split(host, ":")[0]

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
		failedInfo += fmt.Sprintf("%v '%v' init failed, %v%v", ip, domain, val.Msg, multiSeparator)
	}

	failedInfo = strings.TrimSuffix(failedInfo, multiSeparator)

	if total == failureNum {
		return failedInfo, fmt.Errorf("init all domain failed")
	}

	return failedInfo, nil
}

func (gp *Group) initDomain() ([]string, []string, error) {
	var (
		snDomains   = &sync.Map{}
		wg          = sync.WaitGroup{}
		initResults = make([]InitResult, len(config.KeenTune.IPMap))
		warnResults = make([]string, len(config.KeenTune.IPMap))
		errMsg      string
	)

	for _, ip := range gp.IPs {
		wg.Add(1)
		host := fmt.Sprintf("%v:%v", ip, gp.Port)
		req := request{
			ip:      ip,
			host:    host,
			ipIndex: config.KeenTune.IPMap[ip] - 1,
			body: map[string]interface{}{
				"domain_list": gp.initDomains,
				"backup_all":  config.KeenTune.BackupAll,
			},
		}

		go doInitDomain(initResults, req, &wg, snDomains)
	}

	wg.Wait()

	var err error

	for idx := range initResults {
		warnResults[idx] = initResults[idx].Warn
		if len(initResults[idx].Err) > 0 {
			errMsg += fmt.Sprintln(initResults[idx].Err)
		}
	}

	if errMsg != "" {
		err = fmt.Errorf(errMsg)
	}

	unAvlDomains := gp.updateDomains(snDomains)

	return unAvlDomains, warnResults, err
}

func doInitDomain(results []InitResult, req request, w *sync.WaitGroup, domains *sync.Map) {
	defer w.Done()
	result, err := initDomain(req.host, req.body, domains)
	if err != nil {
		results[req.ipIndex].Err = fmt.Sprintf("%v %v", req.ip, err)
	}

	if result != "" {
		results[req.ipIndex].Warn = fmt.Sprintf("%v", result)
	}
}

