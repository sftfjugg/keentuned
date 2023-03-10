package common

import (
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils/http"
	m "keentune/daemon/modules"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type ClientState struct {
	Status string `json:"status"`
}

const (
	// MaxReconnectionTime ...
	MaxReconnectionTime = 3
)

func HeartbeatCheck() error {
	_, _, err := StartCheck()
	if err != nil {
		return err
	}

	log.Info("", "\tKeenTuned heartbeat check : keentuned and client connection success")

	var clientName = new(string)
	go monitorClientStatus(IsClientOffline, clientName, nil)
	return nil
}

func monitorClientStatus(monitor interface{}, clientName *string, group []bool) {
	signalChan := make(chan os.Signal, 1)

	// notify system signal, such as ctrl + C
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	var faultCount int
	ticker := time.NewTicker(time.Duration(config.KeenTune.HeartbeatTime) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if monitorImply(monitor, clientName, group) {
				faultCount++
				log.Infof("", "keentuned detected that '%v' was offline for the %vth time", *clientName, faultCount)
			}

			if faultCount == MaxReconnectionTime {
				log.Info("", "Heartbeat Check the client is offline")
				config.ClientOffline <- true
				return
			}

		case <-config.ProgramNeedExit:
			log.Debug("", "Heartbeat Check program is finish")
			return

		case <-config.ServeTerminate:
			log.Debug("", "Heartbeat Timeout and service stop")
			return

		case <-signalChan:
			log.Debug("", "Heartbeat Check program is interrupt")

			if m.GetRunningTask() != "" {
				ResetJob()
				http.RemoteCall("GET", config.KeenTune.BrainIP+":"+config.KeenTune.BrainPort+"/end", nil)
			}
			return
		}
	}
}

func monitorImply(monitor interface{}, clientName *string, group []bool) bool {
	switch f := monitor.(type) {
	case func(*string) bool:
		return f(clientName)
	case func([]bool, *string) bool:
		return f(group, clientName)
	}

	return false
}

func IsClientOffline(clientName *string) bool {
	var offline = false
	offline = IsTargetOffline(clientName)

	benchStatus := IsBenchOffline(clientName)
	offline = offline || benchStatus

	// check brain
	brainOffline := IsBrainOffline(clientName)
	offline = offline || brainOffline

	*clientName = strings.TrimSuffix(*clientName, ", ")
	return offline
}

func IsTargetOffline(clientName *string) bool {
	var offline bool

	for _, group := range config.KeenTune.Group {
		for index, ip := range group.IPs {
			err := GetStatus(ip, group.Port)
			if err != nil {
				*clientName += fmt.Sprintf("target-group[%v] %v, ", group.GroupNo, index+1)
				offline = true
			}
		}
	}

	return offline
}

func IsBenchOffline(clientName *string) bool {
	var offline bool

	for groupID, benchGroup := range config.KeenTune.BenchGroup {
		for _, benchIp := range benchGroup.SrcIPs {
			benchAvl, _, _ := GetAVLAgentAddr(benchIp, benchGroup.SrcPort, benchGroup.DestIP)
			if !benchAvl {
				*clientName += fmt.Sprintf("bench-group[%v] %v, ", groupID+1, benchIp)
				offline = true
			}
		}
	}

	return offline
}

func StartCheck() (string, []string, error) {
	var okInfo = new(string)
	var warningInfo, domains []string
	checkTarget(&domains, &warningInfo, okInfo)

	checkBenchAndDest(&warningInfo, okInfo)

	checkBrain(&warningInfo, okInfo, domains)

	if len(warningInfo) > 0 {
		return *okInfo, warningInfo, fmt.Errorf("%s", strings.Join(warningInfo, "\t"))
	}

	return *okInfo, warningInfo, nil
}

