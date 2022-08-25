package log

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
)

var log = logrus.New()

func init() {
	fmt.Println("INIT LOGGING...")
	log.SetLevel(logrus.DebugLevel)
	log.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})
	logger := &lumberjack.Logger{
		Filename:   "./youtube-audio.log",
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
	log.Printf(format, args...)
}
