package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
	BENEFIT_ITEM
)

var (
	maxLogSize   int64 = 100000 // 100 kbite
	ginLogsFile  *os.File
	apiLogsFile  *os.File
	ginLogsMutex sync.Mutex
	apiLogsMutex sync.Mutex
	ginMode      string
)

// Opens log files and set log output. Handle logger graceful shutdown and log files rotation
func init() {
	var ok bool
	ginMode, ok = os.LookupEnv("GIN_MODE")
	if !ok {
		ginMode = "debug"
	}

	var err error
	ginLogsFile, err = os.OpenFile("/logs/gin.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalf("Failed to open gin log file: %v", err)
	}
	apiLogsFile, err = os.OpenFile("/logs/api.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalf("Failed to open api log file: %v", err)
	}

	gin.DefaultWriter = io.MultiWriter(ginLogsFile)
	log.SetOutput(io.MultiWriter(apiLogsFile))

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGABRT)
		<-quit
		ginLogsFile.Close()
		apiLogsFile.Close()
	}()

	go func() {
		for {
			if err := rotateLogIfNeeded(&ginLogsFile, &ginLogsMutex, "gin"); err != nil {
				log.Printf("Error rotating gin log: %v", err)
			}
			if err := rotateLogIfNeeded(&apiLogsFile, &apiLogsMutex, "api"); err != nil {
				log.Printf("Error rotating api log: %v", err)
			}
			time.Sleep(1 * time.Hour)
		}
	}()
}

// If size of the logs file, rotates them
func rotateLogIfNeeded(filePtr **os.File, mutex *sync.Mutex, logName string) error {
	mutex.Lock()
	defer mutex.Unlock()

	file := *filePtr
	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("error getting file info: %w", err)
	}
	if info.Size() <= maxLogSize {
		return nil
	}

	file.Close()
	newName := fmt.Sprintf("%s.%s.old", file.Name(), time.Now().Format("2006-01-02_15:04"))
	if err := os.Rename(file.Name(), newName); err != nil {
		return fmt.Errorf("error renaming log file: %w", err)
	}

	newFile, err := os.OpenFile(file.Name(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("error creating new log file: %w", err)
	}

	*filePtr = newFile
	if logName == "gin" {
		gin.DefaultWriter = io.MultiWriter(newFile)
	} else if logName == "api" {
		log.SetOutput(io.MultiWriter(newFile))
	}

	log.Printf("Rotated %s log file", logName)
	return nil
}

func (l LogLevel) String() string {
	return [...]string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL", "BENEFIT_ITEM"}[l]
}

func logWithLevel(level LogLevel, format string, v ...interface{}) {
	_, f, l, _ := runtime.Caller(2)
	fullFuncName := fmt.Sprintf("%s:%d", f, l)
	logMsg := fmt.Sprintf(format, v...)
	log.Printf("[%s]\n%s: %s", level, fullFuncName, logMsg)
}

func WrapError(err error, message string) error {
	_, f, l, _ := runtime.Caller(1)
	if message != "" {
		return fmt.Errorf("\n%s:%d: %s: %w", f, l, message, err)
	}
	return fmt.Errorf("\n%s:%d: %w", f, l, err)
}

func Debug(format string, v ...interface{}) {
	if ginMode == "release" {
		logWithLevel(DEBUG, format, v...)
	}
}

func Info(format string, v ...interface{}) {
	logWithLevel(INFO, format, v...)
}

func Warn(format string, v ...interface{}) {
	logWithLevel(WARN, format, v...)
}

func Error(format string, v ...interface{}) {
	logWithLevel(ERROR, format, v...)
}

func Fatal(format string, v ...interface{}) {
	logWithLevel(FATAL, format, v...)
	os.Exit(1)
}
func BenefitItem(format string, v ...interface{}) {
	logWithLevel(BENEFIT_ITEM, format, v...)
}
