/*
Copyright 2016 The Rook Authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package mgr for the Ceph manager.
package mgr

import (
	"fmt"
	"path"
	"strconv"

	"github.com/coreos/pkg/capnslog"
	cephv1beta1 "github.com/rook/rook/pkg/apis/ceph.rook.io/v1beta1"
	rookalpha "github.com/rook/rook/pkg/apis/rook.io/v1alpha2"
	"github.com/rook/rook/pkg/clusterd"
	"github.com/rook/rook/pkg/daemon/ceph/client"
	opmon "github.com/rook/rook/pkg/operator/ceph/cluster/mon"
	"github.com/rook/rook/pkg/operator/k8sutil"
	"github.com/rook/rook/pkg/util"
	"k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/rook/rook/pkg/daemon/ceph/ceph"
)

var logger = capnslog.NewPackageLogger("github.com/rook/rook", "op-mgr")

const (
	appName              = "rook-ceph-mgr"
	keyringName          = "keyring"
	prometheusModuleName = "prometheus"
	dashboardModuleName  = "dashboard"
	metricsPort          = 9283
	dashboardPort        = 7000
)

var mgrNames = []string{"a", "b"}

// Cluster is the ceph mgr manager
type Cluster struct {
	Namespace   string
	Version     string
	Replicas    int
	placement   rookalpha.Placement
	context     *clusterd.Context
	dataDir     string
	HostNetwork bool
	resources   v1.ResourceRequirements
	ownerRef    metav1.OwnerReference
	dashboard   cephv1beta1.DashboardSpec
}

// New creates an instance of the mgr
func New(context *clusterd.Context, namespace, version string, placement rookalpha.Placement, hostNetwork bool, dashboard cephv1beta1.DashboardSpec,
	resources v1.ResourceRequirements, ownerRef metav1.OwnerReference) *Cluster {
	return &Cluster{
		context:     context,
		Namespace:   namespace,
		placement:   placement,
		Version:     version,
		Replicas:    1,
		dataDir:     k8sutil.DataDir,
		dashboard:   dashboard,
		HostNetwork: hostNetwork,
		resources:   resources,
		ownerRef:    ownerRef,
	}
}

// Start the mgr instance
func (c *Cluster) Start() error {
	logger.Infof("start running mgr")

	for i := 0; i < c.Replicas; i++ {
		if i >= len(mgrNames) {
			logger.Errorf("cannot have more than %d mgrs", len(mgrNames))
			break
		}
		daemonName := mgrNames[i]
		name := fmt.Sprintf("%s-%s", appName, daemonName)
		if err := c.createKeyring(c.Namespace, name, daemonName); err != nil {
			return fmt.Errorf("failed to create %s keyring. %+v", name, err)
		}

		// start the deployment
		// Why do we start multiple deployments instead of starting one deployment with replicas?
		deployment := c.makeDeployment(name, daemonName)
		if _, err := c.context.Clientset.ExtensionsV1beta1().Deployments(c.Namespace).Create(deployment); err != nil {
			if !errors.IsAlreadyExists(err) {
				return fmt.Errorf("failed to create %s deployment. %+v", name, err)
			}
			logger.Infof("%s deployment already exists", name)
		} else {
			logger.Infof("%s deployment started", name)
		}
	}

	if err := c.enablePrometheusModule(c.Namespace); err != nil {
		return fmt.Errorf("failed to enable mgr prometheus module. %+v", err)
	}

	// create the metrics service
	service := c.makeMetricsService(appName)
	if _, err := c.context.Clientset.CoreV1().Services(c.Namespace).Create(service); err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create mgr service. %+v", err)
		}
		logger.Infof("mgr metrics service already exists")
	} else {
		logger.Infof("mgr metrics service started")
	}

	return c.configureDashboard()
}

func (c *Cluster) configureDashboard() error {
	// enable or disable the dashboard module
	if err := c.configureDashboardModule(c.Namespace, c.dashboard.Enabled); err != nil {
		return fmt.Errorf("failed to enable mgr dashboard module. %+v", err)
	}

	dashboardService := c.makeDashboardService(appName)
	if c.dashboard.Enabled {
		// expose the dashboard service
		if _, err := c.context.Clientset.CoreV1().Services(c.Namespace).Create(dashboardService); err != nil {
			if !errors.IsAlreadyExists(err) {
				return fmt.Errorf("failed to create dashboard mgr service. %+v", err)
			}
			logger.Infof("dashboard service already exists")
		} else {
			logger.Infof("dashboard service started")
		}
	} else {
		// delete the dashboard service if it exists
		err := c.context.Clientset.CoreV1().Services(c.Namespace).Delete(dashboardService.Name, &metav1.DeleteOptions{})
		if err != nil && !errors.IsNotFound(err) {
			return fmt.Errorf("failed to delete dashboard service. %+v", err)
		}
	}

	return nil
}

func (c *Cluster) makeMetricsService(name string) *v1.Service {
	labels := c.getLabels()
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: c.Namespace,
			Labels:    labels,
		},
		Spec: v1.ServiceSpec{
			Selector: labels,
			Type:     v1.ServiceTypeClusterIP,
			Ports: []v1.ServicePort{
				{
					Name:     "http-metrics",
					Port:     int32(metricsPort),
					Protocol: v1.ProtocolTCP,
				},
			},
		},
	}

	k8sutil.SetOwnerRef(c.context.Clientset, c.Namespace, &svc.ObjectMeta, &c.ownerRef)
	return svc
}

func (c *Cluster) makeDashboardService(name string) *v1.Service {
	labels := c.getLabels()
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-dashboard", name),
			Namespace: c.Namespace,
			Labels:    labels,
		},
		Spec: v1.ServiceSpec{
			Selector: labels,
			Type:     v1.ServiceTypeClusterIP,
			Ports: []v1.ServicePort{
				{
					Name:     "http-dashboard",
					Port:     int32(dashboardPort),
					Protocol: v1.ProtocolTCP,
				},
			},
		},
	}
	k8sutil.SetOwnerRef(c.context.Clientset, c.Namespace, &svc.ObjectMeta, &c.ownerRef)
	return svc
}

func (c *Cluster) makeDeployment(name, daemonName string) *extensions.Deployment {
	clusterInfo, _, _, _ := opmon.LoadClusterInfo(c.context, c.Namespace)
	// TODO: Swallow any errors for now. Will need to handle this and return an error condition in
	//   makeDeployment before finalizing this work
	// if err != nil {
	// 	return fmt.Errorf("failed to load cluster information from clusters namespace %s: %+v", c.Namespace, err)
	// }

	confFile := getMgrConfFilePath(c.context.ConfigDir, daemonName, clusterInfo.Name)
	util.WriteFileToLog(logger, confFile)

	keyringPath := getMgrKeyringPath(c.context.ConfigDir, daemonName)
	util.WriteFileToLog(logger, keyringPath)

	podSpec := v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: c.getDaemonLabels(daemonName),
			Annotations: map[string]string{"prometheus.io/scrape": "true",
				"prometheus.io/port": strconv.Itoa(metricsPort)},
		},
		Spec: v1.PodSpec{
			Containers:     []v1.Container{c.mgrContainer(name, clusterInfo.Name, confFile, keyringPath)},
			InitContainers: []v1.Container{c.mgrInitContainer(name, daemonName, keyringPath, getMgrConfDir(c.context.ConfigDir, daemonName))},
			RestartPolicy:  v1.RestartPolicyAlways,
			Volumes: []v1.Volume{
				{Name: k8sutil.DataDirVolume, VolumeSource: v1.VolumeSource{EmptyDir: &v1.EmptyDirVolumeSource{}}},
				k8sutil.ConfigOverrideVolume(),
				{Name: "ceph-default-config-dir", VolumeSource: v1.VolumeSource{EmptyDir: &v1.EmptyDirVolumeSource{}}},
			},
			HostNetwork: c.HostNetwork,
		},
	}
	if c.HostNetwork {
		podSpec.Spec.DNSPolicy = v1.DNSClusterFirstWithHostNet
	}
	c.placement.ApplyToPodSpec(&podSpec.Spec)

	replicas := int32(1)
	d := &extensions.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: c.Namespace,
		},
		Spec: extensions.DeploymentSpec{Template: podSpec, Replicas: &replicas},
	}
	k8sutil.SetOwnerRef(c.context.Clientset, c.Namespace, &d.ObjectMeta, &c.ownerRef)
	return d
}

func (c *Cluster) mgrContainer(name, clusterName, confFilePath, keyringPath string) v1.Container {
	return v1.Container{
		Command: []string{
			"/usr/bin/ceph-mgr",
		},
		Args: []string{
			"--foreground",
			fmt.Sprintf("--cluster=%s", clusterName),
			fmt.Sprintf("--conf=%s", confFilePath),
			fmt.Sprintf("--keyring=%s", keyringPath),
		},
		// There is no need to name containers in the pod 'rook-ceph-mgr-<name>'. 'mgr' and
		//   'mgr-init' should be sufficient, and then it's easier to get specific mgr container
		//   logs from any mgr pod since the containers inside will have deterministic names
		Name:  "mgr",
		Image: k8sutil.MakeRookImage(c.Version),
		VolumeMounts: []v1.VolumeMount{
			{Name: k8sutil.DataDirVolume, MountPath: k8sutil.DataDir},
			k8sutil.ConfigOverrideMount(),
			{Name: "ceph-default-config-dir", MountPath: ceph.DefaultConfigDir}
		},
		// TODO: Should all/some of the below be kept so the env vars can give useful info to
		//       admins poking through running containers?
		// Env: []v1.EnvVar{
		// 	// {Name: "ROOK_MGR_NAME", Value: daemonName},
		// 	// {Name: "ROOK_MGR_KEYRING", ValueFrom: &v1.EnvVarSource{SecretKeyRef: &v1.SecretKeySelector{LocalObjectReference: v1.LocalObjectReference{Name: name}, Key: keyringName}}},
		// 	k8sutil.PodIPEnvVar(k8sutil.PrivateIPEnvVar),
		// 	k8sutil.PodIPEnvVar(k8sutil.PublicIPEnvVar),
		// 	// opmon.ClusterNameEnvVar(c.Namespace),
		// 	// opmon.EndpointEnvVar(),
		// 	// opmon.SecretEnvVar(),
		// 	// opmon.AdminSecretEnvVar(),
		// 	// k8sutil.ConfigOverrideEnvVar(), // <-- What does this do? It seems unused
		// },
		Resources: c.resources,
		Ports: []v1.ContainerPort{
			{
				Name:          "mgr",
				ContainerPort: int32(6800),
				Protocol:      v1.ProtocolTCP,
			},
			{
				Name:          "http-metrics",
				ContainerPort: int32(metricsPort),
				Protocol:      v1.ProtocolTCP,
			},
			{
				Name:          "dashboard",
				ContainerPort: int32(dashboardPort),
				Protocol:      v1.ProtocolTCP,
			},
		},
	}
}

func (c *Cluster) mgrInitContainer(name, daemonName, keyringPath, confDir string) v1.Container {
	return v1.Container{
		Args: []string{
			"ceph",
			"mgr",
			fmt.Sprintf("--config-dir=%s", k8sutil.DataDir),
		},
		Name:  "mgr-init",
		Image: k8sutil.MakeRookImage(c.Version),
		VolumeMounts: []v1.VolumeMount{
			{Name: k8sutil.DataDirVolume, MountPath: k8sutil.DataDir},
			k8sutil.ConfigOverrideMount(),
			// Also mount Ceph's default config dir (/etc/ceph) so that when the Rook binary
			// initializes the configuration and keyring and copies it to /etc/ceph, the data will
			// be persisted to the running container as well.
			// Is this going to overwrite any critical files installed by default into /etc/ceph?
			{Name: "ceph-default-config-dir", MountPath: ceph.DefaultConfigDir}
		},
		Env: []v1.EnvVar{
			{Name: "ROOK_MGR_NAME", Value: daemonName},
			{Name: "ROOK_MGR_KEYRING", ValueFrom: &v1.EnvVarSource{SecretKeyRef: &v1.SecretKeySelector{LocalObjectReference: v1.LocalObjectReference{Name: name}, Key: keyringName}}},
			{Name: "ROOK_MGR_KEYRING_PATH", Value: keyringPath},
			{Name: "ROOK_MGR_CONF_DIR", Value: confDir},
			k8sutil.PodIPEnvVar(k8sutil.PrivateIPEnvVar),
			k8sutil.PodIPEnvVar(k8sutil.PublicIPEnvVar),
			opmon.ClusterNameEnvVar(c.Namespace),
			opmon.EndpointEnvVar(),
			opmon.SecretEnvVar(),
			opmon.AdminSecretEnvVar(),
			k8sutil.ConfigOverrideEnvVar(),
		},
		Resources: c.resources,
	}
}

func (c *Cluster) getLabels() map[string]string {
	return map[string]string{
		k8sutil.AppAttr:     appName,
		k8sutil.ClusterAttr: c.Namespace,
	}
}

func (c *Cluster) getDaemonLabels(daemonName string) map[string]string {
	labels := c.getLabels()
	labels["instance"] = daemonName
	return labels
}

func (c *Cluster) createKeyring(clusterName, name, daemonName string) error {
	_, err := c.context.Clientset.CoreV1().Secrets(c.Namespace).Get(name, metav1.GetOptions{})
	if err == nil {
		logger.Infof("the mgr keyring was already generated")
		return nil
	}
	if !errors.IsNotFound(err) {
		return fmt.Errorf("failed to get mgr secrets. %+v", err)
	}

	// get-or-create-key for the user account
	keyring, err := createKeyring(c.context, clusterName, daemonName)
	if err != nil {
		return fmt.Errorf("failed to create mgr keyring. %+v", err)
	}

	// Store the keyring in a secret
	secrets := map[string]string{
		keyringName: keyring,
	}
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: c.Namespace,
		},
		StringData: secrets,
		Type:       k8sutil.RookType,
	}
	k8sutil.SetOwnerRef(c.context.Clientset, c.Namespace, &secret.ObjectMeta, &c.ownerRef)

	_, err = c.context.Clientset.CoreV1().Secrets(c.Namespace).Create(secret)
	if err != nil {
		return fmt.Errorf("failed to save mgr secrets. %+v", err)
	}

	return nil
}

// Ceph docs about the prometheus module: http://docs.ceph.com/docs/master/mgr/prometheus/
func (c *Cluster) enablePrometheusModule(clusterName string) error {
	if err := client.MgrEnableModule(c.context, clusterName, prometheusModuleName, true); err != nil {
		return fmt.Errorf("failed to enable mgr prometheus module. %+v", err)
	}
	return nil
}

// Ceph docs about the dashboard module: http://docs.ceph.com/docs/luminous/mgr/dashboard/
func (c *Cluster) configureDashboardModule(clusterName string, enable bool) error {
	if enable {
		if err := client.MgrEnableModule(c.context, clusterName, dashboardModuleName, true); err != nil {
			return fmt.Errorf("failed to enable mgr dashboard module. %+v", err)
		}
	} else {
		if err := client.MgrDisableModule(c.context, clusterName, dashboardModuleName); err != nil {
			return fmt.Errorf("failed to disable mgr dashboard module. %+v", err)
		}
	}
	return nil
}

func getKeyringProperties(name string) (string, []string) {
	username := fmt.Sprintf("mgr.%s", name)
	access := []string{"mon", "allow *"}
	return username, access
}

// create a keyring for the mds client with a limited set of privileges
func createKeyring(context *clusterd.Context, clusterName, name string) (string, error) {
	// get-or-create-key for the user account
	username, access := getKeyringProperties(name)
	keyring, err := client.AuthGetOrCreateKey(context, clusterName, username, access)
	if err != nil {
		return "", fmt.Errorf("failed to get or create auth key for %s. %+v", username, err)
	}

	return keyring, nil
}

// get the manager config directory for a manager daemon
func getMgrConfDir(rookConfigDir, mgrDaemonName string) string {
	return path.Join(rookConfigDir, fmt.Sprintf("mgr-%s", mgrDaemonName))
}

// get the full path of the manager config file for a manager daemon
func getMgrConfFilePath(rookConfigDir, mgrDaemonName, clusterName string) string {
	return path.Join(getMgrConfDir(rookConfigDir, mgrDaemonName), fmt.Sprintf("%s.config", clusterName))
}

// get the full path of the manager keyring file for a manager daemon
func getMgrKeyringPath(rookConfigDir, mgrDaemonName string) string {
	return path.Join(getMgrConfDir(rookConfigDir, mgrDaemonName), "keyring")
}
