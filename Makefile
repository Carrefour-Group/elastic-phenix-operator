include version.mk
# Image URL to use all building/pushing image targets
REPOSITORY ?= carrefourphx/elastic-phenix-operator
IMG ?= $(REPOSITORY):$(BUILD_VERSION)
LATEST_IMG ?= $(REPOSITORY):latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: manager

# Start docker-compose
docker-up:
	docker-compose -f docker-compose-test.yml up -d

# Stop docker-compose
docker-down:
	docker-compose -f docker-compose-test.yml down

# Run vet inside docker
vet-in-docker:
	docker-compose -f docker-compose-test.yml exec -T build go vet ./...

# Run fmt inside docker
fmt-in-docker:
	@if docker-compose -f docker-compose-test.yml exec -T build bash -c '[ "$$(gofmt -s -l cmd/. pkg/. | wc -l)" -gt 0 ]'; then\
		exit 1;\
	fi

# Run build inside docker
build-in-docker:
	docker-compose -f docker-compose-test.yml exec -T build go build -o bin/manager cmd/manager/main.go

# Run test inside docker
test-in-docker:
	docker-compose -f docker-compose-test.yml exec -T build go test ./... -coverprofile cover.out

# Run all inside docker
all-in-docker: docker-up vet-in-docker build-in-docker test-in-docker docker-down

# Run tests
test: generate fmt vet manifests
	go test ./... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager cmd/manager/main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./cmd/manager/main.go

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests
	kustomize build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default | kubectl apply -f -

# Generate epo-all-in-one.yaml k8s manifests file
generate-all-in-one-manifests:
	docker-compose -f docker-compose-release.yml up -d
	docker-compose -f docker-compose-release.yml exec -T release sh -c "cd config/manager && kustomize edit set image controller=${IMG}"
	docker-compose -f docker-compose-release.yml exec -T release kustomize build config/default > manifests/epo-all-in-one.yaml
	docker-compose -f docker-compose-release.yml down

# Commit epo-all-in-one.yaml k8s manifests file
commit-all-in-one-manifests: generate-all-in-one-manifests
	/usr/bin/git add manifests/epo-all-in-one.yaml
	/usr/bin/git commit -m "all-in-one k8s manifests $(BUILD_VERSION)"
	/usr/bin/git stash

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

git-unlock:
	@gpg --import /gpg/gpg.public || :
	@gpg --allow-secret-key-import --import /gpg/gpg.private || :
	@echo `cat /gpg/gpg.ownertrust` | gpg --import-ownertrust
	@git-crypt unlock
	$(eval DOCKER_USERNAME := `grep 'DOCKER_USERNAME=' docker.secret | sed 's/DOCKER_USERNAME=//'`)
	$(eval DOCKER_PASSWORD := `grep 'DOCKER_PASSWORD=' docker.secret | sed 's/DOCKER_PASSWORD=//'`)

# Build the docker image
docker-build:
	docker build . -t ${IMG}
	docker tag ${IMG} ${LATEST_IMG}

# Push the docker image
docker-push: git-unlock
	@echo "docker login ..."
	@docker login --username=${DOCKER_USERNAME} --password=${DOCKER_PASSWORD}
	docker push ${IMG}
	docker push ${LATEST_IMG}
	docker logout

docker-rmi:
	docker rmi ${IMG} ${LATEST_IMG} || true

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

# for release
package-and-publish: docker-build docker-push commit-all-in-one-manifests docker-rmi
