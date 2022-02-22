package param

import (
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	m "keentune/daemon/modules"
	"fmt"
)

// List run param list service
func (s *Service) List(flag string, reply *string) error {
	defer func() {
		*reply = log.ClientLogMap[log.ParamList]
		log.ClearCliLog(log.ParamList)
	}()

	paramHeader := "Parameter List"
	benchHeader := "Benchmark List"

	paramInfo, err1 := walkAndShow(m.GetParamHomePath(), ".json", false, paramHeader)
	benchInfo, err2 := walkAndShow(m.GetBenchHomePath(), ".json", false, benchHeader)
	if err1 != nil || err2 != nil {
		log.Errorf(log.ParamList, "Walk path failed, err1: %v and err2: %v", err1, err2)
		return nil
	}

	log.Infof(log.ParamList,"%v\n\n%v", paramInfo, benchInfo)

	return nil
}

func walkAndShow(filePath string, match string, isDir bool, header string, separator ...string) (string, error) {
	list, err := file.WalkFilePath(filePath, match, isDir, separator...)
	if err != nil {
		return "", fmt.Errorf("walk path: %v, err: %v", filePath, err)
	}

	return showList(list, header), nil
}

func showList(data []string, header string) string {
	var listInfo string
	for _, value := range data {
		listInfo += fmt.Sprintf("\n\t%v", value)
	}

	if listInfo != "" {
		return fmt.Sprintf("%v%v", header, listInfo)
	}

	return ""
}
