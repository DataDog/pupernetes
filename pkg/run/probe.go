package run

import (
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
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
