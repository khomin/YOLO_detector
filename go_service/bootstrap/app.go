package bootstrap

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Application struct {
	Env     *Env
	Db      *DatabaseUseCase
	LogFile *os.File
}

func App() Application {
	app := &Application{}
	app.Env = NewEnv(os.Args[1])
	app.InitLog()
	app.Db = InitDatabase(app.Env)
	return *app
}

func (a *Application) InitLog() {
	lumberjackLogger := &lumberjack.Logger{
		Filename:   filepath.ToSlash(a.Env.LOG_PATH),
		MaxSize:    1, // MB
		MaxBackups: 2,
		MaxAge:     3,    // days
		Compress:   true, // disabled by default
	}
	// fork writing into two outputs
	multiWriter := io.MultiWriter(os.Stderr, lumberjackLogger)
	logFormatter := new(logrus.TextFormatter)
	logFormatter.TimestampFormat = time.RFC1123Z
	logFormatter.FullTimestamp = true
	logFormatter.ForceColors = true

	logrus.SetFormatter(logFormatter)
	if a.Env.APP_ENV == "development" {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.WarnLevel)
	}
	logrus.SetOutput(multiWriter)
}

func (app *Application) Close() {
	app.LogFile.Close()
}
