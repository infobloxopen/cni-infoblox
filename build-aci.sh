#!/bin/bash

# Build ACI image for infoblox-cni-daemon

acbuild begin

acbuild set-name infoblox.com/infoblox-cni-daemon

acbuild dependency add quay.io/quay/ubuntu:latest

acbuild mount add run-cni /run/cni

acbuild copy infoblox-cni-daemon /usr/local/bin/infoblox-cni-daemon

acbuild set-exec /usr/local/bin/infoblox-cni-daemon

acbuild write --overwrite infoblox-cni-daemon.aci

acbuild end
