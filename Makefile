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


# 静态变量
Date=`date "+%Y-%m-%d %H:%M:%S"`
LD_FLAGS=" \
    -X 'main.Built=${Date}'   \
    -X 'main.Version=${RELEASE_VER}'"

# 编译
all_build: deviceplugin_build deviceexporter_build

init:
	mkdir -p ${BINARY_DIR}

deviceplugin_build: init
	cd ./deviceplugin && go build -ldflags ${LD_FLAGS} -o ${BINARY_DIR} ./...

deviceexporter_build: init
	cd ./deviceexporter && go build -ldflags ${LD_FLAGS} -o ${BINARY_DIR} ./...

# 构建镜像
images: deviceplugin_image deviceexporter_image

deviceplugin_image:
	docker build -t deviceplugin:${RELEASE_VER} -f ./build/application/deviceplugin/dockerfile .

deviceexporter_image:
	docker build -t deviceexporter:${RELEASE_VER} -f ./build/application/deviceexporter/dockerfile .

# 镜像推送
images_push: deviceplugin_image_push deviceexporter_image_push

image_push_init:
	(echo ${DOCKER_HUB_PASSWD} | docker login ${DOCKER_HUB_HOST} -u ${DOCKER_HUB_USERNAME} --password-stdin) 1>/dev/null 2>&1

deviceplugin_image_push: image_push_init
	docker buildx build -t ${DOCKER_HUB_HOST}/${DOCKER_HUB_PROJECT}/deviceplugin:${RELEASE_VER} --platform=linux/arm64,linux/amd64 --provenance=false -f ./build/application/deviceplugin/dockerfile . --push
#	docker tag deviceplugin:${RELEASE_VER} ${DOCKER_HUB_HOST}/${DOCKER_HUB_PROJECT}/deviceplugin:${RELEASE_VER}
#	docker push ${DOCKER_HUB_HOST}/${DOCKER_HUB_PROJECT}/deviceplugin:${RELEASE_VER}

ifneq (${RELEASE_VER}, latest)
ifeq (${NEED_LATEST}, TRUE)
	docker tag deviceplugin:${RELEASE_VER} ${DOCKER_HUB_HOST}/${DOCKER_HUB_PROJECT}/deviceplugin:latest
	docker push ${DOCKER_HUB_HOST}/${DOCKER_HUB_PROJECT}/deviceplugin:latest
endif
endif

deviceexporter_image_push: image_push_init
	docker buildx build -t ${DOCKER_HUB_HOST}/${DOCKER_HUB_PROJECT}/deviceplugin:${RELEASE_VER} --platform=linux/arm64,linux/amd64 --provenance=false -f ./build/application/deviceexporter/dockerfile . --push
#	docker tag deviceexporter:${RELEASE_VER} ${DOCKER_HUB_HOST}/${DOCKER_HUB_PROJECT}/deviceexporter:${RELEASE_VER}
#	docker push ${DOCKER_HUB_HOST}/${DOCKER_HUB_PROJECT}/deviceexporter:${RELEASE_VER}

ifneq (${RELEASE_VER}, latest)
ifeq (${NEED_LATEST}, TRUE)
	docker tag deviceexporter:${RELEASE_VER} ${DOCKER_HUB_HOST}/${DOCKER_HUB_PROJECT}/deviceexporter:latest
	docker push ${DOCKER_HUB_HOST}/${DOCKER_HUB_PROJECT}/deviceexporter:latest
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
	cd ${CHARTS_GIT_DIR} && git config --global user.email ${CHARTS_GIT_USER_EMAIL} && git config --global user.name ${CHARTS_GIT_USER_NAME} && git add index.yaml krakenplug-${RELEASE_VER}.tgz && git commit -m "${RELEASE_VER}" && git push