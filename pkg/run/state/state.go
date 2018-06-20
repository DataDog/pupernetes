package state

import (
	"sync"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
)

type State struct {
	sync.RWMutex

	apiServerProbeLastError string
	ready                   bool

	kubeletProbeFailures  int
	kubeletAPIPodRunning  int
	kubeletLogsPodRunning int

	promVersion prometheus.Gauge

	promStateReady            prometheus.Gauge
	promKubeletAPIPodRunning  prometheus.Gauge
	promKubeletLogsPodRunning prometheus.Gauge

	promKubeletProbeFailures prometheus.Counter
}

func NewState() (*State, error) {
	s := &State{
		promVersion: prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        "pupernetes_versions",
			Help:        "Pupernetes versions",
			ConstLabels: prometheus.Labels{},
			// TODO record all versions
		}),
		promStateReady: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "pupernetes_ready",
			Help: "Boolean for pupernetes readiness",
		}),
		promKubeletAPIPodRunning: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "pupernetes_kubelet_api_pods_running",
			Help: "Number of kubelet API pods running",
		}),
		promKubeletLogsPodRunning: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "pupernetes_kubelet_logs_pods_running",
			Help: "Number of kubelet logs pods running",
		}),
		promKubeletProbeFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "pupernetes_kubelet_probe_failures",
			Help: "Total number of kubelet probe failures",
		}),
	}
	err := prometheus.Register(s.promVersion)
	if err != nil {
		return nil, err
	}
	err = prometheus.Register(s.promStateReady)
	if err != nil {
		return nil, err
	}
	err = prometheus.Register(s.promKubeletAPIPodRunning)
	if err != nil {
		return nil, err
	}
	err = prometheus.Register(s.promKubeletLogsPodRunning)
	if err != nil {
		return nil, err
	}
	err = prometheus.Register(s.promKubeletProbeFailures)
	if err != nil {
		return nil, err
	}
	s.promVersion.Inc()
	return s, nil
}

func (s *State) IsReady() bool {
	s.RLock()
	defer s.RUnlock()
	return s.ready
}

func (s *State) SetReady() {
	s.Lock()
	s.ready = true
	s.Unlock()
	// Ignore errors
	notifySystemd()
	s.promStateReady.Set(1)
}

func (s *State) SetAPIServerProbeLastError(msg string) {
	s.Lock()
	if s.apiServerProbeLastError != msg {
		glog.Infof("Kubenertes apiserver not ready yet: %s", msg)
		s.apiServerProbeLastError = msg
	}
	s.Unlock()
}

func (s *State) IncKubeletProbeFailures() {
	s.Lock()
	s.kubeletProbeFailures++
	s.Unlock()
	s.promKubeletProbeFailures.Inc()
}

func (s *State) GetKubeletProbeFail() int {
	s.RLock()
	defer s.RUnlock()
	return s.kubeletProbeFailures
}

func (s *State) SetKubeletAPIPodRunning(nb int) {
	s.Lock()
	if s.kubeletAPIPodRunning != nb {
		glog.Infof("Kubelet API reports %d running pods", nb)
		s.kubeletAPIPodRunning = nb
	}
	s.Unlock()
	s.promKubeletAPIPodRunning.Set(float64(nb))
}

func (s *State) SetKubeletLogsPodRunning(nb int) {
	s.Lock()
	if s.kubeletLogsPodRunning != nb {
		glog.Infof("Kubelet log reports %d running pods", nb)
		s.kubeletLogsPodRunning = nb
	}
	s.Unlock()
	s.promKubeletLogsPodRunning.Set(float64(nb))
}

func (s *State) GetKubeletLogsPodRunning() int {
	s.RLock()
	defer s.RUnlock()
	return s.kubeletLogsPodRunning
}
