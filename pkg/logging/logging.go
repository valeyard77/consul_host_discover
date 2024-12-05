package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

type Logging struct {
	Debug  bool
	Format string
	Output string
}

/*
	debug - set debug mode [true/false], false by default
	logFormat - set log format [text/json], text by default
	logOutput - set log output mode [stdout/file], file by default
*/
func New(debug bool, logFormat, logOutput string) *Logging {
	if logFormat == "" {
		logFormat = "text"
	}
	if logOutput == "" {
		logOutput = "stdout"
	}
	return &Logging{
		Debug:  debug,
		Format: logFormat,
		Output: logOutput,
	}
}

func (l *Logging) InitLog() *logrus.Logger {
	log := logrus.New()
	if l.Debug {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}

	if l.Output == "stdout" {
		log.SetOutput(os.Stdout)
	} else {
		ex, _ := os.Executable()
		baseDir := filepath.Join(filepath.Dir(ex))
		_ = os.Mkdir(filepath.Join(baseDir, "logs"), 777)
		_, f := filepath.Split(ex)
		log.Debugf("BaseDir: %s", baseDir)
		log.Debugf("exec: %s, path: %s", ex, f)
		if strings.Index(f, ".") != -1 {
			f = f[0:strings.Index(f, ".")] + ".log"
		} else {
			f = fmt.Sprintf("%s.log", f)
		}
		log.Debugf("log_file: %s", f)

		file, err := os.OpenFile(filepath.Join(baseDir, "logs", f), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err == nil {
			log.SetOutput(file)
		} else {
			log.Errorln("Failed to log into file, using stdout")
		}
	}

	switch l.Format {
	case "text":
		log.SetFormatter(&logrus.TextFormatter{
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				filename := filepath.Base(f.File)
				return fmt.Sprintf("%s:%d", filename, f.Line), fmt.Sprintf("%s()", f.Function)
			},
			ForceColors:     false,
			DisableQuote:    true,
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05.000",
		})
	case "json":
		log.SetFormatter(&logrus.JSONFormatter{
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				filename := filepath.Base(f.File)
				return fmt.Sprintf("%s:%d", filename, f.Line), fmt.Sprintf("%s()", f.Function)
			},
			TimestampFormat: "2006-01-02 15:04:05.000",
		})
	}
	return log
}
