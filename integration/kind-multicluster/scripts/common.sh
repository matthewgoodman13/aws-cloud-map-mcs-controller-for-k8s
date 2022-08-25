#!/usr/bin/env bash

export KIND_BIN='./bin/kind'
export KUBECTL_BIN='kubectl'
export LOGS='./integration/kind-test/testlog'
export CONFIGS='./integration/kind-test/configs'
export SCENARIOS='./integration/shared/scenarios'
export NAMESPACE='aws-cloud-map-mcs-kind-e2e'

# Cluster 1
export KIND_SHORT1='e2e-cluster-1'
export CLUSTER1='kind-e2e-cluster-1'
export CLUSTERID1='kind-e2e-clusterid-1'
export CLUSTERSETID1='kind-e2e-clustersetid-1'
export C1YAML='./integration/kind-multicluster/scripts/c1.yaml'

# Cluster 2
export KIND_SHORT2='e2e-cluster-2'
export CLUSTER2='kind-e2e-cluster-2'
export CLUSTERID2='kind-e2e-clusterid-2'
export CLUSTERSETID2='kind-e2e-clustersetid-2'
export C2YAML='./integration/kind-multicluster/scripts/c2.yaml'


export SERVICE='e2e-service'
export ENDPT_PORT=80
export SERVICE_PORT=8080
export IMAGE='kindest/node:v1.20.15@sha256:a6ce604504db064c5e25921c6c0fffea64507109a1f2a512b1b562ac37d652f3'
export EXPECTED_ENDPOINT_COUNT=5
export UPDATED_ENDPOINT_COUNT=6
