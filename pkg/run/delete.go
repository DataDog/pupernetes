// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package run

import (
	"fmt"
	"strings"

	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Runtime) deleteDeployments(namespaces *corev1.NamespaceList) error {
	var errs []string
	for _, ns := range namespaces.Items {
		toDelete, err := r.env.GetKubernetesClient().AppsV1beta1().Deployments(ns.Name).List(v1.ListOptions{})
		if err != nil {
			glog.Errorf("Cannot get Deployments in ns %q: %v", ns.Name, err)
			return err
		}
		glog.V(4).Infof("Deleting %d Deployments in ns %q ...", len(toDelete.Items), ns.Name)
		for _, elt := range toDelete.Items {
			err = r.env.GetKubernetesClient().AppsV1beta1().Deployments(elt.Namespace).Delete(elt.Name, r.kubeDeleteOption)
			if err != nil {
				glog.Errorf("Cannot delete Deployments %s in ns %q: %v", elt.Name, elt.Namespace, err)
				errs = append(errs, err.Error())
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("unexpected errors during delete Deployments: %s", strings.Join(errs, ", "))
	}
	return nil
}

func (r *Runtime) deleteDaemonset(namespaces *corev1.NamespaceList) error {
	var errs []string
	for _, ns := range namespaces.Items {
		toDelete, err := r.env.GetKubernetesClient().AppsV1().DaemonSets(ns.Name).List(v1.ListOptions{})
		if err != nil {
			glog.Errorf("Cannot get DaemonSets in ns %q: %v", ns.Name, err)
			return err
		}
		glog.V(4).Infof("Deleting %d DaemonSets in ns %q ...", len(toDelete.Items), ns.Name)
		for _, elt := range toDelete.Items {
			err = r.env.GetKubernetesClient().AppsV1().DaemonSets(elt.Namespace).Delete(elt.Name, r.kubeDeleteOption)
			if err != nil {
				glog.Errorf("Cannot delete DaemonSets %s in ns %q: %v", elt.Name, elt.Namespace, err)
				errs = append(errs, err.Error())
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("unexpected errors during delete DaemonSets: %s", strings.Join(errs, ", "))
	}
	return nil
}

func (r *Runtime) deleteReplicationController(namespaces *corev1.NamespaceList) error {
	var errs []string
	for _, ns := range namespaces.Items {
		toDelete, err := r.env.GetKubernetesClient().CoreV1().ReplicationControllers(ns.Name).List(v1.ListOptions{})
		if err != nil {
			glog.Errorf("Cannot get ReplicationControllers in ns %q: %v", ns.Name, err)
			return err
		}
		glog.V(4).Infof("Deleting %d ReplicationControllers in ns %q ...", len(toDelete.Items), ns.Name)
		for _, elt := range toDelete.Items {
			err = r.env.GetKubernetesClient().AppsV1().DaemonSets(elt.Namespace).Delete(elt.Name, r.kubeDeleteOption)
			if err != nil {
				glog.Errorf("Cannot delete ReplicationControllers %s in ns %q: %v", elt.Name, elt.Namespace, err)
				errs = append(errs, err.Error())
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("unexpected errors during delete ReplicationControllers: %s", strings.Join(errs, ", "))
	}
	return nil
}

func (r *Runtime) deleteReplicaSets(namespaces *corev1.NamespaceList) error {
	var errs []string
	for _, ns := range namespaces.Items {
		toDelete, err := r.env.GetKubernetesClient().AppsV1().ReplicaSets(ns.Name).List(v1.ListOptions{})
		if err != nil {
			glog.Errorf("Cannot get ReplicaSets in ns %q: %v", ns.Name, err)
			return err
		}
		glog.V(4).Infof("Deleting %d ReplicaSets in ns %q ...", len(toDelete.Items), ns.Name)
		for _, elt := range toDelete.Items {
			err = r.env.GetKubernetesClient().AppsV1().ReplicaSets(elt.Namespace).Delete(elt.Name, r.kubeDeleteOption)
			if err != nil && !errors.IsNotFound(err) {
				glog.Errorf("Cannot delete ReplicaSets %s in ns %q: %v", elt.Name, elt.Namespace, err)
				errs = append(errs, err.Error())
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("unexpected errors during delete ReplicaSets: %s", strings.Join(errs, ", "))
	}
	return nil
}

func (r *Runtime) deleteJobs(namespaces *corev1.NamespaceList) error {
	var errs []string
	for _, ns := range namespaces.Items {
		toDelete, err := r.env.GetKubernetesClient().BatchV1().Jobs(ns.Name).List(v1.ListOptions{})
		if err != nil {
			glog.Errorf("Cannot get Jobs in ns %q: %v", ns.Name, err)
			return err
		}
		glog.V(4).Infof("Deleting %d Jobs in ns %q ...", len(toDelete.Items), ns.Name)
		for _, elt := range toDelete.Items {
			err = r.env.GetKubernetesClient().BatchV1().Jobs(elt.Namespace).Delete(elt.Name, r.kubeDeleteOption)
			if err != nil && !errors.IsNotFound(err) {
				glog.Errorf("Cannot delete Jobs %s in ns %q: %v", elt.Name, elt.Namespace, err)
				errs = append(errs, err.Error())
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("unexpected errors during delete Jobs: %s", strings.Join(errs, ", "))
	}
	return nil
}

func (r *Runtime) deletePods(namespaces *corev1.NamespaceList) error {
	var errs []string
	for _, ns := range namespaces.Items {
		toDelete, err := r.env.GetKubernetesClient().CoreV1().Pods(ns.Name).List(v1.ListOptions{})
		if err != nil {
			glog.Errorf("Cannot get Pods in ns %q: %v", ns.Name, err)
			return err
		}
		glog.V(4).Infof("Deleting %d Pods in ns %s ...", len(toDelete.Items), ns.Name)
		for _, elt := range toDelete.Items {
			err = r.env.GetKubernetesClient().CoreV1().Pods(elt.Namespace).Delete(elt.Name, r.kubeDeleteOption)
			if err != nil && !errors.IsNotFound(err) {
				glog.Errorf("Cannot delete Pods %s in ns %q: %v", elt.Name, elt.Namespace, err)
				errs = append(errs, err.Error())
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("unexpected errors during delete Pods: %s", strings.Join(errs, ", "))
	}
	return nil
}

func (r *Runtime) deleteServices(namespaces *corev1.NamespaceList) error {
	var errs []string
	for _, ns := range namespaces.Items {
		toDelete, err := r.env.GetKubernetesClient().CoreV1().Services(ns.Name).List(v1.ListOptions{})
		if err != nil {
			glog.Errorf("Cannot get Services in ns %q: %v", ns.Name, err)
			return err
		}
		glog.V(4).Infof("Deleting %d Services in ns %s ...", len(toDelete.Items), ns.Name)
		for _, elt := range toDelete.Items {
			err = r.env.GetKubernetesClient().CoreV1().Services(elt.Namespace).Delete(elt.Name, r.kubeDeleteOption)
			if err != nil && !errors.IsNotFound(err) {
				glog.Errorf("Cannot delete Services %s in ns %q: %v", elt.Name, elt.Namespace, err)
				errs = append(errs, err.Error())
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("unexpected errors during delete Services: %s", strings.Join(errs, ", "))
	}
	return nil
}

func (r *Runtime) deleteEndpoints(namespaces *corev1.NamespaceList) error {
	var errs []string
	for _, ns := range namespaces.Items {
		toDelete, err := r.env.GetKubernetesClient().CoreV1().Endpoints(ns.Name).List(v1.ListOptions{})
		if err != nil {
			glog.Errorf("Cannot get Endpoints in ns %q: %v", ns.Name, err)
			return err
		}
		glog.V(4).Infof("Deleting %d Endpoints in ns %s ...", len(toDelete.Items), ns.Name)
		for _, elt := range toDelete.Items {
			err = r.env.GetKubernetesClient().CoreV1().Endpoints(elt.Namespace).Delete(elt.Name, r.kubeDeleteOption)
			if err != nil && !errors.IsNotFound(err) {
				glog.Errorf("Cannot delete Endpoints %s in ns %q: %v", elt.Name, elt.Namespace, err)
				errs = append(errs, err.Error())
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("unexpected errors during delete Endpoints: %s", strings.Join(errs, ", "))
	}
	return nil
}

func (r *Runtime) deleteConfigMaps(namespaces *corev1.NamespaceList) error {
	var errs []string
	for _, ns := range namespaces.Items {
		toDelete, err := r.env.GetKubernetesClient().CoreV1().ConfigMaps(ns.Name).List(v1.ListOptions{})
		if err != nil {
			glog.Errorf("Cannot get ConfigMaps in ns %q: %v", ns.Name, err)
			return err
		}
		glog.V(4).Infof("Deleting %d ConfigMaps in ns %s ...", len(toDelete.Items), ns.Name)
		for _, elt := range toDelete.Items {
			err = r.env.GetKubernetesClient().CoreV1().ConfigMaps(elt.Namespace).Delete(elt.Name, r.kubeDeleteOption)
			if err != nil && !errors.IsNotFound(err) {
				glog.Errorf("Cannot delete ConfigMaps %s in ns %q: %v", elt.Name, elt.Namespace, err)
				errs = append(errs, err.Error())
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("unexpected errors during delete ConfigMaps: %s", strings.Join(errs, ", "))
	}
	return nil
}

func (r *Runtime) deleteSecrets(namespaces *corev1.NamespaceList) error {
	var errs []string
	for _, ns := range namespaces.Items {
		toDelete, err := r.env.GetKubernetesClient().CoreV1().Secrets(ns.Name).List(v1.ListOptions{})
		if err != nil {
			glog.Errorf("Cannot get Secrets in ns %q: %v", ns.Name, err)
			return err
		}
		glog.V(4).Infof("Deleting %d Secrets in ns %s ...", len(toDelete.Items), ns.Name)
		for _, elt := range toDelete.Items {
			err = r.env.GetKubernetesClient().CoreV1().Secrets(elt.Namespace).Delete(elt.Name, r.kubeDeleteOption)
			if err != nil && !errors.IsNotFound(err) {
				glog.Errorf("Cannot delete Secrets %s in ns %q: %v", elt.Name, elt.Namespace, err)
				errs = append(errs, err.Error())
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("unexpected errors during delete Secrets: %s", strings.Join(errs, ", "))
	}
	return nil
}

func (r *Runtime) deleteServiceAccounts(namespaces *corev1.NamespaceList) error {
	var errs []string
	for _, ns := range namespaces.Items {
		toDelete, err := r.env.GetKubernetesClient().CoreV1().ServiceAccounts(ns.Name).List(v1.ListOptions{})
		if err != nil {
			glog.Errorf("Cannot get ServiceAccounts in ns %q: %v", ns.Name, err)
			return err
		}
		glog.V(4).Infof("Deleting %d ServiceAccounts in ns %s ...", len(toDelete.Items), ns.Name)
		for _, elt := range toDelete.Items {
			err = r.env.GetKubernetesClient().CoreV1().ServiceAccounts(elt.Namespace).Delete(elt.Name, r.kubeDeleteOption)
			if err != nil && !errors.IsNotFound(err) {
				glog.Errorf("Cannot delete ServiceAccounts %s in ns %q: %v", elt.Name, elt.Namespace, err)
				errs = append(errs, err.Error())
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("unexpected errors during delete ServiceAccounts: %s", strings.Join(errs, ", "))
	}
	return nil
}

// DeleteAPIManifests execute Kubernetes delete API calls to delete the following resources:
// - services
// - Jobs
// - Deployments
// - Daemonset
// - ReplicationController
// - ReplicaSets
// - Pods
// - Endpoints
// - ConfigMaps
// - Secrets
// - ServiceAccounts
func (r *Runtime) DeleteAPIManifests(namespaces *corev1.NamespaceList) error {
	fnList := []func(ns *corev1.NamespaceList) error{
		r.deleteServices,
		r.deleteJobs,
		r.deleteDeployments,
		r.deleteDaemonset,
		r.deleteReplicationController,
		r.deleteReplicaSets,
		r.deletePods,
		r.deleteEndpoints,
		r.deleteConfigMaps,
		r.deleteSecrets,
		r.deleteServiceAccounts,
	}
	errChan := make(chan error, len(fnList))

	for _, elt := range fnList {
		go func(fn func(ns *corev1.NamespaceList) error) {
			errChan <- fn(namespaces)
		}(elt)
	}
	var err error
	var errMsgList []string
	for i := 0; i < len(fnList); i++ {
		err = <-errChan
		if err != nil {
			errMsgList = append(errMsgList, err.Error())
		}
	}
	close(errChan)

	// Remaining Pods
	err = r.deletePods(namespaces)
	if err != nil {
		errMsgList = append(errMsgList, err.Error())
	}
	if len(errMsgList) == 0 {
		glog.Infof("Graceful deleted API resources in %d namespaces", len(namespaces.Items))
		return nil
	}
	return fmt.Errorf("cannot delete API resources: %v", strings.Join(errMsgList, ","))
}
