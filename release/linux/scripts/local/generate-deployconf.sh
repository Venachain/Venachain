#!/bin/bash

###########################################################################################################
################################################# VRIABLES #################################################
###########################################################################################################
SCRIPT_NAME="$(basename ${0})"
SCRIPT_ALIAS="$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')"
PROJECT_PATH=$(
    cd $(dirname $0)
    cd ../../
    pwd
)
DATA_PATH=${PROJECT_PATH}/data
CONF_PATH=${PROJECT_PATH}/conf

NODE_ID=""
NODE_DIR=""
DEPLOY_CONF=""
AUTO=""

#############################################################################################################
################################################# FUNCTIONS #################################################
#############################################################################################################

################################################# Help #################################################
function help() {
    echo
    echo "
USAGE: ${SCRIPT_NAME}  [options] [value]

        OPTIONS:

           --nodeid, -n             the node's name, must be specified

           --auto, -a               will read exist file

           --help, -h               show help

"
}

################################################# Print Log #################################################
function printLog() {
    if [[ "${1}" == "error" ]]; then
        echo -e "\033[31m[ERROR] [${SCRIPT_ALIAS}] ${2}\033[0m"
    elif [[ "${1}" == "warn" ]]; then
        echo -e "\033[33m[WARN] [${SCRIPT_ALIAS}] ${2}\033[0m"
    elif [[ "${1}" == "success" ]]; then
        echo -e "\033[32m[SUCCESS] [${SCRIPT_ALIAS}] ${2}\033[0m"
    elif [[ "${1}" == "question" ]]; then
        echo -e "\033[36m[${SCRIPT_ALIAS}] ${2}\033[0m"
    else
        echo "[INFO] [${SCRIPT_ALIAS}] ${2}"
    fi
}

################################################# Check Shift Option #################################################
function shiftOption2() {
    if [[ $1 -lt 2 ]]; then
        printLog "error" "MISS OPTION VALUE! PLEASE SET THE VALUE"
        help
        exit
    fi
}

################################################# Save Config #################################################
function saveConf() {
    conf="${NODE_DIR}/deploy_node-${NODE_ID}.conf"
    conf_tmp="${NODE_DIR}/deploy_node-${NODE_ID}.tmp.conf"
    if [[ "${2}" == "" ]]; then
        return
    fi
    cat "${conf}" | sed "s#${1}=.*#${1}=${2}#g" | cat >"${conf_tmp}"
    mv "${conf_tmp}" "${conf}"
}

################################################# Setup Port #################################################
function setupPort() {
    port_name=${1}
    port=${2}

    while [[ 0 -lt 1 ]]; do
        if [[ $(lsof -i:${port}) == "" ]]; then
            saveConf "${port_name}" "${port}"
            return
        else
            port=$(expr ${port} + 1)
        fi
    done
}

################################################# Setup Deploy Config #################################################
function setupDeployConfig() {
    ## set up default conf
    mkdir -p ${NODE_DIR}

    if [[ -f "${DEPLOY_CONF}" ]]; then
        timestamp=$(date '+%Y%m%d%H%M%S')
        mkdir -p "${NODE_DIR}/bak"
        mv "${DEPLOY_CONF}" "${NODE_DIR}/bak/deploy_conf-${NODE_ID}.conf.bak.${timestamp}"
        if [ -f "${NODE_DIR}/bak/deploy_conf-${NODE_ID}.conf.bak.${timestamp}" ]; then
            printLog "info" "Backup ${NODE_DIR}/bak/deploy_conf-${NODE_ID}.conf completed"
        else
            printLog "error" "BACKUP NODE DEPLOY CONF FAILED"
            exit
        fi
    fi

    cp ${CONF_PATH}/deploy.conf.template ${NODE_DIR}/deploy_node-${NODE_ID}.conf
    saveConf "deploy_path" "${PROJECT_PATH}"
    saveConf "user_name" "$(whoami)"

    ## setup ports
    setupPort "rpc_port" "6791"
    setupPort "p2p_port" "16791"
    setupPort "ws_port" "26791"
}

################################################# Main #################################################
function main() {
    printLog "info" "## Node-${NODE_ID} setup deployconf Start ##"
    if [[ "${AUTO}" == "true" ]] && [[ -f "${DEPLOY_CONF}" ]]; then
        printLog "warn" "Deploy Conf Already Exists, Will Read It Automatically"
        exit
    fi
    setupDeployConfig
    printLog "success" "Node-${NODE_ID} setup deployconf succeeded"
}

###########################################################################################################
#################################################  EXECUTE #################################################
###########################################################################################################
if [ $# -eq 0 ]; then
    help
    exit
fi
while [ ! $# -eq 0 ]; do
    case ${1} in
    --nodeid | -n)
        shiftOption2 $#
        NODE_ID=${2}
        NODE_DIR="${DATA_PATH}/node-${NODE_ID}" 
        DEPLOY_CONF="${NODE_DIR}/deploy_node-${NODE_ID}.conf"   
        shift 2
        ;;
    --help | -h)
        help
        exit
        ;;
    --auto | -a)
        AUTO="true"
        shift 1
        ;;
    *)
        printLog "error" "COMMAND \"${1}\" NOT FOUND"
        help
        exit
        ;;
    esac
done
main