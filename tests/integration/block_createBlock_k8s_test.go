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

package integration

import (
	"fmt"
	"testing"

	"github.com/rook/rook/tests/framework/clients"
	"github.com/rook/rook/tests/framework/contracts"
	"github.com/rook/rook/tests/framework/installer"
	"github.com/rook/rook/tests/framework/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"k8s.io/api/core/v1"
)

// Test K8s Block Image Creation Scenarios. These tests work when platform is set to Kubernetes

func TestK8sBlockImageCreateSuite(t *testing.T) {
	s := new(K8sBlockImageCreateSuite)
	defer func(s *K8sBlockImageCreateSuite) {
		HandlePanics(recover(), s.op, s.T)
	}(s)
	suite.Run(t, s)
}

type K8sBlockImageCreateSuite struct {
	suite.Suite
	testClient     *clients.TestClient
	bc             contracts.BlockOperator
	kh             *utils.K8sHelper
	initBlockCount int
	namespace      string
	installer      *installer.InstallHelper
	op             contracts.Setup
}

func (s *K8sBlockImageCreateSuite) SetupSuite() {

	var err error
	s.namespace = "block-k8s-ns"
	s.op, s.kh = NewBaseTestOperations(s.T, s.namespace, "bluestore", "", false, false, 1)
	s.testClient = GetTestClient(s.kh, s.namespace, s.op, s.T)
	s.bc = s.testClient.GetBlockClient()
	initialBlocks, err := s.bc.BlockList()
	assert.Nil(s.T(), err)
	s.initBlockCount = len(initialBlocks)
}

// Test case when persistentvolumeclaim is created for a storage class that doesn't exist
func (s *K8sBlockImageCreateSuite) TestCreatePVCWhenNoStorageClassExists() {
	logger.Infof("Test creating PVC(block images) when storage class is not created")

	//Create PVC
	claimName := "test-no-storage-class-claim"
	poolName := "test-no-storage-class-pool"
	storageClassName := "rook-block"
	defer s.tearDownTest(claimName, poolName, storageClassName, "ReadWriteOnce")

	result, err := installer.BlockResourceOperation(s.kh, installer.GetBlockPvcDef(claimName, storageClassName, "ReadWriteOnce"), "create")
	require.Contains(s.T(), result, fmt.Sprintf("persistentvolumeclaim \"%s\" created", claimName), "Make sure pvc is created. "+result)
	require.NoError(s.T(), err)

	//check status of PVC
	pvcStatus, err := s.kh.GetPVCStatus(defaultNamespace, claimName)
	require.Nil(s.T(), err)
	require.Contains(s.T(), pvcStatus, "Pending", "Makes sure PVC is in Pending state")

	//check block image count
	b, _ := s.bc.BlockList()
	require.Equal(s.T(), s.initBlockCount, len(b), "Make sure new block image is not created")

}

// Test case when persistentvolumeclaim  with ReadWriteOnce access is created for a valid storage class
func (s *K8sBlockImageCreateSuite) TestCreatePReadWriteOnceVCWhenStorageClassExists() {
	logger.Infof("Test creating PVC(block images) when storage class is created")
	claimName := "test-with-storage-class-claim"
	poolName := "test-with-storage-class-pool"
	storageClassName := "rook-block"
	defer s.tearDownTest(claimName, poolName, storageClassName, "ReadWriteOnce")

	//create pool and storageclass
	result0, err0 := installer.BlockResourceOperation(s.kh, installer.GetBlockPoolDef(poolName, s.namespace, "1"), "create")
	require.Contains(s.T(), result0, fmt.Sprintf("pool \"%s\" created", poolName), "Make sure test pool is created")
	require.NoError(s.T(), err0)
	result1, err1 := installer.BlockResourceOperation(s.kh, installer.GetBlockStorageClassDef(poolName, storageClassName, s.namespace), "create")
	require.Contains(s.T(), result1, fmt.Sprintf("storageclass \"%s\" created", storageClassName), "Make sure storageclass is created")
	require.NoError(s.T(), err1)

	//make sure storageclass is created
	present, err := s.kh.IsStorageClassPresent(storageClassName)
	require.Nil(s.T(), err)
	require.True(s.T(), present, "Make sure storageclass is present")

	//create pvc
	result2, err2 := installer.BlockResourceOperation(s.kh, installer.GetBlockPvcDef(claimName, storageClassName, "ReadWriteOnce"), "create")
	require.Contains(s.T(), result2, fmt.Sprintf("persistentvolumeclaim \"%s\" created", claimName), "Make sure pvc is created. "+result2)
	require.NoError(s.T(), err2)

	//check status of PVC
	require.True(s.T(), s.kh.WaitUntilPVCIsBound(defaultNamespace, claimName))
	accessModes, err := s.kh.GetPVCAccessModes(defaultNamespace, claimName)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), accessModes[0], v1.PersistentVolumeAccessMode(v1.ReadWriteOnce))

	//check block image count
	b, _ := s.bc.BlockList()
	require.Equal(s.T(), s.initBlockCount+1, len(b), "Make sure new block image is created")

}

