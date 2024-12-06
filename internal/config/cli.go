package config

import (
	"github.com/spf13/pflag"
)

var (
	//app
	subnet  = pflag.StringP("subnet", "n", "192.168.2.0/24", "Subnet for search, ex: 192.168.2.0/24")
	threads = pflag.Int("threads", 14, "Number of threads, default=14")

	//logger
	output    = pflag.StringP("log.output", "l", "stdout", "Log output mode [stdout/file]")
	logformat = pflag.String("log.format", "text", "Log output format [text/json]")
	debug     = pflag.BoolP("debug", "X", false, "Set debug mode")

	help = pflag.BoolP("help", "h", false, "Help")
)

type Cli struct {
	Subnet    string
	LogOutput string
	LogFormat string
	Debug     bool
	Thread    int
}

func newCli() *Cli {
	pflag.Parse()
	if *help {
		pflag.Usage()
		return nil
	}

	return &Cli{
		Subnet:    *subnet,
		LogOutput: *output,
		LogFormat: *logformat,
		Debug:     *debug,
		Thread:    *threads,
	}
}
