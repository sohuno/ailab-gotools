package logconf

import (
	"flag"
	"github.com/sohuno/gotools/configutils"
	"os"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var DefaultLogConfigFilePath = "log_conf.json"

const LogConfigArgName = "--log_config"

type LogConfig struct {
	LogFile         string
	LogBacktraceAt  string
	LogFileMaxSize  uint64
	Verbose         int
	AddDirHeader    bool
	AlsoLogToStderr bool
	LogToStderr     bool
	SkipHeaders     bool
	SkipLogHeaders  bool
}

// AddFlags adds flags of klog into the specified FlagSet
func (l *LogConfig) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&l.LogFile, "log_file", l.LogFile, "log file name")
	fs.Uint64Var(&l.LogFileMaxSize, "log_file_max_size", l.LogFileMaxSize, "logFileMaxSizeMB")
	fs.BoolVar(&l.LogToStderr, "logtostderr", l.LogToStderr, "log to standard error instead of files")
	fs.BoolVar(&l.AlsoLogToStderr, "alsologtostderr", l.AlsoLogToStderr, "log to standard error as well as files")
	fs.IntVar(&l.Verbose, "v", l.Verbose, "number for the log level verbosity")
	fs.BoolVar(&l.AddDirHeader, "add_dir_header", l.AddDirHeader, "")
	fs.BoolVar(&l.SkipHeaders, "skip_headers", l.SkipHeaders, "If true, avoid header prefixes in the log messages")
	fs.BoolVar(&l.SkipLogHeaders, "skip_log_headers", l.SkipLogHeaders, "If true, avoid headers when opening log files")
	fs.StringVar(&l.LogBacktraceAt, "log_backtrace_at", l.LogBacktraceAt, "when logging hits line file:N, emit a stack trace")
}

func LoadLogConfig(defaultLogFile string) {
	logConfFilePath := parseLogConfFilePath(defaultLogFile)

	v := viper.New()
	configutils.LoadConfig(v, logConfFilePath, "")
	fs := flag.CommandLine
	for _, key := range v.AllKeys() {
		f := fs.Lookup(key)
		if f == nil {
			fs.Set(key, v.GetString(key))
		} else {
			f.Value.Set(v.GetString(key))
		}
	}
}

func parseLogConfFilePath(defaultLogFile string) string {
	if len(os.Args) == 0 {
		return DefaultLogConfigFilePath
	}

	logConfFilePath := DefaultLogConfigFilePath
	for _, value := range os.Args {
		if !strings.HasPrefix(value, LogConfigArgName) {
			continue
		}
		items := strings.Split(value, "=")
		if len(items) != 2 {
			continue
		}
		if strings.Trim(items[0], " ") != LogConfigArgName {
			continue
		}
		logConfFilePath = strings.Trim(items[1], " ")
		break
	}
	if _, err := os.Stat(logConfFilePath); err != nil {
		if _, err2 := os.Stat(defaultLogFile); err2 != nil {
			panic("LogConfigFilePath does not exists: " + logConfFilePath)
		}
		return defaultLogFile
	}
	return logConfFilePath
}
