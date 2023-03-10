package modules

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	"keentune/daemon/common/utils/http"
	"reflect"
	"strings"
	"sync"
	"time"
)

const (
	itemLenWarnFmt      = "find bench.json items length [%v] is not equal to /benchmark response scores length [%v], please check the bench.json and the python script whether matched"
	itemNotFoundWarnFmt = "benchmark response  [%v] detail info not exist, please check the bench.json and the python script whether matched"

	debugScoreInfoFmt = "\n\t[%v]\t(weight: %.1f)\tscores %v,\taverage = %.3f,\t%v"
	positiveWeightFmt = "\n\t[%v]\t(weight: %.1f)\taverage scores = %.3f"
)

// Benchmark define benchmark cmd and host to run
type Benchmark struct {
	Cmd         string                `json:"benchmark_cmd"`
	Host        string                `json:"host"`
	FilePath    string                `json:"local_script_path"`
	Items       map[string]ItemDetail `json:"items"`
	round       int
	verbose     bool
	LogName     string   `json:"-"`
	SortedItems []string `json:"-"`
	isBase      bool
}

// BenchResult benchmark request result
type BenchResult struct {
	Success bool                 `json:"suc"`
	Result  map[string][]float32 `json:"result,omitempty"`
	Message interface{}          `json:"msg,omitempty"`
}

// RunBenchmark : run benchmark script or command in client
func (tuner *Tuner) RunBenchmark(num int, isBase ...bool) (map[string][]float32, string, error) {
	// start
	start := time.Now()
	if len(isBase) == 0 {
		tuner.Benchmark.isBase = false
	} else {
		tuner.Benchmark.isBase = isBase[0]
	}

	// do benchmark
	groupScores, benchResult := tuner.doBenchmark(num)

	// analyse score, Currently, only one group of benches is supported, so idx is 0
	err := tuner.analyseScore(groupScores[0], benchResult, start)
	return groupScores[0], tuner.benchSummary, err
}

func (tuner *Tuner) doBenchmark(num int) ([]map[string][]float32, []*string) {
	var wg sync.WaitGroup
	var benchResult = make([]*string, len(config.KeenTune.BenchIPMap))
	sc := NewSafeChan()
	defer sc.SafeStop()
	var scores = make([]map[string][]float32, len(config.KeenTune.BenchGroup))

	for gpIdx, benchGroup := range config.KeenTune.BenchGroup {
		scores[gpIdx] = make(map[string][]float32)
		benchResult[gpIdx] = new(string)

		for index, benchIP := range benchGroup.SrcIPs {
			wg.Add(1)
			ipIndex := config.KeenTune.Bench.BenchIPMap[benchIP]
			req := request{
				host:    fmt.Sprintf("%s:%s", benchIP, benchGroup.SrcPort),
				id:      index,
				ipIndex: ipIndex,
				body:    getBenchReq(tuner.Benchmark.Cmd, ipIndex),
				groupID: gpIdx,
			}

			go doOneBench(&wg, req, scores, benchResult, num, sc)
		}
	}

	wg.Wait()

	return scores, benchResult
}

func getBenchReq(cmd string, id int) interface{} {
	var requestBody = map[string]interface{}{}
	requestBody["benchmark_cmd"] = cmd
	requestBody["resp_ip"] = config.RealLocalIP
	requestBody["resp_port"] = config.KeenTune.Port
	requestBody["bench_id"] = id
	return requestBody
}

func doOneBench(s *sync.WaitGroup, req request, scores []map[string][]float32, result []*string, num int, sc *SafeChan) {
	var errMsg error
	config.IsInnerBenchRequests[req.ipIndex] = true
	defer func() {
		s.Done()
		config.IsInnerBenchRequests[req.ipIndex] = false
		if errMsg != nil {
			sc.SafeStop()
		}
	}()

	var scoreResult = make(map[string][]float32)
	url := fmt.Sprintf("%v/benchmark", req.host)
	resIdx := req.ipIndex - 1
	for i := 0; i < num; i++ {
		resp, err := http.RemoteCall("POST", url, req.body)
		if err != nil {
			errMsg = err
			*result[resIdx] += fmt.Sprintf("bench.group %v-%v %v", req.groupID+1, req.id+1, err.Error())
			return
		}

		benchScore, err := parseScore(resp, req, sc)
		if err != nil {
			errMsg = err
			*result[resIdx] += fmt.Sprintf("bench.group %v-%v %v", req.groupID+1, req.id+1, err.Error())
			return
		}

		for item, scores := range benchScore {
			scoreResult[item] = append(scoreResult[item], scores...)
		}
	}

	scores[req.groupID] = scoreResult
}

