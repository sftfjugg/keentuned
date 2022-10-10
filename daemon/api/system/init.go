package system

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	com "keentune/daemon/api/common"
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	"strings"
)

type ymlTarget struct {
	IP        string   `yaml:"ip"`
	Available bool     `yaml:"available"`
	Knobs     []string `yaml:"knobs"`
	Domain    []string `yaml:"domain"`
}

type ymlBrain struct {
	BrainIP  string   `yaml:"ip"`
	AlgoTune []string `yaml:"algo_tuning"`
	AlgoSen  []string `yaml:"algo_sensi"`
}

type ymlBench struct {
	IP        string  `yaml:"ip"`
	Available bool    `yaml:"available"`
	Dest      ymlDest `yaml:"destination"`
	BenchConf string  `yaml:"benchmark"`
}

type ymlDest struct {
	IP        string `yaml:"ip"`
	Reachable bool   `yaml:"reachable"`
}

type keenTuneYML struct {
	Bench []ymlBench
	Brain ymlBrain
}

// Init ...
func (s *Service) Init(flag string, reply *string) error {
	result, err := initialize()
	if err != nil {
		log.Errorf("", "keentune init failed: %v", err)
		return err
	}

	if result != "" {
		*reply = fmt.Sprintf("%v  Connection failure Details:\n%v\n", utils.ColorString("yellow", "[Warning]"), result)
		log.Warnf("", "keentune init: %v", result)
		return nil
	}

	*reply = fmt.Sprintf("%v KeenTune Init success\n", utils.ColorString("green", "[OK]"))
	log.Info("", "KeenTune Init success")
	return nil
}

// initialize  KeenTune available test between brain, bench, target and daemon; Init Yaml create or update.
func initialize() (string, error) {
	err := config.CheckAndReloadConf()
	if err != nil {
		return "", err
	}

	var ymlConf = &keenTuneYML{}
	var warningDetail = new(string)
	checkBenchAVL(ymlConf, warningDetail)

	checkBrainAVL(ymlConf, warningDetail)

	targetGroup := checkTargetAVL(warningDetail)

	bytes, err := yaml.Marshal(getYMLConf(ymlConf, targetGroup))
	if err != nil {
		return *warningDetail, err
	}

	err = ioutil.WriteFile(config.KeenTuneYMLFile, bytes, 0644)
	return strings.TrimSuffix(*warningDetail, "\n"), err
}

func checkBrainAVL(ymlConf *keenTuneYML, warningDetail *string) {
	ymlConf.Brain.BrainIP = config.KeenTune.BrainIP
	var err error
	_, ymlConf.Brain.AlgoTune, ymlConf.Brain.AlgoSen, err = com.GetAVLDataAndAlgo()
	if err != nil {
		*warningDetail += fmt.Sprintf("\tbrain %v offline\n", config.KeenTune.BrainIP)
	}
}

func getYMLConf(conf *keenTuneYML, group [][]ymlTarget) interface{} {
	var ret = map[string]interface{}{}
	ret["brain"] = conf.Brain
	ret["bench-group-1"] = conf.Bench

	for idx, target := range config.KeenTune.Group {
		groupName := fmt.Sprintf("target-group-%v", target.GroupNo)
		ret[groupName] = group[idx]
	}

	return ret
}

func checkTargetAVL(warningDetail *string) [][]ymlTarget {
	var err error
	targetGroup := make([][]ymlTarget, len(config.KeenTune.Group))
	for groupIdx, target := range config.KeenTune.Group {
		targetGroup[groupIdx] = make([]ymlTarget, len(target.IPs))
		for ipIdx, ip := range target.IPs {
			var tmpTarget ymlTarget
			knobs := strings.Split(target.ParamConf, ",")
			for _, knob := range knobs {
				tmpTarget.Knobs = append(tmpTarget.Knobs, strings.TrimSpace(knob))
			}

			tmpTarget.Domain, err = com.GetAVLDomain(ip, target.Port)
			tmpTarget.IP = ip
			if err != nil {
				groupInfo := fmt.Sprintf("target-group[%v]:", target.GroupNo)
				*warningDetail += fmt.Sprintf("\t%v %v offline\n", groupInfo, ip)
				targetGroup[groupIdx][ipIdx] = tmpTarget
				continue
			}

			tmpTarget.Available = true
			targetGroup[groupIdx][ipIdx] = tmpTarget
		}
	}

	return targetGroup
}

func checkBenchAVL(ymlConf *keenTuneYML, warningDetail *string) {
	ymlConf.Bench = make([]ymlBench, len(config.KeenTune.BenchIPMap))
	for _, bench := range config.KeenTune.BenchGroup {
		for _, ip := range bench.SrcIPs {
			var tmpBench ymlBench
			var err error

			tmpBench.BenchConf = bench.BenchConf

			tmpBench.Available, tmpBench.Dest.Reachable, err = com.GetAVLAgentAddr(ip, bench.SrcPort, bench.DestIP)

			tmpBench.IP = ip
			tmpBench.Dest.IP = bench.DestIP

			ymlConf.Bench[config.KeenTune.BenchIPMap[ip]-1] = tmpBench
			if err != nil {
				*warningDetail += fmt.Sprintf("%v", err)
			}
		}
	}
}

