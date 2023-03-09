package modules

import (
	"encoding/json"
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	"keentune/daemon/common/utils/http"
	"strings"
	"time"
)

func (tuner *Tuner) getBest() error {
	// get best configuration
	start := time.Now()
	url := config.KeenTune.BrainIP + ":" + config.KeenTune.BrainPort + "/best"
	resp, err := http.RemoteCall("GET", url, nil)
	if err != nil {
		return fmt.Errorf("remote call: %v", err)
	}

	var bestConfig ReceivedConfigure
	err = json.Unmarshal(resp, &bestConfig)
	if err != nil {
		return fmt.Errorf("unmarshal best config: %v", err)
	}

	// time cost
	timeCost := utils.Runtime(start)
	tuner.timeSpend.best += timeCost.Count

	tuner.bestInfo.Round = bestConfig.Iteration
	tuner.bestInfo.Score = bestConfig.Score
	tuner.bestInfo.Parameters = bestConfig.Candidate

	return nil
}

func (tuner *Tuner) verifyBest() error {
	err := tuner.setConfigure()
	if err != nil {
		log.Errorf(log.ParamTune, "best apply configuration failed, details: %v", tuner.applySummary)
		return err
	}

	log.Debugf(log.ParamTune, "Step%v. apply best configuration details: %v", tuner.Step, tuner.applySummary)

	log.Infof(log.ParamTune, "\nStep%v. Tuning is finished, checking benchmark score of best configuration.\n\n", tuner.IncreaseStep())

	if tuner.feedbackScore, _, err = tuner.RunBenchmark(config.KeenTune.AfterRound); err != nil {
		if strings.Contains(err.Error(), "get benchmark is interrupted") {
			log.Infof(log.ParamTune, "Tuning interrupted after step%v, [check best configuration benchmark] stopped.", tuner.Step)
			return fmt.Errorf("run benchmark interrupted")
		}
		log.Errorf(log.ParamTune, "tuning execute best benchmark err:%v\n", err)
		return err
	}

	log.Infof(log.ParamTune, "[BEST] Benchmark result: %v\n", tuner.benchSummary)

	currentRatioInfo := tuner.analyseBestResult()
	if currentRatioInfo != "" {
		log.Infof(log.ParamTune, "[BEST] Tuning improvement: %v\n", currentRatioInfo)
	}

	tuner.end()

	if tuner.Verbose {
		log.Infof(log.ParamTune, "Time cost statistical information:%v", tuner.timeSpend.detailInfo)
	}

	return nil
}

func (tuner *Tuner) dumpBest() error {
	if !config.KeenTune.BestDump {
		return nil
	}

	err := tuner.getBest()
	if err != nil {
		return err
	}

	err = tuner.parseBestParam()
	if err != nil {
		return err
	}

	suffix := "_best.json"
	var fileList string
	jobPath := config.GetTuningPath(tuner.Name)
	for index := range tuner.Group {
		err = tuner.Group[index].Dump.Save(tuner.Name, fmt.Sprintf("_group%v%v", index+1, suffix))
		if err != nil {
			log.Warnf(tuner.logName, "dump best.json failed, %v", err)
			continue
		}

		fileList += fmt.Sprintf("\n\t%v/%v_group%v%v", jobPath, tuner.Name, index+1, suffix)
	}

	log.Infof(tuner.logName, "\nStep%v. Best configuration dump successfully. File list: %v\n\n", tuner.IncreaseStep(), fileList)
	return nil
}

