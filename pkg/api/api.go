// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package api

import (
	"net/http"
	// register pprof handlers with its package init
	_ "net/http/pprof"
	"os"
	"syscall"
	"time"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
	corev1 "k8s.io/api/core/v1"

	"github.com/DataDog/pupernetes/pkg/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	stopRoute  = "/stop"
	applyRoute = "/apply"
	resetRoute = "/reset"
)

// HandlerAPI handles the API calls
type HandlerAPI struct {
	sigChan        chan os.Signal
	resetNamespace func(namespaces *corev1.NamespaceList) error
	isReady        func() bool
	apply          chan struct{}
}

func (h *HandlerAPI) stopHandler(_ http.ResponseWriter, _ *http.Request) {
	h.sigChan <- syscall.SIGTERM
}

func (h *HandlerAPI) applyHandler(_ http.ResponseWriter, _ *http.Request) {
	h.apply <- struct{}{}
}

func (h *HandlerAPI) resetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespaceName, ok := vars["namespace"]
	if !ok || namespaceName == "" {
		glog.Warningf("Invalid namespace %v", vars)
		http.NotFound(w, r)
		return
	}
	namespaceItem := corev1.Namespace{}
	namespaceItem.Name = namespaceName
	glog.Infof("Resetting namespace %q ...", namespaceItem.Name)
	err := h.resetNamespace(&corev1.NamespaceList{
		Items: []corev1.Namespace{namespaceItem},
	})
	if err != nil {
		glog.Errorf("Cannot reset namespace %s: %v", namespaceName, err)
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(200)
}

func (h *HandlerAPI) isReadyHandler(w http.ResponseWriter, _ *http.Request) {
	if h.isReady() {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
		return
	}
	w.WriteHeader(500)
	w.Write([]byte("not ready yet"))
	return
}

// NewAPI returns the API HTTP server
func NewAPI(sigChan chan os.Signal, resetNamespaceFn func(namespaces *corev1.NamespaceList) error, isReadyFn func() bool, apply chan struct{}) *http.Server {
	h := HandlerAPI{
		sigChan:        sigChan,
		resetNamespace: resetNamespaceFn,
		isReady:        isReadyFn,
		apply:          apply,
	}
	r := mux.NewRouter()

	// POSTs
	r.Methods("POST").Path(stopRoute).HandlerFunc(h.stopHandler)
	r.Methods("POST").Path(applyRoute).HandlerFunc(h.applyHandler)
	r.Methods("POST").Path(resetRoute + "/{namespace}").HandlerFunc(h.resetHandler)

	// GETs
	r.Methods("GET").Path("/ready").HandlerFunc(h.isReadyHandler)

	// monitoring
	r.Methods("GET").Path("/metrics").Handler(promhttp.Handler())

	// this is the cleanest way I found to register all pprof routes,
	// see https://stackoverflow.com/questions/19591065/profiling-go-web-application-built-with-gorillas-mux-with-net-http-pprof
	r.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)

	srv := &http.Server{
		Handler:      r,
		Addr:         config.ViperConfig.GetString("bind-address"),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	return srv
}
