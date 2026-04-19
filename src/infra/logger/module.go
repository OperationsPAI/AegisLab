package logger

import (
	"fmt"
	"path"
	"runtime"
	"sync"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

var (
	configureOnce sync.Once

	Module = fx.Module("logger",
		fx.Invoke(Configure),
	)
)

func Configure() {
	configureOnce.Do(func() {
		logrus.SetReportCaller(true)
		logrus.SetFormatter(&nested.Formatter{
			CustomCallerFormatter: func(f *runtime.Frame) string {
				filename := path.Base(f.File)
				return fmt.Sprintf(" (%s:%d)", filename, f.Line)
			},
			FieldsOrder:     []string{"component", "category"},
			HideKeys:        true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
		logrus.SetLevel(logrus.InfoLevel)
		logrus.Info("Logger initialized")
	})
}