// Test case when persistentvolumeclaim is created for a valid storage class twice
func (s *K8sBlockImageCreateSuite) TestCreateSamePVCTwice() {
	logger.Infof("Test creating PVC(create block images) twice")
	claimName := "test-twice-claim"
	poolName := "test-twice-pool"
	storageClassName := "rook-block"
	defer s.tearDownTest(claimName, poolName, storageClassName, "ReadWriteOnce")
	status, _ := s.kh.GetPVCStatus(defaultNamespace, claimName)
	logger.Infof("PVC %s status: %s", claimName, status)
	s.bc.BlockList()

	logger.Infof("create pool and storageclass")
	result0, err0 := installer.BlockResourceOperation(s.kh, installer.GetBlockPoolDef(poolName, s.namespace, "1"), "create")
	require.Contains(s.T(), result0, fmt.Sprintf("pool \"%s\" created", poolName), "Make sure test pool is created")
	require.NoError(s.T(), err0)
	result1, err1 := installer.BlockResourceOperation(s.kh, installer.GetBlockStorageClassDef(poolName, storageClassName, s.namespace), "create")
	require.Contains(s.T(), result1, fmt.Sprintf("storageclass \"%s\" created", storageClassName), "Make sure storageclass is created")
	require.NoError(s.T(), err1)

	logger.Infof("make sure storageclass is created")
	present, err := s.kh.IsStorageClassPresent("rook-block")
	require.Nil(s.T(), err)
	require.True(s.T(), present, "Make sure storageclass is present")

	logger.Infof("create pvc")
	result2, err2 := installer.BlockResourceOperation(s.kh, installer.GetBlockPvcDef(claimName, storageClassName, "ReadWriteOnce"), "create")
	require.Contains(s.T(), result2, fmt.Sprintf("persistentvolumeclaim \"%s\" created", claimName), "Make sure pvc is created. "+result2)
	require.NoError(s.T(), err2)

	logger.Infof("check status of PVC")
	require.True(s.T(), s.kh.WaitUntilPVCIsBound(defaultNamespace, claimName))

	b1, err := s.bc.BlockList()
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), s.initBlockCount+1, len(b1), "Make sure new block image is created")

	logger.Infof("Create same pvc again")
	result3, err3 := installer.BlockResourceOperation(s.kh, installer.GetBlockPvcDef(claimName, storageClassName, "ReadWriteOnce"), "create")
	require.Contains(s.T(), result3, fmt.Sprintf("persistentvolumeclaims \"%s\" already exists", claimName), "make sure PVC is not created again. "+result3)
	require.NoError(s.T(), err3)

	logger.Infof("check status of PVC")
	require.True(s.T(), s.kh.WaitUntilPVCIsBound(defaultNamespace, claimName))

	logger.Infof("check block image count")
	b2, _ := s.bc.BlockList()
	assert.Equal(s.T(), len(b1), len(b2), "Make sure new block image is created")

}

