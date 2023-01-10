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
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const (
	// LimitBytes ...
	LimitBytes = 1024 * 1024 * 5
)

var compiler = "\"([^\"]+)\""

func registerRouter() {
	http.HandleFunc("/benchmark_result", handler)
	http.HandleFunc("/apply_result", handler)
	http.HandleFunc("/sensitize_result", handler)
	http.HandleFunc("/status", status)
	http.HandleFunc("/cmd", command)
	http.HandleFunc("/write", write)
	http.HandleFunc("/read", read)
}

func write(w http.ResponseWriter, r *http.Request) {
	var result = new(string)
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
			w.Write([]byte(fmt.Sprintf("{\"suc\": false, \"msg\": \"%v\"}", getMsg(err.Error(), ""))))
			log.Errorf("", "write operation: %v", getMsg(err.Error(), ""))
			return
		}

		w.Write([]byte(fmt.Sprintf("{\"suc\": true, \"msg\": \"%s\"}", getMsg(*result, ""))))
		log.Infof("", "write operation: %v", *result)
	}()

	bytes, err := ioutil.ReadAll(&io.LimitedReader{R: r.Body, N: LimitBytes})
	if err != nil {
		return
	}

	var req struct {
		Name    string `json:"name"`
		Info    string `json:"info"`
		Replace string `json:"replace"`
	}

	err = json.Unmarshal(bytes, &req)
	if err != nil {
		err = fmt.Errorf("parse request info failed: %v", err)
		return
	}

	if req.Name == file.GetPlainName(config.GetKeenTunedConfPath("")) && strings.Contains(req.Info, "[brain]") {
		*result, err = config.UpdateKeentunedConf(req.Info)
		return
	}

	fullName := getFullPath(req.Name)

	parts := strings.Split(fullName, "/")
	if !file.IsPathExist(strings.Join(parts[:len(parts)-1], "/")) {
		os.MkdirAll(strings.Join(parts[:len(parts)-1], "/"), os.ModePerm)
	}

	err = ioutil.WriteFile(fullName, []byte(req.Info), 0755)
	if err != nil {
		return
	}

	if req.Replace != "" && req.Replace != req.Name {
		os.Remove(getFullPath(req.Replace))
	}

	*result = fmt.Sprintf("write file '%v' successfully.", req.Name)
	return
}

func getFullPath(name string) string {
	var fullName string
	if strings.HasPrefix(name, "/") {
		return name
	}

	if strings.Contains(name, "profile/") {
		fullName = fmt.Sprintf("%v/%v", config.KeenTune.DumpHome, name)
		return fullName
	}

	fullName = fmt.Sprintf("%v/profile/%v", config.KeenTune.DumpHome, name)
	return fullName
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

	if strings.Contains(url, "benchmark_result") {
		var benchResult struct {
			BenchID int `json:"bench_id"`
		}
		err := json.Unmarshal(value, &benchResult)
		if err != nil {
			fmt.Printf("unmarshal bench id err: %v", err)
			return
		}

		if config.IsInnerBenchRequests[benchResult.BenchID] && benchResult.BenchID > 0 {
			config.BenchmarkResultChan[benchResult.BenchID] <- value
		}
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
			config.ApplyResultChan[applyResult.ID] <- value
		}

		return
	}

	if strings.Contains(url, "sensitize_result") && config.IsInnerSensitizeRequests[1] {
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

func command(w http.ResponseWriter, r *http.Request) {
	var result = new(string)
	var err error
	var resp struct {
		Suc bool   `json:"suc"`
		Msg string `json:"msg"`
	}
	defer func() {
		w.WriteHeader(http.StatusOK)
		var errResp, okResp string
		if err != nil {
			resp.Suc = false
			resp.Msg = fmt.Sprint(err)
			errResp = fmt.Sprintf("{\"suc\": false, \"msg\": \"%v\"}", err.Error())
		} else {
			if strings.Contains(*result, "[ERROR]") {
				resp.Suc = false
			} else {
				resp.Suc = true
			}

			resp.Msg = *result
			okResp = fmt.Sprintf("{\"suc\": true, \"msg\": \"%s\"}", *result)
		}

		bytes, err := json.Marshal(resp)
		if err != nil {
			fmt.Println(err)
			switch resp.Suc {
			case true:
				w.Write([]byte(okResp))
			default:
				w.Write([]byte(errResp))
			}
			return
		}

		w.Write(bytes)
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
	stderr, _ := cmd.StderrPipe()
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("can not obtain stdout pipe for command: %s", err)
	}

	if err = cmd.Start(); err != nil {
		return fmt.Errorf("command start err: %v", err)
	}

	bytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		return fmt.Errorf("ReadAll Stdout: %v", err.Error())
	}

	if len(bytes) == 0 {
		bytes, _ = ioutil.ReadAll(stderr)
	}

	if err = cmd.Wait(); err != nil {
		parts := strings.Split(string(bytes), "failed, msg: ")
		if len(parts) > 0 {
			return fmt.Errorf("%s", getMsg(parts[len(parts)-1], inputCmd))
		}

		return fmt.Errorf("%s", getMsg(string(bytes), inputCmd))
	}

	if strings.Contains(string(bytes), "Y(yes)/N(no)") {
		msg := strings.Split(string(bytes), "Y(yes)/N(no)")
		if len(msg) != 2 {
			return fmt.Errorf("get result %v", string(bytes))
		}

		*result = getMsg(msg[1], inputCmd)
		return nil
	}

	*result = getMsg(string(bytes), inputCmd)

	return nil
}

func getMsg(origin, cmd string) string {
	if strings.HasSuffix(cmd, "-h") || strings.Contains(cmd, "jobs") {
		return origin
	}

	// replace color control special chars
	matchStr := "\u001B\\[1;40;3[1-3]m(.*?)\u001B\\[0m"
	pureMSg := origin
	matched, _ := regexp.MatchString(matchStr, pureMSg)
	if matched {
		re := regexp.MustCompile(matchStr)
		pureMSg = re.ReplaceAllString(strings.TrimSpace(origin), "$1")
	}

	// changeLinefeed := strings.ReplaceAll(pureMSg, "\n", "\\n")
	changeTab := strings.ReplaceAll(pureMSg, "\t", " ")
	return strings.ReplaceAll(strings.TrimSuffix(changeTab, "\\n"), "\"", "'")
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

	if strings.Contains(reqInfo.Cmd, "delete") || strings.Contains(reqInfo.Cmd, "dump") {
		return "echo y|" + reqInfo.Cmd, nil
	}

	return reqInfo.Cmd, nil
}


