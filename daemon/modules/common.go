package modules

import (
	"encoding/json"
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/utils/http"
	"os"
	"strings"
	"sync"
)

// StopSig ...
var StopSig chan os.Signal

// Status code
const (
	// SUCCESS status code
	SUCCESS = iota + 1
	WARNING
	FAILED
)

const multiRecordSeparator = "*#++#*"

// backup doesn't exist
const (
	// BackupNotFound error information
	BackupNotFound = "Can not find backup file"
	FileNotExist   = "do not exists"
	NoNeedRollback = "don't need rollback"
	NoBackupFile   = "No such file"
)

func (tuner *Tuner) isInterrupted() bool {
	select {
	case <-StopSig:
		tuner.rollback()
		return true
	default:
		return false
	}
}

// Rollback ...
func Rollback(logName string, callType string) (string, error) {
	tune := new(Tuner)
	tune.logName = logName
	tune.initParams()
	var err error
	if callType == "original" {
		err = tune.original()
	} else {
		err = tune.rollback()
	}

	if err != nil {
		return tune.rollbackFailure, err
	}

	return tune.rollbackDetail, nil
}

func (gp *Group) concurrentSuccess(uri string, request interface{}) (string, bool) {
	wg := sync.WaitGroup{}
	var sucCount = new(int)
	var detailInfo = new(string)
	var failedInfo = new(string)
	unAVLParams := make([]map[string]map[string]string, len(gp.IPs))

	for index, ip := range gp.IPs {
		wg.Add(1)
		id := config.KeenTune.IPMap[ip]
		config.IsInnerApplyRequests[id] = false
		go func(index, groupID int, ip string, wg *sync.WaitGroup) {
			defer wg.Done()
			url := fmt.Sprintf("%v:%v/%v", ip, gp.Port, uri)
			var msg string
			var status int
			if uri != "backup" {
				msg, status = callRollback("POST", url, request)
			} else {
				unAVLParams[index-1], msg, status = callBackup("POST", url, request)
			}

			switch status {
			case SUCCESS:
				*sucCount++
			case WARNING:
				*sucCount++
				*detailInfo += fmt.Sprintf("Group %v Node %v: %v ", groupID, index, ip)
			case FAILED:
				*failedInfo += fmt.Sprintf("\tGroup %v Node %v: %v\n%v\n", groupID, index, ip, msg)
			}

			return
		}(index+1, gp.GroupNo, ip, &wg)
	}

	wg.Wait()

	if uri == "backup" {
		warningInfo, status := gp.deleteUnAVLConf(unAVLParams)
		if status == FAILED {
			return warningInfo, false
		}

		if status == WARNING {
			return warningInfo, true
		}

		return warningInfo, true
	}

	if *sucCount == len(gp.IPs) {
		return *detailInfo, true
	}

	return *failedInfo, false
}

func callRollback(method string, url string, request interface{}) (string, int) {
	resp, err := http.RemoteCall(method, url, request)
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			return "server is offline", FAILED
		}

		return err.Error(), FAILED
	}

	var response map[string]struct {
		Suc bool        `json:"suc"`
		Msg interface{} `json:"msg"`
	}

	if err = json.Unmarshal(resp, &response); err != nil {
		return string(resp), FAILED
	}

	var parseFailedMsg string
	var warnCount int

	for domain, res := range response {
		if parseStatusCode(res.Msg) == WARNING {
			warnCount++
			continue
		}

		if !res.Suc {
			parseFailedMsg += fmt.Sprintf("\n\t'%v' failed msg: %v", domain, res.Msg)
		}
	}

	if parseFailedMsg != "" {
		return parseFailedMsg, FAILED
	}

	if warnCount == len(response) {
		return "No domain needs to be rolled back", WARNING
	}

	return "", SUCCESS
}

func parseStatusCode(msg interface{}) int {
	switch info := msg.(type) {
	case map[string]interface{}:
		var count int
		for _, value := range info {
			message := fmt.Sprint(value)
			if strings.Contains(message, BackupNotFound) || strings.Contains(message, FileNotExist) || strings.Contains(message, NoBackupFile) {
				count++
			}
		}

		if count == len(info) && count > 0 {
			return WARNING
		}
		return SUCCESS
	case string:
		if strings.Contains(info, BackupNotFound) || strings.Contains(info, FileNotExist) || strings.Contains(info, NoBackupFile) {
			return WARNING
		}
		return SUCCESS
	case interface{}:
		if info == nil {
			return SUCCESS
		}
		return WARNING
	}

	return SUCCESS
}

func parseMsg(originMsg interface{}) string {
	var resp map[string]struct {
		Suc bool   `json:"suc"`
		Msg string `json:"msg"`
	}

	msg, _ := json.Marshal(originMsg)
	err := json.Unmarshal(msg, &resp)
	if err != nil {
		return string(msg)
	}

	var retMsg string

	for domain, info := range resp {
		if info.Suc {
			continue
		}

		replaced := strings.ReplaceAll(info.Msg, "\n", "\n\t\t\t")

		retMsg += fmt.Sprintf("\t\t[%s]\t%v\n", domain, replaced)
	}

	if len(retMsg) != 0 {
		return retMsg
	}

	return string(msg)
}

func (tuner *Tuner) original() error {
	var domains []string
	tuner.rollbackReq = map[string]interface{}{
		"domains": domains,
		"all":     true,
	}
	return tuner.concurrent("rollback")
}


