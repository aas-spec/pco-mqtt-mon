// Logger with File Rotation 4.0
// Function:
// LogX, PanicX - Простая запись в лог
// LLogX - запись в лог с указанием в первом аргументе
// уровня логирования(Level)
// PrintX - аналог LogX, для совместимости и горячей замены log
// OutX - аналог предыдущих функций,
// но в качестве первого аргумента принимают LoggerID
// LOutx - второй аргумент Level (уровень логирования)
// SetLogLevel  - устанавливает для логгера уровень логирования
// SetStoreDays - устанавливает для логгера кол-во дней ротации
// UseOwnDir - устанавливает для логеера параметр записи в свою папку

package mlog

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

const (
	DefStoreDays     = 5     // Кол-во хранения файлов по дефолту
	DefLoggerID      = ""    // Идентификатор дефолтного логгера
	DefLevel         = 5     // Уровень логирования по дефолту
	DefUseOwnDir     = false // У каждого логера своя папка
)

// Описывает логгер
type Logger struct {
	ID           string
	BaseFileName string
	log          *log.Logger
	lastFileName string
	StoreDays    int
	Level        int
	UseOwnDir    bool
}

func (logger Logger) getLogFileNameStr(s string) string {
	curPath := logger.BaseFileName
	curPath = filepath.Dir(curPath) + string(os.PathSeparator) + filepath.Base(curPath)
	logfile := curPath[:len(curPath)-len(filepath.Ext(curPath))]
	logfile = logfile + "_" + s + ".log"
	return logfile
}

func (logger Logger) getLogFileName(tm time.Time) string {
	return logger.getLogFileNameStr(tm.Format("2006-01-02"))
}

