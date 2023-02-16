package modules

import (
	"encoding/json"
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	"keentune/daemon/common/utils/http"
	"strings"
	"time"
)

type implyDetail struct {
	useTime         string
	applySummary    string
	applyDetail     string
	benchSummary    string
	backupFailure   string
	backupWarning   string
	rollbackDetail  string
	rollbackFailure string
}

// Tuner define a tuning job include Algorithm, Benchmark, Group
type Tuner struct {
	Name          string // Name job name
	Algorithm     string // 使用的算法
	MAXIteration  int    // 最大执行轮次
	Iteration     int    // 当前轮次
	StartTime     time.Time
	Benchmark     Benchmark
	timeSpend     timeSpend
	ParamConf     config.DBLMap
	Verbose       bool
	Step          int    // tuning process steps
	isSensitize   bool   // sensitive parameter identification mark
	Flag          string // command flag, enum: "collect", "tuning"
	logName       string
	feedbackScore map[string][]float32
	benchScore    map[string]ItemDetail
	Group         []Group
	BrainParam    []Parameter
	ReadConfigure bool
	implyDetail
	bestInfo    Configuration
	allowUpdate bool
	Setter
	Trainer
	rollbackReq map[string]interface{}
}

type timeSpend struct {
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
	tuner.logName = log.ParamTune
	if err = tuner.CreateTuneJob(); err != nil {
		log.Errorf(log.ParamTune, "create tune job failed: %v", err)
		return
	}

	defer func() { tuner.parseTuningError(err) }()

	if err = tuner.init(); err != nil {
		err = fmt.Errorf("prepare for tuning: %v", err)
		return
	}

	log.Infof(log.ParamTune, "\nStep%v. Start tuning, total iteration is %v.\n\n", tuner.IncreaseStep(), tuner.MAXIteration)

	if err = tuner.loop(); err != nil {
		err = fmt.Errorf("loop tuning: %v", err)
		return
	}

	if err = tuner.dumpBest(); err != nil {
		err = fmt.Errorf("dump best configuration: %v", err)
		return
	}

	if err = tuner.verifyBest(); err != nil {
		err = fmt.Errorf("check best configuration: %v", err)
		return
	}
}

func (tuner *Tuner) parseTuningError(err error) {
	defer tuner.end()
	if err == nil {
		tuner.updateStatus(Finish)
		return
	}

	if tuner.Flag == "tuning" {
		tuner.rollback()
	}

	if strings.Contains(err.Error(), "interrupted") {
		tuner.updateStatus(Stop)
		log.Infof(tuner.logName, "parameter optimization job abort!")
		return
	}

	tuner.updateStatus(Err)
	log.Errorf(tuner.logName, "%v", err)
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

	if err = tuner.parseAcquireParam(acquiredInfo); err != nil {
		return false, err
	}

	// check interrupted
	if tuner.isInterrupted() {
		log.Infof(tuner.logName, "Tuning interrupted after step%v, [acquire] round %v finish.", tuner.Step, tuner.Iteration)
		return false, fmt.Errorf("tuning is interrupted")
	}

	return false, nil
}

/*feedback configuration with score to brain*/
func (tuner *Tuner) feedback() error {
	start := time.Now()
	feedbackMap := map[string]interface{}{
		"iteration":   tuner.Iteration - 1,
		"bench_score": tuner.feedbackScore,
	}

	url := config.KeenTune.BrainIP + ":" + config.KeenTune.BrainPort + "/feedback"
	body, err := http.RemoteCall("POST", url, feedbackMap)
	if err != nil {
		return fmt.Errorf("'feedback' remote call err:%v", err)
	}

	var resp struct {
		Suc       bool        `json:"suc"`
		Msg       interface{} `json:"msg"`
		TimeData  string      `json:"time_data"`
		ScoreData string      `json:"score_data"`
	}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		return fmt.Errorf("unmarshal 'feedback' responese failed: %v", err)
	}

	if !resp.Suc {
		return fmt.Errorf("'feedback' failed, msg: %v", resp.Msg)
	}

	scorePath := fmt.Sprintf("%v/score.csv", config.GetTuningPath(tuner.Name))

	err = file.Append(scorePath, strings.Split(resp.ScoreData, ","))
	if err != nil {
		log.Errorf(tuner.logName, "%vth iteration save score value failed: %v", tuner.Iteration, err)
	}

	timePath := fmt.Sprintf("%v/time.csv", config.GetTuningPath(tuner.Name))
	err = file.Append(timePath, strings.Split(resp.TimeData, ","))
	if err != nil {
		log.Errorf(tuner.logName, "%vth iteration save score value failed: %v", tuner.Iteration, err)
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

	var endInfo = make(map[int]interface{})

	if tuner.Flag == "tuning" {
		endInfo[TuneEndIdx] = start.Format(Format)
		endInfo[tuneCostIdx] = endTime(int64(totalTime))
	} else if tuner.Flag == "training" {
		endInfo[TrainEndIdx] = start.Format(Format)
		endInfo[trainCostIdx] = endTime(int64(totalTime))
	}

	tuner.updateJob(endInfo)

	if totalTime == 0.0 || !tuner.Verbose {
		return
	}

	tuner.setTimeSpentDetail(totalTime)
}

func endTime(cost int64) string {
	h := cost / 3600
	min := cost % 3600
	m := min / 60
	sec := min % 60

	if h >= 1 {
		return fmt.Sprintf("%dh%dm%vs", h, m, sec)
	}

	if m >= 1 {
		return fmt.Sprintf("%dm%vs", m, sec)
	}

	return fmt.Sprintf("%vs", sec)
}

func (tuner *Tuner) setTimeSpentDetail(totalTime float64) {
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

func (tuner *Tuner) IncreaseStep(initVal ...int) int {
	if len(initVal) == 0 {
		tuner.Step++
		return tuner.Step
	}

	tuner.Step = initVal[0] + 1
	return tuner.Step
}

// deleteUnAVLParams delete unavailable parameters for brain init
func (tuner *Tuner) deleteUnAVLParams() {
	var newBrainParams []Parameter
	for _, p := range tuner.BrainParam {
		name, idx, err := parseBrainName(p.ParaName)
		if err != nil {
			continue
		}

		// the domain is unavailable
		unavailableParam, exist := tuner.Group[idx].UnAVLParams[p.DomainName]
		if len(tuner.Group[idx].UnAVLParams[p.DomainName]) == 0 && exist {
			continue
		}

		_, find := unavailableParam[name]
		if !find {
			newBrainParams = append(newBrainParams, p)
		}
	}

	tuner.BrainParam = newBrainParams

	if tuner.backupWarning != "" {
		for _, backupWarning := range strings.Split(tuner.backupWarning, multiRecordSeparator) {
			pureInfo := strings.TrimSpace(backupWarning)
			if len(pureInfo) > 0 {
				log.Warn(tuner.logName, backupWarning)
			}
		}
	}
}

