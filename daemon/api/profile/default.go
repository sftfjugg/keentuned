package profile

import (
	"fmt"
	com "keentune/daemon/api/common"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	m "keentune/daemon/modules"
	"regexp"
	"strings"
	"sync"
	"time"
)

type setVars struct {
	ip     string
	wg     *sync.WaitGroup
	target config.Group
	param  config.DBLMap
}

var (
	colorWarn = utils.ColorString("yellow", "[Warning]")
)

func SetDefault() () {
	defaultConf := config.GetProfileHomePath("default.conf")
	_, param, err := m.ConvertConfFileToJson(defaultConf)
	if err != nil {
		log.Errorf("", "read default conf err %v\n", err)
		return
	}

	var wg sync.WaitGroup
	for _, tg := range config.KeenTune.Group {
		for _, ip := range tg.IPs {
			wg.Add(1)
			setter := setVars{
				ip:     ip,
				target: tg,
				param:  param,
				wg:     &wg,
			}

			go setConfigure(setter)
		}
	}

	wg.Wait()
}

func setConfigure(setter setVars) {
	defer setter.wg.Done()

	for {
		domains, err := com.GetAVLDomain(setter.ip, setter.target.Port)
		if err == nil && len(domains) > 0 {
			break
		}
		time.Sleep(1 * time.Second)
	}

	var tuner = &m.Tuner{}

	recConf, find := getSuitableConf(setter, tuner)
	if !find {
		return
	}

	tuner.Group = []m.Group{
		{IPs: []string{setter.ip}, Port: setter.target.Port},
	}

	recommend, result, err := tuner.SetDefault()
	if recommend != "" {
		recPrefix := fmt.Sprintf("[+] Default Set '%v' of '%v' recommendations:", recConf, setter.ip)
		recPrefix = utils.ColorString("green", recPrefix)
		setRecommend := fmt.Sprintf("\n%v\n%v", recPrefix, recommend)
		fmt.Print(setRecommend)
		log.Info("", setRecommend)
	}

	if err != nil {
		log.Errorf("", "'%v' set default '%v' err %v", setter.ip, recConf, err)
		return
	}

	resultPrefix := fmt.Sprintf("[+] Default Set '%v' of '%v' result:", recConf, setter.ip)

	resultPrefix = utils.ColorString("green", resultPrefix)
	setResult := fmt.Sprintf("\n%v\n%v", resultPrefix, result)

	fmt.Print(setResult)
	log.Info("", setResult)
}

func getSuitableConf(setter setVars, tuner *m.Tuner) (string, bool) {
	// First priority: find the last active conf
	activeConf := getActiveConf(setter.target.GroupNo, setter.ip)
	if activeConf != "" {
		fileName := config.GetProfileHomePath(activeConf)
		tuner.Setter.ConfFile = []string{fileName}
		return activeConf, true
	}

	// Check the default configuration switch
	if !config.KeenTune.DefaultSet {
		skipInfo := fmt.Sprintf("Skip default setting for '%v', due to active profile not found and 'DEFAULT_AUTO_SET' is false in keentuned.conf.", setter.ip)
		fmt.Printf("\n[%v] %v\n", utils.ColorString("yellow", "Warning"), skipInfo)
		log.Warn("", skipInfo)
		return "", false
	}

	host := fmt.Sprintf("%v:%v", setter.ip, setter.target.Port)
	envConds, err := m.GetEnvCondition(setter.param, host)
	if err != nil {
		log.Errorf("", "host '%v' get environment condition err %v", setter.ip, err)
		return "", false
	}

	var recConf string
	for recommendConf, compares := range setter.param {
		match := true
		for name, regulation := range compares {
			rule := fmt.Sprint(regulation.(map[string]interface{})["value"])
			res, _ := regexp.MatchString(rule, envConds[name])
			match = match && res
		}

		if match {
			recConf = recommendConf
			recommendConf = fmt.Sprintf("%v.conf", recommendConf)
			fileName := config.GetProfileHomePath(recommendConf)
			tuner.Setter.ConfFile = []string{fileName}
			break
		}
	}

	if len(tuner.Setter.ConfFile) != 1 {
		fmt.Printf("%v No recommended configuration found for '%v'\n", colorWarn, setter.ip)
		log.Warnf("", "No recommended configuration found for '%v'", setter.ip)
		return "", false
	}

	return recConf, true
}

func getActiveConf(groupNo int, ip string) string {
	var fileName string
	groupNoStr := fmt.Sprint(groupNo)
	activeFileName := config.GetProfileWorkPath("active.conf")
	records, _ := file.GetAllRecords(activeFileName)
	for _, record := range records {
		if len(record) == 2 {
			ids := strings.Split(record[1], " ")
			for _, idInfo := range ids {
				if idInfo == ip {
					fileName = record[0]
					return fileName
				}

				if strings.TrimPrefix(idInfo, "group") == groupNoStr {
					fileName = record[0]
					return fileName
				}
			}
		}
	}

	return fileName
}


