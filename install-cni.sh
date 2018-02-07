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

# Polls for file change at /install/config/..data/00-infoblox-ipam.conf

LTIME=`stat -c %Z /install/config/..data/00-infoblox-ipam.conf`
while true
do
   ATIME=`stat -c %Z /install/config/..data/00-infoblox-ipam.conf`

   if [[ "$ATIME" != "$LTIME" ]]
   then
    cp -f /install/config/00-infoblox-ipam.conf /host/etc/cni/net.d/;
    LTIME=$ATIME
   fi
   sleep 5
done
