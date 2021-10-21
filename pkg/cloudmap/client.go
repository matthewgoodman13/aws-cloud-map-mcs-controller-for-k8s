package cloudmap

import (
	"context"
	"fmt"
	"github.com/aws/aws-cloud-map-mcs-controller-for-k8s/pkg/model"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/util/cache"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
)

const (
	defaultNamespaceIdCacheTTL  = 2 * time.Minute
	defaultNamespaceIdCacheSize = 100
	defaultServiceIdCacheTTL    = 2 * time.Minute
	defaultServiceIdCacheSize   = 1024
	defaultEndpointsCacheTTL    = 5 * time.Second
	defaultEndpointsCacheSize   = 1024
)

// ServiceDiscoveryClient provides the service endpoint management functionality required by the AWS Cloud Map
// multi-cluster service discovery for Kubernetes controller. It maintains local caches for all AWS Cloud Map resources.
type ServiceDiscoveryClient interface {
	// ListServices returns all services and their endpoints for a given namespace.
	ListServices(ctx context.Context, namespaceName string) ([]*model.Service, error)

	// CreateService creates a Cloud Map service resource, and namespace if necessary.
	CreateService(ctx context.Context, namespaceName string, serviceName string) error

	// GetService returns a service resource fetched from AWS Cloud Map or nil if not found.
	GetService(ctx context.Context, namespaceName string, serviceName string) (*model.Service, error)

	// RegisterEndpoints registers all endpoints for given service.
	RegisterEndpoints(ctx context.Context, namespaceName string, serviceName string, endpoints []*model.Endpoint) error

	// DeleteEndpoints de-registers all endpoints for given service.
	DeleteEndpoints(ctx context.Context, namespaceName string, serviceName string, endpoints []*model.Endpoint) error
}

type serviceDiscoveryClient struct {
	log              logr.Logger
	sdApi            ServiceDiscoveryApi
	namespaceIdCache *cache.LRUExpireCache
	serviceIdCache   *cache.LRUExpireCache
	endpointCache    *cache.LRUExpireCache
}

// NewServiceDiscoveryClient creates a new service discovery client for AWS Cloud Map from a given AWS client config.
func NewServiceDiscoveryClient(cfg *aws.Config) ServiceDiscoveryClient {
	return &serviceDiscoveryClient{
		log:              ctrl.Log.WithName("cloudmap"),
		sdApi:            NewServiceDiscoveryApiFromConfig(cfg),
		namespaceIdCache: cache.NewLRUExpireCache(defaultNamespaceIdCacheSize),
		serviceIdCache:   cache.NewLRUExpireCache(defaultServiceIdCacheSize),
		endpointCache:    cache.NewLRUExpireCache(defaultEndpointsCacheSize),
	}
}

func (sdc *serviceDiscoveryClient) ListServices(ctx context.Context, nsName string) (svcs []*model.Service, err error) {
	nsId, err := sdc.getNamespaceId(ctx, nsName)
	if err != nil || nsId == "" {
		return svcs, err
	}

	svcSums, err := sdc.sdApi.ListServices(ctx, nsId)

	if err != nil {
		return svcs, err
	}

	for _, svcSum := range svcSums {
		sdc.cacheServiceId(nsName, svcSum.Name, svcSum.Id)

		endpts, endptsErr := sdc.listEndpoints(ctx, svcSum.Id)

		if endptsErr != nil {
			return svcs, endptsErr
		}

		svcs = append(svcs, &model.Service{
			Namespace: nsName,
			Name:      svcSum.Name,
			Endpoints: endpts,
		})
	}

	return svcs, nil
}

func (sdc *serviceDiscoveryClient) CreateService(ctx context.Context, nsName string, svcName string) (err error) {
	sdc.log.Info("creating a new service", "namespace", nsName, "name", svcName)

	nsId, err := sdc.getNamespaceId(ctx, nsName)
	if err != nil {
		return err
	}

	if nsId == "" {
		nsId, err = sdc.createNamespace(ctx, nsName)
	}
	if err != nil {
		return err
	}

	//TODO: Handle non-http namespaces
	svcId, err := sdc.sdApi.CreateService(ctx, nsId, svcName)

	if err != nil {
		return err
	}

	sdc.cacheServiceId(nsName, svcName, svcId)

	return nil
}

func (sdc *serviceDiscoveryClient) GetService(ctx context.Context, nsName string, svcName string) (svc *model.Service, err error) {
	sdc.log.Info("fetching a service", "nsName", nsName, "svcName", svcName)

	svcId, err := sdc.getServiceId(ctx, nsName, svcName)

	if err != nil {
		return nil, err
	}

	if svcId == "" {
		return nil, nil
	}

	endpts, err := sdc.listEndpoints(ctx, svcId)

	if err != nil {
		return nil, err
	}

	svc = &model.Service{
		Namespace: nsName,
		Name:      svcName,
		Endpoints: endpts,
	}

	return svc, nil
}

