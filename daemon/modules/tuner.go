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

// Tuner define a tuning job include Algorithm, Benchmark, Target
type Tuner struct {
	Name                string
	Algorithm           string // 使用的算法
	MAXIteration        int    // 最大执行轮次
	Iteration           int    // 当前轮次
	TargetHost          []string
	BenchmarkHost       string
	StartTime           time.Time
	Benchmark           Benchmark
	BaseConfiguration   []Configuration
	BestConfiguration   Configuration
	nextConfiguration   Configuration
	TargetConfiguration []Configuration
	timeSpend           TimeSpend
	ParamConf           map[string]map[string]interface{}
	Verbose             bool
	Step                int    // tuning process steps
	isSensitize         bool   // sensitive parameter identification mark
	Flag                string // command flag, enum: "collect", "tuning"
	logName             string
	benchScore          map[string][]float32
	Group []Target
	BrainParam []Parameter
	MergedParam []map[string]interface{}
}

type TimeSpend struct {
	init       time.Duration
	acquire    time.Duration
	apply      time.Duration
	send       time.Duration
	benchmark  time.Duration
	feedback   time.Duration
	end        time.Duration
	best       time.Duration
	detailInfo string
}

// Tune : tuning main process
func (tuner *Tuner) Tune() {
	var err error

	defer func() {
		if err != nil {
			tuner.end()
			parseTuningError(log.ParamTune, err)
		}
	}()

	tuner.logName = log.ParamTune
	if err = tuner.init(); err != nil {
		err = fmt.Errorf("[%v] prepare for tuning: %v", utils.ColorString("red", "ERROR"), err)
		return
	}

	log.Infof(log.ParamTune, "\nStep%v. Start tuning, total iteration is %v.\n", tuner.IncreaseStep(), tuner.MAXIteration)

	if err = tuner.loop(); err != nil {
		err = fmt.Errorf("[%v] loop tuning: %v", utils.ColorString("red", "ERROR"), err)
		return
	}

	if err = tuner.getBestConfiguration(); err != nil {
		err = fmt.Errorf("[%v] get best configuration: %v", utils.ColorString("red", "ERROR"), err)
		return
	}

	if err = tuner.checkBestConfiguration(); err != nil {
		err = fmt.Errorf("[%v] check best configuration: %v", utils.ColorString("red", "ERROR"), err)
		return
	}
}

func parseTuningError(logName string, err error) {
	if err == nil {
		return
	}

	Rollback(logName)

	if strings.Contains(err.Error(), "interrupted") {
		log.Infof(logName, "parameter optimization job abort!")
		return
	}
	log.Infof(logName, "%v", err)
}

/*acquire configuration from brain*/
func (tuner *Tuner) acquire() (bool, error) {
	// remote call and parse info
	start := time.Now()
	url := config.KeenTune.BrainIP + ":" + config.KeenTune.BrainPort + "/acquire"
	resp, err := http.RemoteCall("GET", url, nil)
	if err != nil {
		log.Errorf(tuner.logName, "[%vth] remote call acquire configuration err:%v", tuner.Iteration, err)
		return false, err
	}

	var acquiredInfo ReceivedConfigure
	if err = json.Unmarshal(resp, &acquiredInfo); err != nil {
		log.Errorf(tuner.logName, "[%vth] parse acquire unmarshal err:%v\n", tuner.Iteration, err)
		return false, err
	}

	// check interrupted
	if isInterrupted(tuner.logName) {
		log.Infof(tuner.logName, "Tuning interrupted after step%v, [acquire] round %v finish.", tuner.Step, tuner.Iteration)
		return false, fmt.Errorf("tuning is interrupted")
	}

	// check end loop ahead of time
	if acquiredInfo.Iteration < 0 {
		log.Warnf(tuner.logName, "%vth Tuning acquired round is less than zero, the tuning job will end ahead of time", tuner.Iteration)
		return true, nil
	}

	// time cost
	timeCost := utils.Runtime(start)
	tuner.timeSpend.acquire += timeCost.Count
	if tuner.Verbose {
		log.Infof(tuner.logName, "[Iteration %v] Acquire success, %v", tuner.Iteration, timeCost.Desc)
	}

	// assign to nextConfiguration
	tuner.nextConfiguration = Configuration{
		Parameters: acquiredInfo.Candidate,
		Round:      acquiredInfo.Iteration,
		budget:     acquiredInfo.Budget,
	}

	return false, nil
}

/*feedback configuration with score to brain*/
func (tuner *Tuner) feedback() error {
	start := time.Now()
	feedbackMap := map[string]interface{}{
		"iteration":   tuner.Iteration - 1,
		"bench_score": tuner.benchScore,
	}

	err := http.ResponseSuccess("POST", config.KeenTune.BrainIP+":"+config.KeenTune.BrainPort+"/feedback", feedbackMap)
	if err != nil {
		return fmt.Errorf("[feedback] remote call feedback err:%v\n", err)
	}

	timeCost := utils.Runtime(start)
	tuner.timeSpend.feedback += timeCost.Count
	return nil
}

