#外部参数
BINARY_DIR=$(binary_dir)
ifeq (${BINARY_DIR}, )
	BINARY_DIR=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))/bin
endif

RELEASE_VER=$(tag)
ifeq (${RELEASE_VER}, )
	RELEASE_VER=latest
endif

DOCKER_HUB_HOST=$(docker_hub_host)
DOCKER_HUB_USERNAME=$(docker_hub_userame)
DOCKER_HUB_PASSWD=$(docker_hub_passwd)
DOCKER_HUB_PROJECT=$(docker_hub_project)

CHARTS_GIT_DIR=./tmp/gitcharts
CHARTS_GIT_CLONE=$(charts_git_clone)
CHARTS_GIT_RAW=$(charts_git_raw)
CHARTS_GIT_USER_NAME=$(charts_git_user_name)
CHARTS_GIT_USER_EMAIL=$(charts_git_user_email)
NEED_LATEST=$(need_latest)


DRONE_REPO=$(drone_repo)

CLI_VERSION_PACKAGE = openi.pcl.ac.cn/Kraken/KrakenPlug/common/info
# 静态变量
LD_FLAGS=" \
    -X '$(CLI_VERSION_PACKAGE).gitCommit=`git log --pretty=format:'%h' -1`'   \
    -X '$(CLI_VERSION_PACKAGE).version=${RELEASE_VER}'"


# 编译
all_build: deviceplugin_build deviceexporter_build devicediscovery_build

init:
	mkdir -p ${BINARY_DIR}

deviceplugin_build: init
	cd ./deviceplugin && go build -ldflags ${LD_FLAGS} -o ${BINARY_DIR} ./...

deviceexporter_build: init
	cd ./deviceexporter && go build -ldflags ${LD_FLAGS} -o ${BINARY_DIR} ./...

devicediscovery_build: init
	cd ./devicediscovery && go build -ldflags ${LD_FLAGS} -o ${BINARY_DIR} ./...

kpsmi_build: init
	cd ./kpsmi && go build -ldflags ${LD_FLAGS} -o ${BINARY_DIR} ./...

kprunc_build: init
	cd ./kprunc && go build -ldflags ${LD_FLAGS} -o ${BINARY_DIR} ./...

# 构建镜像
images: deviceplugin_image deviceexporter_image devicediscovery_image

deviceplugin_image:
	docker build -t deviceplugin:${RELEASE_VER} -f ./build/application/deviceplugin/dockerfile .

deviceexporter_image:
	docker build -t deviceexporter:${RELEASE_VER} -f ./build/application/deviceexporter/dockerfile .

devicediscovery_image:
	docker build -t devicediscovery:${RELEASE_VER} -f ./build/application/devicediscovery/dockerfile .

# 镜像推送
images_push: deviceplugin_image_push deviceexporter_image_push

image_push_init:
	(echo ${DOCKER_HUB_PASSWD} | docker login ${DOCKER_HUB_HOST} -u ${DOCKER_HUB_USERNAME} --password-stdin) 1>/dev/null 2>&1

deviceplugin_image_push: image_push_init
	docker buildx build --build-arg tag=${RELEASE_VER} -t ${DOCKER_HUB_HOST}/${DOCKER_HUB_PROJECT}/deviceplugin:${RELEASE_VER} --platform=linux/arm64,linux/amd64 --provenance=false -f ./build/application/deviceplugin/dockerfile . --push
#	docker tag deviceplugin:${RELEASE_VER} ${DOCKER_HUB_HOST}/${DOCKER_HUB_PROJECT}/deviceplugin:${RELEASE_VER}
#	docker push ${DOCKER_HUB_HOST}/${DOCKER_HUB_PROJECT}/deviceplugin:${RELEASE_VER}

ifneq (${RELEASE_VER}, latest)
ifeq (${NEED_LATEST}, TRUE)
	docker buildx build --build-arg tag=${RELEASE_VER} -t ${DOCKER_HUB_HOST}/${DOCKER_HUB_PROJECT}/deviceplugin:latest --platform=linux/arm64,linux/amd64 --provenance=false -f ./build/application/deviceplugin/dockerfile . --push
endif
endif

deviceexporter_image_push: image_push_init
	docker buildx build --build-arg tag=${RELEASE_VER} -t ${DOCKER_HUB_HOST}/${DOCKER_HUB_PROJECT}/deviceexporter:${RELEASE_VER} --platform=linux/arm64,linux/amd64 --provenance=false -f ./build/application/deviceexporter/dockerfile . --push
