FROM registry.access.redhat.com/ubi9/ubi:9.2-489

RUN dnf install -y \
      # selinux-policy is a dependency of the cvmfs package.
      # When installing cvmfs on aarch64, an older version of selinux-policy
      # package is selected for installation which then results in 404.
      # As a temporary workaround we install it explicitly until this is fixed.
      selinux-policy \
      http://ecsft.cern.ch/dist/cvmfs/cvmfs-release/cvmfs-release-latest.noarch.rpm \
    && dnf --setopt=install_weak_deps=False install -y cvmfs \
    && dnf --enablerepo=* clean all && rm -rf /var/cache/yum /var/cache/dnf

ARG TARGETARCH
ARG RELEASE
ARG GITREF
ARG CREATED

LABEL org.opencontainers.image.title="cvmfs-csi" \
      org.opencontainers.image.description="The CernVM-FS CSI image contains tools for exposing CVMFS repositories through Container Storage Interface." \
      org.opencontainers.image.authors="CernVM-FS CSI authors" \
      org.opencontainers.image.vendor="CernVM-FS CSI authors" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.documentation="https://github.com/cvmfs-contrib/cvmfs-csi" \
      org.opencontainers.image.source="https://github.com/cvmfs-contrib/cvmfs-csi" \
      org.opencontainers.image.url="https://github.com/cvmfs-contrib/cvmfs-csi" \
      org.opencontainers.image.revision=${GITREF} \
      org.opencontainers.image.version=${RELEASE} \
      org.opencontainers.image.created=${CREATED} \
      org.opencontainers.image.ref.name="" \
      org.opencontainers.image.base.digest="" \
      org.opencontainers.image.base.name=""

COPY bin/linux-${TARGETARCH}/csi-cvmfsplugin /csi-cvmfsplugin
COPY bin/linux-${TARGETARCH}/automount-runner /automount-runner
COPY bin/linux-${TARGETARCH}/singlemount-runner /singlemount-runner
