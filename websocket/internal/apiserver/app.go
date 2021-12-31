package apiserver

import (
	"github.com/sirupsen/logrus"
	"github.com/xiaodulala/component-tools/pkg/app"
	"path"
	"runtime"
	"socket-demo/internal/apiserver/config"
	"socket-demo/internal/apiserver/options"
	pkg_options "socket-demo/internal/pkg/options"
	"strconv"
)

const appName = "apiServer"
const commandDesc = "apiServer command description"

// NewApp 创建应用
func NewApp(basename string) *app.App {
	opts := options.NewOptions()
	return app.NewApp(
		basename, appName,
		app.WithDescription(commandDesc),
		app.WithOptions(opts),
		app.WithRunFunc(run(opts)),
	)
}

func run(opts *options.Options) app.RunFunc {
	return func(basename string) error {
		//结合配置文件，选项参数。生成配置。
		cfg, err := config.CreateConfigFromOptions(opts)
		if err != nil {
			return err
		}

		// 根据配置初始化日志
		if err := initLog(cfg); err != nil {
			return err
		}

		return Run(cfg)
	}
}

func initLog(cfg *config.Config) error {
	level, err := logrus.ParseLevel(cfg.LogOptions.Level)
	if err != nil {
		return err
	}
	logrus.SetLevel(level)
	if cfg.LogOptions.Formatter == pkg_options.FormatJSON {
		logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat:  cfg.LogOptions.TimeFormatter,
			DisableTimestamp: false,
			DataKey:          "",
			FieldMap:         nil,
			CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
				fileName := path.Base(frame.File)
				return frame.Function, fileName + ":" + strconv.Itoa(frame.Line)
			},
			PrettyPrint: false,
		})
	} else if cfg.LogOptions.Formatter == pkg_options.FormatText {
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: cfg.LogOptions.TimeFormatter,
			CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
				fileName := path.Base(frame.File)
				return frame.Function, fileName + ":" + strconv.Itoa(frame.Line)
			},
		})
		logrus.SetReportCaller(true)
	}

	return nil
}
