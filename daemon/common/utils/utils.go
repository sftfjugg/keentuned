/*
Copyright © 2021 KeenTune

Package utils for daemon, this package contains the calculator, parsejson, utils, http. The package contains the parsing of the json file, calculate the suffix expressions, and the final result is saved at the top of the stack.Implement the device ip acquisition and validity judgment, print content string by color.
*/
package utils

import (
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"
)

// TimeSpend struct for time-consuming
type TimeSpend struct {
	Desc  string
	Count time.Duration
}

// Runtime Runtime Statistics
func Runtime(start time.Time) TimeSpend {
	tc := time.Since(start)
	return TimeSpend{
		Desc:  fmt.Sprintf("use time %.3fs", tc.Seconds()),
		Count: tc}
}

// Fluctuation count the fluctuation range of data slice by average
func Fluctuation(data []float32, average float32) string {
	if average == 0 {
		return "fluctuation = [average = 0, can not calculate]"
	}

	var ratios []float32
	for _, value := range data {
		ratios = append(ratios, (value-average)*100/average)
	}

	var minimum, maximum float32
	for _, value := range ratios {
		if value < minimum {
			minimum = value
		}

		if value > maximum {
			maximum = value
		}
	}

	if maximum > 0 {
		if minimum > 0 {
			return fmt.Sprintf("fluctuation = [+%.2f%%, +%.2f%%]", minimum, maximum)
		}

		return fmt.Sprintf("fluctuation = [%.2f%%, +%.2f%%]", minimum, maximum)
	}

	return fmt.Sprintf("fluctuation = [%.2f%%, %.2f%%]", minimum, maximum)
}

// IncreaseRatio calculate the percentage after the change relative to that before the change
func IncreaseRatio(after, before float32) float32 {
	if before == 0 {
		return 0.0
	}

	return (after - before) * 100 / before
}

// ColorString print content string by color
func ColorString(color string, content string) string {
	// 其中0x1B是标记，[开始定义颜色，1代表高亮，40代表黑色背景，32代表绿色前景，0代表恢复默认颜色
	switch strings.ToUpper(color) {
	case "RED":
		return fmt.Sprintf("%c[1;40;31m%s%c[0m", 0x1B, content, 0x1B)
	case "GREEN":
		return fmt.Sprintf("%c[1;40;32m%s%c[0m", 0x1B, content, 0x1B)
	case "YELLOW":
		return fmt.Sprintf("%c[1;40;33m%s%c[0m", 0x1B, content, 0x1B)
	default:
		return content
	}
}

// GetExternalIP Get the real IP address of the device
func GetExternalIP() (string, error) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("net.Interfaces failed, err: %v", err.Error())
	}

	for _, iface := range netInterfaces {
		if iface.Flags&net.FlagUp != 0 {
			addrs, err := iface.Addrs()
			if err != nil {
				return "", err
			}

			for _, address := range addrs {
				if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						return ipnet.IP.String(), nil
					}
				}
			}
		}
	}

	return "", fmt.Errorf("ip not found")
}

// CheckIPValidity argv 0: origin slice; return 0: valid IPs; 1: invalid IPs; 2: repeated IPs
func CheckIPValidity(origin []string) ([]string, []string, []string) {
	var valid, invalid, repeated []string
	var orgMap = make(map[string]bool, len(origin))
	for _, v := range origin {
		pureValue := strings.Trim(v, " ")
		if pureValue == "localhost" || pureValue == "127.0.0.1" || pureValue == "::1" {
			localIP := "localhost"
			if !orgMap[localIP] {
				valid = append(valid, pureValue)
				orgMap[localIP] = true
				continue
			}
			repeated = append(repeated, pureValue)
			continue
		}

		if !orgMap[pureValue] && isIP(pureValue) {
			orgMap[pureValue] = true
			valid = append(valid, pureValue)
			continue
		}
		invalid = append(invalid, pureValue)
	}

	return valid, invalid, repeated
}

func isIP(ip string) bool {
	address := net.ParseIP(ip)
	return address != nil
}

// FormatInTable format content to a table string
func FormatInTable(data [][]string, shift ...string) string {
	if len(data) == 0 || len(data[0]) == 0 {
		return ""
	}

	shiftStr := strings.Join(shift, "")

	var maxes = make([]int, len(data[0]))
	for _, row := range data {
		for j, column := range row {
			if maxes[j] < len(strings.Trim(column, " \t")) {
				maxes[j] = len(strings.Trim(column, " \t"))
			}
		}
	}

	var fmtInfo string
	for index, row := range data {
		if index == 0 {
			fmtInfo += decorateString(maxes, shiftStr)
		}

		if len(row) != len(maxes) {
			continue
		}

		fmtInfo += shiftStr + "|"
		for j, len := range maxes {
			fmtInfo += fmt.Sprintf("%-"+fmt.Sprintf("%v", len)+"s|", strings.Trim(row[j], " \t"))
		}

		fmtInfo += fmt.Sprintln()

		if index == 0 || index == len(data)-1 {
			fmtInfo += decorateString(maxes, shiftStr)
		}
	}

	return fmt.Sprintln() + strings.TrimSuffix(fmtInfo, "\n")
}

func decorateString(maxes []int, shiftStr string) string {
	var decorateStr = shiftStr + "+"
	for _, len := range maxes {
		decorateStr += strings.ReplaceAll(fmt.Sprintf("%-"+fmt.Sprintf("%v", len)+"s+", ""), " ", "-")
	}

	return decorateStr + fmt.Sprintln()
}

// RemoveRepeated ...
func RemoveRepeated(s []string) []string {
	var result []string
	m := make(map[string]bool)
	for _, v := range s {
		if _, ok := m[v]; !ok {
			result = append(result, v)
			m[v] = true
		}
	}
	return result
}

// DeepCopy deep copy from src to dst
func DeepCopy(dst, src interface{}) error {
	bytes, err := json.Marshal(src)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, dst)
}

// Ping detect ip whether is reachable
func Ping(ip string, port string) error {
	if ip != "localhost" {
		ipMatch := net.ParseIP(ip)
		if ipMatch == nil {
			return fmt.Errorf("ip %v is invalid", ip)
		}
	}

	if !isPingExist() {
		return dail(ip, port)
	}

	if pingTong(ip) {
		return nil
	}

	return fmt.Errorf("ip %v offline", ip)
}

func dail(ip, port string) error {
	addr := net.JoinHostPort(ip, port)
	connect, err := net.DialTimeout("tcp", addr, 3*time.Second)
	if err != nil || connect == nil {
		return fmt.Errorf("'%v:%v' is unreachable", ip, port)
	}

	connect.Close()

	return nil
}

func pingTong(ip string) bool {
	var success = "true"
	pingCmd := fmt.Sprintf("ping -c 1 -w 1 %s > /dev/null && echo true || echo false", ip)
	output, _ := exec.Command("/bin/bash", "-c", pingCmd).Output()

	return success == strings.TrimSpace(string(output))
}

func isPingExist() bool {
	var success = "true"
	whichPing := "which ping > /dev/null && echo true || echo false"
	output, _ := exec.Command("/bin/bash", "-c", whichPing).Output()

	return success == strings.TrimSpace(string(output))
}

