package prometheus

import (
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

var (
	// format should be -promAddr 127.0.0.1:9091
	PrometheusAddrFlag = "promAddr"

	prometheusAddr = ""
	// enabled is the flag specifying if metrics are enable or not.
	enabled = false

	defaultRegister *prometheusRegister
)

// Init enables or disables the metrics system. Since we need this to run before
// any other code gets to create meters and timers, we'll actually do an ugly hack
// and peek into the command line args for the metrics flag.
func init() {
	for i, arg := range os.Args {
		if strings.TrimLeft(arg, "-") == PrometheusAddrFlag {
			prometheusAddr = os.Args[i+1]
			enabled = true
			glog.V(logger.Info).Infof("Enabling prometheus exporter:%v", prometheusAddr)
		}
	}
	defaultRegister = newPrometheusRegister(prometheusAddr, "geth", getSystemInfo())
}

type prometheusRegister struct {
	namespace      string
	url            string
	registry       *prometheus.Registry
	labels         map[string]string
	gauges         map[string]prometheus.Gauge
	counters       map[string]prometheus.Counter
	exportDuration time.Duration
}

func newPrometheusRegister(url, namespace string, labels map[string]string) *prometheusRegister {
	r := &prometheusRegister{
		namespace:      namespace,
		url:            url,
		registry:       prometheus.NewRegistry(),
		labels:         labels,
		gauges:         make(map[string]prometheus.Gauge),
		counters:       make(map[string]prometheus.Counter),
		exportDuration: 10 * time.Second,
	}
	go r.export()
	return r
}

func (r *prometheusRegister) GetOrRegisterCounter(subsystem, name string) prometheus.Counter {
	cnt, ok := r.counters[name]
	if !ok {
		cnt = prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: flattenKey(r.namespace),
			Subsystem: flattenKey(subsystem),
			Name:      flattenKey(name),
			Help:      name,
		})
		r.registry.MustRegister(cnt)
		r.counters[name] = cnt
	}
	return cnt
}

func (r *prometheusRegister) export() {
	for _ = range time.Tick(r.exportDuration) {
		push.FromGatherer(r.namespace, r.labels, r.url, r.registry)
	}
}

// NewCounter create a new metrics Counter, either a real one of a NOP stub depending
// on the metrics flag.
func NewCounter(subsystem, name string) Counter {
	if !enabled {
		return &MockCounter{}
	}
	return defaultRegister.GetOrRegisterCounter(subsystem, name)
}

func flattenKey(key string) string {
	key = strings.Replace(key, " ", "_", -1)
	key = strings.Replace(key, ".", "_", -1)
	key = strings.Replace(key, "-", "_", -1)
	key = strings.Replace(key, "=", "_", -1)
	return key
}

func getSystemInfo() map[string]string {
	labels := map[string]string{}
	if host, err := os.Hostname(); err == nil {
		labels["instance"] = host
	}
	return labels
}
