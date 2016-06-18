#!/bin/bash

# Build ACI image for infoblox-daemon

acbuild begin

acbuild set-name infoblox.com/cni-infoblox-daemon

acbuild dependency add quay.io/quay/ubuntu:latest

acbuild mount add run-cni /run/cni

acbuild copy infoblox-daemon /usr/local/bin/infoblox-daemon

acbuild set-exec /usr/local/bin/infoblox-daemon

acbuild write --overwrite infoblox-cni-daemon.aci

acbuild end
