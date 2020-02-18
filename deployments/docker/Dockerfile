FROM centos:7
LABEL description="CernVM-FS CSI Plugin"

RUN yum install -y http://ecsft.cern.ch/dist/cvmfs/cvmfs-release/cvmfs-release-latest.noarch.rpm && \
    yum install -y cvmfs && yum clean all && rm -rf /var/cache/yum

COPY configs/default.local /etc/cvmfs

COPY _dist/linux-amd64/csi-cvmfsplugin_linux_amd64 /csi-cvmfsplugin
RUN chmod +x /csi-cvmfsplugin
ENTRYPOINT ["/csi-cvmfsplugin"]
