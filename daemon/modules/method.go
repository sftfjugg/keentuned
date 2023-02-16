package modules

import (
	"encoding/json"
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/utils/http"
	"regexp"
	"strings"
)

const (
	defDoubleFuncReg       = "\\$\\{f:(.*):\\$\\{f:(.*):(.*):(.*)\\}\\}"
	defFuncWithAssertReg   = "\\$\\{f:(.*):(.*):(.*):(.*)\\}"
	defFuncWithOneArgReg   = "\\$\\{f:(.*):(.*)\\}"
	defFuncWithFourArgsReg = "\\$\\{f:(.*):(.*):(.*):(.*):(.*)\\}"
)

var specVariableName = map[string]bool{
	"no_balance_cores":            true,
	"isolated_cores":              true,
	"isolated_cores_assert_check": true,
	"isolate_managed_irq":         true,
	"netdev_queue_count":          true,
}

var specVariableValue = map[string]string{
	"no_balance_cores":            "2-3",
	"isolated_cores":              "5",
	"isolated_cores_assert_check": "\\2-3",
	"netdev_queue_count":          "4",
	"isolate_managed_irq":         "Y",
}

type methodReq struct {
	Name string        `json:"method_name"`
	Args []interface{} `json:"method_args"`
}

type methodResp struct {
	Suc    bool   `json:"suc"`
	Result string `json:"res"`
}

func getMethodReqByNames(data []string) []methodReq {
	var retReqMethod []methodReq
	for _, methodName := range data {
		retReqMethod = append(retReqMethod, methodReq{
			Name: methodName,
			Args: []interface{}{},
		})
	}

	return retReqMethod
}

func getMethodReqByArg(data map[string]interface{}) ([]string, []methodReq) {
	var retReqMethod []methodReq
	var varNames []string
	for name, arg := range data {
		retReqMethod = append(retReqMethod, arg.(methodReq))
		varNames = append(varNames, name)
	}

	return varNames, retReqMethod
}

func matchString(pattern, s string) bool {
	res, err := regexp.MatchString(pattern, s)
	if err != nil {
		return false
	}

	return res
}

func getVariableReq(line string, varMap map[string]interface{}) {
	parts := strings.Split(line, ":")
	if len(parts) < 2 {
		return
	}
	varName := strings.TrimSpace(parts[0])
	varRegexStr := strings.TrimSpace(strings.Join(parts[1:], ":"))

	switch {
	case matchString(defDoubleFuncReg, varRegexStr):
		req := getDoubleFuncMethodReq(varRegexStr)
		varMap[varName] = req
	case matchString(defFuncWithFourArgsReg, varRegexStr):
		// todo
	case matchString(defFuncWithAssertReg, varRegexStr):
		// todo
	case matchString(defFuncWithOneArgReg, varRegexStr):
		req := getFuncWithOneArgMethodReq(varRegexStr, varMap)
		varMap[varName] = req
	default:
		return
	}
}

func getFuncWithOneArgMethodReq(origin string, varMap map[string]interface{}) methodReq {
	reg := regexp.MustCompile(defFuncWithOneArgReg)
	replaced := reg.ReplaceAllString(origin, "$1#$ $2")
	args := strings.Split(replaced, "#$ ")
	if len(args) != 2 {
		return methodReq{
			Name: args[0],
			Args: []interface{}{},
		}
	}

	if matchString("\\$\\{(.*)\\}", args[1]) {
		varName := strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(args[1]), "${"), "}")
		specValue, find := specVariableValue[varName]
		var arg interface{}
		if find {
			arg = specValue
		} else {
			arg = varMap[varName]
		}

		return methodReq{
			Name: args[0],
			Args: []interface{}{arg},
		}
	}

	return methodReq{
		Name: args[0],
		Args: []interface{}{args[1]},
	}
}

func getDoubleFuncMethodReq(origin string) methodReq {
	reg := regexp.MustCompile(defDoubleFuncReg)
	replaced := reg.ReplaceAllString(origin, "$1#$ $2#$ $3#$ $4")
	args := strings.Split(replaced, "#$ ")
	if len(args) != 4 {
		return methodReq{
			Name: args[0],
			Args: []interface{}{},
		}
	}

	innerReq := methodReq{
		Name: args[1],
		Args: []interface{}{args[2], args[3]},
	}

	return methodReq{
		Name: args[0],
		Args: []interface{}{innerReq},
	}
}

func requestAllVariables(destMap map[string]string, reqMap map[string]interface{}) error {
	varNames, req := getMethodReqByArg(reqMap)
	url := fmt.Sprintf("%v:%v/method", config.KeenTune.BenchGroup[0].DestIP, config.KeenTune.Group[0].Port)
	respByte, err := http.RemoteCall("POST", url, req)
	if err != nil {
		return fmt.Errorf("remote call err:%v", err)
	}

	var resp []methodResp
	err = json.Unmarshal(respByte, &resp)
	if err != nil {
		return fmt.Errorf("unmarshal method response err:%v", err)
	}

	if len(varNames) != len(resp) {
		return fmt.Errorf("method response length is %v, expect %v", len(resp), len(varNames))
	}

	var failedInfo string
	for idx, varName := range varNames {
		result := resp[idx]
		if !result.Suc {
			failedInfo += fmt.Sprintf("variable '%v' response res '%v' is false\n", varName, result.Result)
			continue
		}
		destMap[varName] = result.Result
	}

	if failedInfo != "" {
		return fmt.Errorf("method response failed, %v", failedInfo)
	}

	return nil
}

// GetEnvCondition get environment condition by remote call '/method'
func GetEnvCondition(param []map[string]map[string]string, host string) (map[string]string, error) {
	names, req, err := parseEnvCondReq(param)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%v/method", host)
	respByte, err := http.RemoteCall("POST", url, req)
	if err != nil {
		return nil, fmt.Errorf("remote call err:%v", err)
	}

	var resp []methodResp
	err = json.Unmarshal(respByte, &resp)
	if err != nil {
		return nil, fmt.Errorf("unmarshal method response err:%v", err)
	}

	if len(names) != len(resp) {
		return nil, fmt.Errorf("method response length is %v, expect %v", len(resp), len(names))
	}

	var destMap = make(map[string]string)
	for idx, varName := range names {
		result := resp[idx]
		if !result.Suc {
			destMap[varName] = ""
			continue
		}
		destMap[varName] = result.Result
	}

	return destMap, nil
}

func parseEnvCondReq(param []map[string]map[string]string) ([]string, []methodReq, error) {
	var methodNames = make(map[string]string)
	for _, domainCondDict := range param {
		for _, conds := range domainCondDict {
			for cond := range conds {
				_, find := methodNames[cond]
				if !find {
					methodNames[cond] = "true"
					continue
				}
			}
		}
	}

	if len(methodNames) == 0 {
		return nil, nil, fmt.Errorf("no rules key found")
	}

	var req []methodReq
	var reqNames []string
	for name := range methodNames {
		req = append(req, methodReq{
			Name: name,
			Args: []interface{}{},
		})
		reqNames = append(reqNames, name)
	}

	return reqNames, req, nil
}


