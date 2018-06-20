package run

import (
	"github.com/golang/glog"
	"sync"
)

// State keeps track of the current stats
type State struct {
	sync.RWMutex

	apiServerProbeLastError string
	ready                   bool

	kubeletProbeFailures  int
	kubeletAPIPodRunning  int
	kubeletLogsPodRunning int
}

// IsReady returns if the kube-apiserver is available and the manifests are applied
func (s *State) IsReady() bool {
	s.RLock()
	defer s.RUnlock()
	return s.ready
}

// SetReady is the trigger to mark pupernetes as ready.
// It notify systemd of its readiness
func (s *State) SetReady() {
	s.Lock()
	s.ready = true
	s.Unlock()
	// Ignore errors
	notifySystemd()
}

// SetAPIServerProbeLastError keep track of the latest error message and display only
// if there is a a diff from the last record
func (s *State) SetAPIServerProbeLastError(msg string) {
	s.Lock()
	if s.apiServerProbeLastError != msg {
		glog.Infof("Kubenertes apiserver not ready yet: %s", msg)
		s.apiServerProbeLastError = msg
	}
	s.Unlock()
}

// IncKubeletProbeFailures increment the number of kubelet failures
func (s *State) IncKubeletProbeFailures() {
	s.Lock()
	s.kubeletProbeFailures++
	s.Unlock()
}

// GetKubeletProbeFail returns the number of kubelet failures
func (s *State) GetKubeletProbeFail() int {
	s.RLock()
	defer s.RUnlock()
	return s.kubeletProbeFailures
}

// SetKubeletAPIPodRunning keep track of the number of kubelet Pods and display only
// if there is a a diff from the last record
func (s *State) SetKubeletAPIPodRunning(nb int) {
	s.Lock()
	if s.kubeletAPIPodRunning != nb {
		glog.Infof("Kubelet API reports %d running pods", nb)
		s.kubeletAPIPodRunning = nb
	}
	s.Unlock()
}

// SetKubeletLogsPodRunning keep track of the number of kubelet Pods in /var/log/pods and display only
// if there is a a diff from the last record
func (s *State) SetKubeletLogsPodRunning(nb int) {
	s.Lock()
	if s.kubeletLogsPodRunning != nb {
		glog.Infof("Kubelet log reports %d running pods", nb)
		s.kubeletLogsPodRunning = nb
	}
	s.Unlock()
}

// GetKubeletLogsPodRunning returns the number of kubelet Pods in /var/log/pods
func (s *State) GetKubeletLogsPodRunning() int {
	s.RLock()
	defer s.RUnlock()
	return s.kubeletLogsPodRunning
}
