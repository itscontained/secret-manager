#!/bin/bash

# Copyright 2017 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e
if [ -n "$DEBUG" ]; then
	set -x
fi

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

export NAMESPACE=$1
export HELM_VALUES_FILE=${2:-default}
export RELEASE_NAME=${3:-undefined}

echo "deploying secret-manager in namespace $NAMESPACE"

function on_exit {
    local error_code="$?"

    test $error_code == 0 && return;

    echo "Obtaining secret-manager pod logs..."
    kubectl logs -l app=secret-manager -n "$NAMESPACE"
}
trap on_exit EXIT

echo "Helm values file $HELM_VALUES_FILE is being used for namespace $NAMESPACE"

helm install "$RELEASE_NAME" "$DIR/deploy/charts/secret-manager" \
    --namespace="$NAMESPACE" \
    --values "$DIR/$HELM_VALUES_FILE.yaml" \
    --wait \
    --timeout=5m
