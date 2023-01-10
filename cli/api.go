package main

import (
	"fmt"
	"keentune/daemon/common/utils"
	"log"
	"net/rpc"
	"os"
	"strings"
)

// TuneFlag tune options
type TuneFlag struct {
	Name    string // job specific name
	Round   int
	Verbose bool
	Log     string // log file
}

// DumpFlag ...
type DumpFlag struct {
	Name   string
	Output []string
	Force  bool
}

// GenFlag ...
type GenFlag struct {
	Name   string
	Output string
	Force  bool
}

// SetFlag ...
type SetFlag struct {
	Group    []bool
	ConfFile []string
}

// TrainFlag ...
type TrainFlag struct {
	Job    string
	Data   string
	Trials int
	Force  bool
	Log    string // log file
}

// DeleteFlag ...
type DeleteFlag struct {
	Name  string
	Cmd   string
	Force bool
}

// RollbackFlag ...
type RollbackFlag struct {
	Cmd string
}

// BenchmarkFlag ...
type BenchmarkFlag struct {
	Round     int
	BenchConf string
	Name      string
}

type MigrateFlag struct {
	Filepath string
}

var (
	outputTips = "If the %v name is duplicated, overwrite? Y(yes)/N(no)"
	deleteTips = "Are you sure you want to permanently delete job data"
)

func remoteImpl(callName string, flag interface{}) {
	client, err := rpc.Dial("tcp", "localhost:9870")
	if err != nil {
		log.Fatal("dialing:", err)
	}

	var reply string
	err = client.Call(callName, flag, &reply)
	if err != nil {
		fmt.Printf("%v %v\n", ColorString("red", "[ERROR]"), err)
		os.Exit(1)
	}

	fmt.Printf("%v", reply)
	if strings.Contains(reply, "[ERROR]") ||
		strings.Contains(reply, utils.ColorString("red", "[ERROR]")) ||
		strings.Contains(reply, utils.ColorString("red", "ERROR")) {
		os.Exit(1)
	}

	return
}

// RunTuneRemote ...
func RunTuneRemote(flag TuneFlag) {
	remoteImpl("param.Tune", flag)

	fmt.Printf("%v Running Param Tune Success.\n", ColorString("green", "[ok]"))
	fmt.Printf("\n\titeration: %v\n\tname: %v\n", flag.Round, flag.Name)
	fmt.Printf("\n\tsee more details by log file: \"%v\"\n", flag.Log)
	return
}

// RunDumpRemote ...
func RunDumpRemote(flag DumpFlag) {
	remoteImpl("param.Dump", flag)
}

// RunListRemote ...
func RunListRemote(flag string) {
	remoteImpl(fmt.Sprintf("%s.List", flag), flag)
}

// RunRollbackRemote ...
func RunRollbackRemote(flag RollbackFlag) {
	remoteImpl(fmt.Sprintf("%s.Rollback", flag.Cmd), flag)
}

// RunDeleteRemote ...
func RunDeleteRemote(flag DeleteFlag) {
	remoteImpl(fmt.Sprintf("%s.Delete", flag.Cmd), flag)
}

// RunInfoRemote ...
func RunInfoRemote(flag string) {
	remoteImpl("profile.Info", flag)
}

// RunSetRemote ...
func RunSetRemote(flag SetFlag) {
	remoteImpl("profile.Set", flag)
}

// RunGenerateRemote ...
func RunGenerateRemote(flag GenFlag) {
	remoteImpl("profile.Generate", flag)
}

// RunTrainRemote ...
func RunTrainRemote(flag TrainFlag) {
	remoteImpl("sensitize.Train", flag)

	fmt.Printf("%v Running Sensitize Train Success.\n", ColorString("green", "[ok]"))
	fmt.Printf("\n\ttrials: %v\n\tdata: %v\n\tjob: %v\n", flag.Trials, flag.Data, flag.Job)
	fmt.Printf("\n\tsee more detailsby log file:  \"%v\"\n", flag.Log)
	return
}

// StopRemote ...
func StopRemote(flag string) {
	remoteImpl(fmt.Sprintf("%s.Stop", flag), flag)
}

// RunJobsRemote ...
func RunJobsRemote(flag string) {
	remoteImpl(fmt.Sprintf("%s.Jobs", flag), flag)
}

// RunBenchRemote ...
func RunBenchRemote(flag BenchmarkFlag) {
	remoteImpl("system.Benchmark", flag)
}

// RunRollbackAllRemote ...
func RunRollbackAllRemote() {
	remoteImpl("system.RollbackAll", "")
}

// RunInitRemote ...
func RunInitRemote() {
	remoteImpl("system.Init", "")
}