// Test case when persistentvolumeclaim with ReadOnlyMany
func (s *K8sBlockImageCreateSuite) TestCreateReadOnlyManyPVCWhenStorageClassExists() {
	logger.Infof("Test creating PVC(block images) when storage class is created")
	claimName := "test-with-storage-class-claim-rox"
	poolName := "test-with-storage-class-pool-rox"
	storageClassName := "rook-block"
	defer s.tearDownTest(claimName, poolName, storageClassName, "ReadOnlyMany")

	//create pool and storageclass
	result0, err0 := installer.BlockResourceOperation(s.kh, installer.GetBlockPoolDef(poolName, s.namespace, "1"), "create")
	require.Contains(s.T(), result0, fmt.Sprintf("pool \"%s\" created", poolName), "Make sure test pool is created")
	require.NoError(s.T(), err0)
	result1, err1 := installer.BlockResourceOperation(s.kh, installer.GetBlockStorageClassDef(poolName, storageClassName, s.namespace), "create")
	require.Contains(s.T(), result1, fmt.Sprintf("storageclass \"%s\" created", storageClassName), "Make sure storageclass is created")
	require.NoError(s.T(), err1)

	//make sure storageclass is created
	present, err := s.kh.IsStorageClassPresent(storageClassName)
	require.Nil(s.T(), err)
	require.True(s.T(), present, "Make sure storageclass is present")

	//create pvc
	result2, err2 := installer.BlockResourceOperation(s.kh, installer.GetBlockPvcDef(claimName, storageClassName, "ReadOnlyMany"), "create")
	require.Contains(s.T(), result2, fmt.Sprintf("persistentvolumeclaim \"%s\" created", claimName), "Make sure pvc is created. "+result2)
	require.NoError(s.T(), err2)

	//check status of PVC
	require.True(s.T(), s.kh.WaitUntilPVCIsBound(defaultNamespace, claimName))
	accessModes, err := s.kh.GetPVCAccessModes(defaultNamespace, claimName)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), accessModes[0], v1.PersistentVolumeAccessMode(v1.ReadOnlyMany))

	//check block image count
	b, _ := s.bc.BlockList()
	require.Equal(s.T(), s.initBlockCount+1, len(b), "Make sure new block image is created")

}

// Test case when persistentvolumeclaim with ReadWriteMany
func (s *K8sBlockImageCreateSuite) TestCreateReadWriteManyPVCWhenStorageClassExists() {
	logger.Infof("Test creating PVC(block images) when storage class is created")
	claimName := "test-with-storage-class-claim-rwx"
	poolName := "test-with-storage-class-pool-rwx"
	storageClassName := "rook-block"
	defer s.tearDownTest(claimName, poolName, storageClassName, "ReadWriteMany")

	//create pool and storageclass
	result0, err0 := installer.BlockResourceOperation(s.kh, installer.GetBlockPoolDef(poolName, s.namespace, "1"), "create")
	require.Contains(s.T(), result0, fmt.Sprintf("pool \"%s\" created", poolName), "Make sure test pool is created")
	require.NoError(s.T(), err0)
	result1, err1 := installer.BlockResourceOperation(s.kh, installer.GetBlockStorageClassDef(poolName, storageClassName, s.namespace), "create")
	require.Contains(s.T(), result1, fmt.Sprintf("storageclass \"%s\" created", storageClassName), "Make sure storageclass is created")
	require.NoError(s.T(), err1)

	//make sure storageclass is created
	present, err := s.kh.IsStorageClassPresent(storageClassName)
	require.Nil(s.T(), err)
	require.True(s.T(), present, "Make sure storageclass is present")

	//create pvc
	result2, err2 := installer.BlockResourceOperation(s.kh, installer.GetBlockPvcDef(claimName, storageClassName, "ReadWriteMany"), "create")
	require.Contains(s.T(), result2, fmt.Sprintf("persistentvolumeclaim \"%s\" created", claimName), "Make sure pvc is created. "+result2)
	require.NoError(s.T(), err2)

	//check status of PVC
	require.True(s.T(), s.kh.WaitUntilPVCIsBound(defaultNamespace, claimName))
	accessModes, err := s.kh.GetPVCAccessModes(defaultNamespace, claimName)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), accessModes[0], v1.PersistentVolumeAccessMode(v1.ReadWriteMany))

	//check block image count
	b, _ := s.bc.BlockList()
	require.Equal(s.T(), s.initBlockCount+1, len(b), "Make sure new block image is created")

}

