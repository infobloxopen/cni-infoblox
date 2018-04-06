#! /bin/bash
set -e -x;

checksubnet(){
	subnt=`jq '.ipam.subnet' /install/config/$CONF_FILE_NAME`
    temp="${subnt%\"}"
    temp="${temp#\"}"
	if [ -z ` egrep '^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])(\/([0-9]|[1-2][0-9]|3[0-2]))$' <<< $temp` ]; then
		echo "invalid CIDR mention in ${CONF_FILE_NAME}"
		if [ $1 -eq 1 ]; then # exit the script
			exit 255
		fi
		return 1
	else
		return 0		
    fi
}

checksubnet 1

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
    
   ATIME=`stat -c %Z /install/config/..data/$CONF_FILE_NAME`

   if [[ "$ATIME" != "$LTIME" ]]
   then       
	    checksubnet 0        
        if [ $? -eq 0 ]; then
			cp -f /install/config/$CONF_FILE_NAME /host/etc/cni/net.d/;
	        LTIME=$ATIME
        fi
   fi
   sleep 5
done
