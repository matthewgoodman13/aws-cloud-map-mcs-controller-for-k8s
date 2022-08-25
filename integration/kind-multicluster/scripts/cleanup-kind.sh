#!/usr/bin/env bash

# Deletes Kind clusters used for integration test.

set -eo pipefail
source ./integration/kind-multicluster/scripts/common.sh

$KIND_BIN delete cluster --name "$KIND_SHORT1"
$KIND_BIN delete cluster --name "$KIND_SHORT2"