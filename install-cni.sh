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

# Polls for file change at /install/config/..data/$CONF_FILE_NAME
LTIME=`stat -c %Z /install/config/..data/$CONF_FILE_NAME`
while true
do
   ATIME=`stat -c %Z /install/config/..data/"$CONF_FILE_NAME"`

   if [[ "$ATIME" != "$LTIME" ]]
   then
    cp -f /install/config/$CONF_FILE_NAME /host/etc/cni/net.d/;
    LTIME=$ATIME
   fi
   sleep 5
done
