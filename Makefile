.PHONY: all cvmfsplugin

NAME=csi-cvmfsplugin
IMAGE_VERSION=v1.0.1

all: cvmfsplugin

test:
	go test gitlab.cern.ch/cloud-infrastructure/cvmfs-csi/pkg/... -cover
	go vet gitlab.cern.ch/cloud-infrastructure/cvmfs-csi/pkg/...

cvmfsplugin:
	go version
	go mod tidy -v
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o  _output/$(NAME) ./cvmfs

image:
	cp _output/$(NAME) deploy/docker
	docker build -t $(NAME):$(IMAGE_VERSION) deploy/docker

clean:
	go clean -r -x
	rm -f deploy/docker/$(NAME)
