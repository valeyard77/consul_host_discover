package config

import (
	"github.com/sirupsen/logrus"

	"github.com/valeyard77/consul_host_discover/pkg/logging"
)

type Config struct {
	Logger  *logrus.Logger
	Subnet  string
	Threads int
}

func New() *Config {
	cli := newCli()
	if cli == nil {
		return nil
	}

	logger := logging.New(cli.Debug, cli.LogFormat, cli.LogOutput).InitLog()

	return &Config{
		Logger:  logger,
		Subnet:  cli.Subnet,
		Threads: cli.Thread,
	}

}
