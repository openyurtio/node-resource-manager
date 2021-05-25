# Copyright 2021 The OpenYurt Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#!/usr/bin/env bash
set -x

# usage:
#  ./build.sh ARCH=amd64 VERSION=v1.0

NRM_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)"
NRM_BUILD_DIR=${NRM_ROOT}/build
NRM_OUTPUT_DIR=${NRM_BUILD_DIR}/_output
NRM_BUILD_IMAGE="golang:1.13.3-alpine"

GIT_VERSION="v1.0"
GIT_VERSION=(${VERSION:-${GIT_VERSION}})
GIT_SHA=`git rev-parse --short HEAD || echo "HEAD"`
GIT_BRANCH=`git rev-parse --abbrev-ref HEAD 2>/dev/null`
BUILD_TIME=`date "+%Y-%m-%d-%H:%M:%S"`

IMG_REPO="openyurt/node-resource-manager"
IMG_VERSION=${GIT_VERSION}-${GIT_SHA}

readonly -a SUPPORTED_ARCH=(
    amd64
    arm
    arm64
)

readonly -a target_arch=(${ARCH:-${SUPPORTED_ARCH[@]}})

function build_multi_arch_binaries() {
    local docker_run_opts=(
        "-i"
        "--rm"
        "--network host"
        "-v ${NRM_ROOT}:/opt/src"
        "--env CGO_ENABLED=0"
        "--env GOOS=linux"
        "--env GIT_VERSION=${GIT_VERSION}"
        "--env GIT_SHA=${GIT_SHA}"
        "--env GIT_BRANCH=${GIT_BRANCH}"
        "--env BUILD_TIME=${BUILD_TIME}"
    )
    # use goproxy if build from inside mainland China
    [[ $region == "cn" ]] && docker_run_opts+=("--env GOPROXY=https://goproxy.cn")

    local docker_run_cmd=(
        "/bin/sh"
        "-xe"
        "-c"
    )

    local sub_commands="sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories; \
        apk --no-cache add bash; \
        cd /opt/src; "
    for arch in ${target_arch[@]}; do
        sub_commands+="CGO_ENABLED=0 GOOS=linux GOARCH='${arch}' go build \
  -ldflags '-X main._BRANCH_=${GIT_BRANCH} -X main._VERSION_=${IMG_VERSION} -X main._BUILDTIME_=${BUILD_TIME}' -o build/_output/nrm.${arch} main.go; "
    done

    docker run ${docker_run_opts[@]} ${NRM_BUILD_IMAGE} ${docker_run_cmd[@]} "${sub_commands}"
}

function build_images() {
    for arch in ${target_arch[@]}; do
        local docker_file_path=${NRM_BUILD_DIR}/Dockerfile.$arch
        local docker_build_path=${NRM_BUILD_DIR}
        local nrm_image=$IMG_REPO:${IMG_VERSION}.${arch}
        local base_image
        case $arch in
            amd64)
                base_image="amd64/alpine:3.10"
                ;;
            arm64)
                base_image="arm64v8/alpine:3.10"
                ;;
            arm)
                base_image="arm32v7/alpine:3.10"
                ;;
            *)
                echo unknown arch $arch
                exit 1
        esac
        cat << EOF > $docker_file_path
FROM ${base_image}
LABEL maintainers="OpenYurt Authors"
LABEL description="OpenYurt Node Resource Manager"

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk update && apk add --no-cache ca-certificates file util-linux lvm2 xfsprogs e2fsprogs blkid
COPY entrypoint.sh /entrypoint.sh
COPY _output/nrm.${arch} /bin/nrm
RUN chmod +x /bin/nrm && chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
EOF
        docker build --no-cache -t $nrm_image -f $docker_file_path $docker_build_path
        docker save $nrm_image > ${NRM_OUTPUT_DIR}/node-resource-manager-${arch}.tar
    done
}

rm -rf ${NRM_OUTPUT_DIR}
mkdir -p ${NRM_OUTPUT_DIR}
umask 0022
build_multi_arch_binaries
build_images
