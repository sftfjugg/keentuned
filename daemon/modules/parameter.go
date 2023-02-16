package modules

import (
	"fmt"
	"io/ioutil"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Parameter define a os parameter value scope, operating command and value
type Parameter struct {
	DomainName string        `json:"domain,omitempty"`
	ParaName   string        `json:"name,omitempty"`
	Scope      []interface{} `json:"range,omitempty"`
	Options    []string      `json:"options,omitempty"`
	Sequence   []interface{} `json:"sequence,omitempty"`
	Dtype      string        `json:"dtype"`
	Value      interface{}   `json:"value,omitempty"`
	Msg        string        `json:"msg,omitempty"`
	Step       int           `json:"step,omitempty"`
	Weight     float32       `json:"weight,omitempty"`
	Success    bool          `json:"suc,omitempty"`
	Base       interface{}   `json:"base,omitempty"`
}

const (
	defDetectAllReg = "#(!|\\?)([0-9A-Za-z_]+)#"
	defMarcoString  = "#!([0-9A-Za-z_]+)#"
	defDetectMapReg = "#\\?([0-9A-Za-z_]+)#"
	recommendReg    = "^recommend.*"
)

// updateParameter update the partial param by the total param
func updateParameter(partial, total *Parameter) {
	if partial.Dtype == "" {
		partial.Dtype = total.Dtype
	}

	if len(partial.Options) == 0 {
		partial.Options = total.Options
	}

	if partial.Value == nil {
		partial.Value = total.Value
	}

	if len(partial.Scope) == 0 {
		partial.Scope = total.Scope
	}

	if partial.Step == 0 {
		partial.Step = total.Step
	}
}

// UpdateParams update params by total param
func UpdateParams(userParam config.DBLMap) {
	for domainName, domainMap := range userParam {
		fileName := fmt.Sprintf("%s/parameter/%s.json", config.KeenTune.Home, domainName)

		totalParamMap, err := file.ReadFile2Map(fileName)
		if err != nil {
			log.Warnf("", "Read file: '%v' , err: %v\n", fileName, err)
			continue
		}

		compareMap := utils.Parse2Map(domainName, totalParamMap)
		if len(totalParamMap) == 0 || len(compareMap) == 0 {
			log.Warnf("", "domain [%v] does not exist when update params by parsing '%v'", domainName, fileName)
			continue
		}

		for name, paramValue := range domainMap {
			param, ok := paramValue.(map[string]interface{})
			if !ok {
				log.Warnf("", "parse [%+v] type to map failed", paramValue)
				continue
			}

			err := modifyParam(&param, compareMap, name)
			if err != nil {
				log.Warnf("", "modify parameters err:%v", err)
				continue
			}

			domainMap[name] = param
		}

		userParam[domainName] = domainMap
	}

	return
}

func modifyParam(originMap *map[string]interface{}, compareMap map[string]interface{}, paramName string) error {
	var userParam, totalParam Parameter
	err := utils.Map2Struct(originMap, &userParam)
	if err != nil {
		return fmt.Errorf("modifyParam map to struct err:%v", err)
	}

	if err = utils.Parse2Struct(paramName, compareMap, &totalParam); err != nil {
		return fmt.Errorf("modifyParam parse compareMap %+v to struct err:%v", compareMap, err)
	}

	updateParameter(&userParam, &totalParam)
	*originMap, err = utils.Struct2Map(userParam)
	if err != nil {
		return fmt.Errorf("modifyParam struct %+v To map  err:%v", userParam, err)

	}

	return nil
}

func getExtremeValue(macros []string, detectedMacroValue map[string]string, macroString string, ip string) (int64, error) {
	if len(macros) == 0 {
		return 0, fmt.Errorf("range type is '%v', but macros length is 0", macroString)
	}

	if err := getMacroValue(macros, detectedMacroValue, ip); err != nil {
		return 0, fmt.Errorf("get detect value failed: %v", err)
	}

	express, symbol, compareValue := convertString(macroString, detectedMacroValue)

	calcResult, err := utils.Calculate(express)
	if err != nil || len(compareValue) == 0 {
		return calcResult, err
	}

	switch symbol {
	case "MAX":
		return int64(math.Max(float64(calcResult), compareValue[0])), nil
	case "MIN":
		return int64(math.Min(float64(calcResult), compareValue[0])), nil
	}

	return calcResult, nil
}

func convertString(macroString string, macroMap map[string]string) (string, string, []float64) {
	retStr := strings.ReplaceAll(macroString, " ", "")
	for name, value := range macroMap {
		retStr = strings.ReplaceAll(retStr, name, fmt.Sprint(value))
	}

	var symbol string
	if len(retStr) > 4 {
		switch strings.ToUpper(retStr)[0:4] {
		case "MAX(", "MAX[":
			symbol = "MAX"
		case "MIN(", "MIN[":
			symbol = "MIN"
		default:
			return retStr, "", nil
		}

		macroParts := strings.Split(retStr[4:len(retStr)-1], ",")
		express := ""
		var compareInt []float64
		for _, part := range macroParts {
			value, err := strconv.ParseFloat(part, 64)
			if err != nil {
				express = part
			} else {
				compareInt = append(compareInt, value)
			}
		}

		return express, symbol, compareInt
	}

	return retStr, "", nil
}

func getMacroValue(macros []string, detectedMacroValue map[string]string, ip string) error {
	if len(macros) == 0 {
		return nil
	}

	var macroOrgNames []string
	var macroReqNames []string
	for _, macro := range macros {
		if _, ok := detectedMacroValue[macro]; ok {
			continue
		}

		name := strings.TrimSuffix(strings.TrimPrefix(macro, "#!"), "#")
		lowerName := strings.ToLower(name)

		macroOrgNames = append(macroOrgNames, name)

		macroReqNames = append(macroReqNames, lowerName)
	}

	if len(macroReqNames) == 0 {
		return nil
	}

	return detect(macroReqNames, macroOrgNames, detectedMacroValue, ip)
}

// ConvertConfFileToJson convert conf file to json
func ConvertConfFileToJson(fileName string, ip ...string) (ABNLResult, map[string]map[string]interface{}, error) {
	var abnormal = ABNLResult{}
	replacedStr, err := readConfFile(fileName)
	if err != nil {
		return abnormal, nil, err
	}

	commonDomain, recommendMap, domainMap := parseConfStrToMapSlice(replacedStr, fileName, &abnormal, ip...)

	for key, value := range recommendMap {
		abnormal.Recommend += fmt.Sprintf("\t[%v]\n%v", key, strings.Join(value, ""))
	}

	if len(domainMap) == 0 {
		if abnormal.Recommend != "" {
			return abnormal, nil, nil
		}

		return abnormal, nil, fmt.Errorf("domain '%v' content is empty", commonDomain)
	}

	return changeMapSliceToDBLMap(domainMap, abnormal)
}

func readConfFile(fileName string) (string, error) {
	paramBytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return "", fmt.Errorf("read file err: %v", err)
	}

	if len(paramBytes) == 0 {
		return "", fmt.Errorf("read file is empty")
	}

	replacedStr := strings.ReplaceAll(string(paramBytes), "：", ":")
	return replacedStr, nil
}

