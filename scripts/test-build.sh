#!/bin/bash
# Creates a test build of image and installs this to a cluster.
#
# This aims to make it easier to imperatively test builds when developing
# locally without require updating all the image repository paths in
# the helm chart values file.
#
# Usage: BASE_REPOSITORY=your-registry IMAGE_BUILD_TOOL=docker TARGETS=linux/amd64 IMAGE_ADDITIONAL_ARGS="--platform linux/amd64" ./test-build.sh

set -xeoa pipefail

BASE_REPOSITORY=${BASE_REPOSITORY}
IMAGE_BUILD_TOOL=${IMAGE_BUILD_TOOL:=docker}
IMAGE_ADDITIONAL_ARGS=${IMAGE_ADDITIONAL_ARGS}
TARGETS=${TARGETS:=linux/amd64}

GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
IMAGE_TAG=${GIT_BRANCH//\//-}

BASE_REPOSITORY=$BASE_REPOSITORY \
  IMAGE_ADDITIONAL_ARGS=$IMAGE_ADDITIONAL_ARGS \
  IMAGE_BUILD_TOOL=$IMAGE_BUILD_TOOL \
  TARGETS=$TARGETS \
  make image
docker push $BASE_REPOSITORY:$IMAGE_TAG

helm install -n kube-system \
  cvmfs-csi \
  ./deployments/helm/cvmfs-csi/ \
  --set controllerplugin.plugin.image.repository=$BASE_REPOSITORY \
  --set controllerplugin.plugin.image.tag=$IMAGE_TAG \
  --set nodeplugin.plugin.image.repository=$BASE_REPOSITORY \
  --set nodeplugin.plugin.image.tag=$IMAGE_TAG \
  --set nodeplugin.automount.image.repository=$BASE_REPOSITORY \
  --set nodeplugin.automount.image.tag=$IMAGE_TAG \
  --set nodeplugin.automountReconciler.image.repository=$BASE_REPOSITORY \
  --set nodeplugin.automountReconciler.image.tag=$IMAGE_TAG \
  --set nodeplugin.prefetcher.image.repository=$BASE_REPOSITORY \
  --set nodeplugin.prefetcher.image.tag=$IMAGE_TAG \
  --set nodeplugin.singlemount.image.repository=$BASE_REPOSITORY \
  --set nodeplugin.singlemount.image.tag=$IMAGE_TAG \
  --set automountStorageClass.create=true
