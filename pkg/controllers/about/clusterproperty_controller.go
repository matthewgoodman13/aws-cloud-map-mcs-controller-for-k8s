package controllers

import (
	"context"
	"os"

	"github.com/aws/aws-cloud-map-mcs-controller-for-k8s/pkg/common"
	"sigs.k8s.io/controller-runtime/pkg/client"

	aboutv1alpha1 "github.com/aws/aws-cloud-map-mcs-controller-for-k8s/pkg/apis/about/v1alpha1"
)

const (
	ClusterIdName = "id.k8s.io"
)

type ClusterPropertyController struct {
	Client client.Client
	Log    common.Logger
}

func (c *ClusterPropertyController) Start(ctx context.Context) error {
	c.Log.Info("Starting ClusterProperty controller")

	clusterId := &aboutv1alpha1.ClusterProperty{}
	err := c.Client.Get(ctx, client.ObjectKey{Name: ClusterIdName}, clusterId)
	if err != nil {
		c.Log.Error(err, "Unable to start ClusterProperty controller. No ClusterID defined.")
		os.Exit(1) // Halt MCS controller when no ClusterID is defined
	}
	return nil
}
