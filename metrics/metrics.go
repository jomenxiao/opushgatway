package metrics

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func init() {
	prometheus.MustRegister(version.NewCollector("pushgateway"))
}

func initMetrics(metricsBasePath string, persistenceInterval int) {
	persistenceFile := filepath.Join(metricsBasePath, "metrics.file")
	ms = NewDiskMetricStore(persistenceFile, time.Duration(persistenceInterval))
	prometheus.SetMetricFamilyInjectionHook(ms.GetMetricFamilies)
}

//InstanceMetrics is accept main
func InstanceMetrics(ctx context.Context, metricsBasePath string, listenPort int, persistenceInterval int) error {
	log.Infof("starting metrics")
	if err := os.MkdirAll(metricsBasePath, 0666); err != nil {
		return err
	}

	initMetrics(metricsBasePath, persistenceInterval)
	createHTTP(metricsBasePath, listenPort, persistenceInterval)
	return nil
}

//createHttp HTTP
func createHTTP(metricsBasePath string, listenPort int, persistenceInterval int) {

	// create http router
	r := mux.NewRouter()
	r.Handle("/metrics", prometheus.Handler()).Methods("GET")
	rs := r.PathPrefix("/metrics/job").Subrouter()
	// Handlers for pushing and deleting metrics.
	rs.HandleFunc("/{job}/{rest:.*}", Push).Methods("PUT", "POST", "DELETE")
	rs.HandleFunc("/{job}/{rest:.*}", Delete).Methods("DELETE")
	rs.HandleFunc("/{job}", Push).Methods("PUT", "POST")
	rs.HandleFunc("/{job}/{rest:.*}", Delete).Methods("DELETE")
	// Handlers for the deprecated API.
	rsl := r.PathPrefix("/metrics/jobs").Subrouter()
	rsl.HandleFunc("/{job}/instance/{instance}", LegacyPush).Methods("PUT", "POST")
	rsl.HandleFunc("/{job}/instance/{instance}", LegacyDelete).Methods("DELETE")
	rsl.HandleFunc("/{job}", LegacyPush).Methods("PUT", "POST")
	rsl.HandleFunc("/{job}", LegacyDelete).Methods("DELETE")

	log.Infof("Listening on %d", listenPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", listenPort), r))
	time.Sleep(time.Second)
	if err := ms.Shutdown(); err != nil {
		log.Errorln("Problem shutting down metric storage:", err)
	}
}
