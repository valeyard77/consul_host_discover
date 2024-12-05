package netutils

import (
	logger "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// check http connection
// return status: true/false, node_exporter: true/false
func CheckHTTPConnect(hostname, address string, port int, schema string, ch chan<- map[string]bool) (status bool) {
	res := make(map[string]bool, 1)
	if schema != "http" {
		schema = "https"
	}
	url := schema + "://" + address + ":" + strconv.Itoa(port)
	url_h := schema + "://" + hostname + ":" + strconv.Itoa(port)

	// create http client option
	client := &http.Client{
		Timeout: 1 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	//make get query
	conn, err := client.Get(url)
	if err != nil {
		logger.WithFields(logger.Fields{
			"function": "checkHTTPConnect",
			"address":  url_h,
		}).Debugln(err)

		res["http_result"] = false
		ch <- res
		return false
	}
	defer conn.Body.Close()

	//check service on this port, may be it some prometheus exporter
	svc, _ := getServiceName("tcp", port)
	logger.WithFields(logger.Fields{
		"function": "checkTCPConnect",
		"service":  strings.ToUpper(svc),
	}).Infof("Service %s was found on %s:%d\n", svc, hostname, port)

	bytesv, _ := ioutil.ReadAll(conn.Body)
	httpBody := string(bytesv)

	//check node exporter, ports 9100,9200
	if strings.Index(httpBody, "Node Exporter") != -1 {
		logger.WithFields(logger.Fields{"function": "checkHTTPConnect", "address": url_h}).Info("Find Node Exporter's enpoint")
		//node_exporter was found in endpoint
		res["node_exporter"] = true
		ch <- res
		return true
	} else {
		res["node_exporter"] = false
	}

	//check consul_exporter port 9107
	if strings.Index(httpBody, "Consul Exporter") != -1 {
		logger.WithFields(logger.Fields{"function": "checkHTTPConnect", "address": url_h}).Info("Find Consul Exporter's enpoint")
		//consul_exporter was found in endpoint
		res["consul_exporter"] = true
		ch <- res
		return true
	} else {
		res["consul_exporter"] = false
	}

	//check process_exporter port 9256
	if strings.Index(httpBody, "Process Exporter") != -1 {
		logger.WithFields(logger.Fields{"function": "checkHTTPConnect", "address": url_h}).Info("Find Process Exporter's enpoint")
		//process_exporter was found in endpoint
		res["process_exporter"] = true
		ch <- res
		return true
	} else {
		res["process_exporter"] = false
	}

	//check consul_exporter port 9256
	if port == 15672 {
		logger.WithFields(logger.Fields{"function": "checkHTTPConnect", "address": url_h}).Info("Find RabbitMQ enpoint")
		//process_exporter was found in endpoint
		res["rabbitmq"] = true
		ch <- res
		return true
	} else {
		res["rabbitmq"] = false
	}

	//check vm port 8428
	if port == 8428 {
		logger.WithFields(logger.Fields{"function": "checkHTTPConnect", "address": url_h}).Info("Find VictoriaMetrics enpoint")
		//process_exporter was found in endpoint
		res["victoriametrics"] = true
		ch <- res
		return true
	} else {
		res["victoriametrics"] = false
	}

	//if no exporter's service, return http_result => true
	res["http_result"] = true
	ch <- res
	return true
}
