#! /bin/bash
set -e -x;

CONF_FILE_CONTAINER_PATH="/install/config"
CONF_FILE_HOST_PATH="/host/etc/cni/net.d/"
BIN_FILE_CONTAINER_PATH="/install/bin"
BIN_FILE_HOST_PATH="/host/opt/cni/bin/"
CONF_FILE="${CONF_FILE_CONTAINER_PATH}/..data/${CONF_FILE_NAME}"

# Validating CIDR from network configuration file
checksubnet() {
    SUBNET=$(cat ${CONF_FILE} | jq 'select(.ipam.subnet != null) | .ipam.subnet' | tr -d \")
    # If 'subnet' key it self not provided in network conf
    if [ -z "${SUBNET}" ] ; then
        return 0
    fi
    if [ -z $(egrep '^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])(\/([0-9]|[1-2][0-9]|3[0-2]))$' <<< $SUBNET) ]; then
        echo "Invalid CIDR mentioned in the conf file ${CONF_FILE_NAME}"
  	    if [ $1 -eq 1 ]; then # exit the script
  	        exit 255
  	    fi
        return 1
    fi
        return 0
}

checksubnet 1

if [ -w "${BIN_FILE_HOST_PATH}" ]; then
    cp -f ${BIN_FILE_CONTAINER_PATH}/* ${BIN_FILE_HOST_PATH};
    echo "Wrote Infoblox CNI binaries to ${BIN_FILE_HOST_PATH}";
fi;

if [ -w "${CONF_FILE_HOST_PATH}" ]; then
    cp -f ${CONF_FILE_CONTAINER_PATH}/* ${CONF_FILE_HOST_PATH};
    echo "Wrote network conf to ${CONF_FILE_HOST_PATH}";
fi;

# Polls for file change at ${CONF_FILE}
LAST_MODIFIED_TIME=$(stat -c %Z ${CONF_FILE})
while true
do    
    CURRENT_MODIFIED_TIME=$(stat -c %Z ${CONF_FILE})
    if [[ "$CURRENT_MODIFIED_TIME" != "$LAST_MODIFIED_TIME" ]]; then	    
        checksubnet 0
        if [ $? -eq 0 ]; then
            echo "Network conf is modified so changing..."
            cp -f ${CONF_FILE_CONTAINER_PATH}/${CONF_FILE_NAME} ${CONF_FILE_HOST_PATH};
            LAST_MODIFIED_TIME=$CURRENT_MODIFIED_TIME
        fi
    fi
    sleep 30
done
