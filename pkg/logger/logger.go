package logger

import (
	"github.com/sirupsen/logrus"
)

var (
	Log *logrus.Logger
)

func init() {
	Log = logrus.New()
	Log.Formatter = &logrus.JSONFormatter{}
	Log.SetReportCaller(true)
}
