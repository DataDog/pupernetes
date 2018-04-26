package run

import (
	"github.com/golang/glog"
	"sync"
)

type State struct {
	sync.RWMutex

	apiServerProbeLastError string
	ready                   bool

	kubeletProbeFail  int
	kubeletPodRunning int
}

func (s *State) IsReady() bool {
	s.RLock()
	defer s.RUnlock()
	return s.ready
}

func (s *State) setReady() {
	s.Lock()
	s.ready = true
	s.Unlock()
	// Ignore errors
	notifySystemd()
}

func (s *State) setAPIServerProbeLastError(msg string) {
	s.Lock()
	if s.apiServerProbeLastError != msg {
		glog.Infof("Kubenertes apiserver not ready yet: %s", msg)
		s.apiServerProbeLastError = msg
	}
	s.Unlock()
}

func (s *State) incKubeletProbeFail() {
	s.Lock()
	s.kubeletProbeFail++
	s.Unlock()
}

func (s *State) getKubeletProbeFail() int {
	s.RLock()
	defer s.RUnlock()
	return s.kubeletProbeFail
}

func (s *State) setKubeletPodRunning(nb int) {
	s.Lock()
	if s.kubeletPodRunning != nb {
		glog.Infof("Kubelet is running %d pods", nb)
		s.kubeletPodRunning = nb
	}
	s.Unlock()
}
