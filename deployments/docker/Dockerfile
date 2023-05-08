FROM registry.access.redhat.com/ubi9/ubi:latest

LABEL description="CernVM-FS CSI Plugin"

RUN dnf install -y http://ecsft.cern.ch/dist/cvmfs/cvmfs-release/cvmfs-release-latest.noarch.rpm && \
    dnf --setopt=install_weak_deps=False install -y cvmfs && dnf --enablerepo=* clean all && rm -rf /var/cache/yum /var/cache/dnf

COPY bin/csi-cvmfsplugin /csi-cvmfsplugin
COPY bin/automount-runner /automount-runner

RUN chmod +x /csi-cvmfsplugin /automount-runner
