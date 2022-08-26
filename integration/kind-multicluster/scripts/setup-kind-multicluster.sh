#!/usr/bin/env bash

# Builds the AWS Cloud Map MCS Controller for K8s, provisions a Kubernetes clusters with Kind,
# installs Cloud Map CRDs and controller into the cluster and applies export and deployment configs.

set -e

source ./integration/kind-multicluster/scripts/common.sh

./integration/kind-test/scripts/ensure-jq.sh

MCS_CONTROLLER_IMAGE="ghcr.io/matthewgoodman13/aws-cloud-map-mcs-controller-for-k8s:latest"

# Build Docker Image of Controller
# if USE_EXISTING_IMAGE is set to true, then the image will not be built
if [ ! "$USE_EXISTING_IMAGE" = "true" ]; then
  echo "Building Docker Image of Controller"
  make docker-build
fi

# Cluster 1
$KIND_BIN create cluster --name "$KIND_SHORT1" --image "$IMAGE" --config "$C1YAML"
$KUBECTL_BIN config use-context "$CLUSTER1"
$KUBECTL_BIN create namespace "$NAMESPACE"
make install
$KUBECTL_BIN apply -f "$KIND_CONFIGS/e2e-clusterproperty-1.yaml"
kind load docker-image "${MCS_CONTROLLER_IMAGE}" --name "$KIND_SHORT1"
make deploy IMG=${MCS_CONTROLLER_IMAGE} AWS_REGION=us-west-2


# Cluster 2
$KIND_BIN create cluster --name "$KIND_SHORT2" --image "$IMAGE" --config "$C2YAML"
$KUBECTL_BIN config use-context "$CLUSTER2"
$KUBECTL_BIN create namespace "$NAMESPACE"
make install
$KUBECTL_BIN apply -f "$KIND_CONFIGS/e2e-clusterproperty-2.yaml"
kind load docker-image "${MCS_CONTROLLER_IMAGE}" --name "$KIND_SHORT2"
make deploy IMG=${MCS_CONTROLLER_IMAGE} AWS_REGION=us-west-2

exit 0