func changeMapSliceToDBLMap(domainMap map[string][]map[string]interface{}, abnormal ABNLResult) (ABNLResult, map[string]map[string]interface{}, error) {
	var resultMap = make(map[string]map[string]interface{})
	for domain, paramSlice := range domainMap {
		if len(paramSlice) == 0 {
			return abnormal, nil, fmt.Errorf("domain '%v' content is empty", domain)
		}

		var paramMap = make(map[string]interface{})
		for _, paramInfo := range paramSlice {
			name, ok := paramInfo["name"].(string)
			if !ok {
				abnormal.Warning += fmt.Sprintf("%v name does not exist", paramInfo)
				continue
			}
			delete(paramInfo, "name")
			paramMap[name] = paramInfo
		}
		resultMap[domain] = paramMap
	}
	return abnormal, resultMap, nil
}

// return:
//        0: domain name
//        1: recommend summery info
//        2: map slice, design this data structure to avoid duplication and leakage
// parseConfStrToMapSlice ...
func parseConfStrToMapSlice(replacedStr, fileName string, abnormal *ABNLResult, ip ...string) (string, map[string][]string, map[string][]map[string]interface{}) {
	var deleteDomain string
	var recommendMap = make(map[string][]string)
	var domainMap = make(map[string][]map[string]interface{})
	var includeMap = make(map[string][]map[string]interface{})
	var commonDomain string
	var variableMap = make(map[string]string)
	var variableReq = make(map[string]interface{})
	var isVarReady bool
	for _, line := range strings.Split(replacedStr, "\n") {
		pureLine := strings.TrimSpace(replaceEqualSign(line))
		if len(pureLine) == 0 {
			continue
		}

		if strings.HasPrefix(pureLine, "#") {
			continue
		}

		if strings.HasPrefix(pureLine, "[") && strings.HasSuffix(pureLine, "]") {
			commonDomain = strings.TrimSpace(strings.Trim(strings.Trim(pureLine, "["), "]"))
			continue
		}

		if commonDomain == tunedMainDomain {
			includeMap = parseIncludeConf(pureLine, abnormal, ip...)
			continue
		}

		if commonDomain == deleteDomain {
			continue
		} else if len(deleteDomain) > 0 {
			// empty deleteDomain, after the last deleted section skipped
			deleteDomain = ""
		}

		if commonDomain == tunedBootloaderDomain {
			if len(recommendMap[tunedBootloaderDomain]) == 0 {
				recommendMap[tunedBootloaderDomain] = append(recommendMap[tunedBootloaderDomain], fmt.Sprintf("\t\t%v\n", bootloaderRecommend))
			}

			continue
		}

		if commonDomain == tunedVariableDomain {
			collectConfVariables(pureLine, variableMap, variableReq)
			continue
		}

		if len(variableReq) > 0 && !isVarReady {
			isVarReady = true
			err := requestAllVariables(variableMap, variableReq)
			if err != nil {
				recommend := fmt.Sprintf("errMsg: %v; Please Check the variable in %v\n", err, fileName)
				recommendMap[tunedVariableDomain] = append(recommendMap[tunedVariableDomain], recommend)
				return tunedVariableDomain, recommendMap, nil
			}
		}

		if matchString(defVarReg, pureLine) {
			pureLine = replaceVariables(variableMap, pureLine)
		}

		recommend, condition, param, err := readLine(pureLine, ip...)
		if len(condition) != 0 {
			deleteDomain = commonDomain
			if commonDomain == myConfDomain {
				notMetInfo := fmt.Sprintf(detectENVNotMetFmt, commonDomain, myConfCondition, file.GetPlainName(fileName))
				abnormal.Warning += fmt.Sprintf("%v%v", notMetInfo, multiRecordSeparator)
				continue
			}

			notMetInfo := fmt.Sprintf(detectENVNotMetFmt, commonDomain, condition, file.GetPlainName(fileName))
			abnormal.Warning += fmt.Sprintf("%v%v", notMetInfo, multiRecordSeparator)

			continue
		}

		if err != nil {
			abnormal.Warning += fmt.Sprintf("content '%v' abnormal%v", pureLine, multiRecordSeparator)
			continue
		}

		if len(recommend) != 0 {
			recommendMap[commonDomain] = append(recommendMap[commonDomain], recommend)
			continue
		}

		// when condition is empty, param maybe null
		if param == nil {
			continue
		}

		convertedDomain := convertDomain(commonDomain)

		domainMap[convertedDomain] = append(domainMap[convertedDomain], param)
	}

	if len(includeMap) > 0 {
		return commonDomain, recommendMap, mergedMapSlice(domainMap, includeMap)
	}

	return commonDomain, recommendMap, domainMap
}

