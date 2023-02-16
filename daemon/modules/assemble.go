/*
It is mainly used to assemble and transform the data used for restful request or response with other components.
*/
package modules

import (
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	"strconv"
	"strings"
)

// Group ...
type Group struct {
	IPs            []string
	Params         []config.DBLMap // priority params for set configure
	Port           string
	ReadOnly       bool
	Dump           Configuration
	MergedParam    map[string]interface{}
	AllowUpdate    map[string]bool // prevent map concurrency security problems
	GroupName      string          // target-group-x
	GroupNo        int             // No. x of target-group-x
	ParamTotal     int
	ProfileSetFlag bool
	UnAVLParams    map[string]map[string]string // un available params
}

const brainNameParts = 2

const (
	groupIDPrefix = "group-"
)

func (tuner *Tuner) initParams() error {
	var target *Group
	var err error
	tuner.BrainParam = []Parameter{}
	for index, group := range config.KeenTune.Group {
		target, err = getInitParam(index+1, group.ParamMap, &tuner.BrainParam)
		if err != nil {
			return err
		}

		target.IPs = group.IPs
		target.Port = group.Port
		target.GroupName = group.GroupName
		target.GroupNo = group.GroupNo
		target.mergeParam()

		var updateIP = make(map[string]bool)
		for i := 0; i < len(target.IPs); i++ {
			if i == 0 {
				updateIP[target.IPs[i]] = true
				continue
			}
			updateIP[target.IPs[i]] = false
		}

		target.AllowUpdate = updateIP
		tuner.Group = append(tuner.Group, *target)
	}

	if len(tuner.Group) == 0 {
		return fmt.Errorf("found group is null")
	}

	return nil
}

func getInitParam(groupID int, paramMaps []config.DBLMap, brainParam *[]Parameter) (*Group, error) {
	var target = new(Group)
	var params = make([]config.DBLMap, len(paramMaps))

	var initConf Configuration
	for index, paramMap := range paramMaps {
		if paramMap != nil {
			params[index] = make(config.DBLMap)
		}

		for domain, parameters := range paramMap {
			var temp = make(map[string]interface{})
			for name, value := range parameters {
				origin, ok := value.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("assert %v to parameter failed", value)
				}

				param := deepCopy(origin)

				var nameSaltedParam, originParam Parameter
				if err := utils.Map2Struct(param, &nameSaltedParam); err != nil {
					return nil, fmt.Errorf("map to struct: %v", err)
				}

				paramSuffix := fmt.Sprintf("@%v%v", groupIDPrefix, groupID)
				nameSaltedParam.ParaName = fmt.Sprintf("%v%v", name, paramSuffix)
				nameSaltedParam.DomainName = domain

				if err := detectParam(&nameSaltedParam); err != nil {
					return nil, fmt.Errorf("detect macro defination param:%v", err)
				}

				originParam = nameSaltedParam
				originParam.ParaName = name
				*brainParam = append(*brainParam, nameSaltedParam)
				initConf.Parameters = append(initConf.Parameters, originParam)
				delete(param, "options")
				delete(param, "range")
				delete(param, "step")
				delete(param, "desc")
				temp[name] = param
			}

			params[index][domain] = temp
		}
	}

	target.Params = params
	target.Dump = initConf

	return target, nil
}

func deepCopy(origin interface{}) map[string]interface{} {
	if origin == nil {
		return nil
	}

	var newMap = make(map[string]interface{})
	copyMap, ok := origin.(map[string]interface{})
	if ok {
		for key, value := range copyMap {
			newMap[key] = value
		}

		return newMap
	}

	copyDBLMap, ok := origin.(config.DBLMap)
	if ok {
		for key, value := range copyDBLMap {
			newMap[key] = value
		}

		return newMap
	}

	return newMap
}

// getBrainInitParams get request parameters for brain init
func (tuner *Tuner) getBrainInitParams() error {
	for i := range tuner.BrainParam {
		name, groupID, err := parseBrainName(tuner.BrainParam[i].ParaName)
		if err != nil {
			return err
		}

		tuner.BrainParam[i].Base, err = tuner.Group[groupID].getBase(tuner.BrainParam[i].DomainName, name)
		if err != nil {
			return fmt.Errorf("get base for brain init: %v", err)
		}
	}

	return nil
}

func parseBrainName(originName string) (name string, groupIndex int, err error) {
	names := strings.Split(originName, "@")
	if len(names) < brainNameParts {
		return "", 0, fmt.Errorf("brain param name %v part length is not correct", originName)
	}

	name = names[0]

	groupIDStr := strings.TrimPrefix(names[1], groupIDPrefix)
	groupID, err := strconv.Atoi(groupIDStr)
	groupIndex = groupID - 1
	if groupIndex < 0 || groupIndex >= len(config.KeenTune.Group) {
		return "", 0, fmt.Errorf("parse brain name groupIndex %v %v", groupIDStr, err)
	}

	return name, groupIndex, nil
}

