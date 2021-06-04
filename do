#!/usr/bin/env bash

red='\033[1;31m'
green='\033[1;32m'
normal='\033[0m'

appName='impftermin-notifier'

goflags=""
if [[ "${READ_ONLY:-false}" == "true" ]]; then
    echo "Running in readonly mode"
    goflags="-mod=readonly"
fi

export CGO_ENABLED=0

linter_version=1.40.1

## test
function task_test {
    go test ./... -v -count=1 $goflags && echo -e "${green}TESTS SUCCEEDED${normal}" || (echo -e "${red}!!! TESTS FAILED !!!${normal}"; exit 1)
}

function update_linter {
    pushd /tmp > /dev/null # don't install dependencies of golangci-lint in current module
    GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint@v"${linter_version}" 2>&1
    popd >/dev/null
}

## lint
function task_lint {
    update_linter
    lint_path="${GOPATH:-${HOME}/go}/bin/golangci-lint"
    "${lint_path}" run 1>&2
}

## go-fmt
function task_go_fmt {
    go fmt ./...
}

## build <arch> : default=amd64 arm,arm64
function task_build {
  os="linux"
  arch=${1:-amd64}
  version=6
  GOOS=${os} go build -a ${goflags} -ldflags="-s -w" -o ${appName} cmd/main.go
}

## build-container: builds the container image
function task_build_container {
    docker build -t damek/${appName} .
}

## docker-save: save docker image to tar file
function task_docker_save {
    docker save damek/${appName}:latest -o ./${appName}.tar
}

function task_usage {
    echo "Usage: $0"
    sed -n 's/^##//p' <$0 | column -t -s ':' |  sed -E $'s/^/\t/'
}

CMD=${1:-}
shift || true
RESOLVED_COMMAND=$(echo "task_"$CMD | sed 's/-/_/g')
if [ "$(LC_ALL=C type -t $RESOLVED_COMMAND)" == "function" ]; then
    pushd $(dirname "${BASH_SOURCE[0]}") >/dev/null
    $RESOLVED_COMMAND "$@"
else
    task_usage
fi