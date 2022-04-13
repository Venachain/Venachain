#!/bin/bash

###########################################################################################################
################################################# VRIABLES #################################################
###########################################################################################################

## path
CURRENT_PATH="$(cd "$(dirname "$0")";pwd)"
PROJECT_PATH="$(
    cd $(dirname ${0})
    cd ../
    pwd
)"
BIN_PATH="${PROJECT_PATH}/bin"
DATA_PATH="${PROJECT_PATH}/data"
CONF_PATH="${PROJECT_PATH}/conf"
SCRIPT_PATH="${PROJECT_PATH}/scripts"

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
 
            --nodeid, -n                the specified node id (default: 0)

            --auto                      will use the default node password: 0
                                        to create the account and also to unlock the account

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
        exit 1
    fi
}

################################################# Check Conf #################################################
function checkConf() {
    ref="$(cat "${DEPLOY_CONF}" | grep "${1}=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    if [[ "${ref}" != "" ]]; then
        return 1
    fi
    return 0
}

################################################# Check Env #################################################
function checkEnv() {
    if [[ "${NODE_ID}" == "" ]]; then
        NODE_ID="0"
    fi
    NODE_DIR="${DATA_PATH}/node-${NODE_ID}"
    DEPLOY_CONF="${NODE_DIR}/deploy_node-${NODE_ID}.conf"

    if [ ! -f "${DEPLOY_CONF}" ]; then
        printLog "error" "FILE ${DEPLOY_CONF} NOT FOUND"
        exit 1
    fi

    checkConf "ip_addr"
    if [[ $? -ne 1 ]]; then
        printLog "error" "NODE'S IP NOT SET IN ${DEPLOY_CONF}"
        exit 1
    fi
    checkConf "rpc_port"
    if [[ $? -ne 1 ]]; then
        printLog "error" "NODE'S RPC PORT NOT SET IN ${DEPLOY_CONF}"
        exit 1
    fi
    checkConf "p2p_port"
    if [[ $? -ne 1 ]]; then
        printLog "error" "NODE'S P2P PORT NOT SET IN ${DEPLOY_CONF}"
        exit 1
    fi
}

################################################# Assign Default #################################################
function assignDefault() {
    IP_ADDR="127.0.0.1"
}

################################################# Read File #################################################
function readFile() {
    checkConf "ip_addr"
    if [[ $? -eq 1 ]]; then
        IP_ADDR="$(cat ${DEPLOY_CONF} | grep "ip_addr=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    fi
}

################################################# Deploy System Contract #################################################
function deploySystemContract() {
    flag_auto=""
    if [[ "${AUTO}" == "true" ]]; then
        flag_auto=" --auto "
    fi

    "${SCRIPT_PATH}"/venachainctl.sh createacc -n "${NODE_ID}" -ck ${flag_auto}
    if [[ $? -eq 1 ]]; then
        exit 1
    fi
    "${SCRIPT_PATH}"/venachainctl.sh addadmin -n "${NODE_ID}"
    if [[ $? -eq 1 ]]; then
        exit 1
    fi
    "${SCRIPT_PATH}"/venachainctl.sh addnode -n "${NODE_ID}"
    if [[ $? -eq 1 ]]; then
        exit 1
    fi
    "${SCRIPT_PATH}"/venachainctl.sh updatesys -n "${NODE_ID}"
    if [[ $? -eq 1 ]]; then
        exit 1
    fi
}

################################################# Main #################################################
function main() {
    checkEnv
    assignDefault
    readFile

    deploySystemContract
}

###########################################################################################################
#################################################  EXECUTE #################################################
###########################################################################################################
while [ ! $# -eq 0 ]; do
    case "${1}" in
    --nodeid | -n)
        shiftOption2 $#
        NODE_ID="${2}"
        shift 2
        ;;
    --auto)
        shiftOption2 $#
        AUTO="${2}"
        shift 2
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
