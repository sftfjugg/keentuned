package modules

import (
	"encoding/json"
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	"keentune/daemon/common/utils/http"
	"strings"
)

// Configuration define a group of parameter and benchmark score in this configuration.
type Configuration struct {
	Parameters []Parameter           `json:"parameters"`
	Score      map[string]ItemDetail `json:"score"`
	Round      int                   `json:"current_round"`
	budget     float32
	timeSpend  utils.TimeSpend
	targetIP   []string
}

// ReceivedConfigure Received Configure from brain
type ReceivedConfigure struct {
	Candidate  []Parameter           `json:"candidate"`
	Score      map[string]ItemDetail `json:"bench_score,omitempty"`
	Iteration  int                   `json:"iteration"`
	Budget     float32               `json:"budget"`
	ParamValue string                `json:"parameter_value,omitempty"`
}

// ItemDetail multi item details
type ItemDetail struct {
	Value    float32   `json:"value,omitempty"`
	Negative bool      `json:"negative"`
	Weight   float32   `json:"weight"`
	Strict   bool      `json:"strict"`
	Baseline []float32 `json:"base,omitempty"`
}

type target struct {
	desSeparator string // description separator
	host         string
	ipIndex      int
	tabSeparator string // table separator
	reqBody      interface{}
	respBody     []byte
}

// Save configuration to profile file
func (conf Configuration) Save(fileName, suffix string) error {
	// acquire API return round is 1 less than the actual round value
	conf.Round = conf.Round + 1

	err := file.Dump2File(config.GetTuningPath(fileName), fileName+suffix, conf)
	if err != nil {
		return err
	}
	return err
}

// collectParam collect param change map to struct map and state param success information
func (tg target) collectParam(applyResp map[string]interface{}) (string, map[string]Parameter, error) {
	if len(applyResp) == 0 {
		return "", nil, fmt.Errorf("apply response is null")
	}

	var paramCollection = make(map[string]Parameter)
	var setResult string
	var totalFailed int

	for domain, paramMap := range applyResp {
		var sucCount, failedCount, skippedCount int
		var failedInfoSlice []string

		setResult += fmt.Sprintf("%v[%v] ", tg.desSeparator, domain)

		parameter, _ := paramMap.(map[string]interface{})

		for name, param := range parameter {
			var detail struct {
				Success bool        `json:"suc"`
				Msg     interface{} `json:"msg"`
				Value   interface{} `json:"value"`
			}

			utils.Map2Struct(param, &detail)

			valueStr, ok := detail.Value.(string)
			if ok && strings.Contains(valueStr, "\t") {
				detail.Value = strings.ReplaceAll(valueStr, "\t", " ")
			}

			paramCollection[name] = Parameter{
				DomainName: domain,
				ParaName:   name,
				Value:      detail.Value,
			}

			if detail.Success {
				sucCount++
				continue
			}

			failedCount++
			totalFailed++
			msg := strings.ReplaceAll(fmt.Sprint(detail.Msg), "\n", ". ")
			redFailed := utils.ColorString("red", "FAILED")
			failedInfoSlice = append(failedInfoSlice, fmt.Sprintf("%v[%v] %v: %v", tg.tabSeparator, redFailed, name, msg))
		}

		successInfo := fmt.Sprintf("%v Succeeded, %v Failed, %v Skipped", sucCount, failedCount, skippedCount)
		if failedCount == 0 {
			setResult += fmt.Sprintf("%v\n", successInfo)
			continue
		}

		failedDetail := strings.Join(failedInfoSlice, "\n")
		setResult += fmt.Sprintf("%v\n%s\n", successInfo, failedDetail)
	}

	if totalFailed == len(paramCollection) {
		return setResult, paramCollection, fmt.Errorf("return all failed")
	}

	return setResult, paramCollection, nil
}

func (tg target) getApplyResult() (map[string]interface{}, error) {
	var applyShortRet struct {
		Success bool        `json:"suc"`
		Msg     interface{} `json:"msg"`
	}

	err := json.Unmarshal(tg.respBody, &applyShortRet)
	if err != nil {
		return nil, err
	}

	if !applyShortRet.Success {
		detail, _ := json.Marshal(applyShortRet.Msg)
		if len(detail) != 0 {
			return nil, fmt.Errorf("%s", detail)
		}
		return nil, fmt.Errorf("%v", applyShortRet.Msg)
	}

	var applyResp struct {
		Data map[string]interface{} `json:"data"`
	}

	select {
	case body := <-config.ApplyResultChan[tg.ipIndex]:
		log.Debugf(log.ParamTune, "target id: %v receive apply result :[%v]\n", tg.ipIndex, string(body))
		if err := json.Unmarshal(body, &applyResp); err != nil {
			return nil, fmt.Errorf("Parse apply response Unmarshal err: %v", err)
		}
	case <-StopSig:
		return nil, fmt.Errorf("get apply result is interrupted")
	}

	return applyResp.Data, nil
}

// GetApplyResult get apply result by waiting for target active reports
func (tg target) GetApplyResult() (string, map[string]Parameter, error) {
	applyResp, err := tg.getApplyResult()
	if err != nil {
		return "", nil, err
	}

	return tg.collectParam(applyResp)
}

// Configure ...
func Configure(req interface{}, host string, ipIndex int) (string, error) {
	config.IsInnerApplyRequests[ipIndex] = true
	defer func() { config.IsInnerApplyRequests[ipIndex] = false }()
	desSep := "\t"
	tabSep := "\t\t"

	tgt := newTarget(ipIndex, host, req, []string{desSep, tabSep}...)

	applyResult, _, err := tgt.configure()
	return applyResult, err
}

func newTarget(index int, host string, body interface{}, args ...string) target {
	desSep, tabSep := "\t\t", "\t\t\t"
	if len(args) == 2 {
		desSep = args[0]
		tabSep = args[1]
	}

	return target{
		desSeparator: desSep,
		host:         host,
		ipIndex:      index,
		tabSeparator: tabSep,
		reqBody:      body,
	}
}

func (tg target) configure() (string, map[string]Parameter, error) {
	uri := fmt.Sprintf("%v/configure", tg.host)
	var err error
	tg.respBody, err = http.RemoteCall("POST", uri, tg.reqBody)
	if err != nil {
		return "", nil, fmt.Errorf("remote call: %v", err)
	}

	applyResult, paramInfo, err := tg.GetApplyResult()
	if err != nil {
		if applyResult != "" {
			return applyResult, nil, fmt.Errorf(applyResult)
		}

		return applyResult, nil, err
	}

	return applyResult, paramInfo, nil
}


