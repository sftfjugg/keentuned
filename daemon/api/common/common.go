/*
Copyright © 2021 KeenTune

Package common for daemon, this package contains the common, handle, heartbeat for common function. The variables required for the dynamic and static tuning function are defined, including state detection of the keentuned service and heartbeat packet detection for the other three components.
*/
package common

import (
	"encoding/json"
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	utilhttp "keentune/daemon/common/utils/http"
	m "keentune/daemon/modules"
	"net/http"
	"os"
	"strings"
	"syscall"
)

type deleter struct {
	cmd      string
	fileName string
}

// DeleteFlag ...
type DeleteFlag struct {
	Name  string
	Cmd   string
	Force bool
}

// RollbackFlag ...
type RollbackFlag struct {
	Cmd string
}

// DumpFlag ...
type DumpFlag struct {
	Name   string
	Output []string
	Force  bool
}

var (
	logHome = "/var/log/keentune"
)

// IsDataReady ...
func IsDataReady(name string) bool {
	dataList, _, _, err := GetAVLDataAndAlgo()
	if err != nil {
		return false
	}

	for _, has := range dataList {
		if has == name {
			return true
		}
	}

	return false
}

// GetAVLDataAndAlgo get available data, algo from brain
func GetAVLDataAndAlgo() ([]string, []string, []string, error) {
	resp, err := pingAndConnect(config.KeenTune.BrainIP, config.KeenTune.BrainPort, "avaliable")
	if err != nil {
		return nil, nil, nil, err
	}

	var brainRet struct {
		Success   bool     `json:"suc"`
		Data      []string `json:"data"`
		Explainer []string `json:"explainer"`
		Tune      []string `json:"tune"`
	}

	if err = json.Unmarshal(resp, &brainRet); err != nil {
		return nil, nil, nil, err
	}

	if !brainRet.Success {
		return nil, nil, nil, fmt.Errorf("remotecall available return suc is false")
	}

	return brainRet.Data, brainRet.Tune, brainRet.Explainer, nil
}

// KeenTunedService ...
func KeenTunedService(quit chan os.Signal) {
	// register router
	registerRouter()

	go func() {
		select {
		case sig := <-quit:
			log.Info("", "keentune is interrupted")
			if m.GetRunningTask() != "" {
				ResetJob()
				utilhttp.RemoteCall("GET", config.KeenTune.BrainIP+":"+config.KeenTune.BrainPort+"/end", nil)
			}
			if sig == syscall.SIGTERM {
				os.Exit(0)
			} else {
				os.Exit(1)
			}
		}
	}()

	err := http.ListenAndServe(":"+config.KeenTune.Port, nil)
	if err != nil {
		log.Errorf("", "listen and serve err: %v", err)
		os.Exit(1)
	}
}

// RunDelete run delete file service
func RunDelete(flag DeleteFlag, reply *string) error {
	defer func() {
		*reply = log.ClientLogMap[fmt.Sprintf("%s delete", flag.Cmd)]
		log.ClearCliLog(fmt.Sprintf("%s delete", flag.Cmd))
	}()

	var fullName string
	switch flag.Cmd {
	case "param":
		fullName = GetParameterPath(flag.Name)
	case "profile":
		fullName = config.GetProfilePath(flag.Name)
	default:
		log.Errorf("", "%v is not supported", flag.Cmd)
		return nil
	}

	inst := new(deleter)
	inst.cmd = flag.Cmd
	inst.fileName = fullName

	if err := inst.check(flag.Name); err != nil {
		log.Errorf(fmt.Sprintf("%s delete", inst.cmd), err.Error())
		return fmt.Errorf("Check name failed: %v", err.Error())
	}

	if err := inst.delete(); err != nil {
		log.Errorf(fmt.Sprintf("%s delete", inst.cmd), err.Error())
		return fmt.Errorf("Delete failed: %v", err.Error())
	}

	if flag.Cmd == "param" {
		primaryKeys := []string{flag.Name}
		file.DeleteRow(config.GetDumpPath(config.TuneCsv), primaryKeys)
		os.Remove(fmt.Sprintf("%v/%v.log", logHome, flag.Name))
	}

	log.Infof(fmt.Sprintf("%s delete", inst.cmd), "[ok] %v delete successfully", flag.Name)
	return nil
}

// RunTrainDelete run training delete file service
func RunTrainDelete(flag DeleteFlag, reply *string) error {
	fullName := GetSensitizePath(flag.Name)

	inst := new(deleter)
	inst.cmd = flag.Cmd
	inst.fileName = fullName

	if err := inst.check(flag.Name); err != nil {
		log.Errorf(fmt.Sprintf("%s delete", inst.cmd), err.Error())
		return fmt.Errorf("Check name failed: %v", err.Error())
	}

	if err := inst.delete(); err != nil {
		log.Errorf(fmt.Sprintf("%s delete", inst.cmd), err.Error())
		return fmt.Errorf("Delete failed: %v", err.Error())
	}
	primaryKeys := []string{flag.Name}
	file.DeleteRow(config.GetDumpPath(config.SensitizeCsv), primaryKeys)
	os.Remove(fmt.Sprintf("%v/keentuned-sensitize-train-%v.log", logHome, flag.Name))

	log.Infof(log.SensitizeDel, "[ok] %v delete successfully", flag.Name)
	return nil
}

