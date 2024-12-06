package netutils

import (
	"bytes"
	logger "github.com/sirupsen/logrus"
	"net"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-ping/ping"
)

func Ping(address string) bool {
	pinger, err := ping.NewPinger(address)
	if err != nil {
		return false
	}
	pinger.Count = 2
	pinger.Timeout = 1 * time.Second
	if err = pinger.Run(); err != nil { // Blocks until finished.
		return false
	}
	if lost := pinger.Statistics().PacketLoss; lost > 2 {
		return false
	}

	return true
}

func PingAlive(iplist []string, threads int) []string {
	tokens := make(chan struct{}, threads)
	monitoringhostsvc := make(chan string)
	var monitoringHosts []string

	for _, ip := range iplist {
		go func(ip string, monitoringhostsvc chan<- string) {
			tokens <- struct{}{}
			if Ping(ip) {
				monitoringhostsvc <- ip
			} else {
				monitoringhostsvc <- ""
			}
			<-tokens

		}(ip, monitoringhostsvc)
	}

	var hosts []string
	for _, _ = range iplist {
		select {
		case m := <-monitoringhostsvc:
			hosts = append(hosts, m)
		}
	}

	for _, host := range hosts {
		if host != "" {
			monitoringHosts = append(monitoringHosts, host)
		}
	}

	return monitoringHosts
}

// CheckTCPConnect check tcp connection, return status: true/false,
func CheckTCPConnect(hostname, address string, port int, ch chan<- map[string]bool) (status bool) {
	res := make(map[string]bool, 1)
	conn, err := net.DialTimeout("tcp", address+":"+strconv.Itoa(port), 1*time.Second)
	if err != nil {
		logger.WithFields(logger.Fields{
			"function": "checkTCPConnect",
			"address":  hostname + ":" + strconv.Itoa(port),
		}).Debugln(err)
		res["tcp_check"] = false
		ch <- res
		return false
	}

	conn.Close()
	svc, _ := getServiceName("tcp", port)
	logger.WithFields(logger.Fields{
		"function": "checkTCPConnect",
		"service":  strings.ToUpper(svc),
	}).Infof("Service %s was found on %s:%d\n", svc, hostname, port)
	res["tcp_check"] = true
	ch <- res
	return true
}

func GetDNSZoneInfo(domain string) map[string]string {
	dnslookup := make(map[string]string, 5)
	cmd := exec.Command("/usr/bin/host", "-al", "-tA", domain)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		logger.WithFields(logger.Fields{
			"function":  "getDNSZoneInfo",
			"cmd":       out.String(),
			"cmd.error": stderr.String(),
			"int.error": err,
		}).Errorln(err)
	}
	re := regexp.MustCompile(`(?m)(?P<dns_name>[a-zA-Z.0-9-]+).*\s+(?P<ip>\d+\.\d+\.\d+\.\d+)`)
	dz := strings.Split(out.String(), "\n")
	for idx := 0; idx < len(dz)-2; idx++ {
		for _, match := range re.FindAllStringSubmatch(dz[idx], -1) {
			last := len(match[1]) - 1
			hostname := string(match[1])[:last]
			dnslookup[hostname] = match[2]
		}
	}
	return dnslookup
}
