#!/bin/bash

acbuild begin

acbuild set-name infoblox.com/cni-infoblox-daemon

acbuild dependency add quay.io/quay/ubuntu:latest

acbuild mount add run-cni /run/cni

acbuild mount add rkt-netd /etc/rkt/net.d

acbuild copy infoblox-daemon /usr/local/bin/infoblox-daemon

acbuild set-exec /usr/local/bin/infoblox-daemon

acbuild write --overwrite infoblox-daemon-latest-linux-amd64.aci

acbuild end
