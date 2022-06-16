package common

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	"net/http"
	"strconv"
	"strings"
)

type TuneCmdResp struct {
	Iteration    int    `json:"iteration"`
	BaseRound    int    `json:"baseline_bench_round"`
	TuningRound  int    `json:"tuning_bench_round"`
	RecheckRound int    `json:"recheck_bench_round"`
	Algo         string `json:"algorithm"`
	BenchGroup   string `json:"bench_group"`
	TargetGroup  string `json:"target_group"`
}

type TrainCmdResp struct {
	Trial int    `json:"trial"`
	Epoch int    `json:"epoch"`
	Algo  string `json:"algorithm"`
	Data  string `json:"data"`
}

func read(w http.ResponseWriter, r *http.Request) {
	var result = new(string)
	w.Header().Set("content-type", "text/json")
	if strings.ToUpper(r.Method) != "POST" {
		*result = fmt.Sprintf("request method '%v' is not supported", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(*result))
		return
	}

	var err error
	defer func() {
		w.WriteHeader(http.StatusOK)
		if err != nil {
			w.Write([]byte(fmt.Sprintf("{\"suc\": false, \"msg\": \"%v\"}", err.Error())))
			log.Errorf("", "read operation: %v", err)
			return
		}

		w.Write([]byte(fmt.Sprintf("{\"suc\": true, \"msg\": %s}", *result)))
	}()

	bytes, err := ioutil.ReadAll(&io.LimitedReader{R: r.Body, N: LimitBytes})
	if err != nil {
		return
	}

	var req struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}

	err = json.Unmarshal(bytes, &req)
	if err != nil {
		err = fmt.Errorf("parse request info failed: %v", err)
		return
	}

	if req.Type == "tuning" {
		err = readTuneInfo(req.Name, result)
		return
	}

}

func parseBenchRound(info string, resp *TuneCmdResp) error {
	if strings.Contains(strings.ToLower(info), "baseline_bench_round") {
		num, err := parseRound(info, "baseline_bench_round")
		if err != nil {
			return err
		}

		resp.BaseRound = num
	}

	if strings.Contains(strings.ToLower(info), "tuning_bench_round") {
		num, err := parseRound(info, "tuning_bench_round")
		if err != nil {
			return err
		}
		resp.TuningRound = num
	}

	if strings.Contains(strings.ToLower(info), "recheck_bench_round") {
		num, err := parseRound(info, "recheck_bench_round")
		if err != nil {
			return err
		}
		resp.RecheckRound = num
	}

	return nil
}

func parseRound(info, key string) (int, error) {
	if !strings.Contains(strings.ToLower(info), key) {
		return 0, nil
	}

	flagParts := strings.Split(info, "=")
	if len(flagParts) != 2 {
		return 0, fmt.Errorf("algorithm not found")
	}

	num, err := strconv.Atoi(strings.TrimSpace(flagParts[1]))
	if err != nil {
		return 0, fmt.Errorf("get %v number err %v", key, err)
	}

	return num, nil
}

func readTuneInfo(job string, result *string) error {
	cmd := file.GetRecord(tuningCsv, "name", job, "cmd")
	if cmd == "" {
		return fmt.Errorf("'%v' not exists", job)
	}

	var resp = TuneCmdResp{}
	iterationStr := file.GetRecord(tuningCsv, "name", job, "iteration")
	iteration, err := strconv.Atoi(strings.Trim(iterationStr, " "))
	if err != nil || iteration <= 0 {
		return fmt.Errorf("'%v' not exists", "iteration")
	}

	resp.Iteration = iteration

	matchedConfig, err := parseConfigFlag(cmd)
	if err != nil {
		return err
	}
	for _, info := range strings.Split(matchedConfig, "\n") {
		if strings.TrimSpace(info) == "" {
			continue
		}

		if strings.Contains(strings.ToLower(info), "algorithm") {
			algoPart := strings.Split(info, "=")
			if len(algoPart) != 2 {
				return fmt.Errorf("algorithm not found")
			}
			resp.Algo = strings.Trim(algoPart[1], " ")
		}

		err = parseBenchRound(info, &resp)
		if err != nil {
			return err
		}

		if strings.Contains(strings.ToLower(info), "bench_config") {
			flagParts := strings.Split(info, "=")
			if len(flagParts) != 2 {
				return fmt.Errorf("bench_config not found")
			}
			pathParts := strings.Split(flagParts[1], "/")
			if len(pathParts) < 1 {
				return fmt.Errorf("bench_config '%v' is abnormal", flagParts[1])
			}
			resp.BenchGroup = pathParts[len(pathParts)-1]
		}

	}

	benchGroup, targetGroup, err := config.GetJobGroup(job)
	if err != nil {
		return err
	}

	resp.BenchGroup += benchGroup
	resp.TargetGroup = targetGroup

	bytes, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	*result = string(bytes)
	return nil
}
