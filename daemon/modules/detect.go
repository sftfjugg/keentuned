package modules

import (
	"encoding/json"
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/utils"
	"keentune/daemon/common/utils/http"
	"regexp"
	"strings"
)

const conditionReg = "[|&<>=!∈]"

var (
	detectENVNotMetFmt = "Settings in [%v] domain only suites for %v Env, please set your parameters refer to %v"
)

// domain
const (
	myConfDomain = "my_cnf"
)

const (
	myConfCondition = "8 CPU 32G Memory"
)

// ABNLResult abnormal result
type ABNLResult struct {
	Recommend string
	Warning   string
}

func calculateCondition(content string, macros []string, ip string) (string, bool) {
	// detectedMacroValue used for replace macro with value by func convertString
	detectedMacroValue := make(map[string]string)
	if err := getMacroValue(macros, detectedMacroValue, ip); err != nil {
		return err.Error(), false
	}

	express, _, _ := convertString(content, detectedMacroValue)
	if isExpectedRegx(express) {
		return express, getCondResultWithVar(express)
	}

	return express, utils.CalculateCondExp(express)
}

func isConditionExp(content string) bool {
	regCond := regexp.MustCompile(conditionReg)
	regMacro := regexp.MustCompile(defMarcoString)
	if regCond == nil || regMacro == nil {
		return false
	}

	matchMacro := regMacro.MatchString(content)
	replacedMacro := regMacro.ReplaceAllString(content, "$1")
	matchCond := regCond.MatchString(replacedMacro)
	return matchMacro && matchCond
}

func detectConfValue(re *regexp.Regexp, valueStr string, paramName string, ip string) (string, map[string]interface{}, error) {
	macros := utils.RemoveRepeated(re.FindAllString(strings.ReplaceAll(valueStr, " ", ""), -1))

	if len(macros) == 0 {
		return "", nil, fmt.Errorf("detect '%v' found macros length is 0", valueStr)
	}

	if isDetectMapString(paramName, valueStr) {
		return detectSpecVarMap(paramName, valueStr, ip)
	}
	if isConditionExp(valueStr) {
		expression, condMatched := calculateCondition(valueStr, macros, ip)
		if condMatched {
			return "", nil, nil
		}

		return expression, nil, nil
	}

	detectedMacroValue := make(map[string]string)
	value, err := getExtremeValue(macros, detectedMacroValue, valueStr, ip)
	if err != nil {
		return "", nil, fmt.Errorf("detect '%v'err: %v", paramName, err)
	}

	param := map[string]interface{}{
		"value": value,
		"dtype": "int",
		"name":  paramName,
	}

	return "", param, nil
}

func detectSpecVarMap(paramName string, valueStr string, ip string) (string, map[string]interface{}, error) {
	methodName := detectSpecEnvDict[paramName]
	if methodName == "" {
		return "", nil, fmt.Errorf("methodName '%v' not defined", paramName)
	}

	requestMap := getMethodReqByNames([]string{methodName})
	url := fmt.Sprintf("%v:%v/method", ip, config.KeenTune.Group[0].Port)
	respByte, err := http.RemoteCall("POST", url, requestMap)
	if err != nil {
		return "", nil, fmt.Errorf("remote call err:%v", err)
	}

	var resp []methodResp
	err = json.Unmarshal(respByte, &resp)
	if err != nil || len(resp) == 0 {
		return "", nil, fmt.Errorf("unmarshal method response err:%v", err)
	}

	var detectedValue string
	detectedValue = methodResultDict[methodName][resp[0].Result]
	if detectedValue == "" {
		detectedValue = methodResultDict[methodName]["default"]
	}

	param := genParam(detectedValue, paramName)
	return "", param, nil
}

func isDetectMapString(paramName, valueStr string) bool {
	regMacro := regexp.MustCompile(defDetectMapReg)
	if regMacro == nil {
		return false
	}
	matched := regMacro.MatchString(valueStr)
	replacedMacro := utils.RemoveRepeated(regMacro.FindAllString(valueStr, -1))
	if len(replacedMacro) == 0 {
		return false
	}

	detectVar := strings.TrimSuffix(strings.TrimPrefix(replacedMacro[0], "#?"), "#")
	if matched && detectVar == paramName {
		return true
	}

	return false
}

func detect(data []string, macroNames []string, detectedMacroValue map[string]string, ip string) error {
	requestMap := getMethodReqByNames(data)
	url := fmt.Sprintf("%v:%v/method", ip, config.KeenTune.Group[0].Port)
	respByte, err := http.RemoteCall("POST", url, requestMap)
	if err != nil {
		return fmt.Errorf("remote call err:%v", err)
	}

	var resp []methodResp
	err = json.Unmarshal(respByte, &resp)
	if err != nil {
		return fmt.Errorf("unmarshal method response err:%v", err)
	}

	if len(macroNames) != len(resp) {
		return fmt.Errorf("method response length is %v, expect %v", len(resp), len(macroNames))
	}

	var failedInfo string
	for idx, name := range macroNames {
		result := resp[idx]
		if !result.Suc {
			failedInfo += fmt.Sprintf("macro '%v' response res '%v' is false\n", name, result.Result)
			continue
		}

		macro := fmt.Sprintf("#!%v#", name)
		detectedMacroValue[macro] = result.Result
	}

	if failedInfo != "" {
		return fmt.Errorf("method response failed, %v", failedInfo)
	}

	return nil
}

func detectParam(param *Parameter) error {
	destIP := config.KeenTune.BenchGroup[0].DestIP
	if len(param.Scope) > 0 {
		var range2Int []interface{}
		var detectedMacroValue = make(map[string]string)
		for _, v := range param.Scope {
			value, ok := v.(float64)
			if ok {
				range2Int = append(range2Int, int(value))
				continue
			}
			macroString, ok := v.(string)
			re, _ := regexp.Compile(defMarcoString)
			macros := utils.RemoveRepeated(re.FindAllString(strings.ReplaceAll(macroString, " ", ""), -1))
			calcResult, err := getExtremeValue(macros, detectedMacroValue, macroString, destIP)
			if err != nil {
				return fmt.Errorf("'%v' calculate range err: %v", param.ParaName, err)
			}
			range2Int = append(range2Int, int(calcResult))
		}
		param.Scope = range2Int
	}

	if len(param.Options) > 0 {
		var newOptions []string
		var detectedMacroValue = make(map[string]string)
		for _, v := range param.Options {
			re, _ := regexp.Compile(defMarcoString)
			if !re.MatchString(v) {
				newOptions = append(newOptions, v)
				continue
			}

			macros := utils.RemoveRepeated(re.FindAllString(strings.ReplaceAll(v, " ", ""), -1))
			calcResult, err := getExtremeValue(macros, detectedMacroValue, v, destIP)
			if err != nil {
				return fmt.Errorf("'%v' calculate option err: %v", param.ParaName, err)
			}

			newOptions = append(newOptions, fmt.Sprintf("%v", int(calcResult)))
		}
		param.Options = newOptions
	}

	return nil
}


