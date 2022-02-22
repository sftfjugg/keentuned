package profile

import (
	"fmt"
	com "keentune/daemon/api/common"
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	m "keentune/daemon/modules"
)

// Rollback run profile rollback service
func (s *Service) Rollback(flag com.RollbackFlag, reply *string) error {
	if com.IsApplying() {
		return fmt.Errorf("operation does not support, job %v is running", com.GetRunningTask())
	}

	defer func() {
		*reply = log.ClientLogMap[log.ProfRollback]
		log.ClearCliLog(log.ProfRollback)
	}()

	err := m.Rollback(log.ProfRollback)
	if err!=nil {
		return fmt.Errorf("Rollback details:\n%v", err)
	}

	fileName := config.GetProfileWorkPath("active.conf")
	if err := updateActiveFile(fileName, []byte{}); err != nil {
		log.Errorf(log.ProfRollback, "Update active file failed, err:%v", err)
		return fmt.Errorf("Update active file failed, err:%v", err)
	}

	log.Infof(log.ProfRollback, fmt.Sprintf("[ok] %v rollback successfully", flag.Cmd))
	return nil
}
