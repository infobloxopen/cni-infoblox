#!/bin/bash

DRIVER_NAME="infoblox"
SOCKET_DIR="/run/cni"
GRID_HOST="192.168.124.200"
WAPI_PORT="443"
WAPI_USERNAME=""
WAPI_PASSWORD=""
WAPI_VERSION="2.0"
SSL_VERIFY=false
NETWORK_VIEW="default"
NETWORK_CONTAINER="192.168.0.0/24,192.169.0.0/24"
PREFIX_LENGTH=25


rkt --insecure-options=image --volume run-cni,kind=host,source=/run/cni --mount volume=run-cni,target=/run/cni run docker://infoblox/infoblox-cni-daemon -- --grid-host=${GRID_HOST} --wapi-port=${WAPI_PORT} --wapi-username=${WAPI_USERNAME} --wapi-password=${WAPI_PASSWORD} --wapi-version=${WAPI_VERSION} --socket-dir=${SOCKET_DIR} --driver-name=${DRIVER_NAME} --ssl-verify=${SSL_VERIFY} --network-view=${NETWORK_VIEW} --network-container=${NETWORK_CONTAINER} --prefix-length=${PREFIX_LENGTH}
