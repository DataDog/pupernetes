package run

import (
	"github.com/golang/glog"
	"sync"
)

type State struct {
	sync.RWMutex

	apiServerProbeLastError string
	ready                   bool

	kubeletProbeFailures  int
	kubeletAPIPodRunning  int
	kubeletLogsPodRunning int
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
}

func (s *State) SetKubeletLogsPodRunning(nb int) {
	s.Lock()
	if s.kubeletLogsPodRunning != nb {
		glog.Infof("Kubelet log reports %d running pods", nb)
		s.kubeletLogsPodRunning = nb
	}
	s.Unlock()
}

func (s *State) GetKubeletLogsPodRunning() int {
	s.RLock()
	defer s.RUnlock()
	return s.kubeletLogsPodRunning
}
