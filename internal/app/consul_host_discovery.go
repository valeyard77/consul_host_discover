package app

import (
	"fmt"
	"github.com/valeyard77/consul_host_discover/internal/config"
	"github.com/valeyard77/consul_host_discover/internal/netutils"
)

func Run(cfg *config.Config) {
	threads := make(chan struct{}, cfg.Threads)

	//var monitoringHosts []string

	iplist, err := netutils.ExpandCIDR(cfg.Subnet)
	if err != nil {
		cfg.Logger.Fatalln(err)
	}

	monitoringhostsvc := make(chan string, 0)
	//wg := sync.WaitGroup{}
	for _, ip := range iplist {
		//wg.Add(1)
		go func(ip string, monitoringhostsvc chan<- string) {
			//defer wg.Done()
			threads <- struct{}{}
			if netutils.Ping(ip) {
				monitoringhostsvc <- ip
			}
			<-threads

		}(ip, monitoringhostsvc)

	}
	//wg.Wait()

	for idx := range iplist {
		for m := range monitoringhostsvc {
			fmt.Println(m)
		}
		fmt.Println(idx)
	}

}
