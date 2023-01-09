package config

import (
	"fmt"
	"keentune/daemon/common/file"
	"strings"
)

var TuneTempConf = "/var/keentune/tuning_workspace/temp.conf"
var SensitizeTempConf = "/var/keentune/sensitize_workspace/temp.conf"

const (
	TuneCsv      = "tuning_jobs.csv"
	SensitizeCsv = "sensitize_jobs.csv"
)

func GetTuningWorkPath(fileName string) string {
	return assembleFilePath(KeenTune.DumpHome, "parameter", fileName)
}

func GetGenerateWorkPath(fileName string) string {
	return assembleFilePath(KeenTune.DumpHome, "parameter/generate", fileName)
}

func GetBenchHomePath() string {
	return assembleFilePath(KeenTune.Home, "benchmark", "")
}

func GetProfileWorkPath(fileName string) string {
	return assembleFilePath(KeenTune.DumpHome, "profile", fileName)
}
func GetDumpPath(fileName string) string {
	return assembleFilePath(KeenTune.DumpHome, "", fileName)
}

func GetTuningPath(jobName string) string {
	return assembleFilePath(KeenTune.DumpHome, "tuning_workspace", jobName)
}
func GetSensitizePath(jobName string) string {
	return assembleFilePath(KeenTune.DumpHome, "sensitize_workspace", jobName)
}

func GetSensitizeWorkPath(fileName string) string {
	return assembleFilePath(KeenTune.DumpHome, "sensitize_workspace", fileName)
}

func GetParamHomePath() string {
	return assembleFilePath(KeenTune.Home, "parameter", "") + "/"
}

func GetProfileHomePath(fileName string) string {
	profHome := fmt.Sprintf("%s/%s", KeenTune.Home, "profile")
	if fileName == "" {
		return profHome + "/"
	}

	// absolute path
	if strings.HasPrefix(fileName, "/") && strings.Count(strings.Trim(fileName, "/"), "/") > 1 {
		return fileName
	}

	// file
	parts := strings.Split(fileName, "/")
	if len(parts) == 1 {
		profileFile, err := file.GetWalkPath(GetProfileHomePath(""), fileName)
		if err != nil {
			return fileName
		}

		return profileFile
	}

	// relative path
	return fmt.Sprintf("%v%v", GetProfileHomePath(""), strings.TrimPrefix(fileName, "profile/"))
}

func GetSensitizeHomePath(fileName string) string {
	return assembleFilePath(KeenTune.Home, "sensitize", "") + "/"
}

func GetDumpCSVPath() string {
	return assembleFilePath(KeenTune.DumpHome, "csv", "")
}

func assembleFilePath(prefix, partition, fileName string) string {
	if fileName == "" {
		return fmt.Sprintf("%s/%s", prefix, partition)
	}

	// absolute path
	if strings.HasPrefix(fileName, "/") && strings.Count(strings.Trim(fileName, "/"), "/") > 1 {
		return fileName
	}

	// relative path
	if strings.Contains(fileName, fmt.Sprintf("%v/", partition)) {
		parts := strings.Split(fileName, fmt.Sprintf("%v/", partition))
		return fmt.Sprintf("%s/%s/%s", prefix, partition, parts[len(parts)-1])
	}

	// file
	return fmt.Sprintf("%s/%s/%s", prefix, partition, strings.TrimPrefix(fileName, "/"))
}

/* GetAbsolutePath  fileName support absolute path, relative path, file.
e.g.
	file: param.json
	relative path: parameter/param.json
	absolute path: /etc/keentune/parameter/param.json
*/
func GetAbsolutePath(fileName, class, fileType, extraSufix string) string {
	if fileName == "" {
		return fileName
	}

	// Absolute path, start with "/"
	if string(fileName[0]) == "/" {
		if strings.Contains(fileName, fileType) {
			return fileName
		}

		parts := strings.Split(fileName, "/")
		partLen := len(parts)

		return fmt.Sprintf("%s/%s%s", fileName, parts[partLen-1], extraSufix)
	}

	// Relative path, start with "./" or other
	var relativePath string
	relativePath = strings.Trim(fileName, "./")
	parts := strings.Split(relativePath, "/")
	partLen := len(parts)

	if file.IsPathExist(GetGenerateWorkPath(fmt.Sprintf("%s%s", strings.TrimSuffix(parts[partLen-1], ".json"), ".json"))) && fileType == ".json" {
		return GetGenerateWorkPath(fmt.Sprintf("%s%s", strings.TrimSuffix(parts[partLen-1], ".json"), ".json"))
	}

	var workPath string

	switch partLen {
	// Only a file name, work directory has priority
	case 1:
		if strings.Contains(parts[0], fileType) {
			workPath = fmt.Sprintf("%s/%s/%s", KeenTune.DumpHome, class, parts[0])
			if file.IsPathExist(workPath) {
				return workPath
			}

			return fmt.Sprintf("%s/%s/%s", KeenTune.Home, class, parts[0])
		}

		return fmt.Sprintf("%s/%s/%s/%s%s", KeenTune.DumpHome, class, parts[0], parts[0], extraSufix)
	// File relative path, work directory has priority
	default:
		// If the first element of the split has the same name as the specified class, then it will Trim the class+"/"
		if strings.Contains(parts[partLen-1], fileType) {
			workPath = fmt.Sprintf("%s/%s/%s", KeenTune.DumpHome, class, strings.TrimPrefix(relativePath, fmt.Sprintf("%s/", class)))
			if file.IsPathExist(workPath) {
				return workPath
			}

			return fmt.Sprintf("%s/%s/%s", KeenTune.Home, class, strings.TrimPrefix(relativePath, fmt.Sprintf("%s/", class)))
		}

		return fmt.Sprintf("%s/%s/%s/%s%s", KeenTune.DumpHome, class, strings.TrimPrefix(relativePath, fmt.Sprintf("%s/", class)), parts[partLen-1], extraSufix)
	}
}

func GetBenchJsonPath(fileName string) string {
	if fileName == "" || string(fileName[0]) == "/" {
		return fileName
	}

	parts := strings.Split(fileName, "/")
	if len(parts) == 1 {
		benchPath, err := file.GetWalkPath(GetBenchHomePath(), fileName)
		if err != nil {
			return fileName
		}

		return benchPath
	}

	return fmt.Sprintf("%v/%v", GetBenchHomePath(), strings.TrimPrefix(fileName, "benchmark/"))
}

func GetKeenTunedConfPath(origin string) string {
	if origin == "" {
		return keentuneConfigFile
	}

	if strings.HasPrefix(origin, "/") {
		return origin
	}

	fullPath := fmt.Sprintf("%v/conf/%v", KeenTune.Home, strings.TrimPrefix(strings.TrimPrefix(origin, "keentune/"), "conf/"))
	return fullPath
}

func GetProfilePath(fileName string) string {
	if strings.HasPrefix(fileName, "/") && file.IsPathExist(fileName) {
		return fileName
	}

	workPath := GetProfileWorkPath(fileName)
	if file.IsPathExist(workPath) {
		return workPath
	}

	homePath := GetProfileHomePath(fileName)
	if file.IsPathExist(homePath) {
		return homePath
	}

	return ""
}


