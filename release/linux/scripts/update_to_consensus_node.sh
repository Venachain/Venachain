#!/bin/bash

###########################################################################################################
################################################# VRIABLES #################################################
###########################################################################################################
SCRIPT_NAME="$(basename ${0})"
SCRIPT_ALIAS="$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')"
PROJECT_PATH=$(
    cd $(dirname $0)
    cd ../
    pwd
)
BIN_PATH=${PROJECT_PATH}/bin
DATA_PATH=${PROJECT_PATH}/data
CONF_PATH=${PROJECT_PATH}/conf

NODE_ID=""
NODE_DIR=""
DEPLOY_CONF=""
PUBKEY_FILE=""
KEYFILE=""
PHRASE=""
FIRSTNODE_INFO=""
FIRSTNODE_IP_ADDR=""
FIRSTNODE_RPC_PORT=""

NODE_TYPE=""

#############################################################################################################
################################################# FUNCTIONS #################################################
#############################################################################################################

################################################# Help #################################################
function help() {
    echo
    echo "
USAGE: ${SCRIPT_NAME}  [options] [value]

        OPTIONS:

           --nodeid, -n                 the specified node name. must be specified

           --content, -c                update content (default: consensus)

           --help, -h                   show help
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

################################################# Check Env #################################################
function checkEnv() {
    if [ ! -f ${NODE_DIR}/deploy_node-${NODE_ID}.conf ]; then
        printLog "error" "FILE ${NODE_DIR}/deploy_node-${NODE_ID}.conf NOT FOUND"
        exit
    fi

    PUBKEY_FILE="${NODE_DIR}/node.pubkey"
    if [[ ! -f "${PUBKEY_FILE}" ]]; then
        printLog "error" "FILE ${PUBKEY_FILE} NOT FOUND"
        exit
    fi
    KEYFILE="${CONF_PATH}/keyfile.json"
    if [[ ! -f "${KEYFILE}" ]]; then
        printLog "error" "FILE ${KEYFILE} NOT FOUND"
        exit
    fi
    PHRASE="${CONF_PATH}/keyfile.phrase"
    if [[ ! -f "${PHRASE}" ]]; then
        printLog "error" "FILE ${PHRASE} NOT FOUND"
        exit
    fi
    FIRSTNODE_INFO="${CONF_PATH}/firstnode.info"
    if [[ ! -f "${FIRSTNODE_INFO}" ]]; then
        printLog "error" "FILE ${firstnode_info} NOT FOUND"
        exit
    fi
}

################################################# Assign Default #################################################
function assignDefault() { 
    if [[ "${NODE_TYPE}" == "" ]]; then
        NODE_TYPE="consensus"
    fi
}

################################################# Read File #################################################
function readFile() {
    FIRSTNODE_IP_ADDR=$(cat ${FIRSTNODE_INFO} | grep "ip_addr=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    if [[ "${FIRSTNODE_IP_ADDR}" == "" ]]; then
        printLog "error" "FIRST NODE'S IP NOT FOUND"
        exit
    fi
    FIRSTNODE_RPC_PORT=$(cat ${FIRSTNODE_INFO} | grep "rpc_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    if [[ "${FIRSTNODE_RPC_PORT}" == "" ]]; then
        printLog "error" "FIRST NODE'S RPC PORT NOT FOUND"
        exit
    fi
}

################################################# Update To Consensus Node #################################################
function updateToConsensusNode() {
    res_update_node=""

    ${BIN_PATH}/venachaincli node update "${NODE_ID}" --type "${NODE_TYPE}" --keyfile "${KEYFILE}" --url "${FIRSTNODE_IP_ADDR}:${FIRSTNODE_RPC_PORT}" <"${PHRASE}"
    timer=0
    while [ ${timer} -lt 10 ]; do
        res_update_node=$(${BIN_PATH}/venachaincli node query --type ${NODE_TYPE} --name ${NODE_ID} --url "${FIRSTNODE_IP_ADDR}:${FIRSTNODE_RPC_PORT}") >/dev/null 2>&1
        if [[ $(echo ${res_update_node} | grep "success") != "" ]]; then
            break
        fi
        sleep 1
        let timer++
    done
    if [[ $(echo ${res_update_node} | grep "success") == "" ]]; then
        printLog "error" "UPDATE NODE-${NODE_ID} FAILED"
        exit
    fi
}

################################################# Main #################################################
function main() {
    printLog "info" "## Update Node-${NODE_ID} Start ##"
    checkEnv
    assignDefault
    readFile
    updateToConsensusNode
    printLog "success" "Update Node-${NODE_ID} succeeded"
}

###########################################################################################################
#################################################  EXECUTE #################################################
###########################################################################################################
if [ $# -eq 0 ]; then
    help
    exit
fi
while [ ! $# -eq 0 ]; do
    case "$1" in
    --nodeid | -n)
        shiftOption2 $#
        NODE_ID="${2}"
        NODE_DIR="${DATA_PATH}/node-${NODE_ID}"
        DEPLOY_CONF="${NODE_DIR}/deploy_node-${NODE_ID}.conf"
        shift 2
        ;;
    --content | -c)
        NODE_TYPE=${2}
        if [[ "${NODE_TYPE}" != "consensus" ]] && [[ "${NODE_TYPE}" != "observer" ]]; then
            printLog "error" "NODE TYPE ${NODE_TYPE} NOT FOUND"
        fi
        shift 2
        ;;
    *)
        printLog "error" "COMMAND \"$1\" NOT FOUND"
        help
        exit
        ;;
    esac
done
main
