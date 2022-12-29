package profile

import (
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	"strings"
)

// List run profile list service
func (s *Service) List(flag string, reply *string) error {
	defer func() {
		*reply = log.ClientLogMap[log.ProfList]
		log.ClearCliLog(log.ProfList)
	}()

	repeatedNameInfo, proFileList, err := walkProfileAllFiles()
	if err != nil {
		log.Errorf(log.ProfList, "Walk file path failed: %v", err)
		return fmt.Errorf("Walk file path failed: %v", err)
	}

	var fileListInfo string
	activeFileName := config.GetProfileWorkPath("active.conf")
	records, _ := file.GetAllRecords(activeFileName)

	for _, value := range proFileList {
		if value == "active.conf" || value == "default.conf" {
			continue
		}

		activeFlag := false
		for _, record := range records {
			if len(record) == 2 && record[0] == value {
				activeInfo := fmt.Sprintf("[active]\t%v", strings.Join(record, "\ttarget_info: "))
				activeFlag = true
				fileListInfo += fmt.Sprintln(utils.ColorString("GREEN", activeInfo))
				break
			}
		}

		if !activeFlag {
			fileListInfo += fmt.Sprintf("[available]\t%v\n", value)
		}
	}

	if len(fileListInfo) == 0 {
		log.Info(log.ProfList, "There is no profile configuration file exists, please execute keentune param dump first.")
		return nil
	}

	log.Infof(log.ProfList, "%v", strings.TrimSuffix(fileListInfo, "\n"))

	if len(repeatedNameInfo) != 0 {
		log.Warnf(log.ProfList, "Found the same name files exist. Please handle it manually. See details:\n %v", repeatedNameInfo)
	}

	return nil
}

func walkProfileAllFiles() (string, []string, error) {
	_, proFileList, err := file.WalkFilePath(config.GetProfileWorkPath(""), "")
	if err != nil {
		return "", proFileList, fmt.Errorf("walk dump folder failed :%v", err)
	}

	fullPaths, homeFileList, err := file.WalkFilePath(config.GetProfileHomePath(""), ".conf")
	if err != nil {
		return "", proFileList, fmt.Errorf("walk home folder failed :%v", err)
	}

	repeatedNameInfo := getRepeatedNameInfo(homeFileList, fullPaths)

	proFileList = append(proFileList, homeFileList...)
	return repeatedNameInfo, proFileList, nil
}

func getRepeatedNameInfo(names, fullPaths []string) string {
	fileNameDict := make(map[string][]string)
	for curIdx, name := range names {
		fileNameDict[name] = append(fileNameDict[name], fullPaths[curIdx])
	}

	var warningInfo string
	for name, paths := range fileNameDict {
		if len(paths) > 1 {
			warningInfo += fmt.Sprintf("\t %v found in %v\n", name, strings.Join(paths, ", "))
		}
	}

	return warningInfo
}


