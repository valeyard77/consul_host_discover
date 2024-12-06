package app

import (
	"fmt"
	"github.com/valeyard77/consul_host_discover/internal/config"
	"github.com/valeyard77/consul_host_discover/internal/netutils"
)

func Run(cfg *config.Config) {
	iplist, err := netutils.ExpandCIDR(cfg.Subnet)
	if err != nil {
		cfg.Logger.Fatalln(err)
	}

	monitoringHosts := netutils.PingAlive(iplist, cfg.Threads)
	fmt.Println(monitoringHosts)

}
