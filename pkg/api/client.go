package api

import (
	"fmt"
	"github.com/golang/glog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	namespacePrefix = "namespace/"
)

func ResetNamespace(apiAddress, namespace string) error {
	if strings.HasPrefix(namespace, namespacePrefix) {
		glog.V(4).Infof("Stripping namespace %q", namespace)
		namespace = namespace[len(namespacePrefix):]
		glog.V(4).Infof("Namespace renamed as: %q", namespace)
	}
	if namespace == "" {
		err := fmt.Errorf("empty namespace")
		glog.Infof("Cannot continue: %v", err)
		return err
	}
	glog.Infof("Resetting namespace %q ...", namespace)
	c := &http.Client{}
	c.Timeout = time.Second * 5

	u, err := url.Parse(fmt.Sprintf("http://%s/reset/%s", apiAddress, namespace))
	if err != nil {
		glog.Errorf("Error during urlParse: %v", err)
		return err
	}
	glog.V(3).Infof("Using url: %s", u.String())
	resp, err := c.Post(u.String(), "application/json", nil)
	if err != nil {
		glog.Errorf("Unexpected error during reset namespace %s: %v", namespace, err)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("non OK status code when deleting ns %s: %d", namespace, resp.StatusCode)
		glog.Errorf("Cannot delete namespace: %v", err)
		return err
	}
	glog.Infof("Namespace %q successfully reset", namespace)
	return nil
}

func ReApply(apiAddress string) error {
	glog.Infof("Re-applying ...")
	c := &http.Client{}
	c.Timeout = time.Second * 5

	u, err := url.Parse(fmt.Sprintf("http://%s/re-apply", apiAddress))
	if err != nil {
		glog.Errorf("Error during urlParse: %v", err)
		return err
	}
	glog.V(3).Infof("Using url: %s", u.String())
	resp, err := c.Post(u.String(), "application/json", nil)
	if err != nil {
		glog.Errorf("Unexpected error during re-applying: %v", err)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("non OK status code when re-applying: %d", resp.StatusCode)
		glog.Errorf("Cannot re-apply: %v", err)
		return err
	}
	glog.Infof("Re-apply successful")
	return nil
}
