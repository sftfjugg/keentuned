package common

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	m "keentune/daemon/modules"
	"net/http"
	"strconv"
	"strings"
)

// TuneCmdResp ...
type TuneCmdResp struct {
	Iteration    int    `json:"iteration"`
	BaseRound    int    `json:"baseline_bench_round"`
	TuningRound  int    `json:"tuning_bench_round"`
	RecheckRound int    `json:"recheck_bench_round"`
	Algo         string `json:"algorithm"`
}

// TrainCmdResp ...
type TrainCmdResp struct {
	Trial int    `json:"trial"`
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

	switch strings.ToLower(req.Type) {
	case "tuning":
		err = readTuneInfo(req.Name, result)
		return
	case "training":
		err = readTrainInfo(req.Name, result)
		return
	case "param-bench":
		err = readParamAndBenchConf(req.Name, result)
		return
	case "target-group":
		err = readTargetGroup(req.Name, result)
		return
	default:
		err = fmt.Errorf("type '%v' is not supported", req.Type)
		return
	}
}

// readParamAndBenchConf read parameter and benchmark config path from job backup conf
func readParamAndBenchConf(job string, result *string) error {
	params, bench, err := config.GetJobParamConfig(job)
	if err != nil {
		return err
	}
	var resp struct {
		Params string `json:"parameters"`
		Bench  string `json:"benchmark"`
	}

	resp.Params = params
	resp.Bench = bench

	bytes, err := json.Marshal(resp)
	* result = string(bytes)
	return err
}

// readTrainInfo read training rerun info from job records
func readTrainInfo(job string, result *string) error {
	record, err := file.GetOneRecord(config.GetDumpPath(config.SensitizeCsv), job, "name")
	if err != nil {
		return fmt.Errorf("search '%v' err: %v", job, err)
	}

	if len(record) != m.TrainCols {
		return fmt.Errorf("search '%v' record '%v' Incomplete", job, strings.Join(record, ","))
	}

	var resp TrainCmdResp
	resp.Trial, _ = strconv.Atoi(record[m.TrainTrialsIdx])
	resp.Algo = record[m.TrainAlgoIdx]
	resp.Data = record[m.TrainDataIdx]
	bytes, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	*result = string(bytes)
	return nil
}

// readTuneInfo read tuning rerun info from job records and backup conf
func readTuneInfo(job string, result *string) error {
	records, err := file.GetOneRecord(config.GetDumpPath(config.TuneCsv), job, "name")
	if err != nil {
		return fmt.Errorf("'%v' not exists", job)
	}

	if len(records) != m.TuneCols {
		return fmt.Errorf("invalid record '%v': column size %v, expected %v", job, len(records), m.TuneCols)
	}

	var resp = TuneCmdResp{}
	iteration, err := strconv.Atoi(strings.Trim(records[m.TuneRoundIdx], " "))
	if err != nil || iteration <= 0 {
		return fmt.Errorf("'%v' not exists", "iteration")
	}

	resp.Algo = records[m.TuneAlgoIdx]
	resp.Iteration = iteration

	m, err := config.GetRerunConf(records[m.TuneWSPIdx] + "/keentuned.conf")
	if err != nil {
		return err
	}

	resp.BaseRound = m["BaseRound"].(int)
	resp.TuningRound = m["TuningRound"].(int)
	resp.RecheckRound = m["RecheckRound"].(int)

	bytes, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	*result = string(bytes)
	return nil
}

// readTargetGroup  read target group info by grep condition
// 1) grep is empty, represents getting the total group;
// 2) grep is non-empty, e.g.: "cpu_high_load.conf", an active status profile name,
//        represents grep the profile set target group(s).
func readTargetGroup(grep string, result *string) error {
	defer func() {
		if len(*result) > 0 {
			*result = fmt.Sprintf("\"%v\"", *result)
		}
	}()

	if grep == "" {
		for _, group := range config.KeenTune.Target.Group {
			*result += fmt.Sprintf("[target-group-%v]\\n", group.GroupNo)
			*result += fmt.Sprintf("TARGET_IP = %v\\n", strings.Join(group.IPs, ","))
		}

		return nil
	}

	filePath := config.GetProfileWorkPath("active.conf")
	activeGroup := file.GetRecord(filePath, "name", grep, "group_info")
	if len(activeGroup) == 0 {
		return fmt.Errorf("No records found")
	}

	actives := strings.Split(activeGroup, " ")
	for _, group := range config.KeenTune.Target.Group {
		for _, info := range actives {
			if strings.Contains(group.GroupName, strings.Trim(info, "group")) {
				*result += fmt.Sprintf("[target-group-%v]\\n", group.GroupNo)
				*result += fmt.Sprintf("TARGET_IP = %v\\n", strings.Join(group.IPs, ","))
				break
			} else {
				for _, ip := range group.IPs {
					if info == ip {
						*result += fmt.Sprintf("[target-group-%v]\\n", group.GroupNo)
						*result += fmt.Sprintf("TARGET_IP = %v\\n", ip)
					}
				}
			}
		}
	}

	if *result == "" {
		return fmt.Errorf("No matched group found")
	}

	return nil
}


