// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package api

import (
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/DataDog/pupernetes/pkg/config"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
	corev1 "k8s.io/api/core/v1"
)

// HandlerAPI handles the API calls
type HandlerAPI struct {
	sigChan        chan os.Signal
	resetNamespace func(namespaces *corev1.NamespaceList) error
	isReady        func() bool
}

func (h *HandlerAPI) stopHandler(_ http.ResponseWriter, _ *http.Request) {
	h.sigChan <- syscall.SIGTERM
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
	glog.Infof("Resetting namespace %s ...", namespaceItem.Name)
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
func NewAPI(sigChan chan os.Signal, resetNamespaceFn func(namespaces *corev1.NamespaceList) error, isReadyFn func() bool) *http.Server {
	h := HandlerAPI{
		sigChan:        sigChan,
		resetNamespace: resetNamespaceFn,
		isReady:        isReadyFn,
	}
	r := mux.NewRouter()
	r.Methods("POST").Path("/stop").HandlerFunc(h.stopHandler)
	r.Methods("POST").Path("/reset/{namespace}").HandlerFunc(h.resetHandler)
	r.Methods("GET").Path("/ready").HandlerFunc(h.isReadyHandler)

	srv := &http.Server{
		Handler:      r,
		Addr:         config.ViperConfig.GetString("bind-address"),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	return srv
}
