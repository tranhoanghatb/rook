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
package mgr

import (
	"fmt"

	"github.com/coreos/pkg/capnslog"
	"github.com/rook/rook/pkg/clusterd"
	"github.com/rook/rook/pkg/daemon/ceph/mon"
)

var (
	logger          = capnslog.NewPackageLogger("github.com/rook/rook", "cephmgr")
	keyringTemplate = `
[mgr.%s]
	key = %s
	caps mon = "allow profile mgr"
	caps mds = "allow *"
	caps osd = "allow *"
`
)

// const (
// 	cephmgr = "ceph-mgr"
// )

type Config struct {
	ClusterInfo *mon.ClusterInfo
	Name        string // name of this mgr
	Keyring     string // this mgr's keyring
	KeyringPath string // path where the keyring is written
	ConfDir     string // dir where this mgr's config is stored
}

// Rename this to 'Prepare' or something similar? Or is 'Run' a go-ism?
func Run(context *clusterd.Context, config *Config) error {
	//logger.Infof("Starting MGR %s with keyring %s", config.Name, config.Keyring)
	logger.Infof("Preparing MGR %s with keyring %s", config.Name, config.Keyring)
	// init container
	if err := generateConfigFiles(context, config); err != nil {
		return fmt.Errorf("failed to generate mgr config files. %+v", err)
	}

	// if err := startMgr(context, config); err != nil {
	// 	return fmt.Errorf("failed to run mgr. %+v", err)
	// }

	logger.Infof("MGR preparation complete")
	return nil
}

// Do in init container
func generateConfigFiles(context *clusterd.Context, config *Config) error {

	keyringPath := config.KeyringPath // getMgrKeyringPath(context.ConfigDir, config.Name)
	confDir := config.ConfDir         // getMgrConfDir(context.ConfigDir, config.Name)
	username := fmt.Sprintf("mgr.%s", config.Name)
	settings := map[string]string{
		"mgr data": confDir,
	}
	logger.Infof("Conf files: dir=%s keyring=%s", confDir, keyringPath)
	_, err := mon.GenerateConfigFile(context, config.ClusterInfo, confDir,
		username, keyringPath, nil, settings)
	if err != nil {
		return fmt.Errorf("failed to create config file. %+v", err)
	}

	keyringEval := func(key string) string {
		return fmt.Sprintf(keyringTemplate, config.Name, key)
	}

	err = mon.WriteKeyring(keyringPath, config.Keyring, keyringEval)
	if err != nil {
		return fmt.Errorf("failed to create mds keyring. %+v", err)
	}

	return nil
}

// // This should be what is put into the running container
// // ceph-mgr --foreground --cluster=<config.ClusterInfo.Name> --conf=<configFile>
// //          --keyring=<keyringPath> --id <config.Name>
// func startMgr(context *clusterd.Context, config *Config) error {

// 	// start the mgr daemon in the foreground with the given config
// 	logger.Infof("starting ceph-mgr")

// 	if err := context.Executor.ExecuteCommand(false, cephmgr, cephmgr, args...); err != nil {
// 		return fmt.Errorf("failed to start mgr: %+v", err)
// 	}
// 	return nil
// }

//
// We don't want to duplicate the below code items from the operator; pass these in as vars somehow
//

// func getMgrConfDir(dir, name string) string {
// 	return path.Join(dir, fmt.Sprintf("mgr-%s", name))
// }

// func getMgrConfFilePath(dir, name, clusterName string) string {
// 	return path.Join(getMgrConfDir(dir, name), fmt.Sprintf("%s.config", clusterName))
// }

// func getMgrKeyringPath(dir, name string) string {
// 	return path.Join(getMgrConfDir(dir, name), "keyring")
// }
