.PHONY: test clean qtest deploy ovn del-ovn dist
APP_VERSION:=$(shell cat VERSION | head -1)
GIT_COMMIT:=$(shell git describe --dirty --always)
GIT_BRANCH:=$(shell git rev-parse --abbrev-ref HEAD -- | head -1)
BUILD_USER:=$(shell whoami)
BUILD_DATE:=$(shell date +"%Y-%m-%d")
BINARY:=ovn-exporter
VERBOSE:=-v
PROJECT=github.com/forward53/ovn_exporter
PKG_DIR=pkg/ovn_exporter

all:
	@echo "Version: $(APP_VERSION), Branch: $(GIT_BRANCH), Revision: $(GIT_COMMIT)"
	@echo "Build on $(BUILD_DATE) by $(BUILD_USER)"
	@mkdir -p bin/
	@rm -rf ./bin/*
	@CGO_ENABLED=0 go build -o ./bin/$(BINARY) $(VERBOSE) \
		-ldflags="-w -s \
		-X github.com/prometheus/common/version.Version=$(APP_VERSION) \
		-X github.com/prometheus/common/version.Revision=$(GIT_COMMIT) \
		-X github.com/prometheus/common/version.Branch=$(GIT_BRANCH) \
		-X github.com/prometheus/common/version.BuildUser=$(BUILD_USER) \
		-X github.com/prometheus/common/version.BuildDate=$(BUILD_DATE) \
		-X $(PROJECT)/$(PKG_DIR).appName=$(BINARY) \
		-X $(PROJECT)/$(PKG_DIR).appVersion=$(APP_VERSION) \
		-X $(PROJECT)/$(PKG_DIR).gitBranch=$(GIT_BRANCH) \
		-X $(PROJECT)/$(PKG_DIR).gitCommit=$(GIT_COMMIT) \
		-X $(PROJECT)/$(PKG_DIR).buildUser=$(BUILD_USER) \
		-X $(PROJECT)/$(PKG_DIR).buildDate=$(BUILD_DATE)" \
		-gcflags="all=-trimpath=$(GOPATH)/src" \
		-asmflags="all=-trimpath $(GOPATH)/src" \
		./cmd/ovn_exporter/*.go
	@echo "Done!"

test: all
	@go test -v ./$(PKG_DIR)/*.go
	@echo "PASS: core tests"
	@echo "OK: all tests passed!"

clean:
	@rm -rf bin/
	@rm -rf dist/
	@echo "OK: clean up completed"

deploy:
	@sudo rm -rf /usr/sbin/$(BINARY)
	@sudo cp ./bin/$(BINARY) /usr/sbin/$(BINARY)
	@sudo usermod -a -G openvswitch ovn_exporter
	@sudo chmod g+w /var/run/openvswitch/db.sock
	@sudo setcap cap_sys_admin,cap_sys_nice,cap_dac_override+ep /usr/sbin/$(BINARY)

qtest:
	@./bin/$(BINARY) -version
	@sudo ./bin/$(BINARY) -web.listen-address 0.0.0.0:5000 -log.level debug -ovn.poll-interval 5

dist: all
	@mkdir -p ./dist
	@rm -rf ./dist/*
	@mkdir -p ./dist/$(BINARY)-$(APP_VERSION).linux-amd64
	@cp ./bin/$(BINARY) ./dist/$(BINARY)-$(APP_VERSION).linux-amd64/
	@cp ./README.md ./dist/$(BINARY)-$(APP_VERSION).linux-amd64/
	@cp LICENSE ./dist/$(BINARY)-$(APP_VERSION).linux-amd64/
	@cp assets/systemd/add_service.sh ./dist/$(BINARY)-$(APP_VERSION).linux-amd64/install.sh
	@chmod +x ./dist/$(BINARY)-$(APP_VERSION).linux-amd64/*.sh
	@cd ./dist/ && tar -cvzf ./$(BINARY)-$(APP_VERSION).linux-amd64.tar.gz ./$(BINARY)-$(APP_VERSION).linux-amd64

ovn:
	@sudo ovn-nbctl \
		ls-add 19a05268b5eb3df10e2d50b8220505ea0026679bb62eb39d71c8707dd5165248 -- \
        set Logical_Switch 19a05268b5eb3df10e2d50b8220505ea0026679bb62eb39d71c8707dd5165248 \
        external_ids:subnet=10.10.10.0/24 \
        external_ids:gateway_ip=10.10.10.1 \
        external_ids:subnet_context=default || true
	@sudo ovsdb-client transact unix:/run/openvswitch/ovnsb_db.sock '["OVN_Southbound",{"op":"update","table":"Datapath_Binding","where":[["external_ids","includes",["map",[["name","19a05268b5eb3df10e2d50b8220505ea0026679bb62eb39d71c8707dd5165248"]]]]],"row":{"tunnel_key":6500120}}]'
	@echo "*** ADD CHASSIS: gateway nyrtr1"
	@sudo ovn-sbctl chassis-add nyrtr1-6500120-vlan-20 vxlan 172.16.10.1 || true
	@sudo ovsdb-client transact unix:/run/openvswitch/ovnsb_db.sock '["OVN_Southbound",{"op":"update","table":"Chassis","where":[["name","==","nyrtr1-6500120-vlan-20"]],"row":{"hostname":"nyrtr1-v20"}}]'
	@sudo ovn-nbctl lsp-add 19a05268b5eb3df10e2d50b8220505ea0026679bb62eb39d71c8707dd5165248 nyrtr1-6500120-vlan-20-p1 || true
	@sudo ovn-nbctl lsp-set-addresses nyrtr1-6500120-vlan-20-p1 "b7:65:9d:71:22:07 10.10.10.1"
	@sudo ovn-sbctl lsp-bind nyrtr1-6500120-vlan-20-p1 nyrtr1-6500120-vlan-20 || true
	@echo "*** ADD CHASSIS: gateway nyrtr2"
	@sudo ovn-sbctl chassis-add nyrtr2-6500120-vlan-20 vxlan 172.16.10.2 || true
	@sudo ovsdb-client transact unix:/run/openvswitch/ovnsb_db.sock '["OVN_Southbound",{"op":"update","table":"Chassis","where":[["name","==","nyrtr2-6500120-vlan-20"]], "row":{"hostname":"nyrtr2-v20"}}]'
	@sudo ovn-nbctl lsp-add 19a05268b5eb3df10e2d50b8220505ea0026679bb62eb39d71c8707dd5165248 nyrtr2-6500120-vlan-20-p1 || true
	@sudo ovn-nbctl lsp-set-addresses nyrtr2-6500120-vlan-20-p1 "b7:65:9d:71:22:07 10.10.10.2"
	@sudo ovn-sbctl lsp-bind nyrtr2-6500120-vlan-20-p1 nyrtr2-6500120-vlan-20 || true
	@echo "*** ADD CHASSIS: host nyhost1"
	@sudo ovn-sbctl chassis-add 7592b50a-c201-48ea-8737-4748c185237f geneve 172.16.10.10 || true
	@sudo ovsdb-client transact unix:/run/openvswitch/ovnsb_db.sock '["OVN_Southbound",{"op":"update","table":"Chassis","where":[["name","==","7592b50a-c201-48ea-8737-4748c185237f"]],"row":{"hostname":"nyhost1"}}]'
	@echo "*** ADD CHASSIS: host nyhost2"
	@sudo ovn-sbctl chassis-add fa2b92b1-83ff-47a4-ad4a-da219df28a91 geneve 172.16.10.20 || true
	@sudo ovsdb-client transact unix:/run/openvswitch/ovnsb_db.sock '["OVN_Southbound",{"op":"update","table":"Chassis","where":[["name","==","fa2b92b1-83ff-47a4-ad4a-da219df28a91"]],"row":{"hostname":"nyhost2"}}]'
	@sudo ovn-nbctl lsp-add 19a05268b5eb3df10e2d50b8220505ea0026679bb62eb39d71c8707dd5165248 9da77936277dcf536dd03fa0351578948aaf9b9e599063fb9b305e4b2ef977a8 || true
	@sudo ovn-nbctl lsp-set-addresses 9da77936277dcf536dd03fa0351578948aaf9b9e599063fb9b305e4b2ef977a8 "02:54:b4:11:3b:e6 10.10.10.111"
	@sudo ovn-sbctl lsp-bind 9da77936277dcf536dd03fa0351578948aaf9b9e599063fb9b305e4b2ef977a8 7592b50a-c201-48ea-8737-4748c185237f || true
	@sudo ovn-nbctl lsp-add 19a05268b5eb3df10e2d50b8220505ea0026679bb62eb39d71c8707dd5165248 0375c97d5224fbe7cd2d10bfe4c14340b89b288ad20bd759b4a1e385fbb81395 || true
	@sudo ovn-nbctl lsp-set-addresses 0375c97d5224fbe7cd2d10bfe4c14340b89b288ad20bd759b4a1e385fbb81395 "02:27:5b:bd:a9:70 10.10.10.112"
	@sudo ovn-sbctl lsp-bind 0375c97d5224fbe7cd2d10bfe4c14340b89b288ad20bd759b4a1e385fbb81395 7592b50a-c201-48ea-8737-4748c185237f || true
	@sudo ovn-nbctl lsp-add 19a05268b5eb3df10e2d50b8220505ea0026679bb62eb39d71c8707dd5165248 024126f4fe4cc21a95e8aa686d9e30767f3222a2f2adac5d9d1c85f63c27bfd7 || true
	@sudo ovn-nbctl lsp-set-addresses 024126f4fe4cc21a95e8aa686d9e30767f3222a2f2adac5d9d1c85f63c27bfd7 "02:e9:3d:b0:f9:f5 10.10.10.121"
	@sudo ovn-sbctl lsp-bind 024126f4fe4cc21a95e8aa686d9e30767f3222a2f2adac5d9d1c85f63c27bfd7 fa2b92b1-83ff-47a4-ad4a-da219df28a91 || true
	@sudo ovn-nbctl lsp-add 19a05268b5eb3df10e2d50b8220505ea0026679bb62eb39d71c8707dd5165248 2b2bf7e475a74b6f48b8c92c750a13188b1c09f4c7e6342ed8aaa62a00628969 || true
	@sudo ovn-nbctl lsp-set-addresses 2b2bf7e475a74b6f48b8c92c750a13188b1c09f4c7e6342ed8aaa62a00628969 "02:32:6b:ca:25:9e 10.10.10.122"
	@sudo ovn-sbctl lsp-bind 2b2bf7e475a74b6f48b8c92c750a13188b1c09f4c7e6342ed8aaa62a00628969 fa2b92b1-83ff-47a4-ad4a-da219df28a91 || true
	@sudo ovn-nbctl show
	@sudo ovn-sbctl show

del-ovn:
	@sudo ovn-sbctl lsp-unbind nyrtr1-6500120-vlan-20-p1 || true
	@sudo ovn-nbctl lsp-del nyrtr1-6500120-vlan-20-p1 || true
	@sudo ovn-sbctl lsp-unbind nyrtr2-6500120-vlan-20-p1 || true
	@sudo ovn-nbctl lsp-del nyrtr2-6500120-vlan-20-p1 || true
	@sudo ovn-sbctl lsp-unbind 9da77936277dcf536dd03fa0351578948aaf9b9e599063fb9b305e4b2ef977a8 || true
	@sudo ovn-nbctl lsp-del 9da77936277dcf536dd03fa0351578948aaf9b9e599063fb9b305e4b2ef977a8 || true
	@sudo ovn-sbctl lsp-unbind 0375c97d5224fbe7cd2d10bfe4c14340b89b288ad20bd759b4a1e385fbb81395 || true
	@sudo ovn-nbctl lsp-del 0375c97d5224fbe7cd2d10bfe4c14340b89b288ad20bd759b4a1e385fbb81395 || true
	@sudo ovn-sbctl lsp-unbind 024126f4fe4cc21a95e8aa686d9e30767f3222a2f2adac5d9d1c85f63c27bfd7 || true
	@sudo ovn-nbctl lsp-del 024126f4fe4cc21a95e8aa686d9e30767f3222a2f2adac5d9d1c85f63c27bfd7 || true
	@sudo ovn-sbctl lsp-unbind 2b2bf7e475a74b6f48b8c92c750a13188b1c09f4c7e6342ed8aaa62a00628969 || true
	@sudo ovn-nbctl lsp-del 2b2bf7e475a74b6f48b8c92c750a13188b1c09f4c7e6342ed8aaa62a00628969 || true
	@sudo ovn-sbctl chassis-del nyrtr1-6500120-vlan-20 || true
	@sudo ovn-sbctl chassis-del nyrtr2-6500120-vlan-20 || true
	@sudo ovn-sbctl chassis-del 7592b50a-c201-48ea-8737-4748c185237f || true
	@sudo ovn-sbctl chassis-del fa2b92b1-83ff-47a4-ad4a-da219df28a91 || true
	@sudo ovn-nbctl ls-del 19a05268b5eb3df10e2d50b8220505ea0026679bb62eb39d71c8707dd5165248 || true
	@sudo ovn-nbctl show
	@sudo ovn-sbctl show
