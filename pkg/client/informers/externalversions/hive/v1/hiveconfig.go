// Code generated by informer-gen. DO NOT EDIT.

package v1

import (
	"context"
	time "time"

	hivev1 "github.com/openshift/hive/apis/hive/v1"
	versioned "github.com/openshift/hive/pkg/client/clientset/versioned"
	internalinterfaces "github.com/openshift/hive/pkg/client/informers/externalversions/internalinterfaces"
	v1 "github.com/openshift/hive/pkg/client/listers/hive/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// HiveConfigInformer provides access to a shared informer and lister for
// HiveConfigs.
type HiveConfigInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1.HiveConfigLister
}

type hiveConfigInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewHiveConfigInformer constructs a new informer for HiveConfig type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewHiveConfigInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredHiveConfigInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredHiveConfigInformer constructs a new informer for HiveConfig type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredHiveConfigInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.HiveV1().HiveConfigs().List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.HiveV1().HiveConfigs().Watch(context.TODO(), options)
			},
		},
		&hivev1.HiveConfig{},
		resyncPeriod,
		indexers,
	)
}

func (f *hiveConfigInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredHiveConfigInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *hiveConfigInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&hivev1.HiveConfig{}, f.defaultInformer)
}

func (f *hiveConfigInformer) Lister() v1.HiveConfigLister {
	return v1.NewHiveConfigLister(f.Informer().GetIndexer())
}
