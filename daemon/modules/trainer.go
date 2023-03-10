package modules

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	"keentune/daemon/common/utils/http"
	"strings"
	"time"
)

// Tuner define a tuning job include Explainer, Benchmark, Group
type Trainer struct {
	Data   string
	Job    string
	Trials int
	Config string
}

// Tune : training main process
func (tuner *Tuner) Train() {
	var err error
	tuner.logName = log.SensitizeTrain
	if err = tuner.CreateTrainJob(); err != nil {
		log.Errorf(log.SensitizeTrain, "create sensitize train job failed: %v", err)
		return
	}

	defer func() { tuner.parseTuningError(err) }()

	if err = tuner.initiateSensitization(); err != nil {
		return
	}

	tuner.copyKnobsFile()

	log.Infof(log.SensitizeTrain, "\nStep2. Initiate sensitization success.\n")

	resultString, err := tuner.getSensitivityResult()
	if err != nil {
		log.Errorf(log.SensitizeTrain, "Get sensitivity result failed, err:%v", err)
		return
	}

	log.Infof(log.SensitizeTrain, "Step3. Get sensitive parameter identification results successfully, and the details are as follows.%v", resultString)
}

func (tuner *Tuner) CreateTrainJob() error {
	log := fmt.Sprintf("%v/%v-%v.log", "/var/log/keentune", "keentuned-sensitize-train", tuner.Job)

	jobInfo := []string{
		tuner.Job, tuner.StartTime.Format(Format), NA, NA, fmt.Sprint(tuner.Trials), Run,
		log, config.GetSensitizeWorkPath(tuner.Job), tuner.Algorithm, tuner.Data,
	}

	tuner.backupConfFile()
	return file.Insert(getSensitizeJobFile(), jobInfo)
}

func (tuner *Tuner) initiateSensitization() error {
	uri := config.KeenTune.BrainIP + ":" + config.KeenTune.BrainPort + "/sensitize"

	ip, err := utils.GetExternalIP()
	if err != nil {
		log.Errorf(log.SensitizeTrain, "get local external ip err:%v", err)
		return err
	}

	reqInfo := map[string]interface{}{
		"data":      tuner.Data,
		"resp_ip":   ip,
		"resp_port": config.KeenTune.Port,
		"trials":    tuner.Trials,
		"explainer": tuner.Algorithm,
	}

	err = http.ResponseSuccess("POST", uri, reqInfo)
	if err != nil {
		log.Errorf(log.SensitizeTrain, "RemoteCall [POST] sensitize err:%v", err)
		return err
	}

	return nil
}

func (tuner *Tuner) getSensitivityResult() (string, error) {
	var sensitizeParams struct {
		Success bool        `json:"suc"`
		Head    string      `json:"head"`
		Result  [][]float64 `json:"result"`
		Msg     interface{} `json:"msg"`
	}

	config.IsInnerSensitizeRequests[1] = true
	defer func() { config.IsInnerSensitizeRequests[1] = false }()

	ticker := time.NewTicker(time.Duration(config.KeenTune.SensTimeout) * time.Minute)
	defer ticker.Stop()

	host := config.KeenTune.BrainIP + ":" + config.KeenTune.BrainPort

	errTimeout := fmt.Errorf("training wait for %v minutes timeout", config.KeenTune.SensTimeout)

	select {
	case resultBytes := <-config.SensitizeResultChan:
		log.Debugf(log.SensitizeTrain, "get sensitivity result:%s", resultBytes)
		if len(resultBytes) == 0 {
			return "", fmt.Errorf("get sensitivity result is nil")
		}

		if err := json.Unmarshal(resultBytes, &sensitizeParams); err != nil {
			return "", err
		}
	case <-StopSig:
		terminate(host)
		return "", fmt.Errorf("training is interrupted")
	case <-ticker.C:
		config.ServeTerminate <- true
		terminate(host)
		return "", errTimeout

	case <-config.ClientOffline:
		terminate(host)
		return "", errTimeout
	}

	if !sensitizeParams.Success {
		return "", fmt.Errorf("error msg:%v", sensitizeParams.Msg)
	}

	sensiResultCsv := fmt.Sprintf("%v/sensi_result.csv", config.GetSensitizePath(tuner.Job))
	sensiResultHeader := strings.Split(sensitizeParams.Head, ",")
	if len(sensitizeParams.Result) == 0 || len(sensitizeParams.Result[0]) != len(sensiResultHeader) {
		return "", fmt.Errorf("error msg:Header does not match param")
	}

	if !file.IsPathExist(sensiResultCsv) {
		err := file.CreatCSV(sensiResultCsv, sensiResultHeader)
		if err != nil {
			return "", fmt.Errorf("create sensitize jobs csv file: %v", err)
		}
	}

	resultSlice := saveSensitiveResult(sensitizeParams.Result, sensiResultHeader, sensiResultCsv)
	return utils.FormatInTable(resultSlice), nil

}

func saveSensitiveResult(result [][]float64, sensiResultHeader []string, sensiResultCsv string) [][]string {
	transposeCol := len(result) + 1
	transposeRow := len(sensiResultHeader) + 1
	var resultSlice = make([][]string, transposeRow)
	initSensitiveTable(transposeCol, resultSlice, sensiResultHeader)

	for colIdx, paramSlice := range result {
		var endInfo []string
		for rowIdx, param := range paramSlice {
			endInfo = append(endInfo, fmt.Sprint(param))
			if len(paramSlice) != transposeRow-1 {
				continue
			}
			resultSlice[rowIdx+1][colIdx+1] = fmt.Sprintf("%.4f", param)
		}
		file.Insert(sensiResultCsv, endInfo)
	}
	return resultSlice
}

func initSensitiveTable(transposeCol int, resultSlice [][]string, sensiResultHeader []string) {
	var transposeHeader []string
	for i := 0; i < transposeCol; i++ {
		if i == 0 {
			transposeHeader = append(transposeHeader, "param name")
			continue
		}
		transposeHeader = append(transposeHeader, fmt.Sprintf("round %v", i))
	}

	resultSlice[0] = transposeHeader

	for row, paramName := range sensiResultHeader {
		resultSlice[row+1] = make([]string, transposeCol)
		resultSlice[row+1][0] = paramName
	}
}

func (tuner *Tuner) dumpSensitivityResult(resultMap map[string]interface{}, recordName string) error {
	fileName := "sensi-" + recordName + ".json"
	if err := file.Dump2File(config.GetSensitizePath(""), fileName, resultMap); err != nil {
		log.Errorf(log.SensitizeTrain, "dump sensitivity result to file [%v] err:[%v] ", fileName, err)
		return err
	}

	return nil
}

func (tuner *Tuner) copyKnobsFile() error {
	tuneknobsFile := fmt.Sprintf("%v/knobs.json", config.GetTuningPath(tuner.Data))
	sensiknobsFile := fmt.Sprintf("%v/knobs.json", config.GetSensitizePath(tuner.Job))

	input, err := ioutil.ReadFile(tuneknobsFile)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = ioutil.WriteFile(sensiknobsFile, input, 0666)
	if err != nil {
		fmt.Println("Error creating", sensiknobsFile)
		fmt.Println(err)
		return err
	}
	return nil
}

