/*
Copyright © 2021 KeenTune

Package log for daemon, this package contains the log. The package defines the command name, defines the log level, mainly using warning, error, info three levels, to achieve the log output screen and file two ways.
*/
package log

import (
	"fmt"
	"io"
	"keentune/daemon/common/config"
	"keentune/daemon/common/utils"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
)

// ConsoleLevel Console Log Level
const ConsoleLevel = "ERROR"

var fLogger *logrus.Logger
var cLogger *logrus.Logger

// fLogInst file log instance
var fLogInst Logger

// cLogInst console log instance
var cLogInst Logger

// Logger log instance struct
type Logger struct {
	entry *logrus.Entry
}

// logFormater Log custom format
type logFormater struct {
	TimestampFormat string
	LogFormat       string
	file            string
	funcName        string
	line            int
}

// consoleLogFormater console log custom format
type consoleLogFormater struct {
	LogFormat string
	file      string
	funcName  string
	line      int
}

// command name
const (
	ParamDump     = "param dump"
	ParamList     = "param list"
	ParamDel      = "param delete"
	ParamRollback = "param rollback"
	ParamJobs     = "param jobs"

	ProfInfo     = "profile info"
	ProfGenerate = "profile generate"
	ProfSet      = "profile set"
	ProfList     = "profile list"
	ProfDel      = "profile delete"
	ProfRollback = "profile rollback"

	SensitizeDel  = "sensitize delete"
	SensitizeJobs = "sensitize jobs"
	Benchmark     = "benchmark"
	RollbackAll   = "rollback all"
)

// record task to Log file 
var (
	ParamTune      = "param tune"
	SensitizeTrain = "sensitize train"
)

// ClientLogMap ...
var ClientLogMap = make(map[string]string)

func getLevel(lvl string) logrus.Level {
	ret, err := logrus.ParseLevel(lvl)
	if err != nil {
		fmt.Printf("[%v] is not as expected, return default info level]\n", lvl)
		return logrus.InfoLevel
	}
	return ret
}

//  Init ...
func Init() {
	fLogger = &logrus.Logger{Level: getLevel(config.KeenTune.LogFileLvl)}
	cLogger = &logrus.Logger{Level: getLevel(ConsoleLevel)}
	fLogInst = Logger{entry: logrus.NewEntry(fLogger)}
	cLogInst = Logger{entry: logrus.NewEntry(cLogger)}

	fileName := fmt.Sprintf("%v/log/keentune/%s", "/var", config.KeenTune.FileName)

	// 1 set log segmentation method
	writer, err := rotatelogs.New(
		fileName+".%Y-%m-%d",
		rotatelogs.WithLinkName(fileName),
		rotatelogs.WithRotationTime(time.Duration(config.KeenTune.Interval*24)*time.Hour),
		rotatelogs.WithRotationCount(uint(config.KeenTune.BackupCount)),
	)
	if err != nil {
		fmt.Printf("failed to create rotatelogs: %s", err)
		os.Exit(1)
	}

	// 2 set file log
	fLogger.SetOutput(writer)

	// 3 set standard output log
	cLogger.SetOutput(os.Stdout)
}

//  Format define the file log detail
func (s *logFormater) Format(entry *logrus.Entry) ([]byte, error) {
	level := strings.ToUpper(entry.Level.String())
	msg := fmt.Sprintf(s.LogFormat, level, s.TimestampFormat, entry.Message, strings.Trim(s.file+s.funcName, "."), s.line)

	return []byte(removeColor(msg)), nil
}

func removeColor(msg string) string {
	// replace color control special chars
	matchStr := "\u001B\\[1;40;3[1-3]m(.*?)\u001B\\[0m"
	pureMSg := msg
	matched, _ := regexp.MatchString(matchStr, pureMSg)
	if matched {
		re := regexp.MustCompile(matchStr)
		pureMSg = re.ReplaceAllString(msg, "$1")
	}

	return pureMSg
}