func (d *deleter) check(inputName string) error {
	if strings.Contains(d.fileName, config.KeenTune.Home) {
		return fmt.Errorf("%v is not supported to delete", d.fileName)
	}

	if d.cmd != "profile" {
		return nil
	}

	activeFileName := config.GetProfileWorkPath("active.conf")
	if !file.IsPathExist(activeFileName) {
		return nil
	}

	if file.HasRecord(activeFileName, "name", inputName) {
		return fmt.Errorf("%v is active profile, please run \"keentune profile rollback\" before delete", inputName)
	}

	return nil
}

func (d *deleter) delete() error {
	return os.RemoveAll(d.fileName)
}

// GetParameterPath ...
func GetParameterPath(fileName string) string {
	workPath := config.GetTuningPath(fileName)
	if file.IsPathExist(workPath) {
		return workPath
	}

	homePath := config.GetParamHomePath() + fileName
	if file.IsPathExist(homePath) {
		return homePath
	}

	generatePath := config.GetGenerateWorkPath(fmt.Sprintf("%s%s", strings.TrimSuffix(fileName, ".json"), ".json"))
	if file.IsPathExist(generatePath) {
		return generatePath
	}

	return ""
}

// GetSensitizePath ...
func GetSensitizePath(fileName string) string {
	workPath := config.GetSensitizePath(fileName)
	if file.IsPathExist(workPath) {
		return workPath
	}

	homePath := config.GetParamHomePath() + fileName
	if file.IsPathExist(homePath) {
		return homePath
	}

	generatePath := config.GetGenerateWorkPath(fmt.Sprintf("%s%s", strings.TrimSuffix(fileName, ".json"), ".json"))
	if file.IsPathExist(generatePath) {
		return generatePath
	}

	return ""
}

// IsApplying ...
func IsApplying() bool {
	job := m.GetRunningTask()
	if job == "" || len(strings.Split(job, " ")) < 2 {
		return false
	}

	return (strings.Split(job, " ")[0] == m.JobProfile) || (strings.Split(job, " ")[0] == m.JobTuning)
}

// ResetJob ...
func ResetJob() {
	m.ClearTask()

	tuningJob := file.GetRecord(config.GetDumpPath(config.TuneCsv), "status", "running", "name")
	if tuningJob != "" {
		file.UpdateRow(config.GetDumpPath(config.TuneCsv), tuningJob, map[int]interface{}{m.TuneStatusIdx: m.Kill})
	}

	sensitizeJob := file.GetRecord(config.GetDumpPath(config.SensitizeCsv), "status", "running", "name")
	if sensitizeJob != "" {
		file.UpdateRow(config.GetDumpPath(config.SensitizeCsv), sensitizeJob, map[int]interface{}{m.TrainStatusIdx: m.Kill})
	}
}

// GetAVLDomain get available domain
func GetAVLDomain(ip, port string) ([]string, error) {
	resp, err := pingAndConnect(ip, port, "avaliable")
	if err != nil {
		return nil, err
	}

	var ret struct {
		Domains []string `json:"result"`
	}

	err = json.Unmarshal(resp, &ret)
	if err != nil {
		return nil, err
	}

	return ret.Domains, nil
}

// GetStatus get status
func GetStatus(ip, port string) error {
	resp, err := pingAndConnect(ip, port, "status")
	if err != nil {
		return err
	}

	var ret struct {
		Status string `json:"status"`
	}

	err = json.Unmarshal(resp, &ret)
	if err != nil {
		return err
	}

	if ret.Status != "alive" {
		return fmt.Errorf("offline")
	}

	return nil
}

func pingAndConnect(ip, port, uri string, request ...interface{}) ([]byte, error) {
	err := utils.Ping(ip, port)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%v:%v/%v", ip, port, uri)
	if request == nil {
		return utilhttp.RemoteCall("GET", url, nil)
	}

	return utilhttp.RemoteCall("POST", url, request[0])
}

// GetAVLAgentAddr ...
// return:
//       0: benchmark host available? true:false;
//       1: agent ip reachable? true:false;
//       2: err msg
func GetAVLAgentAddr(ip, port, agent string) (bool, bool, error) {
	request := map[string]interface{}{
		"agent_address": agent,
	}

	resp, err := pingAndConnect(ip, port, "avaliable", request)
	if err != nil {
		return false, false, fmt.Errorf("\tbench source %v offline\n", ip)
	}

	var ret map[string]bool

	err = json.Unmarshal(resp, &ret)
	if err != nil {
		return false, false, err
	}

	if ret[agent] {
		return true, true, nil
	}

	return true, false, fmt.Errorf("\tbench destination %v unreachable\n", agent)
}


