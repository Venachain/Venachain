#!/bin/bash

###########################################################################################################
################################################# VRIABLES #################################################
###########################################################################################################

## path
CURRENT_PATH="$(cd "$(dirname "$0")";pwd)"
PROJECT_PATH="$(
    cd $(dirname ${0})
    cd ../../
    pwd
)"
DATA_PATH="${PROJECT_PATH}/data"
CONF_PATH="${PROJECT_PATH}/conf"

## global
SCRIPT_NAME="$(basename ${0})"
SCRIPT_ALIAS="$(echo ${CURRENT_PATH}/${SCRIPT_NAME} | sed -e 's/\(.*\)\/scripts\/\(.*\).sh/\2/g')"
NODE_ID=""
AUTO=""

NODE_DIR=""
DEPLOY_CONF=""

#############################################################################################################
################################################# FUNCTIONS #################################################
#############################################################################################################

################################################# Help #################################################
function help() {
    echo
    echo "
USAGE: ${SCRIPT_NAME}  [options] [value]

        OPTIONS:

            --nodeid, -n                the node's name, must be specified

            --auto                      will read exist file

            --help, -h                  show help
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
        exit 1
    fi
}

################################################# Yes Or No #################################################
function yesOrNo() {
    read -p "" anw
    case "${anw}" in
    [Yy][Ee][Ss] | [yY])
        return 1
        ;;
    [Nn][Oo] | [Nn])
        return 0
        ;;
    esac
    return 0
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
    port_name="${1}"
    port="${2}"

    while [[ 0 -lt 1 ]]; do
        if [[ $(lsof -i:${port}) == "" ]]; then
            saveConf "${port_name}" "${port}"
            return
        else
            port=$(expr ${port} + 1)
        fi
    done
}

################################################# Check Env #################################################
function checkEnv() {
    NODE_DIR="${DATA_PATH}/node-${NODE_ID}" 
    DEPLOY_CONF="${NODE_DIR}/deploy_node-${NODE_ID}.conf"  

    if [[ "${NODE_ID}" == "" ]]; then
        printLog "error" "NODE NAME NOT SET"
        exit 1
    fi 

    if [ ! -f "${CONF_PATH}/deploy.conf.template" ]; then
        printLog "error" "FILE ${CONF_PATH}/deploy.conf.template NOT FOUND"
        exit 1
    fi
}

################################################# Generate Deploy Config #################################################
function generateDeployConfig() {
    if [ -f "${DEPLOY_CONF}" ]; then
        timestamp=$(date '+%Y%m%d%H%M%S')
        mkdir -p "${NODE_DIR}/bak"
        mv "${DEPLOY_CONF}" "${NODE_DIR}/bak/deploy_node-${NODE_ID}.conf.bak.${timestamp}"
        if [ -f "${NODE_DIR}/bak/deploy_node-${NODE_ID}.conf.bak.${timestamp}" ]; then
            printLog "info" "Backup ${NODE_DIR}/bak/deploy_node-${NODE_ID}.conf completed"
        else
            printLog "error" "BACKUP NODE DEPLOY CONF FAILED"
            exit 1
        fi
    fi

    cp "${CONF_PATH}/deploy.conf.template" "${NODE_DIR}/deploy_node-${NODE_ID}.conf"
    saveConf "deploy_path" "${PROJECT_PATH}"
    saveConf "user_name" "$(whoami)"

    ## setup ports
    setupPort "rpc_port" "6791"
    setupPort "p2p_port" "16791"
    setupPort "ws_port" "26791"
}

################################################# Setup Deploy Config #################################################
function setupDeployConfig() {
    ## set up default conf
    mkdir -p "${NODE_DIR}"

    if [[ "${AUTO}" == "true" ]]; then
        if [ ! -f "${DEPLOY_CONF}" ]; then
            generateDeployConfig
        else
            printLog "warn" "Deploy Conf Already Exists, Will Read It Automatically"
            exit 0
        fi
    else 
        printLog "question" "Do You What To Create a deploy conf ? Yes or No(y/n):"
        yesOrNo
        if [ $? -eq 1 ]; then
            if [ -f "${NODE_DIR}/deploy_node-${NODE_ID}.conf" ]; then
                printLog "question" "Deploy conf already exists, overwrite it? Yes or No(y/n):"
                yesOrNo
                if [ $? -ne 1 ]; then
                    exit 2
                fi
            fi
            generateDeployConfig
        else
            exit 2
        fi
    fi
}

################################################# Main #################################################
function main() {
    printLog "info" "## Setup node-${NODE_ID} deployconf Start ##"
    checkEnv

    setupDeployConfig
    printLog "success" "Setup node-${NODE_ID} deployconf succeeded"
}

###########################################################################################################
#################################################  EXECUTE #################################################
###########################################################################################################
if [ $# -eq 0 ]; then
    help
    exit 1
fi
while [ ! $# -eq 0 ]; do
    case "${1}" in
    --nodeid | -n)
        shiftOption2 $#
        NODE_ID="${2}"
        shift 2
        ;;
    --auto)
        AUTO="true"
        shift 1
        ;;
    --help | -h)
        help
        exit 1
        ;;
    *)
        printLog "error" "COMMAND \"${1}\" NOT FOUND"
        help
        exit 1
        ;;
    esac
done
main