func (benchmark *Benchmark) getScore(scores map[string][]float32, start time.Time, benchTime *time.Duration) (map[string]ItemDetail, string, error) {
	benchScoreResult := map[string]ItemDetail{}
	var average float32
	if len(scores) == 0 {
		return nil, "", fmt.Errorf("execute %v rounds all benchmark failed", benchmark.round)
	}

	if len(benchmark.Items) != len(scores) {
		log.Warnf("", itemLenWarnFmt, len(benchmark.Items), len(scores))
	}

	resultString := ""
	for _, name := range benchmark.SortedItems {
		info, _ := benchmark.Items[name]

		scoreSlice, ok := scores[name]
		if !ok {
			log.Warnf("", itemNotFoundWarnFmt, name)
			continue
		}

		num := len(scoreSlice)
		var sumScore float32
		for i := 0; i < num; i++ {
			sumScore += scoreSlice[i]
		}

		average = sumScore / float32(len(scoreSlice))

		if benchmark.verbose {
			resultString += fmt.Sprintf(debugScoreInfoFmt, name, info.Weight, scoreSlice, average, utils.Fluctuation(scoreSlice, average))
		}

		if !benchmark.verbose && info.Weight > 0.0 {
			resultString += fmt.Sprintf(positiveWeightFmt, name, info.Weight, average)
		}

		var items ItemDetail
		items.Negative = info.Negative
		items.Weight = info.Weight
		items.Strict = info.Strict

		if benchmark.isBase {
			items.Baseline = scoreSlice
		} else {
			items.Value = average
		}

		benchScoreResult[name] = items

		timeCost := utils.Runtime(start)
		*benchTime += timeCost.Count

		if benchmark.verbose {
			resultString = fmt.Sprintf("%v, %v", resultString, timeCost.Desc)
		}
	}
	return benchScoreResult, resultString, nil
}

// SendScript : send script file to client
func (benchmark Benchmark) SendScript(sendTime *time.Duration, Host string) (bool, string, error) {
	start := time.Now()
	benchBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", config.KeenTune.Home, benchmark.FilePath))
	if err != nil {
		return false, "", fmt.Errorf("SendScript readFile err:%v", err)
	}

	requestBody := map[string]interface{}{
		"file_name":   benchmark.FilePath,
		"body":        string(benchBytes),
		"encode_type": "utf-8",
	}

	err = http.ResponseSuccess("POST", Host+"/sendfile", requestBody)
	if err != nil {
		return false, "", fmt.Errorf("SendScript remote call err:%v", err)
	}

	timeCost := utils.Runtime(start)
	*sendTime += timeCost.Count

	return true, timeCost.Desc, nil
}

func (tuner *Tuner) analyseScore(scores map[string][]float32, benchResults []*string, start time.Time) error {
	var errMsg string
	for _, status := range benchResults {
		if *status != "" {
			errMsg += fmt.Sprintf("%v;", *status)
		}
	}

	if errMsg != "" {
		return fmt.Errorf(strings.TrimSuffix(errMsg, ";"))
	}

	benchScoreResult, summaryInfo, err := tuner.Benchmark.getScore(scores, start, &tuner.timeSpend.benchmark)
	if err != nil {
		return err
	}

	if tuner.Benchmark.isBase {
		tuner.baseScore = benchScoreResult
	} else {
		for key := range benchScoreResult {
			item := benchScoreResult[key]
			item.Baseline = tuner.baseScore[key].Baseline
			benchScoreResult[key] = item
		}
		tuner.benchScore = benchScoreResult
	}

	tuner.benchSummary = summaryInfo

	return nil
}

func parseScore(body []byte, req request, sc *SafeChan) (map[string][]float32, error) {
	var benchResult BenchResult
	err := json.Unmarshal(body, &benchResult)
	if err != nil {
		return nil, fmt.Errorf("parse score err:%v", err)
	}

	if !benchResult.Success {
		return nil, fmt.Errorf("parse score failed, msg :%v", benchResult.Message)
	}

	ticker := time.NewTicker(time.Duration(config.KeenTune.BenchTimeout) * time.Minute)
	defer ticker.Stop()
	errTimeout := fmt.Errorf("benchmark wait for %v minutes timeout", config.KeenTune.BenchTimeout)
	select {
	case bytes := <-config.BenchmarkResultChan[req.ipIndex]:
		log.Debugf("", "get benchmark result:%s", bytes)
		if err = json.Unmarshal(bytes, &benchResult); err != nil {
			return nil, fmt.Errorf("unmarshal request info err:%v", err)
		}

		if !benchResult.Success {
			return nil, fmt.Errorf("msg:%v", benchResult.Message)
		}

		result, ok := benchResult.Message.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("msg: assert message type failed, type is %v", reflect.TypeOf(benchResult.Message))
		}

		benchResult.Result = map[string][]float32{}
		for key, value := range result {
			scoreI, ok := value.([]interface{})
			if !ok {
				return nil, fmt.Errorf("msg: assert score slice failed, type is %v", reflect.TypeOf(value))
			}

			var scores []float32
			for _, score := range scoreI {
				value, ok := score.(float64)
				if !ok {
					return nil, fmt.Errorf("msg: assert score to float64 failed, type is %v", reflect.TypeOf(value))
				}
				scores = append(scores, float32(value))
			}

			benchResult.Result[key] = scores
		}

		break
	case <-ticker.C:
		config.ServeTerminate <- true
		terminate(req.host)
		return nil, errTimeout
	case <-config.ClientOffline:
		terminate(req.host)
		return nil, errTimeout
	case <-StopSig:
		terminate(req.host)
		closeChan()
		sc.SafeStop()
		return nil, fmt.Errorf("get benchmark is interrupted")
	case _, ok := <-sc.C:
		if !ok {
			return nil, fmt.Errorf("get benchmark is interrupted")
		}
	}

	if len(benchResult.Result) == 0 {
		return nil, fmt.Errorf("get benchmark result is nil")
	}

	return benchResult.Result, nil
}

func closeChan() {
	for i := range config.IsInnerBenchRequests {
		config.IsInnerBenchRequests[i] = false
	}
}

