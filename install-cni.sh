#!/bin/sh

set -e -x;

if [ -w "/host/opt/cni/bin/" ]; then
    cp -f /install/bin/* /host/opt/cni/bin/;
    echo "Wrote CNI binaries to /host/opt/cni/bin/";
fi;

if [ -w "/host/etc/cni/net.d/" ]; then
    cp -f /install/config/* /host/etc/cni/net.d/;
    echo "Wrote netconf to /host/etc/cni/net.d/";
fi;

while :; do sleep 3600; done;