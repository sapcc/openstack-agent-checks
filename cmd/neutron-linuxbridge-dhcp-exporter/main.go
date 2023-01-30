package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	listenAddress  = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9116").String()
	updateInterval = kingpin.Flag("update-interval", "Interval of checking for new linux bridges").Default("30").Int()
	metricPrefix   = kingpin.Flag("metric-prefix", "Prefix the metric name").Default("").String()
	setCurrentTime = kingpin.Flag("set-current-time", "Instead of binary gauges, set the time as timeticks").Default("false").Bool()
	bridgesPath    = kingpin.Flag("bridges-path", "Virtual filesystem path exposing kernel network interfaces").Default("/sys/class/net").String()
	netNsPath      = kingpin.Flag("netns-path", "Virtual filesystem path network namespaces").Default("/var/run/netns/").String()
)

func loopFileSystem(hostname string, path string, prefix string, label string, interval int, setCurrentTime bool, m *prometheus.GaugeVec, sig chan os.Signal) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	for {
		select {
		case <-ticker.C:
			log.Printf("Re-reading %s\n", path)
			files, err := ioutil.ReadDir(path)
			if err != nil {
				log.Fatalf("Failed reading from %s: %s", path, err)
			}
			// reset the vector so old/removed bridges do not appear in the scrape
			m.Reset()
			for _, f := range files {
				if strings.HasPrefix(f.Name(), prefix) {
					label := prometheus.Labels{"hostname": hostname, label: strings.TrimPrefix(f.Name(), prefix)}
					if setCurrentTime {
						m.With(label).SetToCurrentTime()
					} else {
						m.With(label).Set(1)
					}
				}
			}
		case received := <-sig:
			log.Printf("Shutting down %s collector", path)
			ticker.Stop()
			// add the signal again so the other goroutine also gets it and stops
			sig <- received
			return
		}
	}
}

func main() {

	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal("Unable to determine hostame", err)
		os.Exit(1)
	}

	reg := prometheus.NewRegistry()
	bridge_metric := prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: *metricPrefix + "linux_bridge_present", Help: "Indicates if a bridge is present on the agent"},
		[]string{"hostname", "bridge"})
	netns_metric := prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: *metricPrefix + "netns_present", Help: "Indicates if a network namespace is present on the agent"},
		[]string{"hostname", "netns"})
	reg.MustRegister(bridge_metric)
	reg.MustRegister(netns_metric)

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	server := &http.Server{Addr: *listenAddress}

	go func() {
		log.Printf("Starting prometheus http server on %s\n", *listenAddress)
		if err := server.ListenAndServe(); err != nil && err.Error() != "http: Server closed" {
			log.Fatalf("Error starting prometheus http server: %s", err)
			os.Exit(-1)
		}
	}()

	sig := make(chan os.Signal)
	loopSig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	signal.Notify(loopSig, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	go loopFileSystem(hostname, *bridgesPath, "brq", "bridge", *updateInterval, *setCurrentTime, bridge_metric, loopSig)
	go loopFileSystem(hostname, *netNsPath, "qdhcp-", "netns", *updateInterval, *setCurrentTime, netns_metric, loopSig)

	<-sig
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
	log.Println("http server stopped")
}
