package v1alpha1

import (
	"context"

	"github.com/aws/aws-cloud-map-mcs-controller-for-k8s/pkg/common"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var (
	clusterUtils *common.ClusterUtils
	log          = logf.Log.WithName("serviceexport-resource")
)

func (r *ServiceExport) SetupWebhookWithManager(mgr ctrl.Manager, cUtils common.ClusterUtils) error {
	clusterUtils = &cUtils
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-multicluster-x-k8s-io-v1alpha1-serviceexport,mutating=true,failurePolicy=fail,sideEffects=None,groups=multicluster.x-k8s.io,resources=serviceexports,verbs=create;update,versions=v1alpha1,name=mserviceexport.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &ServiceExport{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *ServiceExport) Default() {
	log.Info("default", "name", r.Name)
}

//+kubebuilder:webhook:path=/validate-multicluster-x-k8s-io-v1alpha1-serviceexport,mutating=false,failurePolicy=fail,sideEffects=None,groups=multicluster.x-k8s.io,resources=serviceexports,verbs=create;update,versions=v1alpha1,name=vserviceexport.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &ServiceExport{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *ServiceExport) ValidateCreate() error {
	log.Info("validate create", "name", r.Name)
	return r.ValidateServiceExport()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *ServiceExport) ValidateUpdate(old runtime.Object) error {
	log.Info("validate update", "name", r.Name)
	return r.ValidateServiceExport()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *ServiceExport) ValidateDelete() error {
	log.Info("validate delete", "name", r.Name)
	return nil
}

func (r *ServiceExport) ValidateServiceExport() error {
	// ServiceExport will be rejected if cluster does not have clusterId and clusterSetId defined
	clusterId, err := clusterUtils.GetClusterId(context.TODO())
	if err != nil {
		return err
	}
	clusterSetId, err := clusterUtils.GetClusterSetId(context.TODO())
	if err != nil {
		return err
	}
	log.Info("Validated ServiceExport with name %s, ClusterId: %s, and ClusterSetId: %s\n", r.Name, clusterId, clusterSetId)
	r.Status = ServiceExportStatus{
		Conditions: []metav1.Condition{
			{
				Type:   ServiceExportValid,
				Status: metav1.ConditionTrue,
			},
		},
	}

	// TODO: Check Cloud Map for existing service with same clusterId and clusterSetId

	return nil
}
