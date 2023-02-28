package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"github.com/go-ini/ini"
)

// UpdateKeentunedConf ...
func UpdateKeentunedConf(info string) (string, error) {
	details := strings.Split(info, "\n")
	if len(details) == 0 {
		return "", fmt.Errorf("info is empty")
	}

	var mutex = &sync.RWMutex{}
	mutex.Lock()
	defer mutex.Unlock()

	var result string
	cfg, usrCfg, err := loadCfg(info)
	if err != nil {
		return "", fmt.Errorf("load conf err, %v", err)
	}

	result = setCfg(details, cfg, usrCfg)

	if strings.Contains(info, "-group-") {
		compareCfg(cfg, usrCfg)

		err = checkInitGroup(cfg)
		if err != nil {
			return "", err
		}
	}

	err = cfg.SaveTo(keentuneConfigFile)
	if err != nil {
		return result, err
	}

	if result != "" {
		result = fmt.Sprintf("[Warning] %v", result)
		return result, nil
	}

	result = "keentuned configure save success"

	return result, nil
}

// compareCfg compare with user configuration and delete useless target group from itself
func compareCfg(cfg *ini.File, usrCfg *ini.File) {
	cfgSecNames := cfg.SectionStrings()
	for _, sec := range cfgSecNames {
		// Delete useless group sections before
		if strings.Contains(sec, TargetSectionPrefix) {
			userSec, err := usrCfg.GetSection(sec)
			if err != nil || userSec == nil {
				cfg.DeleteSection(sec)
				continue
			}
		}
	}
}

func loadCfg(info string) (*ini.File, *ini.File, error) {
	cfg, err := ini.Load(keentuneConfigFile)
	if err != nil || cfg == nil {
		return nil, nil, fmt.Errorf("load %s, %v", "keentuned.conf", err)
	}

	tempPath := fmt.Sprintf("%v/temp.conf", KeenTune.DumpHome)
	ioutil.WriteFile(tempPath, []byte(info), 0644)
	defer os.Remove(tempPath)
	usrCfg, err := ini.Load(tempPath)
	if err != nil || usrCfg == nil {
		return nil, nil, fmt.Errorf("parse request to conf: %v", err)
	}

	return cfg, usrCfg, nil
}

func setCfg(details []string, cfg *ini.File, usrCfg *ini.File) string {
	var domain, result string
	var benchGroupCount int
	var isNeedRemove bool
	for _, line := range details {
		pureLine := strings.TrimSpace(line)
		if len(pureLine) == 0 {
			continue
		}

		if strings.Contains(pureLine, "[") {
			if isNeedRemove {
				isNeedRemove = false
			}

			domain = strings.Trim(strings.Trim(strings.TrimSpace(line), "]"), "[")
			// Currently, only supported one bench group, redundant will be deleted
			if strings.Contains(domain, BenchSectionPrefix) {
				benchGroupCount++
				if benchGroupCount >= 2 {
					result += fmt.Sprintf("[%v] is removed, only one bench group is supported.\n", domain)
					usrCfg.DeleteSection(domain)
					isNeedRemove = true
				}
			}

			continue
		}

		// if isNeedRemove is true, it will be set to false, until the next domain.
		if isNeedRemove {
			continue
		}

		kvs := strings.Split(pureLine, "=")
		if len(kvs) != 2 {
			result += fmt.Sprintln(pureLine)
			continue
		}

		value := strings.TrimSpace(strings.Trim(kvs[1], "\""))
		cfg.Section(domain).Key(strings.TrimSpace(strings.ToUpper(kvs[0]))).SetValue(value)

	}

	return result
}

func checkInitGroup(cfg *ini.File) error {
	kdConf := new(KeentunedConf)
	if err := kdConf.getTargetGroup(cfg); err != nil {
		return err
	}

	if err := kdConf.getBenchGroup(cfg); err != nil {
		return err
	}

	return nil
}

// ReloadConf ...
func ReloadConf() error {
	var mutex = &sync.RWMutex{}
	mutex.Lock()
	defer mutex.Unlock()

	KeenTune = new(KeentunedConf)
	err := KeenTune.Save()
	if err != nil {
		return err
	}

	initChanAndIPMap()
	return nil
}

// CheckAndReloadConf check md5 of ENV keentuned.conf and reload conf to cache
func CheckAndReloadConf() error {
	md5Hash := GetKeenTuneConfFileMD5()
	if KeenTuneConfMD5 != md5Hash {
		KeenTuneConfMD5 = md5Hash
		err := ReloadConf()
		if err != nil {
			return fmt.Errorf("reload conf failed: %v", err)
		}
	}

	return nil
}

// GetRerunConf get rerun configuration
func GetRerunConf(jobConf string) (map[string]interface{}, error) {
	cfg, err := ini.InsensitiveLoad(jobConf)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s, %v", keentuneConfigFile, err)
	}

	var retConf = make(map[string]interface{})
	kd := cfg.Section("keentuned")
	retConf["BaseRound"] = kd.Key("BASELINE_BENCH_ROUND").MustInt()
	retConf["TuningRound"] = kd.Key("TUNING_BENCH_ROUND").MustInt()
	retConf["RecheckRound"] = kd.Key("RECHECK_BENCH_ROUND").MustInt()

	return retConf, nil
}