func (gp *Group) getBase(domain string, name string) (interface{}, error) {
	index := config.PriorityList[domain]
	if index < 0 || index >= config.PRILevel {
		return nil, fmt.Errorf("param priority index %v is out of range [0, 1]", index)
	}

	param, ok := gp.Params[index][domain][name]
	if !ok {
		return nil, fmt.Errorf("%v not found in %vth param", name, index)
	}

	return utils.ParseKey("value", param)
}

// parseAcquireParam parse acquire response value for apply request
func (tuner *Tuner) parseAcquireParam(resp ReceivedConfigure) error {
	for _, param := range resp.Candidate {
		paramName, groupID, err := parseBrainName(param.ParaName)
		if err != nil {
			return err
		}

		param.ParaName = paramName
		if err := tuner.Group[groupID].updateValue(param); err != nil {
			return fmt.Errorf("update %v value %v", paramName, err)
		}
	}

	for index := range tuner.Group {
		tuner.Group[index].Dump.Round = resp.Iteration
		tuner.Group[index].Dump.budget = resp.Budget
	}

	paramPath := fmt.Sprintf("%v/parameters_value.csv", config.GetTuningPath(tuner.Name))
	err := file.Append(paramPath, strings.Split(resp.ParamValue, ","))
	if err != nil {
		log.Errorf(tuner.logName, "%vth iteration save parameters_value failed: %v", tuner.Iteration, err)
	}

	return nil
}

// parseBestParam parse best response value for best dump
func (tuner *Tuner) parseBestParam() error {
	var bestParams = make([][]Parameter, len(tuner.Group))
	for _, param := range tuner.bestInfo.Parameters {
		paramName, groupID, err := parseBrainName(param.ParaName)
		if err != nil {
			return err
		}

		param.ParaName = paramName
		bestParams[groupID] = append(bestParams[groupID], param)
	}

	for index := range tuner.Group {
		tuner.Group[index].Dump.Round = tuner.bestInfo.Round
		tuner.Group[index].Dump.Score = tuner.bestInfo.Score
		tuner.Group[index].Dump.Parameters = bestParams[index]
		for _, parameter := range bestParams[index] {
			err := tuner.Group[index].updateValue(parameter)
			if err != nil {
				return fmt.Errorf("update best param %v", err)
			}
		}
	}

	return nil
}

// updateParams update param values by apply result
func (gp *Group) updateParams(params map[string]Parameter) error {
	for name, param := range params {
		param.ParaName = name
		err := gp.updateValue(param)
		if err != nil {
			return err
		}
	}

	return nil
}

func (gp *Group) updateValue(param Parameter) error {
	index, ok := config.PriorityList[param.DomainName]
	if !ok {
		return fmt.Errorf("add '%v' priority white list first", param.DomainName)
	}

	if index < 0 || index >= config.PRILevel {
		return fmt.Errorf("priority id %v is out of range [0, 1]", index)
	}
	name := param.ParaName
	value, ok := gp.Params[index][param.DomainName][name]
	if !ok {
		return fmt.Errorf("%v not found in %vth param", name, index)
	}

	detail, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("assert %v to parameter failed", param)
	}

	detail["value"] = param.Value

	gp.Params[index][param.DomainName][name] = detail
	return nil
}

func (gp *Group) mergeParam() {
	gp.MergedParam = make(map[string]interface{})
	for _, paramMaps := range gp.Params {
		for domain, paramMap := range paramMaps {
			gp.MergedParam[domain] = deepCopy(paramMap)
			gp.ParamTotal += len(paramMap)
		}
	}
}

func (gp *Group) applyReq(ip string, params interface{}, id ...int) map[string]interface{} {
	retRequest := map[string]interface{}{}
	retRequest["data"] = params
	retRequest["resp_ip"] = config.RealLocalIP
	retRequest["resp_port"] = config.KeenTune.Port
	if len(id) > 0 {
		retRequest["target_id"] = id[0]
	} else {
		retRequest["target_id"] = config.KeenTune.IPMap[ip]
	}

	retRequest["readonly"] = gp.ReadOnly
	return retRequest
}

func (gp *Group) updateDump(param map[string]Parameter) {
	for i := range gp.Dump.Parameters {
		name := gp.Dump.Parameters[i].ParaName
		domain := gp.Dump.Parameters[i].DomainName
		info, ok := param[name]
		if ok && domain == info.DomainName {
			gp.Dump.Parameters[i].Value = info.Value
		}
	}
}

