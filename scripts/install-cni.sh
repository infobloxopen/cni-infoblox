#! /bin/bash
set -e -x;

CONF_FILE_PATH="/install/config"
BIN_FILE_PATH="/install/bin"
CONF_FILE_PATH_HOST="/host/etc/cni/net.d/"
BIN_FILE_PATH_HOST="/host/opt/cni/bin/"
CONF_FILE="${CONF_FILE_PATH}/${CONF_FILE_NAME}"

# Validating CIDR from network configuration file
checksubnet(){
	#reads subnet from network configuration file
	subnt=`jq '.ipam.subnet' $CONF_FILE`
    temp="${subnt%\"}"
    temp="${temp#\"}"
	if [ -z ` egrep '^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])(\/([0-9]|[1-2][0-9]|3[0-2]))$' <<< $temp` ]; then
		echo "Invalid CIDR mentioned in the conf file ${CONF_FILE_NAME}"
		if [ $1 -eq 1 ]; then # exit the script
			exit 255
		fi
		return 1
	else
		return 0		
    fi
}

checksubnet 1

if [ -w "$BIN_FILE_PATH_HOST" ]; then
    cp -f $BIN_FILE_PATH/* $BIN_FILE_PATH_HOST;
    echo "Wrote CNI binaries to $BIN_FILE_PATH_HOST";
fi;

if [ -w "$CONF_FILE_PATH_HOST" ]; then
    cp -f $CONF_FILE_PATH/* $CONF_FILE_PATH_HOST;
    echo "Wrote netconf to $CONF_FILE_PATH_HOST";
fi;

# Polls for file change at $CONF_FILE_PATH/..data/$CONF_FILE_NAME
LAST_MODIFIED_TIME=`stat -c %Z $CONF_FILE_PATH/..data/$CONF_FILE_NAME`
while true
do
    
    CURRENT_MODIFIED_TIME=`stat -c %Z $CONF_FILE_PATH/..data/$CONF_FILE_NAME`

    if [[ "$CURRENT_MODIFIED_TIME" != "$LAST_MODIFIED_TIME" ]]
    then
	    
		checksubnet 0
		     
        if [ $? -eq 0 ]; then
			cp -f $CONF_FILE_PATH/$CONF_FILE_NAME $CONF_FILE_PATH_HOST;
	        LAST_MODIFIED_TIME=$CURRENT_MODIFIED_TIME
        fi
   fi
   sleep 5
done
