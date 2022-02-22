/*
It is mainly used to assemble and transform the data used for restful request or response with other components.
*/
package modules

import (
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/utils"
	"strconv"
	"strings"
)

// Group ...
type Group struct {
	IPs         []string
	Params      []config.DBLMap
	Port        string
	ReadOnly    bool
	Dump        Configuration
	MergedParam map[string]interface{}
	AllowUpdate map[string]bool // prevent map concurrency security problems
}

const brainNameParts = 3

const (
	groupIDPrefix    = "group-"
	priorityIDPrefix = "pri-"
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
	for i := range params {
		params[i] = make(config.DBLMap)
	}

	var nameSaltedParam, originParam Parameter
	var initConf Configuration
	for index, paramMap := range paramMaps {
		for domain, parameters := range paramMap {
			var temp = make(map[string]interface{})
			for name, value := range parameters {
				param, ok := value.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("assert %v to parameter failed", value)
				}

				if err := utils.Map2Struct(value, &nameSaltedParam); err != nil {
					return nil, fmt.Errorf("map to struct: %v", err)
				}

				priID, ok := config.PriorityList[domain]
				if !ok {
					priID = 1
				}
				paramSuffix := fmt.Sprintf("@%v%v@%v%v", groupIDPrefix, groupID, priorityIDPrefix, priID)
				nameSaltedParam.ParaName = fmt.Sprintf("%v%v", name, paramSuffix)
				nameSaltedParam.DomainName = domain

				originParam = nameSaltedParam
				originParam.ParaName = name
				*brainParam = append(*brainParam, nameSaltedParam)
				initConf.Parameters = append(initConf.Parameters, originParam)
				delete(param, "options")
				delete(param, "range")
				delete(param, "step")
				temp[name] = param
			}

			params[index][domain] = temp
		}
	}

	target.Params = params
	target.Dump = initConf

	return target, nil
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
	index := config.PriorityList[param.DomainName]
	if index < 0 || index >= config.PRILevel {
		return fmt.Errorf("ip %v priority id %v is out of range [0, 1]", index)
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
			gp.MergedParam[domain] = paramMap
		}
	}
}

func (gp *Group) applyReq(ip string, params interface{}) map[string]interface{} {
	retRequest := map[string]interface{}{}
	retRequest["data"] = params
	retRequest["resp_ip"] = config.RealLocalIP
	retRequest["resp_port"] = gp.Port
	retRequest["target_id"] = config.KeenTune.IPMap[ip]
	retRequest["readonly"] = gp.ReadOnly
	return retRequest
}

func (gp *Group) Get(ip string, ipIndex int) error {
	gp.ReadOnly = true
	return gp.Configure(ip, ipIndex, gp.applyReq(ip, gp.MergedParam))
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

