#!/bin/bash
# Copyright CERN.
#
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
#

deployment_base="${1}"

if [[ -z $deployment_base ]]; then
	deployment_base="../deploy/kubernetes"
fi

cd "$deployment_base" || exit 1

objects=(namespace csi-attacher-rbac csi-provisioner-rbac csi-nodeplugin-rbac csi-cvmfsplugin-attacher csi-cvmfsplugin-provisioner csi-cvmfsplugin)

for obj in ${objects[@]}; do
	kubectl create -f "./$obj.yaml"
done
