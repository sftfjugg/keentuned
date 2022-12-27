package modules

import (
	"keentune/daemon/common/utils"
	"regexp"
	"strings"
)

var specialDomain = map[string]string{
	"scheduler.amd":   "scheduler",
	"scheduler_amd":   "scheduler",
	"sysctl.thunderx": "sysctl",
	"sysctl_thunderx": "sysctl",
	"vm.thunderx":     "vm",
	"vm_thunderx":     "vm",
}

const (
	tunedVariableDomain   = "variables"
	tunedMainDomain       = "main"
	tunedIncludeField     = "include"
	tunedBootloaderDomain = "bootloader"
)

const (
	bootloaderRecommend = "recommend: bootloader is recommended to set manually."
	notSupportRecommend = "it is not supported, skipping."
)

var skipParamDict = map[string]bool{
	"channels":       true,
	"isolated_cores": true,
	"ps_blacklist":   true,
}

var expectedRegx = map[string]string{
	"thunderx_cpuinfo_regex": "CPU part\\s+:\\s+(0x0?516)|(0x0?af)|(0x0?a[0-3])|(0x0?b8)\\b",
	"amd_cpuinfo_regex":      "model name\\s+:.*\\bAMD\\b",
}

const defVarReg = "\\$\\{(.*?)\\}"

func convertDomain(domain string) string {
	matchedDomain, find := specialDomain[domain]
	if find {
		return matchedDomain
	}

	return domain
}

func replaceVariables(variableMap map[string]string, line string) string {
	reg := regexp.MustCompile(defVarReg)
	if reg == nil {
		return line
	}

	variables := reg.FindAllString(line, -1)
	for _, originVar := range variables {
		varName := strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(originVar), "${"), "}")
		value := variableMap[varName]
		line = strings.ReplaceAll(line, originVar, value)
	}

	return line
}

func isExpectedRegx(origin string) bool {
	variablesRegex := "([0-9A-Za-z_]+)\\s*=\\s*\\$\\{(.*_regex|.*_REGEX)\\}"
	reg := regexp.MustCompile(variablesRegex)
	if reg == nil {
		return false
	}

	return reg.MatchString(origin)
}

func getCondResultWithVar(express string) bool {
	switch {
	case strings.Contains(express, "|"):
		return calculateByLogic(express, "or")
	case strings.Contains(express, "&"):
		return calculateByLogic(express, "and")
	default:
		return calculateSingleCond(express)
	}

}

func calculateSingleCond(express string) bool {
	expParts := strings.Split(express, "=")
	if len(expParts) != 2 {
		return false
	}

	matchReg := expectedRegx[strings.TrimSpace(expParts[1])]
	reg := regexp.MustCompile(matchReg)
	if reg == nil {
		return false
	}

	return reg.MatchString(strings.TrimSpace(expParts[0]))
}

func calculateByLogic(express string, logic string) bool {
	var logicResult bool
	var variableExps []string
	var subExps []string
	logicString := "|"
	if logic == "and" {
		logicString = "&"
		logicResult = true
	}

	exps := strings.Split(express, logicString)
	for _, exp := range exps {
		if isExpectedRegx(exp) {
			variableExps = append(variableExps, exp)
			continue
		}

		subExps = append(subExps, exp)
	}

	logicResult = utils.CalculateCondExp(strings.Join(subExps, logicString))

	for _, variableExp := range variableExps {
		expParts := strings.Split(variableExp, "=")
		if len(expParts) != 2 {
			continue
		}

		variableName := strings.TrimSuffix(strings.TrimPrefix(expParts[1], "${"), "}")
		matchReg := expectedRegx[strings.TrimSpace(variableName)]
		reg := regexp.MustCompile(matchReg)
		if reg == nil {
			continue
		}

		if logic == "and" {
			logicResult = logicResult && reg.MatchString(strings.TrimSpace(expParts[0]))
			continue
		}

		logicResult = logicResult || reg.MatchString(strings.TrimSpace(expParts[0]))
	}

	return logicResult
}


