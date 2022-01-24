package system

import (
	"fmt"
	com "keentune/daemon/api/common"
	"keentune/daemon/api/param"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	"strings"
	"time"
)

type Service struct{}

type BenchmarkFlag struct {
	Round     int
	BenchConf string
	Name      string
}

func (s *Service) Benchmark(flag BenchmarkFlag, reply *string) error {
	if com.GetRunningTask() != "" {
		return fmt.Errorf("Job %v is running, please wait for it to finish or stop it and retry.", com.GetRunningTask())
	}

	com.SetRunningTask(com.JobBenchmark, flag.Name)

	defer func() {
		*reply = log.ClientLogMap[log.Benchmark]
		log.ClearCliLog(log.Benchmark)
		com.ClearTask()
	}()

	inst, err := param.GetBenchmarkInst(flag.BenchConf)
	if err != nil {
		return err
	}
	inst.LogName = log.Benchmark
	var step int
	step++
	log.Infof(log.Benchmark, "Step%v. Get benchmark instance successfully.\n", step)

	var sendTimeSpend, benchTimeSpend time.Duration
	success, _, err := inst.SendScript(&sendTimeSpend)
	if err != nil || !success {
		log.Errorf(log.Benchmark, "send script file  result: %v err:%v", success, err)
		return fmt.Errorf("send file failed")
	}

	step++
	log.Infof(log.Benchmark, "Step%v. Send benchmark script successfully.\n", step)

	step++
	log.Infof(log.Benchmark, "Step%v. Run [%v] round benchmark ...\n", step, flag.Round)

	var score map[string][]float32
	var benchmarkResult string
	scores := make(map[string][]float32)

	for i := 1; i <= flag.Round; i++ {
		if score, _, benchmarkResult, err = inst.RunBenchmark(1, &benchTimeSpend, false); err != nil {
			if err.Error() == "get benchmark is interrupted" {
				return fmt.Errorf("run [%v] round benchmark positive stopped", i)
			}

			log.Errorf(log.Benchmark, "Run Benchmark err:%v\n", err)
			return fmt.Errorf("Run Benchmark err:%v", err)
		}

		log.Infof(log.Benchmark, "[Iteration %v] Benchmark result:%v", i, strings.Replace(benchmarkResult, "average ", "", 1))

		for key, value := range score {
			scores[key] = append(scores[key], value...)
		}
	}

	if err = file.Save2CSV(config.GetDumpCSVPath(), flag.Name+".csv", scores); err != nil {
		log.Warnf(log.ParamTune, "Save  Baseline benchmark  to file %v failed, reason:[%v]", flag.Name, err)
	}

	step++
	log.Infof(log.Benchmark, "\nStep%v. Benchmark result dump to %v susccessfully.", step, fmt.Sprintf("%v/%v.csv", config.GetDumpCSVPath(), flag.Name))
	return nil
}
