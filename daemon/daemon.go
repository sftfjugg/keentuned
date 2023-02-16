package main

import (
	"fmt"
	com "keentune/daemon/api/common"
	"keentune/daemon/api/param"
	"keentune/daemon/api/profile"
	"keentune/daemon/api/sensitize"
	"keentune/daemon/api/system"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	m "keentune/daemon/modules"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	config.Init()
	log.Init()
	com.ResetJob()

	m.StopSig = make(chan os.Signal, 1)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	go com.KeenTunedService(quit)
	rpc.RegisterName("system", new(system.Service))
	rpc.RegisterName("param", new(param.Service))
	rpc.RegisterName("profile", new(profile.Service))
	rpc.RegisterName("sensitize", new(sensitize.Service))

	listener, err := net.Listen("tcp", ":9870")
	if err != nil {
		log.Errorf(log.ParamTune, "ListenTCP error:%v", err)
		os.Exit(1)
	}

	go mkWorkDir()

	go showStart()

	go profile.SetDefault()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Errorf(log.ParamTune, "Accept error:%v", err)
			continue
		}
		rpc.ServeConn(conn)
	}
}

func mkWorkDir() {
	if !file.IsPathExist(config.GetProfileWorkPath("")) {
		os.MkdirAll(config.GetProfileWorkPath(""), os.ModePerm)
	}

	if !file.IsPathExist(config.GetTuningWorkPath("")) {
		os.MkdirAll(config.GetTuningWorkPath(""), os.ModePerm)
	}
	if !file.IsPathExist(config.GetTuningPath("")) {
		os.MkdirAll(config.GetTuningPath(""), os.ModePerm)
	}

	if !file.IsPathExist(config.GetSensitizePath("")) {
		os.MkdirAll(config.GetSensitizePath(""), os.ModePerm)
	}
	tuningCsv := config.GetDumpPath(config.TuneCsv)
	if !file.IsPathExist(tuningCsv) {
		err := file.CreatCSV(tuningCsv, m.TuneJobHeader)
		if err != nil {
			fmt.Printf("%v create tuning jobs csv file: %v", utils.ColorString("red", "[ERROR]"), err)
			os.Exit(1)
		}
	}
	sensitizeCsv := config.GetDumpPath(config.SensitizeCsv)
	if !file.IsPathExist(sensitizeCsv) {
		err := file.CreatCSV(sensitizeCsv, m.SensitizeJobHeader)
		if err != nil {
			fmt.Printf("%v create sensitize jobs csv file: %v", utils.ColorString("red", "[ERROR]"), err)
			os.Exit(1)
		}
	}
	activeConf := config.GetProfileWorkPath("active.conf")
	if !file.IsPathExist(activeConf) {
		fp, _ := os.Create(activeConf)
		if fp != nil {
			fp.Close()
		}
	}
}

func showStart() {
	fmt.Printf("Keentune Home: %v\nKeentune Workspace: %v\n", config.KeenTune.Home, config.KeenTune.DumpHome)

	suc, warn, _ := com.StartCheck()

	if len(suc) > 0 {
		fmt.Printf("%v  Loading succeeded:\n%v\n", utils.ColorString("green", "[ok]"), suc)
	}

	for _, detail := range warn {
		fmt.Printf("%v %v", utils.ColorString("yellow", "[Warning]"), detail)
	}

	notifySystemd()
	fmt.Println("KeenTune daemon running...")

}

func notifySystemd() {
	socketAddr := &net.UnixAddr{
		Name: os.Getenv("NOTIFY_SOCKET"),
		Net:  "unixgram",
	}

	if socketAddr.Name == "" {
		return
	}

	conn, err := net.DialUnix(socketAddr.Net, nil, socketAddr)
	if err != nil {
		return
	}

	_, err = conn.Write([]byte("READY=1"))
	if err != nil {
		return
	}
}


