package run

import (
	"fmt"
	"github.com/DataDog/pupernetes/pkg/logging"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"strings"
)

func (r *Runtime) httpProbe(url string) error {
	resp, err := r.httpClient.Get(url)
	if err != nil {
		glog.V(5).Infof("HTTP probe %s failed: %v", url, err)
		return err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf("Unexpected error when reading body of %s: %s", url, err)
		return err
	}
	content := string(b)
	defer resp.Body.Close()
	glog.V(10).Infof("%s %q", url, content)
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status code for %s: %d", url, resp.StatusCode)
		glog.V(5).Infof("HTTP probe %s failed: %v", url, err)
		return err
	}
	return nil
}

func (r *Runtime) probeUnitStatuses() error {
	units, err := r.env.GetDBUSClient().ListUnitsByNames(r.env.GetSystemdUnits())
	if err != nil {
		glog.Errorf("Unexpected error: %v", err)
		return err
	}
	var errs []string
	r.journalTailerMutex.Lock()
	defer r.journalTailerMutex.Unlock()
	for _, u := range units {
		s := fmt.Sprintf("unit %q with load state %q is %q", u.Name, u.LoadState, u.SubState)
		glog.V(3).Infof("%s", s)
		switch u.SubState {
		case "running":
			continue
		case "start":
			continue
		}
		glog.Errorf("Unexpected state of: %s", s)
		errs = append(errs, s)
		jt, err := logging.NewJournalTailer(u.Name, r.runTimestamp)
		if err != nil {
			glog.Errorf("Fail to create the journal tailer for %s: %v", u.Name, err)
			continue
		}
		r.journalTailers[u.Name] = jt
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, ", "))
	}
	return nil
}
