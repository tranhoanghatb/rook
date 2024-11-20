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

// Code generated by lister-gen. DO NOT EDIT.

package v1

import (
	v1 "github.com/rook/rook/pkg/apis/ceph.rook.io/v1"
<<<<<<< HEAD
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/listers"
=======
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
>>>>>>> fc08e87d4 (Revert "object: create cosi user for each object store")
	"k8s.io/client-go/tools/cache"
)

// CephNFSLister helps list CephNFSes.
// All objects returned here must be treated as read-only.
type CephNFSLister interface {
	// List lists all CephNFSes in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.CephNFS, err error)
	// CephNFSes returns an object that can list and get CephNFSes.
	CephNFSes(namespace string) CephNFSNamespaceLister
	CephNFSListerExpansion
}

// cephNFSLister implements the CephNFSLister interface.
type cephNFSLister struct {
<<<<<<< HEAD
	listers.ResourceIndexer[*v1.CephNFS]
=======
	indexer cache.Indexer
>>>>>>> fc08e87d4 (Revert "object: create cosi user for each object store")
}

// NewCephNFSLister returns a new CephNFSLister.
func NewCephNFSLister(indexer cache.Indexer) CephNFSLister {
<<<<<<< HEAD
	return &cephNFSLister{listers.New[*v1.CephNFS](indexer, v1.Resource("cephnfs"))}
=======
	return &cephNFSLister{indexer: indexer}
}

// List lists all CephNFSes in the indexer.
func (s *cephNFSLister) List(selector labels.Selector) (ret []*v1.CephNFS, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.CephNFS))
	})
	return ret, err
>>>>>>> fc08e87d4 (Revert "object: create cosi user for each object store")
}

// CephNFSes returns an object that can list and get CephNFSes.
func (s *cephNFSLister) CephNFSes(namespace string) CephNFSNamespaceLister {
<<<<<<< HEAD
	return cephNFSNamespaceLister{listers.NewNamespaced[*v1.CephNFS](s.ResourceIndexer, namespace)}
=======
	return cephNFSNamespaceLister{indexer: s.indexer, namespace: namespace}
>>>>>>> fc08e87d4 (Revert "object: create cosi user for each object store")
}

// CephNFSNamespaceLister helps list and get CephNFSes.
// All objects returned here must be treated as read-only.
type CephNFSNamespaceLister interface {
	// List lists all CephNFSes in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.CephNFS, err error)
	// Get retrieves the CephNFS from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1.CephNFS, error)
	CephNFSNamespaceListerExpansion
}

// cephNFSNamespaceLister implements the CephNFSNamespaceLister
// interface.
type cephNFSNamespaceLister struct {
<<<<<<< HEAD
	listers.ResourceIndexer[*v1.CephNFS]
=======
	indexer   cache.Indexer
	namespace string
}

// List lists all CephNFSes in the indexer for a given namespace.
func (s cephNFSNamespaceLister) List(selector labels.Selector) (ret []*v1.CephNFS, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.CephNFS))
	})
	return ret, err
}

// Get retrieves the CephNFS from the indexer for a given namespace and name.
func (s cephNFSNamespaceLister) Get(name string) (*v1.CephNFS, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("cephnfs"), name)
	}
	return obj.(*v1.CephNFS), nil
>>>>>>> fc08e87d4 (Revert "object: create cosi user for each object store")
}
