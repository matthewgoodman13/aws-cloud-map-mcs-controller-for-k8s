package controllers

import (
	"context"
	"os"
	"os/exec"

	aboutv1alpha1 "github.com/aws/aws-cloud-map-mcs-controller-for-k8s/pkg/apis/about/v1alpha1"
	multiclusterv1alpha1 "github.com/aws/aws-cloud-map-mcs-controller-for-k8s/pkg/apis/multicluster/v1alpha1"
	"github.com/aws/aws-cloud-map-mcs-controller-for-k8s/pkg/common"
	"github.com/go-logr/logr/testr"
	"github.com/stretchr/testify/assert"

	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestClusterIDDoesNotExist(t *testing.T) {
	fakeClient := fake.NewClientBuilder().
		WithScheme(getClusterPropertyScheme()).
		Build()

	cpController := getClusterPropertyController(t, fakeClient)

	// Run the crashing code when BE_CRASHER is set
	if os.Getenv("BE_CRASHER") == "1" {
		cpController.Start(context.TODO())
		return
	}

	// Run the test in a subprocess
	cmd := exec.Command(os.Args[0], "-test.run=TestClusterIDDoesNotExist")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()

	// Cast the error as *exec.ExitError and compare the result
	e, ok := err.(*exec.ExitError)
	expectedErrorString := "exit status 1"
	assert.Equal(t, true, ok)
	assert.Equal(t, expectedErrorString, e.Error())
}

func TestClusterIDExists(t *testing.T) {
	fakeClient := fake.NewClientBuilder().
		WithScheme(getClusterPropertyScheme()).
		WithObjects(clusterIdForTest()).
		Build()

	cpController := getClusterPropertyController(t, fakeClient)
	cpController.Start(context.Background())

	// Run the crashing code when BE_CRASHER is set
	if os.Getenv("BE_CRASHER") == "1" {
		cpController.Start(context.TODO())
		return
	}

	// Run the test in a subprocess
	cmd := exec.Command(os.Args[0], "-test.run=TestClusterIDExists")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()

	assert.Nil(t, err)
}

func getClusterPropertyController(t *testing.T, client client.Client) *ClusterPropertyController {
	return &ClusterPropertyController{
		Client: client,
		Log:    common.NewLoggerWithLogr(testr.New(t)),
	}
}

func getClusterPropertyScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(multiclusterv1alpha1.GroupVersion, &aboutv1alpha1.ClusterProperty{})
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &v1.Service{})
	return scheme
}

func clusterIdForTest() *aboutv1alpha1.ClusterProperty {
	return &aboutv1alpha1.ClusterProperty{
		ObjectMeta: metav1.ObjectMeta{
			Name: ClusterIdName,
		},
		Spec: aboutv1alpha1.ClusterPropertySpec{
			Value: "test_clusterid_uuid",
		},
	}
}