func (s *K8sBlockImageCreateSuite) TestBlockStorageMountUnMountForStatefulSets() {
	poolName := "stspool"
	storageClassName := "stssc"
	statefulSetName := "block-stateful-set"
	statefulPodsName := "ststest"

	defer s.statefulSetDataCleanup(poolName, storageClassName, statefulSetName, statefulPodsName)
	logger.Infof("Test case when block persistent volumes are scaled up and down along with StatefulSet")
	logger.Info("Step 1: Create pool and storageClass")
	_, cbErr := installer.BlockResourceOperation(s.kh, installer.GetBlockPoolStorageClass(s.namespace, poolName, storageClassName), "create")
	assert.Nil(s.T(), cbErr)

	logger.Info("Step 2 : Deploy statefulSet with 1X replication")
	_, sderr := installer.BlockResourceOperation(s.kh, getBlockStatefulSetDefintion(statefulSetName, statefulPodsName, storageClassName), "create")
	assert.Nil(s.T(), sderr)
	assert.True(s.T(), s.kh.CheckPodCountAndState(statefulSetName, defaultNamespace, 1, "Running"))
	assert.True(s.T(), s.kh.CheckPvcCountAndStatus(statefulSetName, defaultNamespace, 1, "Bound"))

	logger.Info("Step 3 : Scale up replication on statefulSet")
	s.kh.ScaleStatefulSet(statefulPodsName, 2)
	assert.True(s.T(), s.kh.CheckPodCountAndState(statefulSetName, defaultNamespace, 2, "Running"))
	assert.True(s.T(), s.kh.CheckPvcCountAndStatus(statefulSetName, defaultNamespace, 2, "Bound"))

	logger.Info("Step 4 : Scale down replication on statefulSet")
	s.kh.ScaleStatefulSet(statefulPodsName, 1)
	assert.True(s.T(), s.kh.CheckPodCountAndState(statefulSetName, defaultNamespace, 1, "Running"))
	assert.True(s.T(), s.kh.CheckPvcCountAndStatus(statefulSetName, defaultNamespace, 2, "Bound"))

	logger.Info("Step 5 : Delete statefulSet")
	_, sddelerr := installer.BlockResourceOperation(s.kh, getBlockStatefulSetDefintion(statefulSetName, statefulPodsName, storageClassName), "delete")
	assert.Nil(s.T(), sddelerr)
	assert.True(s.T(), s.kh.WaitUntilPodWithLabelDeleted(fmt.Sprintf("app=%s", statefulSetName), defaultNamespace))
	assert.True(s.T(), s.kh.CheckPvcCountAndStatus(statefulSetName, defaultNamespace, 2, "Bound"))
}

func (s *K8sBlockImageCreateSuite) statefulSetDataCleanup(poolName, storageClassName, statefulSetName, statefulPodsName string) {

	//Delete stateful set
	installer.BlockResourceOperation(s.kh, getBlockStatefulSetDefintion(statefulSetName, statefulPodsName, storageClassName), "delete")
	//Delete all PVCs
	s.kh.DeletePvcWithLabel(defaultNamespace, statefulSetName)
	//Delete storageclass and pool
	installer.BlockResourceOperation(s.kh, installer.GetBlockPoolStorageClass(s.namespace, poolName, storageClassName), "delete")

}

func (s *K8sBlockImageCreateSuite) tearDownTest(claimName string, poolName string, storageClassName string, accessMode string) {
	installer.BlockResourceOperation(s.kh, installer.GetBlockPvcDef(claimName, storageClassName, accessMode), "delete")
	installer.BlockResourceOperation(s.kh, installer.GetBlockPoolDef(poolName, s.namespace, "1"), "delete")
	installer.BlockResourceOperation(s.kh, installer.GetBlockStorageClassDef(poolName, storageClassName, s.namespace), "delete")

}

func (s *K8sBlockImageCreateSuite) TearDownSuite() {
	s.op.TearDown()
}
