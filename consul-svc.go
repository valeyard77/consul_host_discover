package main

import (
	consulapi "github.com/hashicorp/consul/api"
	logger "github.com/sirupsen/logrus"
	"reflect"
	"strconv"
	"strings"
)

var location string
var group string = "server"
var job string = "consul_blackbox_icmp_autodiscovery"
var svcName string = "icmp-check"
var service consulapi.AgentServiceRegistration

func initConsul(consulURL, token, datacenter string) (client *consulapi.Client, err error){
	config := consulapi.DefaultConfig()
	config.Address = consulURL
	config.Datacenter = datacenter
	config.Token = token
	consulClient, err := consulapi.NewClient(config)
	if err != nil {
		logger.WithFields(logger.Fields{
			"function": "consul-svc.go/initConsul",
			"consulURL": consulURL,
			"consulDC": datacenter,
		}).Error(err)
		return nil, err
	}
	return consulClient, nil
}

func setICMPSvc(consulClient *consulapi.Client, dns_name, ip, consulURL string) {
	svcCheck:= new(consulapi.AgentServiceCheck)
	svcCheck.CheckID = "ping_"+dns_name
	svcCheck.Name = "Ping test: "+dns_name
	svcCheck.Args = []string {"ping","-c2", ip}
	svcCheck.Interval = "5m"
	svcCheck.Timeout = "2s"
	svcCheck.Status = "passing"
	svcCheck.FailuresBeforeCritical = FailuresBeforeCritical
	svcCheck.DeregisterCriticalServiceAfter = DeregisterServiceTime
	service.Check = svcCheck

	service.ID = "icmp_"+dns_name
	service.Name = "prometheus_blackbox_icmp_exporter"
	service.Tags = []string  {"icmp:"+dns_name, "prometheus-icmp" }

	err := consulClient.Agent().ServiceRegister(&service)
	if err != nil {
		logger.WithFields(logger.Fields{
			"function": "consul-svc.go/setICMPSvc/ServiceRegister()",
			"consulURL": consulURL,
			"svcName": "prometheus_blackbox_icmp_exporter",
			"svcID": "icmp_"+dns_name,
		}).Errorln(err)
	} else {
		logger.Infof("ServiceID %s on %s - registration: OK", "icmp-check", dns_name)
	}
}

func setSvc(consulClient *consulapi.Client, dns_name, ip, consulURL string, port int, mode string) {
	svcCheck:= new(consulapi.AgentServiceCheck)
	svcCheck.CheckID = mode + "_check_" + dns_name + "_" + strconv.Itoa(port)
	svcCheck.Name = strings.ToUpper(mode) + " test: " + dns_name + " [" + strconv.Itoa(port) + "]"
	switch mode {
		case "tcp":	svcCheck.TCP = ip + ":" + strconv.Itoa(port)
		case "http": svcCheck.HTTP = "http://"+ip+ ":" + strconv.Itoa(port)
	}
	svcCheck.Interval = "5m"
	svcCheck.Timeout = "10s"
	svcCheck.FailuresBeforeCritical = FailuresBeforeCritical
	svcCheck.DeregisterCriticalServiceAfter = DeregisterServiceTime
	if strings.Index(dns_name, "mpwr") == -1 {
		service.Check = svcCheck
	}
	service.ID = mode + "_"+ dns_name + "_" + strconv.Itoa(port)
	service.Name = "prometheus_blackbox_"+mode+"_exporter"
	service.Tags = []string  {mode+":"+dns_name+":"+strconv.Itoa(port), "prometheus-"+mode }

	err := consulClient.Agent().ServiceRegister(&service)
	if err != nil {
		logger.WithFields(logger.Fields{
			"function": "consul-svc.go/setSvc/ServiceRegister()",
			"consulURL": consulURL,
			"svcName": "prometheus_blackbox_"+mode+"_exporter",
			"svcID": mode+"_"+dns_name,
		}).Errorln(err)
	} else {
		logger.Infof("ServiceID %s on %s (port: %d) - registration: OK", mode+"-check", dns_name, port)
	}
}

