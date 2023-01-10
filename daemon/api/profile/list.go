package profile

import (
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	"sort"
	"strings"
)

const (
	customPathName  = "- custom"
	defaultPathName = "- default"
)

// List run profile list service
func (s *Service) List(flag string, reply *string) error {
	defer func() {
		*reply = log.ClientLogMap[log.ProfList]
		log.ClearCliLog(log.ProfList)
	}()

	repeatedNameInfo, dirNames, proFileList, err := walkProfileAllFiles()
	if err != nil {
		log.Errorf(log.ProfList, "Walk file path failed: %v", err)
		return fmt.Errorf("Walk file path failed: %v", err)
	}

	activeDict := getActiveDict()

	var fileListInfo string
	for idx, name := range dirNames {
		if name == defaultPathName && len(proFileList[idx]) <= 1 {
			continue
		}

		if name != "" {
			fileListInfo += fmt.Sprintln(name)
		}

		var activeList, availableList string
		for _, value := range proFileList[idx] {
			if value == "active.conf" || value == "default.conf" {
				continue
			}

			activeInfo, find := activeDict[value]
			if find {
				activeList += fmt.Sprintln(utils.ColorString("GREEN", activeInfo))
				continue
			}

			availableList += fmt.Sprintf("\t[available]\t%v\n", value)
		}

		fileListInfo += activeList + availableList
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

func getActiveDict() map[string]string {
	activeFileName := config.GetProfileWorkPath("active.conf")
	records, _ := file.GetAllRecords(activeFileName)
	var dict = make(map[string]string)
	for _, record := range records {
		if len(record) == 2 {
			dict[record[0]] = fmt.Sprintf("\t[active]\t%v", strings.Join(record, "\ttarget_info: "))
		}
	}

	return dict
}

func walkProfileAllFiles() (string, []string, [][]string, error) {
	_, proFileList, err := file.WalkFilePath(config.GetProfileWorkPath(""), "")
	if err != nil {
		return "", nil, nil, fmt.Errorf("walk dump folder failed :%v", err)
	}

	fullPaths, homeFileList, err := file.WalkFilePath(config.GetProfileHomePath(""), ".conf")
	if err != nil {
		return "", nil, nil, fmt.Errorf("walk home folder failed :%v", err)
	}

	repeatedNameInfo := getRepeatedNameInfo(homeFileList, fullPaths)

	dirNames, totalList := getProfileDirTree(proFileList, fullPaths)
	if len(dirNames) != len(totalList) {
		return repeatedNameInfo, dirNames, totalList, fmt.Errorf("get profile dir tree failed")
	}

	return repeatedNameInfo, dirNames, totalList, nil
}

func getProfileDirTree(wspProfiles, paths []string) ([]string, [][]string) {
	homePath := config.GetProfileHomePath("")
	var profileDict = make(map[string][]string)
	var dirNames, homeDirNames []string

	// list custom profile group file first
	profileDict[customPathName] = wspProfiles
	dirNames = append(dirNames, customPathName)

	for _, path := range paths {
		parts := strings.Split(strings.TrimPrefix(path, homePath), "/")
		length := len(parts)
		switch length {
		case 0, -1:
			continue
		case 1:
			if len(profileDict[defaultPathName]) == 0 {
				homeDirNames = append(homeDirNames, defaultPathName)
			}

			profileDict[defaultPathName] = append(profileDict[defaultPathName], parts[0])
		case 2:
			dirName := fmt.Sprintf("- %v", parts[0])
			if len(profileDict[dirName]) == 0 {
				homeDirNames = append(homeDirNames, dirName)
			}

			profileDict[dirName] = append(profileDict[dirName], parts[1])
		default:
			dirName := fmt.Sprintf("- %v", parts[length-2])
			if len(profileDict[dirName]) == 0 {
				homeDirNames = append(homeDirNames, dirName)
			}

			profileDict[dirName] = append(profileDict[dirName], parts[length-1])
		}
	}

	sort.Strings(homeDirNames)
	dirNames = append(dirNames, homeDirNames...)

	var totalFileList [][]string
	for _, name := range dirNames {
		totalFileList = append(totalFileList, profileDict[name])
	}

	return dirNames, totalFileList
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