//  Format define the console log detail
func (s *consoleLogFormater) Format(entry *logrus.Entry) ([]byte, error) {
	level := strings.ToUpper(entry.Level.String())
	var msg string
	switch level {
	case "ERROR", "WARNING":
		msg = fmt.Sprintf(s.LogFormat, level, entry.Message, strings.Trim(s.file+s.funcName, "."), s.line)
	default:
		msg = fmt.Sprintf(s.LogFormat, entry.Message)
	}

	return []byte(msg), nil
}

func updateClientLog(cmd, msg string) {
	// update tune, collect, train log client log to file
	if (strings.Contains(cmd, ParamTune) || strings.Contains(cmd, SensitizeTrain)) && msg != "" {
		cmdParts := strings.Split(cmd, ":")
		if len(cmdParts) != 2 {
			return
		}

		appendMsg := removeColor(msg)
		fullPath := cmdParts[1]
		f, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Printf("[ERROR] OpenFile %v err: %v", fullPath, err)
			return
		}

		defer f.Close()
		_, err = io.WriteString(f, appendMsg)
		if err != nil {
			fmt.Printf("[ERROR] Write %v log err: %v", cmdParts[0], err)
			return
		}

		return
	}

	// update other log to memory
	if msg != "" {
		ClientLogMap[cmd] += msg
	}
}

// ClearCliLog clear cache log info
func ClearCliLog(cmd string) {
	// clear other log from memory
	if !strings.Contains(cmd, ParamTune) && !strings.Contains(cmd, SensitizeTrain) {
		ClientLogMap[cmd] = ""
	}
}

// setFormatter dynamic set formatter  for log
func (logger *Logger) setFormatter(mode string, level ...string) *logrus.Entry {
	pc, fileName, line, ok := runtime.Caller(2)
	var funcName string
	var file string
	if ok {
		file = getFile(fileName)
		// get func name
		pcName := strings.Split(runtime.FuncForPC(pc).Name(), ".")
		if len(pcName)-1 > 0 {
			funcName = pcName[len(pcName)-1]
		}
	}

	if mode == "console" {
		if len(level) > 0 {
			cLogger.SetFormatter(logger.getFormater(file, funcName, line, level[0]))
		} else {
			fmt.Printf("Custom console log format level is empty\n")
		}
	}

	if mode == "file" {
		fLogger.SetFormatter(
			&logFormater{
				TimestampFormat: time.Now().Local().Format("2006-01-02 15:04:05"),
				LogFormat:       "[%s] %s %s ... [%s, %d]\n",
				file:            file,
				funcName:        funcName,
				line:            line,
			})
	}

	return logger.entry
}

func getFile(fileName string) string {
	var file string
	// get relative path info
	fileSplit := strings.Split(fileName, "daemon/")
	if len(fileSplit)-1 > 0 {
		// remove the file sufix
		file = strings.ReplaceAll(strings.ReplaceAll(fileSplit[1], "/", "."), "go", "")
	}

	if len(fileSplit)-1 == 0 {
		// remove the file sufix
		file = strings.ReplaceAll(strings.ReplaceAll(fileSplit[0], "/", "."), "go", "")
	}

	return file
}

func (logger *Logger) getFormater(file, funcName string, line int, level string) *consoleLogFormater {
	level = strings.ToUpper(level)
	switch level {
	case "ERROR", "WARNING":
		return &consoleLogFormater{
			LogFormat: "[%s] %s ... [%s, %d]\n",
			file:      file,
			funcName:  funcName,
			line:      line,
		}
	default:
		return &consoleLogFormater{
			LogFormat: "%s\n",
		}
	}
}

// Info method log info level args
func Info(cmd string, args ...interface{}) {
	cacheLog(cmd, "info", "", args...)
	fLogInst.setFormatter("file").Info(args...)
	cLogInst.setFormatter("console", "info").Info(args...)
}