func (tuner *Tuner) initProfiles() error {
	for groupIdx, group := range config.KeenTune.Group {
		isSetting := tuner.Setter.Group[groupIdx]
		if !isSetting {
			continue
		}

		var target = new(Group)
		target.IPs = group.IPs
		confFile := tuner.Setter.ConfFile[groupIdx]
		abnormal, err := target.getConfigParam(confFile)
		if !strings.Contains(tuner.recommend, abnormal.Recommend) {
			tuner.recommend += abnormal.Recommend
		}

		if abnormal.Warning != "" {
			tuner.initWarning += abnormal.Warning
		}

		if err != nil {
			return err
		}

		target.Port = group.Port
		target.GroupName = group.GroupName
		target.GroupNo = group.GroupNo
		tuner.Group = append(tuner.Group, *target)
	}

	if len(tuner.Group) == 0 {
		return fmt.Errorf("found group is null")
	}

	return nil
}

func (gp *Group) getConfigParam(fileName string) (ABNLResult, error) {
	abnormal, resultMap, err := ConvertConfFileToJson(fileName, gp.IPs[0])
	if err != nil {
		return abnormal, fmt.Errorf("convert file '%v' %v", fileName, err)
	}

	if len(resultMap) == 0 {
		return abnormal, fmt.Errorf("No valid domain can be used. Please check and set valid configurations in %v", fileName)
	}

	gp.Params, err = config.GetPriorityParams(resultMap)
	if err != nil {
		return abnormal, err
	}

	gp.mergeParam()
	return abnormal, nil
}

func (gp *Group) deleteUnAVLConf(unAVLParams []map[string]map[string]string) (string, int) {
	var totalWarningInfo = new(string)
	var isAllUnAVL bool
	var retWarnInfo []string
	var domains []string

	gp.UnAVLParams = make(map[string]map[string]string)

	for _, params := range unAVLParams {
		var unAVLCount int
		for domain, kv := range params {
			domains = append(domains, domain)
			_, exist := gp.UnAVLParams[domain]
			if !exist {
				gp.UnAVLParams[domain] = make(map[string]string)
			}

			if params[domain] == nil {
				// domain is all available, skip
				continue
			}

			notMetInfo, ok := params[domain][domain]
			if ok {
				// domain is all unavailable
				unAVLCount += len(gp.MergedParam[domain].(map[string]interface{}))
				for priorityIdx := range gp.Params {
					delete(gp.Params[priorityIdx], domain)
				}

				decoratedWarnInfo := fmt.Sprintf("%v%v", notMetInfo, multiRecordSeparator)
				retWarnInfo = append(retWarnInfo, decoratedWarnInfo)
				isAllUnAVL = isAllUnAVL || unAVLCount == gp.ParamTotal
				continue
			}

			var oneDomainWarning string
			unAVLCount += len(kv)

			oneDomainWarning = gp.tidyUnavailableParams(kv, domain, totalWarningInfo)

			if len(oneDomainWarning) > 0 {
				*totalWarningInfo += oneDomainWarning
				retWarnInfo = append(retWarnInfo, oneDomainWarning)
			}

			isAllUnAVL = isAllUnAVL || unAVLCount == gp.ParamTotal
		}
	}

	if isAllUnAVL {
		return strings.Join(retWarnInfo, ""), FAILED
	}

	if len(retWarnInfo) > 0 {
		return strings.Join(retWarnInfo, ""), WARNING
	}

	return "", SUCCESS
}

// tidyUnavailableParams delete unavailable parameter domain from group.Params and cache them to UnAVLParams
func (gp *Group) tidyUnavailableParams(kv map[string]string, domain string, warningInfo *string) string {
	var oneDomainWarning string
	for name, msg := range kv {
		// cache unavailable params to group
		gp.UnAVLParams[domain][name] = msg
		for priorityIdx := range gp.Params {
			_, exists := gp.Params[priorityIdx][domain][name]
			if exists {
				// delete unavailable configure params
				delete(gp.Params[priorityIdx][domain], name)

				tmpWarn := fmt.Sprintf("[%v] %v: %v%v", domain, name, msg, multiRecordSeparator)
				// discard repeated warning
				if !strings.Contains(*warningInfo, tmpWarn) {
					oneDomainWarning += tmpWarn
				}

				if len(gp.Params[priorityIdx][domain]) == 0 {
					delete(gp.Params[priorityIdx], domain)
				}
				continue
			}
		}
	}

	return oneDomainWarning
}