func (logger *Logger) checkLogRotation() (res string) {
	curLogFileName := logger.getLogFileName(time.Now())
	if logger.lastFileName != curLogFileName { // Переоткрыть новый файл
		// Проверяю и создаю каталог для записи
		curPath := logger.BaseFileName
		curDir := filepath.Dir(curPath)
		curPath = curDir + string(os.PathSeparator) + filepath.Base(curPath)
		_, err := os.Stat(filepath.Dir(curPath))
		if err != nil && os.IsNotExist(err) {
			os.MkdirAll(filepath.Dir(curPath), 0777)
		}

		// Открываю Log
		logFile, err := os.OpenFile(curLogFileName,
			os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
		if err != nil {
			log.Panicf("Unable to open file %v : %s", curLogFileName, err)
		}

		logger.lastFileName = curLogFileName
		disableStdout := os.Getenv("DisableStdout")

		logger.log.SetFlags(0)

		if disableStdout != "1" {
			mw := io.MultiWriter(os.Stdout, logFile)
			logger.log.SetOutput(mw)
		} else {
			mw := io.MultiWriter(logFile)
			logger.log.SetOutput(mw)
		}

		// Очищаю старые фалйы при смене имени лог файла
		files, err := filepath.Glob(logger.getLogFileNameStr("????-??-??"))
		if err != nil {
			res += fmt.Sprintf("%s unable to expand mask: %s \n", GetTimeStamp(), err)
		}
		sort.Strings(files)
		for len(files) > logger.StoreDays {
			fileToDelete := files[0]
			files = files[1:]

			res += fmt.Sprintf("%s delete: %s \n", GetTimeStamp(), fileToDelete)
			err := os.Remove(fileToDelete)
			if err != nil {
				res += fmt.Sprintf("%s unable to delete file: %s \n", GetTimeStamp(), err)
			}
		}
	}
	return res
}

func GetTimeStamp() string {
	tm := time.Now()
	res := tm.Format("2006-01-02 15:04:05")
	return res
}

func (logger *Logger) Logln(level int, v ...interface{}) {
	if level > logger.Level {
		return
	}
	s := logger.checkLogRotation()
	s += GetTimeStamp() + " " + fmt.Sprintln(v...)
	logger.log.Print(s)
}

func (logger *Logger) Log(level int, v ...interface{}) {
	if level > logger.Level {
		return
	}
	s := logger.checkLogRotation()
	s += GetTimeStamp() + " " + fmt.Sprint(v...)
	logger.log.Print(s)
}

func (logger *Logger) Logf(level int, format string, v ...interface{}) {
	if level > logger.Level {
		return
	}
	s := logger.checkLogRotation()
	s += GetTimeStamp() + " " + fmt.Sprintf(format, v...)
	logger.log.Print(s)
}

func (logger *Logger) Panic(v ...interface{}) {
	s := GetTimeStamp() + " " + fmt.Sprint(v...)
	logger.log.Panic(s)
}

// Возвращает базовое имя файла для логгера по дефолту
// Например c:\Blabla\My\My.exe
// Лог будет c:\Blabla\My\logs\My-LogID.log
func getDefLoggerFileName(logID string, useOwnDir bool) string {
	curPath := os.Args[0]
	curDir := filepath.Dir(curPath)
	newPath := curDir + string(os.PathSeparator) + "logs" + string(os.PathSeparator)
	if useOwnDir && logID != "" {
		newPath += logID + string(os.PathSeparator)
	}
	curPath = newPath + filepath.Base(curPath)
	logfile := curPath[:len(curPath)-len(filepath.Ext(curPath))]
	if logID != "" && !useOwnDir {
		logfile += "-" + logID
	}
	logfile += ".log"
	return logfile
}

var defLogger = newLogger(DefLoggerID, getDefLoggerFileName(DefLoggerID, DefUseOwnDir), DefStoreDays, DefLevel, DefUseOwnDir)

// Тип для списка логгеров
type loggerStore struct {
	Sync  sync.Mutex
	Items map[string]Logger
}

// Список логгеров
var loggers = loggerStore{
	Sync:  sync.Mutex{},
	Items: make(map[string]Logger),
}

///////////////////////////////////////////////////
// Добавляет новый логгер
///////////////////////////////////////////////////
func newLogger(id string, baseFileName string, storeDays int, level int, useOwnDir bool) Logger {
	loggers.Sync.Lock()
	defer func() {
		loggers.Sync.Unlock()
	}()
	logger := Logger{
		ID:           id,
		BaseFileName: baseFileName,
		log:          log.New(os.Stdout, "", log.LstdFlags),
		StoreDays:    storeDays,
		Level:        level,
		UseOwnDir:    useOwnDir,
	}
	loggers.Items[id] = logger
	return logger
}

func SetStoreDays(loggerID string, storeDays int) {
	loggers.Sync.Lock()
	logger, found := loggers.Items[loggerID]
	if !found {
		loggers.Sync.Unlock()
		logger = newLogger(loggerID, getDefLoggerFileName(loggerID, DefUseOwnDir), storeDays, DefLevel, DefUseOwnDir)
		loggers.Sync.Lock()
	} else {
		logger.StoreDays = storeDays
	}
	loggers.Items[loggerID] = logger
	if loggerID == DefLoggerID {
		defLogger = logger
	}
	loggers.Sync.Unlock()
}

func SetLogLevel(loggerID string, level int) {
	loggers.Sync.Lock()
	logger, found := loggers.Items[loggerID]
	if !found {
		loggers.Sync.Unlock()
		logger = newLogger(loggerID, getDefLoggerFileName(loggerID, DefUseOwnDir), DefStoreDays, level, DefUseOwnDir)
		loggers.Sync.Lock()
	} else {
		logger.Level = level
	}
	loggers.Items[loggerID] = logger
	if loggerID == DefLoggerID {
		defLogger = logger
	}
	loggers.Sync.Unlock()
}

func SetLogUseOwnDir(loggerID string, useOwnDir bool) {
	loggers.Sync.Lock()
	logger, found := loggers.Items[loggerID]
	if !found {
		loggers.Sync.Unlock()
		logger = newLogger(loggerID, getDefLoggerFileName(loggerID, useOwnDir), DefStoreDays, DefLevel, useOwnDir)
		loggers.Sync.Lock()
	} else {
		logger.UseOwnDir = useOwnDir
		logger.BaseFileName = getDefLoggerFileName(loggerID, useOwnDir)
	}
	loggers.Items[loggerID] = logger
	if loggerID == DefLoggerID {
		defLogger = logger
	}
	loggers.Sync.Unlock()
}

func getLogger(loggerID string) Logger {
	logger, found := loggers.Items[loggerID]
	if !found {
		logger = newLogger(loggerID, getDefLoggerFileName(loggerID, DefUseOwnDir), DefStoreDays, DefLevel, DefUseOwnDir)
	}
	return logger
}

func Outln(loggerID string, v ...interface{}) {
	logger := getLogger(loggerID)
	logger.Logln(0, v...)
}

func Out(loggerID string, v ...interface{}) {
	logger := getLogger(loggerID)
	logger.Log(0, v...)
}

func Outf(loggerID string, format string, v ...interface{}) {
	logger := getLogger(loggerID)
	logger.Logf(0, format, v...)
}

func LOutln(loggerID string, level int, v ...interface{}) {
	logger := getLogger(loggerID)
	logger.Logln(level, v...)
}

func LOut(loggerID string, level int, v ...interface{}) {
	logger := getLogger(loggerID)
	logger.Log(level, v...)
}

func LOutf(loggerID string, level int, format string, v ...interface{}) {
	logger := getLogger(loggerID)
	logger.Logf(level, format, v...)
}

///////////////////////////////////////////////////
// Функции используются с логгером по-умолчанию
///////////////////////////////////////////////////
func Logln(v ...interface{}) {
	defLogger.Logln(0, v...)
}
func Log(v ...interface{}) {
	defLogger.Log(0, v...)
}
func Logf(format string, v ...interface{}) {
	defLogger.Logf(0, format, v...)
}

func Panic(v ...interface{}) {
	defLogger.Panic(v...)
}
func LLogln(level int, v ...interface{}) {
	defLogger.Logln(level, v...)
}

func LLog(level int, v ...interface{}) {
	defLogger.Log(level, v...)
}
func LLogf(level int, format string, v ...interface{}) {
	defLogger.Logf(level, format, v...)
}

func Println(v ...interface{}) {
	defLogger.Logln(0, v...)
}

func Print(v ...interface{}) {
	defLogger.Log(0, v...)
}
func Printf(format string, v ...interface{}) {
	defLogger.Logf(0, format, v...)
}

func LPrintln(level int, v ...interface{}) {
	defLogger.Logln(level, v...)
}

func LPrint(level int, v ...interface{}) {
	defLogger.Log(level, v...)
}
func LPrintf(level int, format string, v ...interface{}) {
	defLogger.Logf(level, format, v...)
}

func StdPrintf(format string, v ... interface{}) {
	s := GetTimeStamp() + " " + fmt.Sprintf(format, v...)
	fmt.Println( s)
}
