package options

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"strings"
)

const (
	// FormatJSON json格式
	FormatJSON = "json"
	// FormatText test格式
	FormatText = "text"
)

var logDefaults = LogOptions{
	Formatter:     "text",
	TimeFormatter: "2006-01-02 15:05:06.000",
	Level:         "debug",
}

// LogOptions 日志选项参数
type LogOptions struct {
	Formatter     string `json:"formatter" mapstructure:"formatter"`
	TimeFormatter string `json:"time-formatter" mapstructure:"time-formatter"`
	Level         string `json:"level" mapstructure:"level"`
}

// NewLogOptions return LogOptions
func NewLogOptions() *LogOptions {
	return &LogOptions{
		Formatter:     logDefaults.Formatter,
		TimeFormatter: logDefaults.TimeFormatter,
		Level:         logDefaults.Level,
	}
}

// Validate 验证选项参数
func (lo *LogOptions) Validate() []error {
	var errs []error

	var logLevel logrus.Level
	if err := logLevel.UnmarshalText([]byte(lo.Level)); err != nil {
		errs = append(errs, err)
	}

	format := strings.ToLower(lo.Formatter)
	if format != FormatJSON && format != FormatText {
		errs = append(errs, fmt.Errorf("not a valid log format:%q,only support %q or %q", lo.Formatter, FormatJSON, FormatText))
	}

	return errs
}

// AddFlags 添加选项参数
func (lo *LogOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&lo.Level, "log.level", lo.Level, "log level")
	fs.StringVar(&lo.Formatter, "log.formatter", lo.Formatter, "log formatter support json or text")
	fs.StringVar(&lo.TimeFormatter, "log.time-formatter", lo.TimeFormatter, "log time formatter")
}
