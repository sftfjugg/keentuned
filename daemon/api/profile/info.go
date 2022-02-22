package profile

import (
	"fmt"
	"io/ioutil"
	com "keentune/daemon/api/common"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
)

// Info run profile info service
func (s *Service) Info(fileName string, reply *string) error {
	defer func() {
		*reply = log.ClientLogMap[log.ProfInfo]
		log.ClearCliLog(log.ProfInfo)
	}()

	fullName := com.GetAbsolutePath(fileName, "profile", ".conf", "")
	if !file.IsPathExist(fullName) {
		log.Errorf(log.ProfInfo, "File %v is non-existent.", fileName)
		return fmt.Errorf("File %v is non-existent.", fileName)
	}

	infoDetialBytes, err := ioutil.ReadFile(fullName)
	if err != nil {
		log.Errorf(log.ProfInfo, "Read file: %v, err:%v\n", fullName, err)
		return fmt.Errorf("Read file: %v, err:%v", fullName, err)
	}

	log.Infof(log.ProfInfo, "%s", infoDetialBytes)
	return nil
}
