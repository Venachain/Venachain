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
BIN_PATH="${PROJECT_PATH}/bin"
DATA_PATH="${PROJECT_PATH}/data"
CONF_PATH="${PROJECT_PATH}/conf"

## global
SCRIPT_NAME="$(basename ${0})"
SCRIPT_ALIAS="$(echo ${CURRENT_PATH}/${SCRIPT_NAME} | sed -e 's/\(.*\)\/scripts\/\(.*\).sh/\2/g')"
NODE_ID=""

NODE_DIR=""
DEPLOY_CONF=""
KEYFILE=""
PHRASE=""

FIRSTNODE_INFO=""
FIRSTNODE_IP_ADDR=""
FIRSTNODE_RPC_PORT=""

#############################################################################################################
################################################# FUNCTIONS #################################################
#############################################################################################################

################################################# Help #################################################
function help() {
    echo
    echo "
USAGE: ${SCRIPTS_NAME}  [options] [value]

        OPTIONS:

            --nodeid, -n                the specified node id, must be specified

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
        return 1
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
    NODE_DIR="${DATA_PATH}/node-${NODE_ID}"
    DEPLOY_CONF="${NODE_DIR}/deploy_node-${NODE_ID}.conf"

    if [[ "${NODE_ID}" == "" ]]; then
        printLog "error" "NODE NAME NOT SET"
        exit 1
    fi 

    if [ ! -f "${CONF_PATH}/genesis.json" ]; then
        printLog "error" "FILE ${CONF_PATH}/genesis.json NOT FOUND"
        exit 1
    fi
    KEYFILE="${CONF_PATH}/keyfile.json"
    if [[ ! -f "${KEYFILE}" ]]; then
        printLog "error" "FILE ${KEYFILE} NOT FOUND"
        exit 1
    fi
    PHRASE="${CONF_PATH}/keyfile.phrase"
    if [[ ! -f "${PHRASE}" ]]; then
        printLog "error" "FILE ${PHRASE} NOT FOUND"
        exit 1
    fi
    FIRSTNODE_INFO="${CONF_PATH}/firstnode.info"
    if [[ ! -f "${FIRSTNODE_INFO}" ]]; then
        printLog "error" "FILE ${FIRSTNODE_INFO} NOT FOUND"
        exit 1
    fi
    if [ ! -f "${BIN_PATH}/vcl" ]; then
        printLog "error" "FILE ${BIN_PATH}/vcl NOT FOUND"
        exit 1
    fi
}

################################################# Read File #################################################
function readFile() {

    FIRSTNODE_IP_ADDR="$(cat ${FIRSTNODE_INFO} | grep "ip_addr=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    if [[ "${FIRSTNODE_IP_ADDR}" == "" ]]; then
        printLog "error" "FIRST NODE'S IP NOT SET IN ${FIRSTNODE_INFO}"
        exit 1
    fi
    FIRSTNODE_RPC_PORT="$(cat ${FIRSTNODE_INFO} | grep "rpc_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    if [[ "${FIRSTNODE_RPC_PORT}" == "" ]]; then
        printLog "error" "FIRST NODE'S RPC PORT NOT SET IN ${FIRSTNODE_INFO}"
        exit 1
    fi
}


################################################# Delete Node #################################################
function deleteNode() {
    res_delete_node="$(${BIN_PATH}/vcl node delete ${NODE_ID} --keyfile ${KEYFILE} --url ${FIRSTNODE_IP_ADDR}:${FIRSTNODE_RPC_PORT} <${PHRASE})"
    if [[ $(echo "${res_delete_node}" | grep "success") == "" ]]; then
        printLog "error" "DELETE NODE-${NODE_ID} FAILED, MAY BE IS DOWN"
        exit 1
    fi
    printLog "info" "Delete node-${NODE_ID} completed"
}

################################################# Main #################################################
function main() {
    checkEnv
    readFile

    deleteNode
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
