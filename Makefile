.PHONY: all cvmfsplugin

NAME=csi-cvmfsplugin
IMAGE_VERSION=v0.3.0

all: cvmfsplugin

test:
	go test gitlab.cern.ch/cloud-infrastructure/cvmfs-csi/pkg/... -cover
	go vet gitlab.cern.ch/cloud-infrastructure/cvmfs-csi/pkg/...

cvmfsplugin:
	if [ ! -d ./vendor ]; then dep ensure; fi
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o  _output/$(NAME) ./cvmfs

image: cvmfsplugin 
	cp _output/$(NAME) deploy/docker
	docker build -t $(NAME):$(IMAGE_VERSION) deploy/docker

clean:
	go clean -r -x
	rm -f deploy/docker/rbdplugin
