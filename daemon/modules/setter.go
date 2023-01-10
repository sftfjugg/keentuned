package modules

import (
	"fmt"
	"io/ioutil"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	"os"
	"strings"
)

// Setter ...
type Setter struct {
	Group       []bool
	ConfFile    []string
	recommend   string
	initWarning string
	prefixReco  string // prefix for recommendation
}

// Set profile set  main process
func (tuner *Tuner) Set() {
	var err error
	tuner.logName = log.ProfSet
	err = tuner.initProfiles()
	// ps: must show recommendation before check err
	tuner.showReco()
	if err != nil {
		tuner.showPrefixReco()
		log.Error(log.ProfSet, err)
		return
	}

	defer func() {
		if err != nil {
			tuner.rollback()
		}
	}()

	if err = tuner.prepareBeforeSet(); err != nil {
		tuner.showPrefixReco()
		log.Errorf(log.ProfSet, err.Error())
		return
	}

	groupSetResult := fmt.Sprintf("%v\n", utils.ColorString("green", "[+] Profile Result (Auto Settings)"))
	if len(log.ClientLogMap[log.ProfSet]) > 0 {
		groupSetResult = fmt.Sprintf("\n%v", groupSetResult)
	}

	log.Info(log.ProfSet, groupSetResult)

	err = tuner.setConfigure()
	if err != nil {
		errMsg := strings.Replace(err.Error(), applyFailureResult, "apply all failed", 1)
		log.Error(log.ProfSet, errMsg)
		return
	}

	err = tuner.updateActive()
	if err != nil {
		return
	}

	log.Info(log.ProfSet, tuner.applySummary)

	return
}

// showReco show recommendation
func (tuner *Tuner) showReco() {
	tuner.prefixReco = fmt.Sprintf("%v\n", utils.ColorString("green", "[+] Recommendation (Manual Settings)"))

	if len(tuner.recommend) > 0 {
		fmtStr := fmt.Sprintf("%v%v\n", tuner.prefixReco, tuner.recommend)
		log.Info(log.ProfSet, fmtStr)
	}

	if len(tuner.initWarning) > 0 {
		tuner.showPrefixReco()

		for _, preWarning := range strings.Split(tuner.initWarning, multiRecordSeparator) {
			pureInfo := strings.TrimSpace(preWarning)
			if len(pureInfo) > 0 {
				log.Warn(log.ProfSet, preWarning)
			}
		}
	}
}

// showPrefixReco show prefix for recommendation log
func (tuner *Tuner) showPrefixReco() {
	if len(log.ClientLogMap[log.ProfSet]) == 0 {
		log.Info(tuner.logName, tuner.prefixReco)
	}
}

func (tuner *Tuner) updateActive() error {
	activeFile := config.GetProfileWorkPath("active.conf")
	// 先拼接，再写入
	var fileSet = fmt.Sprintln("name,group_info")
	var activeInfo = make(map[string][]string)
	for groupIndex, settable := range tuner.Setter.Group {
		if settable {
			fileName := file.GetPlainName(tuner.Setter.ConfFile[groupIndex])
			activeInfo[fileName] = append(activeInfo[fileName], fmt.Sprintf("group%v", groupIndex+1))
		}
	}

	for name, info := range activeInfo {
		fileSet += fmt.Sprintf("%s,%s\n", name, strings.Join(info, " "))
	}

	if err := UpdateActiveFile(activeFile, []byte(fileSet)); err != nil {
		log.Errorf(log.ProfSet, "Update active file err:%v", err)
		return fmt.Errorf("update active file err %v", err)
	}

	return nil
}

func (tuner *Tuner) prepareBeforeSet() error {
	// step1. rollback the target machine
	err := tuner.rollback()
	if err != nil {
		return fmt.Errorf("rollback failed:\n%v", tuner.rollbackFailure)
	}

	// step2. clear the active file
	fileName := config.GetProfileWorkPath("active.conf")
	if err = UpdateActiveFile(fileName, []byte{}); err != nil {
		return fmt.Errorf("update active file failed, err:%v", err)
	}

	// step3. backup the target machine
	err = tuner.backup()
	if tuner.backupWarning != "" {
		tuner.showPrefixReco()
		for _, backupWarning := range strings.Split(tuner.backupWarning, multiRecordSeparator) {
			pureInfo := strings.TrimSpace(backupWarning)
			if len(pureInfo) > 0 {
				log.Warn(tuner.logName, backupWarning)
			}
		}
	}

	if err != nil {
		return fmt.Errorf("%v", tuner.backupFailure)
	}
	return nil
}

// UpdateActiveFile ...
func UpdateActiveFile(fileName string, info []byte) error {
	if err := ioutil.WriteFile(fileName, info, os.ModePerm); err != nil {
		return err
	}

	return nil
}

// SetDefault set default configuration
func (tuner *Tuner) SetDefault() (string, string, error) {
	var recommend string
	ip := tuner.Group[0].IPs[0]
	abn, param, err := ConvertConfFileToJson(tuner.ConfFile[0], ip)
	colorWarn := utils.ColorString("yellow", "[Warning]")
	if abn.Recommend != "" {
		recs := strings.Split(abn.Recommend, multiRecordSeparator)
		for _, rec := range recs {
			if strings.TrimSpace(rec) != "" {
				recommend += fmt.Sprintf("\t%v %v\n", colorWarn, rec)
			}
		}
	}

	if abn.Warning != "" {
		warns := strings.Split(abn.Warning, multiRecordSeparator)
		for _, warn := range warns {
			if strings.TrimSpace(warn) != "" {
				recommend += fmt.Sprintf("\t%v %v\n", colorWarn, warn)
			}
		}
	}

	if err != nil {
		return recommend, "", err
	}

	port := tuner.Group[0].Port
	ipIndex := config.KeenTune.IPMap[ip] * 2
	host := fmt.Sprintf("%v:%v", ip, port)
	tuner.updateGroup(param)

	err = tuner.prepareBeforeSet()
	if tuner.backupWarning != "" {
		for _, backupWarning := range strings.Split(tuner.backupWarning, multiRecordSeparator) {
			if strings.TrimSpace(backupWarning) != "" {
				recommend += fmt.Sprintf("\t%v %v\n", colorWarn, backupWarning)
			}
		}
	}

	if err != nil {
		return recommend, "", err
	}

	reqParam := tuner.Group[0].Params[0]
	reqBody := tuner.Group[0].applyReq(ip, reqParam, ipIndex)
	ret, err := Configure(reqBody, host, ipIndex)

	if err != nil {
		return recommend, ret, err
	}

	activeFile := config.GetProfileWorkPath("active.conf")
	var fileSet = fmt.Sprintln("name,group_info")
	name := file.GetPlainName(tuner.Setter.ConfFile[0])
	fileSet += fmt.Sprintf("%s,%s\n", name, ip)
	UpdateActiveFile(activeFile, []byte(fileSet))
	return recommend, ret, nil
}

func (tuner *Tuner) updateGroup(param map[string]map[string]interface{}) {
	gp := new(Group)
	gp.ReadOnly = false
	gp.Params = []config.DBLMap{param}
	gp.mergeParam()
	gp.IPs = tuner.Group[0].IPs
	gp.Port = tuner.Group[0].Port
	tuner.Group[0] = *gp
}


