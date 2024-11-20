/*
<<<<<<< HEAD
Copyright 2018 The Rook Authors. All rights reserved.
=======
Copyright The Kubernetes Authors.
>>>>>>> fc08e87d4 (Revert "object: create cosi user for each object store")

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

<<<<<<< HEAD
    http://www.apache.org/licenses/LICENSE-2.0
=======
    http://www.apache.org/licenses/LICENSE-2.0
>>>>>>> fc08e87d4 (Revert "object: create cosi user for each object store")

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by informer-gen. DO NOT EDIT.

package v1

import (
	"context"
	time "time"

	cephrookiov1 "github.com/rook/rook/pkg/apis/ceph.rook.io/v1"
	versioned "github.com/rook/rook/pkg/client/clientset/versioned"
	internalinterfaces "github.com/rook/rook/pkg/client/informers/externalversions/internalinterfaces"
	v1 "github.com/rook/rook/pkg/client/listers/ceph.rook.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// CephObjectRealmInformer provides access to a shared informer and lister for
// CephObjectRealms.
type CephObjectRealmInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1.CephObjectRealmLister
}

type cephObjectRealmInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewCephObjectRealmInformer constructs a new informer for CephObjectRealm type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewCephObjectRealmInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredCephObjectRealmInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredCephObjectRealmInformer constructs a new informer for CephObjectRealm type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredCephObjectRealmInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.CephV1().CephObjectRealms(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.CephV1().CephObjectRealms(namespace).Watch(context.TODO(), options)
			},
		},
		&cephrookiov1.CephObjectRealm{},
		resyncPeriod,
		indexers,
	)
}

func (f *cephObjectRealmInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredCephObjectRealmInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *cephObjectRealmInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&cephrookiov1.CephObjectRealm{}, f.defaultInformer)
}

func (f *cephObjectRealmInformer) Lister() v1.CephObjectRealmLister {
	return v1.NewCephObjectRealmLister(f.Informer().GetIndexer())
}
