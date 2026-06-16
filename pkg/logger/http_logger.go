package logger

import (
	"context"
	"os"
	"runtime"

	"github.com/SergeiGD/testify-profile/config"
	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Entry
}

func NewLogger(cfg *config.Config) *Logger {
	l := logrus.New()
	l.SetReportCaller(true)

	if cfg.App.Debug {
		l.Formatter = &logrus.TextFormatter{
			// CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			// 	return "", ""
			// },
			DisableColors: false,
			FullTimestamp: true,
			ForceColors:   true,
		}

		l.SetLevel(logrus.InfoLevel)
	} else {
		l.Formatter = &logrus.JSONFormatter{
			CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
				return "", ""
			},
		}

		l.SetLevel(logrus.InfoLevel)
	}

	l.SetOutput(os.Stdout)

	e := logrus.NewEntry(l)
	return &Logger{e}
}

func NewSilentTestsLogger() *Logger {
	l := logrus.New()
	l.SetReportCaller(true)

	l.SetOutput(nil)

	e := logrus.NewEntry(l)
	return &Logger{e}
}

func (logger *Logger) WithFields(fields logrus.Fields) *Logger {
	return &Logger{
		logger.Entry.WithFields(fields),
	}
}

func (logger *Logger) CtxInfo(ctx context.Context, msg interface{}) {

	requestId := ctx.Value("requestID")
	userId := ctx.Value("userId")

	logger.WithFields(logrus.Fields{
		"requestId": requestId,
		"userId":    userId,
	}).Info(msg)

}

func (logger *Logger) CtxError(ctx context.Context, msg interface{}) {

	requestId := ctx.Value("requestID")
	userId := ctx.Value("userId")

	logger.WithFields(logrus.Fields{
		"requestId": requestId,
		"userId":    userId,
	}).Error(msg)

}
