package modules

import (
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	"strings"
	"sync"
	"time"
)

const applyFailureResult = "apply result exists failure"

type request struct {
	params  []map[string]interface{}
	ip      string
	id      int // ip index in its' own group
	groupID int
	ipIndex int // id of ip in total targets
	body    interface{}
}

func (tuner *Tuner) getConfigure() error {
	tuner.ReadConfigure = true
	return tuner.configure()
}

func (tuner *Tuner) setConfigure() error {
	tuner.ReadConfigure = false
	return tuner.configure()
}

func (tuner *Tuner) configure() error {
	wg := sync.WaitGroup{}
	sc := NewSafeChan()
	defer sc.SafeStop()
	var applySummary string
	var targetFinishStatus = make([]string, len(config.KeenTune.IPMap))
	start := time.Now()
	var errDetail string
	for groupID, group := range tuner.Group {
		req := group.newRequester(groupID)
		for index, ip := range group.IPs {
			wg.Add(1)
			req.id = index + 1
			req.ip = ip
			go func(req request) {
				err := tuner.apply(&wg, targetFinishStatus, req)
				if err != nil && strings.Contains(err.Error(), "interrupted") {
					log.Infof("", "safe stop id: %v", req.id)
					sc.SafeStop()
					errDetail = "apply is interrupted"
				}
			}(req)
		}
	}

	wg.Wait()

	for _, status := range targetFinishStatus {
		if strings.Contains(status, applyFailureResult) {
			errDetail += status
			continue
		}
		applySummary += status
	}

	if errDetail != "" {
		return fmt.Errorf(strings.TrimSuffix(errDetail, "\n"))
	}

	tuner.applySummary = applySummary
	timeCost := utils.Runtime(start)
	tuner.timeSpend.apply += timeCost.Count

	if tuner.Verbose && !tuner.ReadConfigure {
		log.Infof(tuner.logName, "[Iteration %v] Apply success, %v", tuner.Iteration, timeCost.Desc)
	}

	return nil
}

func (tuner *Tuner) apply(wg *sync.WaitGroup, targetFinishStatus []string, req request) error {
	start := time.Now()
	var errMsg error
	req.ipIndex = config.KeenTune.IPMap[req.ip]
	config.IsInnerApplyRequests[req.ipIndex] = true
	identity := fmt.Sprintf("Group %v Node %v: %v", tuner.Group[req.groupID].GroupNo, req.id, req.ip)
	defer func() {
		wg.Done()
		config.IsInnerApplyRequests[req.ipIndex] = false
		if errMsg != nil {
			targetFinishStatus[req.ipIndex-1] = fmt.Sprintf("%v %v:\n%v", identity, applyFailureResult, errMsg)
		}
	}()

	var applyResult string
	if tuner.ReadConfigure {
		applyResult, errMsg = tuner.Group[req.groupID].Get(req)
	} else {
		applyResult, errMsg = tuner.Group[req.groupID].Set(req)
	}

	if errMsg != nil {
		return errMsg
	}

	targetFinishStatus[req.ipIndex-1] = fmt.Sprintf("\t%v\n%v\n", identity, applyResult)
	tuner.timeSpend.apply += utils.Runtime(start).Count
	return nil
}

// Set set configure
func (gp *Group) Set(req request) (string, error) {
	var setResult string
	for index := range gp.Params {
		if gp.Params[index] == nil {
			continue
		}

		gp.ReadOnly = false
		req.body = gp.applyReq(req.ip, req.params[index])
		result, err := gp.Configure(req)
		if err != nil {
			return result, err
		}

		setResult += result
	}

	return setResult, nil
}

// Configure handle configure request
func (gp *Group) Configure(req request) (string, error) {
	host := fmt.Sprintf("%s:%s", req.ip, gp.Port)
	applyResult, paramInfo, err := newTarget(req.ipIndex, host, req.body).configure()
	if err != nil {
		return applyResult, err
	}

	// pay attention to: the results in the same group are the same and only need to be updated once to prevent map concurrency security problems
	if gp.AllowUpdate[req.ip] && gp.ReadOnly {
		err = gp.updateParams(paramInfo)
		if err != nil {
			return applyResult, fmt.Errorf("update apply result: %v", err)
		}

		gp.updateDump(paramInfo)
	}

	return applyResult, nil
}

// Get get configure
func (gp *Group) Get(req request) (string, error) {
	gp.ReadOnly = true
	req.body = gp.applyReq(req.ip, gp.MergedParam)
	return gp.Configure(req)
}

func (gp *Group) newRequester(id int) request {
	var data = make([]map[string]interface{}, len(gp.Params))
	for idx, params := range gp.Params {
		data[idx] = deepCopy(params)
	}

	return request{params: data, groupID: id}
}


