# Copyright (c) 2016 Intel Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
GOBIN=$(GOPATH)/bin
APP_DIR_LIST=$(shell go list ./... | grep -v /vendor/)
APP_NAME=tap-container-broker

build: verify_gopath
	go fmt $(APP_DIR_LIST)
	CGO_ENABLED=0 go install -ldflags "-w" -tags netgo .
	mkdir -p application && cp -f $(GOBIN)/$(APP_NAME) ./application/$(APP_NAME)

verify_gopath:
	$(call verify_env_var,GOPATH)

verify_repository_url:
	$(call verify_env_var,REPOSITORY_URL)

docker_build: build_anywhere
	docker build -t $(APP_NAME) .

push_docker: verify_repository_url docker_build
	docker tag $(APP_NAME) $(REPOSITORY_URL)/tap-container-broker:latest
	docker push $(REPOSITORY_URL)/tap-container-broker:latest

kubernetes_deploy: docker_build
	kubectl create -f configmap.yaml
	kubectl create -f service.yaml
	kubectl create -f deployment.yaml

kubernetes_update: docker_build
	kubectl delete -f deployment.yaml
	kubectl create -f deployment.yaml

deps_fetch_specific: bin/govendor
	$(call verify_env_var,DEP_URL)
	@echo "Fetching specific dependency in newest versions"
	$(GOBIN)/govendor fetch -v $(DEP_URL)

deps_update_tap: verify_gopath
	$(GOBIN)/govendor update github.com/trustedanalytics-ng/...
	$(GOBIN)/govendor remove github.com/trustedanalytics-ng/$(APP_NAME)/...
	@echo "Done"

prepare_dirs:
	mkdir -p ./temp/src/github.com/trustedanalytics-ng/tap-container-broker
	$(eval REPOFILES=$(shell pwd)/*)
	ln -sf $(REPOFILES) temp/src/github.com/trustedanalytics-ng/tap-container-broker

build_anywhere: prepare_dirs
	$(eval GOPATH=$(shell cd ./temp; pwd))
	$(eval APP_DIR_LIST=$(shell GOPATH=$(GOPATH) go list ./temp/src/github.com/trustedanalytics-ng/tap-container-broker/... | grep -v /vendor/))
	GOPATH=$(GOPATH) CGO_ENABLED=0 go build -tags netgo $(APP_DIR_LIST)
	rm -Rf application && mkdir application
	cp ./tap-container-broker ./application/tap-container-broker
	rm -Rf ./temp

install_mockgen:
	scripts/install_mockgen.sh

mock_update:
	$(GOBIN)/mockgen -source=k8s/k8sfabricator.go -package=mocks -destination=mocks/k8sfabricator_mock.go
	$(GOBIN)/mockgen -source=vendor/github.com/trustedanalytics-ng/tap-ceph-broker/client/client.go -package=mocks -destination=mocks/ceph_mock.go
	$(GOBIN)/mockgen -source=vendor/github.com/trustedanalytics-ng/tap-catalog/client/client.go -package=mocks -destination=mocks/catalog_mock.go
	$(GOBIN)/mockgen -source=vendor/github.com/trustedanalytics-ng/tap-template-repository/client/client_api.go -package=mocks -destination=mocks/template_mock.go
	$(GOBIN)/mockgen -source=vendor/github.com/trustedanalytics-ng/tap-ca/client/client.go -package=mocks -destination=mocks/ca_mock.go
	$(GOBIN)/mockgen -source=engine/engine.go -package=mocks -destination=mocks/engine_mock.go
	$(GOBIN)/mockgen -source=engine/processor/processor.go -package=mocks -destination=mocks/processor_mock.go

	./add_license.sh

test: verify_gopath
	CGO_ENABLED=0 go test -ldflags "-w" -tags netgo --cover $(APP_DIR_LIST)

define verify_env_var
	@if [ -z $($(1)) ] || [ $($(1)) = "" ]; then\
		echo "$(1) not set. You need to set $(1) before run this command";\
		exit 1 ;\
	fi
endef
