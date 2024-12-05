/*
Copyright 2018 The Rook Authors. All rights reserved.

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

// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	"context"

	v1 "github.com/rook/rook/pkg/apis/ceph.rook.io/v1"
	scheme "github.com/rook/rook/pkg/client/clientset/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	gentype "k8s.io/client-go/gentype"
)

// CephFilesystemSubVolumeGroupsGetter has a method to return a CephFilesystemSubVolumeGroupInterface.
// A group's client should implement this interface.
type CephFilesystemSubVolumeGroupsGetter interface {
	CephFilesystemSubVolumeGroups(namespace string) CephFilesystemSubVolumeGroupInterface
}

// CephFilesystemSubVolumeGroupInterface has methods to work with CephFilesystemSubVolumeGroup resources.
type CephFilesystemSubVolumeGroupInterface interface {
	Create(ctx context.Context, cephFilesystemSubVolumeGroup *v1.CephFilesystemSubVolumeGroup, opts metav1.CreateOptions) (*v1.CephFilesystemSubVolumeGroup, error)
	Update(ctx context.Context, cephFilesystemSubVolumeGroup *v1.CephFilesystemSubVolumeGroup, opts metav1.UpdateOptions) (*v1.CephFilesystemSubVolumeGroup, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.CephFilesystemSubVolumeGroup, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.CephFilesystemSubVolumeGroupList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.CephFilesystemSubVolumeGroup, err error)
	CephFilesystemSubVolumeGroupExpansion
}

// cephFilesystemSubVolumeGroups implements CephFilesystemSubVolumeGroupInterface
type cephFilesystemSubVolumeGroups struct {
	*gentype.ClientWithList[*v1.CephFilesystemSubVolumeGroup, *v1.CephFilesystemSubVolumeGroupList]
}

// newCephFilesystemSubVolumeGroups returns a CephFilesystemSubVolumeGroups
func newCephFilesystemSubVolumeGroups(c *CephV1Client, namespace string) *cephFilesystemSubVolumeGroups {
	return &cephFilesystemSubVolumeGroups{
		gentype.NewClientWithList[*v1.CephFilesystemSubVolumeGroup, *v1.CephFilesystemSubVolumeGroupList](
			"cephfilesystemsubvolumegroups",
			c.RESTClient(),
			scheme.ParameterCodec,
			namespace,
			func() *v1.CephFilesystemSubVolumeGroup { return &v1.CephFilesystemSubVolumeGroup{} },
			func() *v1.CephFilesystemSubVolumeGroupList { return &v1.CephFilesystemSubVolumeGroupList{} }),
	}
}
