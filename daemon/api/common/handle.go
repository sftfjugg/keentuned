package common

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	"net/http"
	"os/exec"
	"strings"
)

const (
	// LimitBytes
	LimitBytes = 1024 * 1024 * 5
)

func registerRouter() {
	http.HandleFunc("/benchmark_result", handler)
	http.HandleFunc("/apply_result", handler)
	http.HandleFunc("/sensitize_result", handler)
	http.HandleFunc("/status", status)
	http.HandleFunc("/cmd", command)
}

func command(w http.ResponseWriter, r *http.Request) {
	var result = new(string)
	var err error
	defer func() {
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("{\"suc\": false, \"msg\": \"%v\"}", err.Error())))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("{\"suc\": true, \"msg\": \"%s\"}", *result)))
		return
	}()

	if strings.ToUpper(r.Method) != "POST" {
		err = fmt.Errorf("request method \"%v\" is not supported", r.Method)
		return
	}

	var cmd string
	cmd, err = getCmd(r.Body)
	if err != nil {
		return
	}

	err = execCmd(cmd, result)
	if err != nil {
		return
	}
}

func execCmd(inputCmd string, result *string) error {
	cmd := exec.Command("/bin/bash", "-c", inputCmd)
	// Create get command output pipeline
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("can not obtain stdout pipe for command:%s\n", err)
	}

	if err = cmd.Start(); err != nil {
		return fmt.Errorf("command start err: %v", err)
	}

	bytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		return fmt.Errorf("ReadAll Stdout: %v", err.Error())
	}

	if err = cmd.Wait(); err != nil {
		return fmt.Errorf("wait: %v, output %s\n", err.Error(), bytes)
	}	

	if strings.Contains(string(bytes), "Y(yes)/N(no)") {
		msg := strings.Split(string(bytes), "Y(yes)/N(no)")
		if len(msg) != 2 {
			return fmt.Errorf("get result %v", string(bytes))
		}

		*result = msg[1]
		return nil
	}

	*result = getMsg(string(bytes), inputCmd)
	return nil
}

func getMsg(origin, cmd string) string {
	if strings.Contains(cmd, "-h") || strings.Contains(cmd, "sensitize list") {
		return origin
	}

	pureMSg := strings.ReplaceAll(
		strings.ReplaceAll(
			strings.ReplaceAll(
				origin, "\x1b[1;40;32m", ""),
			"\x1b[0m", ""),
		"\x1b[1;40;31m", "")

	changeLinefeed := strings.ReplaceAll(pureMSg, "\n", "\\n")
	changeTab := strings.ReplaceAll(changeLinefeed, "\t", " ")
	return strings.TrimSuffix(changeTab, "\\n")
}

func getCmd(body io.ReadCloser) (string, error) {
	bytes, err := ioutil.ReadAll(&io.LimitedReader{R: body, N: LimitBytes})
	if err != nil {
		return "", err
	}

	var reqInfo struct {
		Cmd string `json:"cmd"`
	}

	err = json.Unmarshal(bytes, &reqInfo)
	if err != nil {
		return "", err
	}

	if strings.Contains(reqInfo.Cmd, "delete") {
		return "echo y|" + reqInfo.Cmd, nil
	}

	return reqInfo.Cmd, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	// check request method
	var msg string
	if strings.ToUpper(r.Method) != "POST" {
		msg = fmt.Sprintf("request method [%v] is not found", r.Method)
		log.Error("", msg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(msg))
		return
	}

	bytes, err := ioutil.ReadAll(&io.LimitedReader{R: r.Body, N: LimitBytes})
	defer report(r.URL.Path, bytes, err)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"suc": true, "msg": ""}`))
	return
}

func report(url string, value []byte, err error) {
	if err != nil {
		msg := fmt.Sprintf("read request info err:%v", err)
		log.Error("", "report value to chan err:%v", msg)
	}

	if strings.Contains(url, "benchmark_result") && config.IsInnerBenchRequests[1] {
		config.IsInnerBenchRequests[1] = false
		config.BenchmarkResultChan <- value
		return
	}

	if strings.Contains(url, "apply_result") {
		var applyResult struct {
			ID int `json:"target_id"`
		}
		err := json.Unmarshal(value, &applyResult)
		if err != nil {
			fmt.Printf("unmarshal apply target id err: %v", err)
			return
		}

		if config.IsInnerApplyRequests[applyResult.ID] && applyResult.ID > 0 {
			config.IsInnerApplyRequests[applyResult.ID] = false
			config.ApplyResultChan[applyResult.ID] <- value
		}

		return
	}

	if strings.Contains(url, "sensitize_result") && config.IsInnerSensitizeRequests[1] {
		config.IsInnerSensitizeRequests[1] = false
		config.SensitizeResultChan <- value
		return
	}
}

func status(w http.ResponseWriter, r *http.Request) {
	// check request method
	var msg string
	if strings.ToUpper(r.Method) != "GET" {
		msg = fmt.Sprintf("request method [%v] is not found", r.Method)
		log.Error("", msg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(msg))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "alive"}`))
	return
}
