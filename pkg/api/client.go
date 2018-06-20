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

// ResetNamespace executes an API call to the pupernetes API to reset
// the namespace in parameter. The namespace can be like ns/default or just default
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
	return doPOST(apiAddress, fmt.Sprintf("%s/%s", resetRoute, namespace))
}

// Apply executes an API call to the pupernetes API to force an apply of the "manifest-api" directory
func Apply(apiAddress string) error {
	glog.Infof("Applying ...")
	return doPOST(apiAddress, applyRoute)
}

func doPOST(apiAddress, apiRoute string) error {
	glog.Infof("Calling POST %s ...", apiRoute)
	c := &http.Client{}
	c.Timeout = time.Second * 5

	u, err := url.Parse(fmt.Sprintf("http://%s%s", apiAddress, apiRoute))
	if err != nil {
		glog.Errorf("Error during urlParse: %v", err)
		return err
	}
	glog.V(3).Infof("Using url: %s", u.String())
	resp, err := c.Post(u.String(), "application/json", nil)
	if err != nil {
		glog.Errorf("Unexpected error during POST %s: %v", u.String(), err)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("non OK status code when POST %s: %d", u.String(), resp.StatusCode)
		glog.Errorf("Cannot POST: %v", err)
		return err
	}
	glog.Infof("POST on %s successfully executed: %d", u.String(), resp.StatusCode)
	return nil
}
