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

func calculateCondition(content string, macros []string) (string, bool) {
	// detectedMacroValue used for replace macro with value by func convertString
	detectedMacroValue := make(map[string]string)
	if err := getMacroValue(macros, detectedMacroValue); err != nil {
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

func detectConfValue(re *regexp.Regexp, valueStr string, paramName string) (string, map[string]interface{}, error) {
	macros := utils.RemoveRepeated(re.FindAllString(strings.ReplaceAll(valueStr, " ", ""), -1))
	detectedMacroValue := make(map[string]string)

	if isConditionExp(valueStr) && len(macros) > 0 {
		expression, condMatched := calculateCondition(valueStr, macros)
		if condMatched {
			return "", nil, nil
		}

		return expression, nil, nil
	}

	value, err := getExtremeValue(macros, detectedMacroValue, valueStr)
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

func detect(data []string, macroNames []string, detectedMacroValue map[string]string) error {
	requestMap := getMethodReqByNames(data)
	url := fmt.Sprintf("%v:%v/method", config.KeenTune.BenchGroup[0].DestIP, config.KeenTune.Group[0].Port)
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
			calcResult, err := getExtremeValue(macros, detectedMacroValue, macroString)
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
			calcResult, err := getExtremeValue(macros, detectedMacroValue, v)
			if err != nil {
				return fmt.Errorf("'%v' calculate option err: %v", param.ParaName, err)
			}

			newOptions = append(newOptions, fmt.Sprintf("%v", int(calcResult)))
		}
		param.Options = newOptions
	}

	return nil
}


