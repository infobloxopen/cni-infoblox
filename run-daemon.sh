#!/bin/bash

# rm /run/cni/infoblox.sock

DRIVER_NAME="infoblox"
PLUGIN_DIR="/run/cni"
# GRID_HOST="192.168.124.200"
GRID_HOST="infoblox.localdomain"
WAPI_PORT="443"
WAPI_USERNAME="cloudadmin"
WAPI_PASSWORD="cloudadmin"
WAPI_VERSION="2.0"
#SSL_VERIFY="./infoblox-localdomain.crt"
SSL_VERIFY=false
#SSL_VERIFY="./crap.crt"
GLOBAL_VIEW="default"
GLOBAL_CONTAINER="172.18.0.0/16"
GLOBAL_PREFIX=24
LOCAL_VIEW="yko_openstack"
LOCAL_CONTAINER="192.168.0.0/24,192.169.0.0/24"
LOCAL_PREFIX=25



./cni-infoblox --daemon=true --grid-host=${GRID_HOST} --wapi-port=${WAPI_PORT} --wapi-username=${WAPI_USERNAME} --wapi-password=${WAPI_PASSWORD} --wapi-version=${WAPI_VERSION} --plugin-dir=${PLUGIN_DIR} --driver-name=${DRIVER_NAME} --ssl-verify=${SSL_VERIFY}