func (sdc *serviceDiscoveryClient) RegisterEndpoints(ctx context.Context, nsName string, svcName string, endpts []*model.Endpoint) (err error) {
	if len(endpts) == 0 {
		sdc.log.Info("skipping endpoint registration for empty endpoint list", "serviceName", svcName)
		return nil
	}

	sdc.log.Info("registering endpoints", "namespaceName", nsName, "serviceName", svcName, "endpoints", endpts)

	svcId, err := sdc.getServiceId(ctx, nsName, svcName)
	if err != nil {
		return err
	}

	opCollector := NewOperationCollector()

	for _, endpt := range endpts {
		opCollector.Add(func() (opId string, err error) {
			return sdc.sdApi.RegisterInstance(ctx, svcId, endpt.Id, endpt.GetCloudMapAttributes())
		})
	}

	err = NewRegisterInstancePoller(sdc.sdApi, svcId, opCollector.Collect(), opCollector.GetStartTime()).Poll(ctx)

	// Evict cache entry so next list call reflects changes
	sdc.evictEndpoints(svcId)

	if err != nil {
		return err
	}

	if !opCollector.IsAllOperationsCreated() {
		return fmt.Errorf("failure while registering endpoints")
	}

	return nil
}

func (sdc *serviceDiscoveryClient) DeleteEndpoints(ctx context.Context, nsName string, svcName string, endpts []*model.Endpoint) (err error) {
	if len(endpts) == 0 {
		sdc.log.Info("skipping endpoint deletion for empty endpoint list", "serviceName", svcName)
		return nil
	}

	sdc.log.Info("deleting endpoints", "namespaceName", nsName,
		"serviceName", svcName, "endpoints", endpts)

	svcId, err := sdc.getServiceId(ctx, nsName, svcName)
	if err != nil {
		return err
	}

	opCollector := NewOperationCollector()

	for _, endpt := range endpts {
		opCollector.Add(func() (opId string, err error) {
			return sdc.sdApi.DeregisterInstance(ctx, svcId, endpt.Id)
		})
	}

	err = NewDeregisterInstancePoller(sdc.sdApi, svcId, opCollector.Collect(), opCollector.GetStartTime()).Poll(ctx)

	// Evict cache entry so next list call reflects changes
	sdc.evictEndpoints(svcId)

	if err != nil {
		return err
	}

	if !opCollector.IsAllOperationsCreated() {
		return fmt.Errorf("failure while de-registering endpoints")
	}

	return nil
}

func (sdc *serviceDiscoveryClient) listEndpoints(ctx context.Context, serviceId string) (endpts []*model.Endpoint, err error) {

	if cachedValue, exists := sdc.endpointCache.Get(serviceId); exists {
		return cachedValue.([]*model.Endpoint), nil
	}

	endpts, err = sdc.sdApi.ListInstances(ctx, serviceId)

	if err != nil {
		return nil, err
	}

	sdc.cacheEndpoints(serviceId, endpts)

	return endpts, nil
}

func (sdc *serviceDiscoveryClient) getNamespaceId(ctx context.Context, nsName string) (nsId string, err error) {
	// We are assuming a unique namespace name per account
	if cachedValue, exists := sdc.namespaceIdCache.Get(nsName); exists {
		return cachedValue.(string), nil
	}

	namespaces, err := sdc.sdApi.ListNamespaces(ctx)

	if err != nil {
		return "", err
	}

	for _, ns := range namespaces {
		sdc.cacheNamespaceId(ns.Name, ns.Id)
		if nsName == ns.Name {
			nsId = ns.Id
		}
	}

	// This will cache empty namespace IDs for namespaces not in Cloud Map
	sdc.cacheNamespaceId(nsName, nsId)
	return nsId, nil
}

func (sdc *serviceDiscoveryClient) getServiceId(ctx context.Context, nsName string, svcName string) (svcId string, err error) {
	cacheKey := sdc.buildServiceIdCacheKey(nsName, svcName)

	if cachedValue, exists := sdc.serviceIdCache.Get(cacheKey); exists {
		return cachedValue.(string), nil
	}

	nsId, err := sdc.getNamespaceId(ctx, nsName)

	if err != nil {
		return "", err
	}

	if nsId == "" {
		return "", nil
	}

	svcs, err := sdc.sdApi.ListServices(ctx, nsId)

	if err != nil {
		return "", err
	}

	for _, svc := range svcs {
		sdc.cacheServiceId(nsName, svcName, svc.Id)
		if svc.Name == svcName {
			svcId = svc.Id
		}
	}

	return svcId, nil
}

func (sdc *serviceDiscoveryClient) createNamespace(ctx context.Context, nsName string) (nsId string, err error) {
	sdc.log.Info("creating a new namespace", "namespace", nsName)
	opId, err := sdc.sdApi.CreateHttpNamespace(ctx, nsName)

	if err != nil {
		return "", err
	}

	nsId, err = sdc.sdApi.PollCreateNamespace(ctx, opId)

	if err != nil {
		return "", err
	}

	sdc.cacheNamespaceId(nsName, nsId)

	return nsId, nil
}

func (sdc *serviceDiscoveryClient) cacheNamespaceId(nsName string, nsId string) {
	sdc.namespaceIdCache.Add(nsName, nsId, defaultNamespaceIdCacheTTL)
}

func (sdc *serviceDiscoveryClient) cacheServiceId(nsName string, svcName string, svcId string) {
	cacheKey := sdc.buildServiceIdCacheKey(nsName, svcName)
	sdc.serviceIdCache.Add(cacheKey, svcId, defaultServiceIdCacheTTL)
}

func (sdc *serviceDiscoveryClient) cacheEndpoints(svcId string, endpts []*model.Endpoint) {
	sdc.endpointCache.Add(svcId, endpts, defaultEndpointsCacheTTL)
}

func (sdc *serviceDiscoveryClient) evictEndpoints(svcId string) {
	sdc.endpointCache.Remove(svcId)
}

func (sdc *serviceDiscoveryClient) buildServiceIdCacheKey(nsName string, svcName string) (cacheKey string) {
	return fmt.Sprintf("%s/%s", nsName, svcName)
}
