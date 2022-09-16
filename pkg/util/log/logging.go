package log

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"time"
)

const (
	FileNameSuffixFormat string = "-2006-01-02-15-04"
	FileNamePrefix       string = "youtube-audio"
	FileNameExt          string = ".log"
	FileNameDir          string = "logs/"
)

var logAsJSON bool

func EnableJSONFormat() {
	logAsJSON = true
}

var log = logrus.New()
var LoggingFilePath = getLogFilePath()

func getLogFilePath() string {
	return FileNameDir + getLogFileName()
}

func getLogFileName() string {
	return FileNamePrefix + time.Now().Format(FileNameSuffixFormat) + FileNameExt
}

func InitLogging() {
	fmt.Println("INIT LOGGING...")
	log.SetLevel(logrus.DebugLevel)
	log.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})
	logger := &lumberjack.Logger{
		Filename:   LoggingFilePath,
		MaxSize:    30,    // 日志文件大小，单位是 MB
		MaxBackups: 3,     // 最大过期日志保留个数
		MaxAge:     28,    // 保留过期文件最大时间，单位 天
		Compress:   false, // 是否压缩日志，默认是 false 不压缩
	}
	mw := io.MultiWriter(os.Stdout, logger)
	log.SetOutput(mw)
}

func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	log.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

func Printf(format string, args ...interface{}) {
	if logAsJSON {
		logJSON("normal", fmt.Sprintf(format, args...))
	} else {
		log.Printf(format, args...)
	}
}

func logJSON(status, msg string) {
	type jsonLog struct {
		Time    time.Time `json:"time"`
		Status  string    `json:"status"`
		Message string    `json:"msg"`
	}

	l := jsonLog{
		Time:    time.Now().UTC(),
		Status:  status,
		Message: msg,
	}
	jsonBytes, err := json.Marshal(&l)
	if err != nil {
		_, err2 := fmt.Fprintf(os.Stderr, msg)
		if err2 != nil {
			fmt.Printf("Error2: %v\n", err)
		}
		return
	}

	_, err = fmt.Fprintf(os.Stdout, "%s\n", string(jsonBytes))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