func collectConfVariables(pureLine string, variableMap map[string]string, variableReq map[string]interface{}) {
	variableParts := strings.Split(pureLine, ":")
	if len(variableParts) <= 1 {
		return
	}

	varName := strings.TrimSpace(variableParts[0])

	// skip include field in variable
	if tunedIncludeField == varName {
		return
	}

	if strings.Contains(varName, "assert") {
		return
	}

	varValue := strings.TrimSpace(strings.Join(variableParts[1:], ":"))
	value, find := expectedRegx[varName]
	if find && value != varValue {
		expectedRegx[varName] = varValue
	}

	if specVariableName[varName] {
		variableMap[varName] = specVariableValue[varName]
		return
	}

	if matchString("\\$\\{(.*)\\}", varValue) {
		getVariableReq(pureLine, variableReq)
		return
	}

	variableMap[varName] = varValue
}

func parseIncludeConf(pureLine string, abnormal *ABNLResult, ip ...string) map[string][]map[string]interface{} {
	if !strings.Contains(pureLine, tunedIncludeField) {
		return nil
	}

	pairs := strings.Split(pureLine, ":")
	if len(pairs) != 2 {
		return nil
	}

	includeFile := fmt.Sprintf("%v.conf", strings.TrimSuffix(strings.TrimSpace(pairs[1]), ".conf"))
	includeInfo, err := readConfFile(config.GetProfileHomePath(includeFile))
	if err != nil {
		abnormal.Warning += fmt.Sprintf("Read include file '%v' failed%v", pairs[1], multiRecordSeparator)
		return nil
	}

	_, _, includeMap := parseConfStrToMapSlice(includeInfo, includeFile, abnormal, ip...)
	return includeMap
}

func mergedMapSlice(domainMap map[string][]map[string]interface{}, includeMap map[string][]map[string]interface{}) map[string][]map[string]interface{} {
	var mergedMap = make(map[string][]map[string]interface{})
	for domainName, params := range includeMap {
		mergedMap[domainName] = append(mergedMap[domainName], params...)
	}

	for domainName, params := range domainMap {
		mergedMap[domainName] = append(mergedMap[domainName], params...)
	}

	return mergedMap
}

