package netutils

import (
	logger "github.com/sirupsen/logrus"
	"os"
	"regexp"
	"strconv"
)

func getServiceName(proto string, port int) (serviceName string, err error) {
	//if port eq one of then set return
	switch port {
	case 3000:
		{
			return "grafana-server", nil
		}
	case 8123:
		{
			return "home-assistant", nil
		}
	case 15672:
		{
			return "rabbitMQ", nil
		}
	case 9256:
		{
			return "process_exporter", nil
		}
	case 9100, 9200:
		{
			return "node_exporter", nil
		}
	case 9107:
		{
			return "consul_exporter", nil
		}
	case 9219:
		{
			return "ssl_exporter", nil
		}
	}

	// else search in /etc/services
	b, err := os.ReadFile("/etc/services")
	if err != nil {
		logger.WithFields(logger.Fields{
			"function":      "service",
			"module":        "services.go",
			"file":          "/etc/services",
			"port/protocol": strconv.Itoa(port) + "/" + proto,
		}).Debugln(err)
		return "Unknown", err
	}

	re := regexp.MustCompile(`(?m)(?P<service>[^ \n]+)\s+(?P<port>\d+)\/(?P<proto>\w+)`)
	for _, match := range re.FindAllStringSubmatch(string(b), -1) {
		name := match[1]
		n, _ := strconv.Atoi(match[2])
		pr := match[3]
		if proto == pr && port == n {
			return name, nil
		}
	}

	return "Unknown", nil
}

func RemoveDuplicateIP(dnsZone map[string]string) map[string]string {
	unixDNSZone := make(map[string]string, 5)
	var l []string
	var ipc bool
	for hostname, ip := range dnsZone {
		ipc = false
		for idx := 0; idx < len(l); idx++ {
			if ip == l[idx] {
				ipc = true
			}
		}
		if ipc == false {
			l = append(l, ip)
			unixDNSZone[hostname] = ip
		}
	}
	return unixDNSZone
}
