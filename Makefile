IMAGE_REPO ?= quay.io/slucidi/ebpf-iomon
IMAGE_TAG  ?= latest

.PHONY: generate build test image push deploy undeploy clean lint fmt

generate:
	go generate ./pkg/ebpf/...

build: generate
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
		go build -ldflags="-s -w" -o bin/ebpf-iomon ./cmd/ebpf-iomon/

test:
	go test -v -count=1 ./pkg/device/... ./pkg/metrics/... ./pkg/config/... ./pkg/k8s/...

image:
	podman build \
		--build-arg KERNEL_VERSION=$(KERNEL_VERSION) \
		-t $(IMAGE_REPO):$(IMAGE_TAG) .

push: image
	podman push $(IMAGE_REPO):$(IMAGE_TAG)

deploy:
	oc apply -k deploy/

undeploy:
	oc delete -k deploy/

clean:
	rm -rf bin/
	rm -f pkg/ebpf/*_bpfel.go pkg/ebpf/*_bpfeb.go pkg/ebpf/*.o

lint:
	golangci-lint run ./...

fmt:
	gofmt -s -w .
