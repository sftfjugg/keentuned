package modules

import (
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	"strings"
)

func (tuner *Tuner) loop() error {
	var err error
	var aheadStop bool
	for i := 1; i <= tuner.MAXIteration; i++ {
		tuner.Iteration = i
		tuner.updateJob(map[int]interface{}{tuneCurRoundIdx: fmt.Sprint(tuner.Iteration)})
		// 1. acquire
		if aheadStop, err = tuner.acquire(); err != nil {
			return err
		}

		if aheadStop {
			break
		}

		// 2. apply
		if err = tuner.setConfigure(); err != nil {
			return err
		}

		log.Debugf(log.ParamTune, "Step%v. loop %vth set configuration details: %v", tuner.Step, tuner.Iteration, tuner.applySummary)

		// 3. benchmark
		if err = tuner.benchmark(); err != nil {
			return err
		}

		// 4. feedback
		if err = tuner.feedback(); err != nil {
			return fmt.Errorf("feedback %vth configuration:%v", i, err)
		}

		// 5. analyse
		optimalRatioInfo := tuner.analyseResult()
		if optimalRatioInfo != "" {
			log.Infof(tuner.logName, "\tCurrent optimal iteration: %v\n", optimalRatioInfo)
		}

		if tuner.isInterrupted() {
			log.Infof(tuner.logName, "Tuning interrupted after step%v, [loop tuning] round %v finish.", tuner.Step, i)
			return fmt.Errorf("tuning is interrupted")
		}
	}

	return nil
}

func (tuner *Tuner) benchmark() error {
	// get round of execution benchmark
	var round int
	if int(tuner.Group[0].Dump.budget) != 0 {
		round = int(tuner.Group[0].Dump.budget)
	} else {
		round = config.KeenTune.ExecRound
	}

	var err error

	// execution benchmark
	tuner.Benchmark.LogName = tuner.logName
	tuner.feedbackScore, _, err = tuner.RunBenchmark(round)
	if err != nil {
		if strings.Contains(err.Error(), "get benchmark is interrupted") {
			log.Infof(tuner.logName, "Tuning interrupted after step%v, [run benchmark] round %v stopped.", tuner.Step, tuner.Iteration)
			return fmt.Errorf("run benchmark interrupted")
		}
		return fmt.Errorf("tuning execute %vth benchmark err:%v", tuner.Iteration, err)
	}

	log.Infof(tuner.logName, "[Iteration %v] Benchmark result: %v", tuner.Iteration, tuner.benchSummary)
	// dump benchmark result of current tuning Iteration
	if !tuner.isSensitize {
		tuner.dump(processOpt)
	}

	return nil
}