func (tuner *Tuner) end() {
	start := time.Now()
	http.RemoteCall("GET", config.KeenTune.BrainIP+":"+config.KeenTune.BrainPort+"/end", nil)
	timeCost := utils.Runtime(start)
	tuner.timeSpend.end += timeCost.Count

	totalTime := utils.Runtime(tuner.StartTime).Count.Seconds()

	if totalTime == 0.0 || !tuner.Verbose {
		return
	}

	tuner.setTimeCostToTableString(totalTime)
}

func (tuner *Tuner) setTimeCostToTableString(totalTime float64) {
	initRatio := fmt.Sprintf("%.2f%%", tuner.timeSpend.init.Seconds()*100/totalTime)
	applyRatio := fmt.Sprintf("%.2f%%", tuner.timeSpend.apply.Seconds()*100/totalTime)
	acquireRatio := fmt.Sprintf("%.2f%%", tuner.timeSpend.acquire.Seconds()*100/totalTime)
	benchmarkRatio := fmt.Sprintf("%.2f%%", tuner.timeSpend.benchmark.Seconds()*100/totalTime)
	feedbackRatio := fmt.Sprintf("%.2f%%", tuner.timeSpend.feedback.Seconds()*100/totalTime)

	var detailSlice [][]string
	header := []string{"Process", "Execution Count", "Total Time", "The Share of Total Time"}
	detailSlice = append(detailSlice, header)

	initTime := fmt.Sprintf("%.3fs", tuner.timeSpend.init.Seconds())
	detailSlice = append(detailSlice, []string{"init", "1", initTime, initRatio})

	applyRound := fmt.Sprint(tuner.MAXIteration + 2)
	applyTime := fmt.Sprintf("%.3fs", tuner.timeSpend.apply.Seconds())
	detailSlice = append(detailSlice, []string{"apply", applyRound, applyTime, applyRatio})

	maxRound := fmt.Sprint(tuner.MAXIteration)
	acquireTime := fmt.Sprintf("%.3fs", tuner.timeSpend.acquire.Seconds())
	detailSlice = append(detailSlice, []string{"acquire", maxRound, acquireTime, acquireRatio})

	benchRound := fmt.Sprint(tuner.MAXIteration*config.KeenTune.ExecRound + config.KeenTune.BaseRound + config.KeenTune.AfterRound)
	benchTime := fmt.Sprintf("%.3fs", tuner.timeSpend.benchmark.Seconds())
	detailSlice = append(detailSlice, []string{"benchmark", benchRound, benchTime, benchmarkRatio})

	feedbackTime := fmt.Sprintf("%.3fs", tuner.timeSpend.feedback.Seconds())
	detailSlice = append(detailSlice, []string{"feedback", maxRound, feedbackTime, feedbackRatio})

	tuner.timeSpend.detailInfo = utils.FormatInTable(detailSlice)
}

// Collect Sensitive parameters
func (tuner *Tuner) Collect() {
	var err error

	defer func() {
		tuner.end()
		parseTuningError(log.SensitizeCollect, err)
	}()

	tuner.logName = log.SensitizeCollect
	if err = tuner.init(); err != nil {
		err = fmt.Errorf("[%v] init Collect: %v", utils.ColorString("red", "ERROR"), err)
		return
	}

	log.Infof(log.SensitizeCollect, "\nStep%v. Collect init success.", tuner.IncreaseStep(1))
	log.Infof(log.SensitizeCollect, "\nStep%v. Start sensitization collection, total iteration is %v.\n", tuner.IncreaseStep(), tuner.MAXIteration)

	tuner.isSensitize = true
	if err = tuner.loop(); err != nil {
		err = fmt.Errorf("[%v] loop collect: %v\n", utils.ColorString("red", "ERROR"), err)
		return
	}

	log.Infof(log.SensitizeCollect, "\nStep%v. Sensitization collection finished, you can get the result by the command \"keentune sensitize train\" (see more details: keentune sensitize train -h).", tuner.IncreaseStep())
}

func (tuner *Tuner) IncreaseStep(initVal ...int) int {
	if len(initVal) == 0 {
		tuner.Step++
		return tuner.Step
	}

	tuner.Step = initVal[0] + 1
	return tuner.Step
}

func backup(logName string, conf *Configuration) error {
	requestInfo, err := conf.assembleApplyRequestMap()
	if err != nil {
		return fmt.Errorf("get backup request err: %v", err)
	}

	backupReq := utils.Parse2Map("data", requestInfo)
	if backupReq == nil {
		return fmt.Errorf("get backup request is null")
	}

	details, suc := Backup(logName, backupReq)
	if !suc {
		return fmt.Errorf("backup detail:\n%v", details)
	}

	return nil
}