func readLine(line string, ip ...string) (string, string, map[string]interface{}, error) {
	paramSlice := strings.Split(line, ":")
	partLen := len(paramSlice)
	switch {
	case partLen <= 1:
		return "", "", nil, fmt.Errorf("param %v length %v is invalid, required: 2", paramSlice, len(paramSlice))
	case partLen == 2:
		return getParam(paramSlice, ip...)
	default:
		newSlice := []string{paramSlice[0]}
		newSlice = append(newSlice, strings.Join(paramSlice[1:], ":"))
		return getParam(newSlice, ip...)
	}
}

func getParam(paramSlice []string, ip ...string) (string, string, map[string]interface{}, error) {
	paramName := strings.TrimSpace(paramSlice[0])
	valueStr := strings.ReplaceAll(strings.TrimSpace(paramSlice[1]), "\"", "")

	recommend, skip := isRecommend(valueStr, paramName)
	if skip {
		return recommend, "", nil, nil
	}

	re, _ := regexp.Compile(defDetectAllReg)
	if re != nil && re.MatchString(valueStr) && len(ip) > 0 {
		expression, param, err := detectConfValue(re, valueStr, paramName, ip[0])
		// replace expression to real condition when expression is not empty
		if expression != "" {
			return "", valueStr, nil, nil
		}

		return "", "", param, err
	}

	param := genParam(valueStr, paramName)
	return "", "", param, nil
}

func genParam(valueStr string, paramName string) map[string]interface{} {
	var param map[string]interface{}
	// remove inline comments
	var rmCommentVal string
	if strings.Index(valueStr, "#") > 0 {
		rmCommentVal = strings.TrimSpace(strings.Split(valueStr, "#")[0])
	} else {
		rmCommentVal = valueStr
	}

	value, err := strconv.ParseInt(rmCommentVal, 10, 64)
	if err != nil {
		param = map[string]interface{}{
			"value": valueStr,
			"dtype": "string",
			"name":  paramName,
		}
		return param
	}

	param = map[string]interface{}{
		"value": value,
		"dtype": "int",
		"name":  paramName,
	}
	return param
}

func isRecommend(valueStr string, paramName string) (string, bool) {
	var recommend string

	if skipParamDict[paramName] {
		recommend = fmt.Sprintf("\t\t%v: %v\n", paramName, notSupportRecommend)
		return recommend, true
	}

	matched, _ := regexp.MatchString(recommendReg, strings.ToLower(valueStr))
	if matched {
		forceWrapLine := strings.Replace(valueStr, ". Please", ".\n\t\t\tPlease", 1)
		recommend = fmt.Sprintf("\t\t%v: %v\n", paramName, strings.TrimPrefix(forceWrapLine, "recommend:"))
		return recommend, true
	}

	return "", false
}

func replaceEqualSign(origin string) string {
	equalIdx := strings.Index(origin, "=")
	colonIdx := strings.Index(origin, ":")
	// First, '=' exists; if ':' not exist or '=' before ':', replace '=' by ':'
	if equalIdx > 0 && (colonIdx < 0 || equalIdx < colonIdx) {
		return strings.Replace(origin, "=", ":", 1)
	}

	return origin
}

// ConvertToSequentialDict ...
func ConvertToSequentialDict(fileName string) ([]map[string]map[string]string, error) {
	replacedStr, err := readConfFile(fileName)
	if err != nil {
		return nil, err
	}
	var domain string
	var seqDict []map[string]map[string]string
	var paramDict = make(map[string]map[string]string)
	var seqDomains []string
	for _, originLine := range strings.Split(replacedStr, "\n") {
		pureLine := strings.TrimSpace(replaceEqualSign(originLine))
		if len(pureLine) == 0 {
			continue
		}

		if strings.HasPrefix(pureLine, "#") {
			continue
		}

		if strings.HasPrefix(pureLine, "[") && strings.HasSuffix(pureLine, "]") {
			domain = strings.TrimSpace(strings.Trim(strings.Trim(pureLine, "["), "]"))
			seqDomains = append(seqDomains, domain)
			paramDict[domain] = make(map[string]string)
			continue
		}

		paramSlice := strings.Split(pureLine, ":")
		partLen := len(paramSlice)
		switch {
		case partLen <= 1:
			continue
		case partLen == 2:
			name := paramSlice[0]
			value := paramSlice[1]
			paramDict[domain][name] = value
		default:
			newSlice := []string{paramSlice[0]}
			newSlice = append(newSlice, strings.Join(paramSlice[1:], ":"))
			name := newSlice[0]
			value := newSlice[1]
			paramDict[domain][name] = value
		}
	}

	for _, domain := range seqDomains {
		params := paramDict[domain]
		seqDict = append(seqDict, map[string]map[string]string{domain: params})
	}

	return seqDict, nil
}


