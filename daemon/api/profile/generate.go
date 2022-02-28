package profile

import (
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	m "keentune/daemon/modules"
)

// GenFlag ...
type GenFlag struct {
	Name   string
	Output string
	Force  bool
}

// Generate run profile generate service
func (s *Service) Generate(flag GenFlag, reply *string) error {
	defer func() {
		*reply = log.ClientLogMap[log.ProfGenerate]
		log.ClearCliLog(log.ProfGenerate)
	}()

	if file.IsPathExist(config.GetGenerateWorkPath(flag.Output)) && !flag.Force {
		log.Errorf(log.ProfGenerate, "Output File: %v exist and you have given up to overwrite it\n", flag.Name)
		return fmt.Errorf("Output File: %v  exist and you have given up to overwrite it", flag.Name)

	}

	fullName := config.GetProfileWorkPath(flag.Name)
	readMap, err := file.ConvertConfFileToJson(fullName)
	if err != nil {
		log.Errorf(log.ProfGenerate, "Convert file: %v, err:%v\n", flag.Name, err)
		return fmt.Errorf("Convert file: %v, err:%v", flag.Name, err)
	}

	totalParamMap, err := file.ReadFile2Map(fmt.Sprintf("%s/%s", config.KeenTune.Home, config.ParamAllFile))
	if err != nil {
		log.Errorf(log.ProfGenerate, "Read file: %v, err:%v\n", fmt.Sprintf("%s/%s", config.KeenTune.Home, config.ParamAllFile), err)
		return fmt.Errorf("Read file: %v, err:%v", fmt.Sprintf("%s/%s", config.KeenTune.Home, config.ParamAllFile), err)
	}

	m.AssembleParams(readMap, totalParamMap)

	if err := file.Dump2File(config.GetGenerateWorkPath(""), flag.Output, readMap); err != nil {
		log.Errorf(log.ProfGenerate, "Dump config info to json file [%v] err: %v", flag.Output, err)
		return fmt.Errorf("Dump json file: %v, err: %v", flag.Output, err)
	}

	log.Infof(log.ProfGenerate, "[ok] %v generate successfully", config.GetGenerateWorkPath(flag.Output))
	return nil
}
