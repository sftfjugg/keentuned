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
		time.Sleep(3 * time.Second)
		domains, err := com.GetAVLDomain(setter.ip, setter.target.Port)
		if err == nil && len(domains) > 0 {
			break
		}
	}

	host := fmt.Sprintf("%v:%v", setter.ip, setter.target.Port)
	envConds, err := m.GetEnvCondition(setter.param, host)
	if err != nil {
		log.Errorf("", "host '%v' get environment condition err %v", setter.ip, err)
		return
	}

	var tuner = &m.Tuner{}
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
		return
	}

	tuner.Group = []m.Group{
		{IPs: []string{setter.ip}, Port: setter.target.Port},
	}

	recommend, result, err := tuner.SetDefault()
	if recommend != "" {
		rec := fmt.Sprintf("[+] set '%v' recommendations:", setter.ip)
		printRec := fmt.Sprintf("%v\n%v", utils.ColorString("green", rec), recommend)
		fmt.Println(printRec)
		log.Infof("", "set '%v' recommendations:\n%v", setter.ip, recommend)
	}

	if err != nil {
		log.Errorf("", "host '%v' set default '%v' err %v", setter.ip, file.GetPlainName(tuner.Setter.ConfFile[0]), err)
		return
	}

	resultPrefix := fmt.Sprintf("[+] Set '%v'  default conf '%v' result:", setter.ip, recConf)

	resultPrefix = utils.ColorString("green", resultPrefix)

	fmt.Printf("%v\n%v\n", resultPrefix, result)
	log.Infof("", "%v\n%v", resultPrefix, result)
}


