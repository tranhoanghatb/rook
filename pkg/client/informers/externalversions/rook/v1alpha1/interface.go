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

// This file was automatically generated by informer-gen

package v1alpha1

import (
	internalinterfaces "github.com/rook/rook/pkg/client/informers/externalversions/internalinterfaces"
)

// Interface provides access to all the informers in this group version.
type Interface interface {
	// Clusters returns a ClusterInformer.
	Clusters() ClusterInformer
	// Filesystems returns a FilesystemInformer.
	Filesystems() FilesystemInformer
	// ObjectStores returns a ObjectStoreInformer.
	ObjectStores() ObjectStoreInformer
	// Pools returns a PoolInformer.
	Pools() PoolInformer
	// VolumeAttachments returns a VolumeAttachmentInformer.
	VolumeAttachments() VolumeAttachmentInformer
}

type version struct {
	internalinterfaces.SharedInformerFactory
}

// New returns a new Interface.
func New(f internalinterfaces.SharedInformerFactory) Interface {
	return &version{f}
}

// Clusters returns a ClusterInformer.
func (v *version) Clusters() ClusterInformer {
	return &clusterInformer{factory: v.SharedInformerFactory}
}

// Filesystems returns a FilesystemInformer.
func (v *version) Filesystems() FilesystemInformer {
	return &filesystemInformer{factory: v.SharedInformerFactory}
}

// ObjectStores returns a ObjectStoreInformer.
func (v *version) ObjectStores() ObjectStoreInformer {
	return &objectStoreInformer{factory: v.SharedInformerFactory}
}

// Pools returns a PoolInformer.
func (v *version) Pools() PoolInformer {
	return &poolInformer{factory: v.SharedInformerFactory}
}

// VolumeAttachments returns a VolumeAttachmentInformer.
func (v *version) VolumeAttachments() VolumeAttachmentInformer {
	return &volumeAttachmentInformer{factory: v.SharedInformerFactory}
}