// Infoln method log info level args
func Infoln(cmd string, args ...interface{}) {
	cacheLog(cmd, "info", "", args...)
	fLogInst.setFormatter("file").Infoln(args...)
	cLogInst.setFormatter("console", "info").Infoln(args...)
}

// Infof method log info level message with format string
func Infof(cmd string, format string, args ...interface{}) {
	cacheLog(cmd, "info", format, args...)
	fLogInst.setFormatter("file").Infof(format, args...)
	cLogInst.setFormatter("console", "info").Infof(format, args...)
}

// Warn method log warn args
func Warn(cmd string, args ...interface{}) {
	cacheLog(cmd, "warning", "", args...)
	fLogInst.setFormatter("file").Warn(args...)
	cLogInst.setFormatter("console", "warning").Warn(args...)
}

// Warnln method log warn args
func Warnln(cmd string, args ...interface{}) {
	cacheLog(cmd, "warning", "", args...)
	fLogInst.setFormatter("file").Warnln(args...)
	cLogInst.setFormatter("console", "warning").Warnln(args...)
}

// Warnf method log warn level message with format string
func Warnf(cmd string, format string, args ...interface{}) {
	cacheLog(cmd, "warning", format, args...)
	fLogInst.setFormatter("file").Warnf(format, args...)
	cLogInst.setFormatter("console", "warning").Warnf(format, args...)
}

// Error method log error args
func Error(cmd string, args ...interface{}) {
	cacheLog(cmd, "error", "", args...)
	fLogInst.setFormatter("file").Error(args...)
	cLogInst.setFormatter("console", "error").Error(args...)
}

// Errorln method log error args
func Errorln(cmd string, args ...interface{}) {
	cacheLog(cmd, "error", "", args...)
	fLogInst.setFormatter("file").Errorln(args...)
	cLogInst.setFormatter("console", "error").Errorln(args...)
}

// Errorf method log error level message with format string
func Errorf(cmd string, format string, args ...interface{}) {
	cacheLog(cmd, "error", format, args...)
	fLogInst.setFormatter("file").Errorf(format, args...)
	cLogInst.setFormatter("console", "error").Errorf(format, args...)
}

// Debug method log debug args
func Debug(cmd string, args ...interface{}) {
	fLogInst.setFormatter("file").Debug(args...)
	cLogInst.setFormatter("console", "debug").Debug(args...)
}

// Debugln method log debug args
func Debugln(cmd string, args ...interface{}) {
	fLogInst.setFormatter("file").Debugln(args...)
	cLogInst.setFormatter("console", "debug").Debugln(args...)
}

// Debugf method log debug level message with format string
func Debugf(cmd string, format string, args ...interface{}) {
	fLogInst.setFormatter("file").Debugf(format, args...)
	cLogInst.setFormatter("console", "debug").Debugf(format, args...)
}

func cacheLog(cmd, level, format string, args ...interface{}) {
	level = strings.ToUpper(level)
	var msg string
	if format == "" {
		msg = fmt.Sprint(args...)
	} else {
		msg = fmt.Sprintf(format, args...)
	}

	trimMsg := strings.TrimSuffix(msg, "\n")
	switch level {
	case "ERROR", "WARNING":
		if cmd == ProfSet {
			msg = fmt.Sprintf("\t[%s] %s\n", colorLevel(level), trimMsg)
			break
		}
		msg = fmt.Sprintf("[%s] %s\n", colorLevel(level), trimMsg)
	case "INFO":
		msg = fmt.Sprintf("%s\n", trimMsg)
	default:
		return
	}

	if cmd != "" {
		updateClientLog(cmd, msg)
	}
}

func colorLevel(lev string) string {
	switch lev {
	case "ERROR":
		return utils.ColorString("red", lev)
	case "WARNING":
		return utils.ColorString("yellow", lev)
	default:
		return lev
	}
}