#	docker tag deviceexporter:${RELEASE_VER} ${DOCKER_HUB_HOST}/${DOCKER_HUB_PROJECT}/deviceexporter:${RELEASE_VER}
#	docker push ${DOCKER_HUB_HOST}/${DOCKER_HUB_PROJECT}/deviceexporter:${RELEASE_VER}

ifneq (${RELEASE_VER}, latest)
ifeq (${NEED_LATEST}, TRUE)
	docker buildx build --build-arg tag=${RELEASE_VER} -t ${DOCKER_HUB_HOST}/${DOCKER_HUB_PROJECT}/deviceexporter:latest --platform=linux/arm64,linux/amd64 --provenance=false -f ./build/application/deviceexporter/dockerfile . --push
endif
endif

devicediscovery_image_push: image_push_init
	docker buildx build --build-arg tag=${RELEASE_VER} -t ${DOCKER_HUB_HOST}/${DOCKER_HUB_PROJECT}/devicediscovery:${RELEASE_VER} --platform=linux/arm64,linux/amd64 --provenance=false -f ./build/application/devicediscovery/dockerfile . --push

ifneq (${RELEASE_VER}, latest)
ifeq (${NEED_LATEST}, TRUE)
	docker buildx build --build-arg tag=${RELEASE_VER} -t ${DOCKER_HUB_HOST}/${DOCKER_HUB_PROJECT}/devicediscovery:latest --platform=linux/arm64,linux/amd64 --provenance=false -f ./build/application/devicediscovery/dockerfile . --push
endif
endif


# helm chart
charts: charts_build charts_push

charts_build:
	-mkdir -p ./tmp/charts
	helm package ./deploy/charts/krakenplug --version ${RELEASE_VER} --app-version ${RELEASE_VER} -d ./tmp/charts

charts_push:
	git clone ${CHARTS_GIT_CLONE} ${CHARTS_GIT_DIR}
	cp ./tmp/charts/krakenplug-${RELEASE_VER}.tgz ${CHARTS_GIT_DIR}
	helm repo index ${CHARTS_GIT_DIR} --url ${CHARTS_GIT_RAW}
	cd ${CHARTS_GIT_DIR} && git config --global user.email ${CHARTS_GIT_USER_EMAIL} && git config --global user.name ${CHARTS_GIT_USER_NAME} && git add index.yaml krakenplug-${RELEASE_VER}.tgz && git commit -m "${RELEASE_VER}" &&  git pull && git push


# run
runpkg_push:
	mkdir -p kptools
	CGO_ENABLED=1 GOARCH=amd64 go build -ldflags ${LD_FLAGS} -o kptools/kpsmi ./kpsmi/cmd/kpsmi/main.go
	CGO_ENABLED=1 GOARCH=amd64 go build -ldflags ${LD_FLAGS} -o kptools/kprunc ./kprunc/cmd/kprunc/main.go
	cp build/application/kprunc/config.yaml kptools
	tar -zcvf kptools.tar.gz kptools
	cp build/application/kprunc/config.yaml kptools
	cat build/script/install_run.sh kptools.tar.gz > krakenplug-${RELEASE_VER}-amd64.run
	rm kptools.tar.gz
	rm -rf kptools

	mkdir -p kptools
	CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc GOARCH=arm64 go build -ldflags ${LD_FLAGS} -o ./kptools/kpsmi ./kpsmi/cmd/kpsmi/main.go
	CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc GOARCH=arm64 go build -ldflags ${LD_FLAGS} -o ./kptools/kprunc ./kprunc/cmd/kprunc/main.go
	cp build/application/kprunc/config.yaml kptools
	tar -zcvf kptools.tar.gz kptools
	cat build/script/install_run.sh kptools.tar.gz > krakenplug-${RELEASE_VER}-arm64.run
	rm kptools.tar.gz
	rm -rf kptools

	git clone ${CHARTS_GIT_CLONE} ${CHARTS_GIT_DIR}
	cp krakenplug-${RELEASE_VER}-amd64.run krakenplug-${RELEASE_VER}-arm64.run ${CHARTS_GIT_DIR}
	cd ${CHARTS_GIT_DIR} && git config --global user.email ${CHARTS_GIT_USER_EMAIL} && git config --global user.name ${CHARTS_GIT_USER_NAME} && git add krakenplug-${RELEASE_VER}-amd64.run krakenplug-${RELEASE_VER}-arm64.run && git commit -m "${RELEASE_VER}" && git pull && git push



