package main

import (
	"fmt"
	logger "github.com/sirupsen/logrus"
	"github.com/valeyard77/consul_host_discover/internal/netutils"
	"os"
	"time"
)

const (
	ConsulSever            = "app-consul:8500"
	Token                  = "j+sZskbe21jYYzwrnJcsJM6Ee6uPD7kDe8PWViIRExM="
	Datacenter             = "ex"
	DeregisterServiceTime  = "48h"
	FailuresBeforeCritical = 3
)

type consulHostSvc struct {
	Svc struct {
		HOSTNAME string `json:"HOSTNAME"`
		IP       string `json:"IP"`
		TCPCheck struct {
			Ports []int `json:"Ports"`
		} `json:"TCPCheck"`
		HTTP struct {
			Ports []int `json:"Ports"`
		} `json:"HTTP"`
		Exporters [1]struct {
			NodeExporter    int `json:"node_exporter"`
			ProcessExporter int `json:"process_exporter"`
			ConsulExporter  int `json:"consul_exporter"`
			SslExporter     int `json:"ssl_exporter"`
			RabbitMQ        int `json:"rabbitMQ"`
			VictoriaMetrics int `json:"victoriametrics"`
		} `json:"Exporters"`
		RabbitMQ struct {
			Port int `json:"Port"`
		} `json:"RabbitMQ"`
	} `json:"Svc"`
}

type hosts struct {
	tcpPorts  []int
	httpPorts []int
}

func init() {
	// Log as JSON instead of the default ASCII formatter.
	logger.SetFormatter(&logger.TextFormatter{
		ForceColors:     true,
		DisableQuote:    true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	logger.SetOutput(os.Stdout)

	logger.SetLevel(logger.InfoLevel)
}

func setConsulCheckParams(dns_zone map[string]string) *[]consulHostSvc {
	ch := make(chan map[string]bool, 4)
	l := []consulHostSvc{}

	/*
		    6379 - redis
			9100, 9200 - node_exporter
			9256 - process_exporter
			9219 - ssl_exporter
			9107 - consul_exporter
		    8428 - victoriametrics
			15672 - rabbitMQ /api/metrics
	*/
	hp := hosts{
		tcpPorts:  []int{21, 22, 1883, 3306, 5432, 6379},
		httpPorts: []int{80, 9100, 9200, 8123, 3000, 9256, 9107, 8428},
	}

	//dns_zone = make(map[string]string)
	//dns_zone["hc.hm.net"] = "192.168.1.4"
	//dns_zone["printer-nkr.hm.net"] = "192.168.0.251"
	//dns_zone["hc2.hm.net"] = "192.168.0.4"
	//dns_zone["mpwr-kt200-sc3.dev.hm.net"] = "192.168.1.200"

	for hostname, ip := range dns_zone {
		alive := netutils.Ping(ip)
		if alive == true {
			logger.Infof("Host %s/%s is alive \n", hostname, ip)
			var hsvc consulHostSvc

			hsvc.Svc.HOSTNAME = hostname
			hsvc.Svc.IP = ip
			for _, port := range hp.tcpPorts {
				go netutils.CheckTCPConnect(hostname, ip, port, ch)
				st := <-ch
				if st["tcp_check"] == true {
					hsvc.Svc.TCPCheck.Ports = append(hsvc.Svc.TCPCheck.Ports, port)
				}
				//time.Sleep(100 * time.Millisecond)
			}

			for _, port := range hp.httpPorts {
				go netutils.CheckHTTPConnect(hostname, ip, port, "http", ch)
				stmap := <-ch
				if stmap["http_result"] == true {
					hsvc.Svc.HTTP.Ports = append(hsvc.Svc.HTTP.Ports, port)
				}
				if stmap["node_exporter"] == true {
					hsvc.Svc.Exporters[0].NodeExporter = port
				}
				if stmap["consul_exporter"] == true {
					hsvc.Svc.Exporters[0].ConsulExporter = port
				}
				if stmap["process_exporter"] == true {
					hsvc.Svc.Exporters[0].ProcessExporter = port
				}
				if stmap["rabbitmq"] == true {
					hsvc.Svc.Exporters[0].RabbitMQ = port
				}
				if stmap["victoriametrics"] == true {
					hsvc.Svc.Exporters[0].VictoriaMetrics = port
				}
			}
			l = append(l, hsvc)
		} else {
			logger.Infof("Host %s/%s is not alive\n", hostname, ip)
		}
	}
	defer close(ch)
	return &l
}

func main() {
	start := time.Now().Unix()
	fmt.Println("Get zone info from hm.net")
	t := netutils.GetDNSZoneInfo("hm.net")
	fmt.Printf("Recieved %d hosts from dns, let's deduplicate data in dns zone hm.net\n", len(t))
	dns_zone := netutils.RemoveDuplicateIP(t)
	fmt.Printf("Deduplicate complete, now is %d hosts in dns zone\n", len(dns_zone))
	fmt.Println("Create service params for consul from hosts")

	cp := setConsulCheckParams(dns_zone)
	setConsulSVC(ConsulSever, Token, Datacenter, cp)

	stop := time.Now().Unix()
	//fmt.Printf("%d | %d\n", start, stop)
	fmt.Printf("Execution time: %d seconds", stop-start)
}
