package modules

import (
	"encoding/json"
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	"keentune/daemon/common/utils/http"
	"time"
)

func (tuner *Tuner) getBest() error {
	// get best configuration
	start := time.Now()
	url := config.KeenTune.BrainIP + ":" + config.KeenTune.BrainPort + "/best"
	resp, err := http.RemoteCall("GET", url, nil)
	if err != nil {
		return fmt.Errorf("remote call: %v\n", err)
	}

	var bestConfig ReceivedConfigure
	err = json.Unmarshal(resp, &bestConfig)
	if err != nil {
		return fmt.Errorf("unmarshal best config: %v\n", err)
	}

	// time cost
	timeCost := utils.Runtime(start)
	tuner.timeSpend.best += timeCost.Count

	tuner.bestInfo.Round = bestConfig.Iteration
	tuner.bestInfo.Score = bestConfig.Score
	tuner.bestInfo.Parameters=bestConfig.Candidate

	return nil
}

func (tuner *Tuner) verifyBest() error {
	err := tuner.setConfigure()
	if err != nil {
		log.Errorf(log.ParamTune, "best apply configuration failed:%v, details: %v", tuner.applySummary)
		return err
	}

	log.Debugf(log.ParamTune, "Step%v. apply configuration details: %v", tuner.Step, tuner.applyDetail)

	log.Infof(log.ParamTune, "Step%v. Tuning is finished, checking benchmark score of best configuration.\n", tuner.IncreaseStep())

	if tuner.feedbackScore, _, tuner.benchSummary, err = tuner.RunBenchmark(config.KeenTune.AfterRound); err != nil {
		if err.Error() == "get benchmark is interrupted" {
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
