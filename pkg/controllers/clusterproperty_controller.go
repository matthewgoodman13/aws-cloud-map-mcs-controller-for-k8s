/*
Copyright 2022 The Kubernetes Authors.

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

package controllers

import (
	"context"

	"github.com/aws/aws-cloud-map-mcs-controller-for-k8s/pkg/api/v1alpha1"
	"github.com/aws/aws-cloud-map-mcs-controller-for-k8s/pkg/common"
	"github.com/google/uuid"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ClusterID_Name = "id.k8s.io"
)

// ClusterPropertyReconciler reconciles state of the properties of the cluster
type ClusterPropertyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    common.Logger
}

// +kubebuilder:rbac:groups=about.k8s.io,resources=clusterproperties,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=about.k8s.io,resources=clusterproperties/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=about.k8s.io,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=about.k8s.io,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=about.k8s.io,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=api.autoscaling.v2beta1,resources=hpa,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=about.k8s.io,resources=clusterproperties/finalizers,verbs=update

// Start implements manager.Runnable
func (r *ClusterPropertyReconciler) Start(ctx context.Context) error {
	r.Log.Info("Starting ClusterPropertyReconciler")
	res, err := r.Reconcile(ctx, ctrl.Request{})
	if err != nil {
		return err
	}
	r.Log.Info("Reconcile result", "result", res)

	return nil
}

func (r *ClusterPropertyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info("Reconciling ClusterProperty")

	// Check if the clusterID is already present, if not create it
	var clusterProperty v1alpha1.ClusterProperty
	if err := r.Get(ctx, client.ObjectKey{Name: ClusterID_Name}, &clusterProperty); err != nil {
		r.Log.Info("Unable to find ClusterID. Creating instead.")
		if err := r.CreateClusterID(ctx); err != nil {
			return ctrl.Result{}, err
		}
	}
	r.Log.Info("ClusterID ClusterProperty found", "clusterID", clusterProperty.Spec.Value)

	r.DisplayAll(ctx)

	return ctrl.Result{}, nil
}

// Get ClusterID ClusterProperty
func (r *ClusterPropertyReconciler) GetClusterID(ctx context.Context) (string, error) {
	var clusterProperty v1alpha1.ClusterProperty
	if err := r.Get(ctx, client.ObjectKey{Name: ClusterID_Name}, &clusterProperty); err != nil {
		return "", err
	}
	return clusterProperty.Spec.Value, nil
}

// Create ClusterID ClusterProperty
func (r *ClusterPropertyReconciler) CreateClusterID(ctx context.Context) error {
	clusterID := uuid.NewString()

	clusterProperty := &v1alpha1.ClusterProperty{
		ObjectMeta: metav1.ObjectMeta{
			Name: ClusterID_Name,
		},
		Spec: v1alpha1.ClusterPropertySpec{
			Value: clusterID,
		},
	}

	if err := r.Create(ctx, clusterProperty); err != nil {
		r.Log.Error(err, "Unable to create ClusterID clusterProperty")
		return err
	}

	r.Log.Info("ClusterID ClusterProperty created", "clusterID", clusterID)

	return nil
}

// Logs all clusterProperties
func (r *ClusterPropertyReconciler) DisplayAll(ctx context.Context) {
	var clusterProperties = v1alpha1.ClusterPropertyList{}
	if err := r.List(ctx, &clusterProperties); err != nil {
		r.Log.Error(err, "Unable to list the ClusterProperties")
		return
	}
	r.Log.Info("ClusterProperties", "clusterProperties", clusterProperties)

	// Delete all clusterProperties
	// for _, clusterProperty := range clusterProperties.Items {
	// 	r.Log.Info("Deleting ClusterProperty", "name", clusterProperty.Name)
	// 	if err := r.Delete(ctx, &clusterProperty); err != nil {
	// 		r.Log.Error(err, "Unable to delete ClusterProperty")
	// 		return ctrl.Result{}, err
	// 	}
	// }
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterPropertyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.ClusterProperty{}).
		Complete(r)
}