func checkBrain(warningInfo *[]string, okInfo *string, domains []string) {
	_, algos, explainers, err := GetAVLDataAndAlgo()
	if err != nil {
		*warningInfo = append(*warningInfo, fmt.Sprintf("brain offline: %v\n", config.KeenTune.BrainIP))
	} else {
		*okInfo += fmt.Sprintf("brain: %v\n", config.KeenTune.BrainIP)
		if len(domains) > 0 {
			*okInfo += fmt.Sprintf("Avaliable parameter domain: %v\n", strings.Join(domains, ", "))
		}

		*okInfo += fmt.Sprintf("Avaliable algorithm : %v\n"+
			"Avaliable sensitivity algorithm : %v\n",
			strings.Join(algos, ", "), strings.Join(explainers, ", "))
	}
}

func checkBenchAndDest(warningInfo *[]string, okInfo *string) {
	for groupIdx, bench := range config.KeenTune.BenchGroup {
		var okIPs []string
		for _, ip := range bench.SrcIPs {
			isBenchOk, _, err := GetAVLAgentAddr(ip, bench.SrcPort, bench.DestIP)

			if err != nil || !isBenchOk {
				warning := fmt.Sprintf("bench-group[%v] source offline: %v\n", groupIdx+1, ip)
				*warningInfo = append(*warningInfo, warning)
				continue
			}

			okIPs = append(okIPs, ip)
		}

		if len(okIPs) > 0 {
			*okInfo += fmt.Sprintf("bench-group[%v]:\n"+
				"\tsource: %v\n"+
				"\tdestination: %v\n",
				groupIdx+1, strings.Join(okIPs, ", "), bench.DestIP)
			continue
		}

		warning := fmt.Sprintf("bench-group[%v] destination unreachable: %v\n", groupIdx+1, bench.DestIP)
		*warningInfo = append(*warningInfo, warning)
	}
}

func checkTarget(domains, warningInfo *[]string, okInfo *string) {
	for _, tg := range config.KeenTune.Group {
		var okIPs []string
		for _, ip := range tg.IPs {
			err := GetStatus(ip, tg.Port)
			if err != nil {
				warning := fmt.Sprintf("target-group[%v] offline: %v\n", tg.GroupNo, ip)
				*warningInfo = append(*warningInfo, warning)
				continue
			}

			okIPs = append(okIPs, ip)
		}

		if len(okIPs) > 0 {
			*okInfo += fmt.Sprintf("target-group[%v]: %v\n", tg.GroupNo, strings.Join(okIPs, ", "))
		}
	}
}

func IsBrainOffline(clientName *string) bool {
	_, _, _, err := GetAVLDataAndAlgo()
	if err != nil {
		*clientName += fmt.Sprintf("brain offline: %v", config.KeenTune.BrainIP)
		return true
	}

	return false
}

func checkSettableTarget(group []bool) error {
	inputLen := len(group)
	hasLen := len(config.KeenTune.Group)
	if inputLen != hasLen {
		return fmt.Errorf("target group out of range, has count %v, input count %v", hasLen, inputLen)
	}

	return nil
}

func IsSetTargetOffline(group []bool, clientName *string) bool {
	err := checkSettableTarget(group)
	if err != nil {
		*clientName += err.Error()
		return true
	}

	var offline bool
	for i, settable := range group {
		if !settable {
			continue
		}

		for _, ip := range config.KeenTune.Group[i].IPs {
			err := GetStatus(ip, config.KeenTune.Group[i].Port)
			if err != nil {
				*clientName += fmt.Sprintf("target-group[%v] %v, ", config.KeenTune.Group[i].GroupNo, ip)
				offline = true
			}
		}

	}

	return offline
}

func CheckBrainClient() error {
	clientName := new(string)
	if IsBrainOffline(clientName) {
		return fmt.Errorf("brain %v offline, please get it ready", config.KeenTune.BrainIP)
	}

	go monitorClientStatus(IsBrainOffline, clientName, nil)
	return nil
}

