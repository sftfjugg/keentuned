package main

import (
	"context"
	"fmt"
	"log"
	"net/rpc"
	"os"
)

// TuneFlag tune options
type TuneFlag struct {
	ParamMap  string
	Name      string // job specific name
	Round     int
	BenchConf string
	ParamConf []string
	Verbose   bool
	Log       string // log file
}

// DumpFlag ...
type DumpFlag struct {
	Name   string
	Output string
	Force  bool
}

type SetFlag struct {
	Name string
}

type TrainFlag struct {
	Output string
	Data   string
	Trials int
	Force  bool
	Log    string // log file
}

type DeleteFlag struct {
	Name  string
	Cmd   string
	Force bool
}

type RollbackFlag struct {
	Cmd string
}

type BenchmarkFlag struct {
	Round     int
	BenchConf string
	Name      string
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
		fmt.Printf("%v %v failed, msg: %v\n", ColorString("red", "[ERROR]"), callName, err)
		os.Exit(1)
	}

	fmt.Printf("%v", reply)
	return
}

func RunTuneRemote(ctx context.Context, flag TuneFlag) {
	remoteImpl("param.Tune", flag)

	fmt.Printf("%v Running Param Tune Success.\n", ColorString("green", "[ok]"))
	fmt.Printf("\n\titeration: %v\n\tname: %v\n\tparam: %v\n\tbench: %v\n", flag.Round, flag.Name, flag.ParamConf, flag.BenchConf)
	fmt.Printf("\n\tsee more details by log file:  \"%v\"\n", flag.Log)
	return
}

func RunDumpRemote(ctx context.Context, flag DumpFlag) {
	remoteImpl("param.Dump", flag)
}

func RunListRemote(ctx context.Context, flag string) {
	remoteImpl(fmt.Sprintf("%s.List", flag), flag)
}

func RunRollbackRemote(ctx context.Context, flag RollbackFlag) {
	remoteImpl(fmt.Sprintf("%s.Rollback", flag.Cmd), flag)
}

func RunDeleteRemote(ctx context.Context, flag DeleteFlag) {
	fmt.Printf("%s %s '%s' ?Y(yes)/N(no)", ColorString("yellow", "[Warning]"), deleteTips, flag.Name)
	if !confirm() {
		fmt.Println("[-] Give Up Delete")
		return
	}

	remoteImpl(fmt.Sprintf("%s.Delete", flag.Cmd), flag)
}

func RunInfoRemote(ctx context.Context, flag string) {
	remoteImpl("profile.Info", flag)
}

func RunSetRemote(ctx context.Context, flag SetFlag) {
	remoteImpl("profile.Set", flag)
}

func RunGenerateRemote(ctx context.Context, flag DumpFlag) {
	remoteImpl("profile.Generate", flag)
}

func RunCollectRemote(ctx context.Context, flag TuneFlag) {
	remoteImpl("sensitize.Collect", flag)

	fmt.Printf("%v Running Sensitize Collect Success.\n", ColorString("green", "[ok]"))
	fmt.Printf("\n\titeration: %v\n\tname: %v\n\tparam: %v\n\tbench: %v\n", flag.Round, flag.Name, flag.ParamConf, flag.BenchConf)
	fmt.Printf("\n\tsee more details by log file:  \"%v\"\n", flag.Log)
	return
}

func RunTrainRemote(ctx context.Context, flag TrainFlag) {
	fmt.Printf("%s %s", ColorString("yellow", "[Warning]"), fmt.Sprintf(outputTips, "trained result"))
	flag.Force = confirm()
	remoteImpl("sensitize.Train", flag)

	fmt.Printf("%v Running Sensitize Train Success.\n", ColorString("green", "[ok]"))
	fmt.Printf("\n\ttrials: %v\n\tdata: %v\n\toutput: %v\n", flag.Trials, flag.Data, flag.Output)
	fmt.Printf("\n\tsee more detailsby log file:  \"%v\"\n", flag.Log)
	return
}

func StopRemote(ctx context.Context, flag string) {
	var job string
	if flag == "param" {
		job = "parameter optimization"
	}

	if flag == "sensitize" {
		job = "sensibility identification"
	}

	fmt.Printf("%v Abort %v job.\n", ColorString("yellow", "[Warning]"), job)
	remoteImpl(fmt.Sprintf("%s.Stop", flag), flag)
}

func RunJobsRemote(ctx context.Context) {
	remoteImpl("param.Jobs", "")
}

func RunBenchRemote(ctx context.Context, flag BenchmarkFlag) {
	remoteImpl("system.Benchmark", flag)
}

