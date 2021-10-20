package param

import (
	"keentune/daemon/common/log"
	m "keentune/daemon/modules"
)

// Jobs run param jobs service
func (s *Service) Jobs(flag string, reply *string) error {
	defer func() {
		*reply = log.ClientLogMap[log.ParamJobs]
		log.ClearCliLog(log.ParamJobs)
	}()

	jobHeader := "Tune Jobs"	

	tuneJob, err := walkAndShow(m.GetTuningWorkPath("") + "/", "", true, jobHeader, "/generate/")
	if err != nil {
		log.Errorf(log.ParamJobs, "Walk path failed, err: %v", err)
		return nil
	}

	if tuneJob == "" {
		log.Infof(log.ParamJobs,"No job found")
		return nil
	}
	
	log.Infof(log.ParamJobs,"%v", tuneJob)
	return nil
}