func setExportsSvc(consulClient *consulapi.Client, dns_name, ip, consulURL string, port int, mode string) {
	svcCheck:= new(consulapi.AgentServiceCheck)
	svcCheck.CheckID = mode + "_check_" + dns_name + "_" + strconv.Itoa(port)
	svcCheck.Name = mode + " test: " + dns_name + "[" + strconv.Itoa(port) + "]"
	svcCheck.HTTP = "http://"+ip + ":" + strconv.Itoa(port)
	svcCheck.Interval = "5m"
	svcCheck.Timeout = "10s"
	svcCheck.FailuresBeforeCritical = FailuresBeforeCritical
	svcCheck.DeregisterCriticalServiceAfter = DeregisterServiceTime
	service.Check = svcCheck

	service.ID = mode+"_"+dns_name
	service.Name = "prometheus_"+mode
	service.Address = dns_name
	service.Port = port
	service.Tags = []string  {mode, "prometheus-" + mode }

	err := consulClient.Agent().ServiceRegister(&service)
	if err != nil {
		logger.WithFields(logger.Fields{
			"function": "consul-svc.go/setExportsSvc/ServiceRegister()",
			"consulURL": consulURL,
			"svcName": "consul_" + mode,
			"svcID": mode+"_"+dns_name,
		}).Errorln(err)
	} else {
		logger.Infof("ServiceID %s on %s (port: %d) - registration: OK", mode, dns_name, port)
	}
}

func setConsulSVC(consulURL, token, datacenter string, listHostServices *[]consulHostSvc) {
	var mode string
	consulClient, err:= initConsul(consulURL, token, datacenter)
	if err != nil {
		logger.WithFields(logger.Fields {
			"function": "consul-svc.go/setConsulSVC",
			"consulURL": consulURL,
			"consulDC": datacenter,
		}).Fatalln(err)
	}

	for _, data:= range *listHostServices {
		dns_name := data.Svc.HOSTNAME
		ip := data.Svc.IP
		if ip == "192.168.1.4" { dns_name = "ha.hm.net" }
		if ip == "192.168.0.4" { dns_name = "mqtt-nkr.hm.net" }

		if strings.Index(ip, "192.168.1") != -1 { location = "himki" }
		if strings.Index(ip, "192.168.2") != -1 { location = "klin" }
		if strings.Index(ip, "192.168.0") != -1 { location = "nekrasovka" }
		if strings.Index(ip, "192.168.3") != -1 { location = "noginsk" }

		if strings.Index(dns_name,"light") != -1 { group = "light" }
		if strings.Index(dns_name, "ipcam") != -1 { group = "ipcam" }
		if strings.Index(dns_name, "hs") != -1 { group = "sockets" }
		if strings.Index(dns_name,"mpwr") != -1 { group = "sockets" }
		if strings.Index(dns_name,"uc") != -1 { group = "unicontroller" }
		if strings.Index(dns_name, "vacuum") != -1 { group = "vacuum" }
		if strings.Index(dns_name, "gw") != -1 { group = "netdevice" }
		if strings.Index(dns_name, "mikrotik") != -1 { group = "netdevice" }
		if strings.Index(dns_name, "ha") != -1 { group = "home-assistant" }
		if strings.Index(dns_name, "qnap") != -1 { group = "qnap" }
		if strings.Index(dns_name, "mqtt") != -1 { group = "mqtt-server" }
		if strings.Index(dns_name, "printer") != -1 { group = "printer" }
		if strings.Index(dns_name, "openhab") != -1 { group = "openhab" }

		mode = "icmp"
		svcName = mode+"-check"
		service.Meta = map[string]string {
			"job":      job,
			"service":  svcName,
			"location": location,
			"group":    group,
			"ip": ip,
		}

		//Set ICMP checking
		setICMPSvc(consulClient, dns_name, ip, consulURL)

		//Set simple TCP checking
		for _, tcpport:= range data.Svc.TCPCheck.Ports {
			mode = "tcp"
			svcName = mode+"-check"
			service.Meta["job"] =  "consul_blackbox_tcp_autodiscovery"
			service.Meta["service"] =  svcName

			setSvc(consulClient, dns_name, ip, consulURL, tcpport, mode)
		}

		//set svc for http ports
		for _, httpport:= range data.Svc.HTTP.Ports {
			mode = "http"
			svcName = mode+"-check"
			service.Meta["job"] =  "consul_blackbox_http_autodiscovery"
			service.Meta["service"] =  svcName

			setSvc(consulClient, dns_name, ip, consulURL, httpport, mode)
		}

		//set svc for found exporters
		expStruct := data.Svc.Exporters[0]
		e:= reflect.ValueOf(&expStruct).Elem()
		for i:=0; i < e.NumField(); i ++ {
			exporter := e.Type().Field(i).Name
			ePort := e.Field(i).Interface()
			if ePort.(int) != 0 {
				t := strings.ToLower(exporter)
				dind := strings.Index(t, "exporter")
				if dind != -1 {
					mode = t[:dind] + "_" + t[dind:]
				} else {
					mode = t
				}
				service.Meta["job"] =  "consul_+" + mode + "+_autodiscovery"
				service.Meta["service"] =  svcName
				setExportsSvc(consulClient, dns_name, ip, consulURL, ePort.(int), mode)
			}
		}

	}
}


