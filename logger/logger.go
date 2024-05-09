package logger

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var once sync.Once
var instance *logrus.Logger

// GetLogger 返回配置了自定义设置的记录器的单一实例。
func GetLogger() *logrus.Logger {
	once.Do(func() {
		instance = configureLogger()
	})
	return instance
}

type customFormatter struct {
	logrus.Formatter
}

func (f *customFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.999")
	msg := fmt.Sprintf("[%s] [func:%s] [%s:%d] [%s]: %s\n", timestamp, entry.Caller.Function, path.Base(entry.Caller.File), entry.Caller.Line, strings.ToUpper(entry.Level.String()), entry.Message)
	return []byte(msg), nil
}

func configureLogger() *logrus.Logger {
	log := logrus.New()
	log.SetFormatter(&customFormatter{})
	log.SetReportCaller(true)
	// 创建log目录
	if _, err := os.Stat("log"); os.IsNotExist(err) {
		err = os.MkdirAll("log", 0755)
		if err != nil {
			log.Errorf("创建日志目录失败： %v", err)
		}
	}

	currentTime := time.Now().Format("2006-01-02 15-04-05")
	logFile, err := os.OpenFile("log/"+currentTime+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Errorf("无法打开日志文件： %v", err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
	//log.SetOutput(os.Stdout)

	//默认日志级别为info
	log.SetLevel(logrus.InfoLevel)
	return log
}
