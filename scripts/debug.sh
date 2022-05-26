#!/usr/bin/env bash
if [ -z "$GOPATH" ];then
    export GOPATH=$HOME/go
fi
export PATH=$PATH:$GOPATH/bin
ROOT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." >/dev/null 2>&1 && pwd )"

SCRIPT_DIR="${ROOT_DIR}/scripts"
CMD_DIR="${ROOT_DIR}/cmd"
REFLEX=$(which reflex)      # Enable Live reloading of service

MAIN="server/main.go"

CMD="${CMD_DIR}/${MAIN}"
OPTS="-config config/access.toml -type toml"

# Load env file
if [ ! -e "$ROOT_DIR/.env" ];then
    echo "$ROOT_DIR/.env does not exists"
    exit -1
fi



#VERSION=$($ROOT_DIR/version.sh)
#GIT_COMMIT=$(git rev-parse HEAD)
#GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
#GIT_REPO=$(git config --get remote.origin.url)
BUILD_TIMESTAMP=$(date '+%a %b %d %T %Z %Y')
# Build Params
BUILD_PARAMS="-X 'github.com/100mslive/packages/version.BuildTimestamp=${BUILD_TIMESTAMP}'"
BUILD_PARAMS="${BUILD_PARAMS} -X  'github.com/100mslive/packages/version.GitCommit=${GIT_COMMIT}'"
BUILD_PARAMS="${BUILD_PARAMS} -X 'github.com/100mslive/packages/version.GitBranch=${GIT_BRANCH}'"
BUILD_PARAMS="${BUILD_PARAMS} -X 'github.com/100mslive/packages/version.GitRepo=${GIT_REPO}'"
BUILD_PARAMS="${BUILD_PARAMS} -X 'github.com/100mslive/packages/version.VersionInfo=${VERSION}'"


#source $ROOT_DIR/.env
#export BRYTECAM_MONGO_HOST=${MONGO_HOST:-localhost}
#export BRYTECAM_MONGO_USER=${MONGO_USER:-brytecam}
#export BRYTECAM_MONGO_PASSWORD=${MONGO_PASSWORD:-VL1WPPzk42wu75V}
#export BRYTECAM_MONGO_DATABASE=${MONGO_DATABASE:-brytecam}
#export BRYTECAM_MONGO_SRV=${MONGO_SRV:-yes}
#export BRYTECAM_MONGO_OPTS=${MONGO_URL_OPTS}

AUTH=$(kubectl get pods -n ion2 | grep auth | awk '{print $1; exit}' | xargs)

kubectl port-forward "$AUTH" 8001:8001 -n ion2 &

if [ -n "${REFLEX}" ];then
    ${REFLEX} -v -s -r '\.go$$'  -- go run -ldflags="${BUILD_PARAMS}"   "${CMD}" ${OPTS}
else
    echo "Running without reflex package, install reflex for auto reload"
    go run -ldflags="${BUILD_PARAMS}" "${CMD}" ${OPTS}
fi
