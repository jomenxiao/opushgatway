package metrics

import (
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/ngaut/log"
	"github.com/prometheus/pushgateway/storage"
)

func Delete(w http.ResponseWriter, r *http.Request) {
	var mtx sync.Mutex // Protects ps.
	mtx.Lock()
	job := mux.Vars(r)["job"]
	restpath := mux.Vars(r)["rest"]
	mtx.Unlock()

	labels, err := splitLabels(restpath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Debugf("Failed to parse URL: %v, %v", restpath, err.Error())
		return
	}
	if job == "" {
		http.Error(w, "job name is required", http.StatusBadRequest)
		log.Debug("job name is required")
		return
	}
	labels["job"] = job
	ms.SubmitWriteRequest(storage.WriteRequest{
		Labels:    labels,
		Timestamp: time.Now(),
	})
	w.WriteHeader(http.StatusAccepted)
}

func LegacyDelete(w http.ResponseWriter, r *http.Request) {
	var mtx sync.Mutex // Protects ps.
	mtx.Lock()
	job := mux.Vars(r)["job"]
	instance := mux.Vars(r)["instance"]
	mtx.Unlock()

	if job == "" {
		http.Error(w, "job name is required", http.StatusBadRequest)
		log.Debug("job name is required")
		return
	}
	labels := map[string]string{"job": job}
	if instance != "" {
		labels["instance"] = instance
	}
	ms.SubmitWriteRequest(storage.WriteRequest{
		Labels:    labels,
		Timestamp: time.Now(),
	})
	w.WriteHeader(http.StatusAccepted)

}
