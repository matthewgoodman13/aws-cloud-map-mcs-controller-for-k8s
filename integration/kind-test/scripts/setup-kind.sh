#!/usr/bin/env bash

. ./util.sh

# Builds the AWS Cloud Map MCS Controller for K8s, provisions Kubernetes clusters with Kind,
# installs Cloud Map CRDs and controller into the clusters and applies export and deployment configs.

set -e

source ./integration/kind-test/scripts/common.sh

./integration/kind-test/scripts/ensure-jq.sh

$KIND_BIN create cluster --name "$KIND_SHORT" --image "$IMAGE"
$KUBECTL_BIN config use-context "$CLUSTER"
$KUBECTL_BIN create namespace "$NAMESPACE"
make install

exit 0
