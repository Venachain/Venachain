#!/bin/bash

###########################################################################################################
################################################# VRIABLES #################################################
###########################################################################################################

## path
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

## global
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
 
        --nodeid, -n                 the specified node id (default: 0)

        --auto, -a                   will use the default node password: 0
                                     to create the account and also to unlock the account

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
        exit
    fi
}

################################################# Check Conf #################################################
function checkConf() {
    ref=$(cat "${DEPLOY_CONF}" | grep "$1"= | sed -e 's/\(.*\)=\(.*\)/\2/g')
    if [[ "${ref}" != "" ]]; then
        return 1
    fi
    return 0
}

################################################# Check Env #################################################
function checkEnv() {
    if [ ! -f "${DEPLOY_CONF}" ]; then
        printLog "error" "FILE ${DEPLOY_CONF} NOT FOUND"
        exit
    fi

    checkConf "ip_addr"
    if [[ $? -ne 1 ]]; then
        printLog "error" "NODE'S IP HAVE NOT BEEN SET"
        exit
    fi
    checkConf "rpc_port"
    if [[ $? -ne 1 ]]; then
        printLog "error" "NODE'S RPC PORT HAVE NOT BEEN SET"
        exit
    fi
    checkConf "p2p_port"
    if [[ $? -ne 1 ]]; then
        printLog "error" "NODE'S P2P PORT HAVE NOT BEEN SET"
        exit
    fi
}

################################################# Assign Default #################################################
function assignDefault() {
    IP_ADDR=127.0.0.1
    RPC_PORT=6791
    P2P_PORT=16791
}

################################################# Read File #################################################
function readFile() {
    checkConf "ip_addr"
    if [[ $? -eq 1 ]]; then
        IP_ADDR=$(cat ${DEPLOY_CONF} | grep "ip_addr=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    fi
    checkConf "rpc_addr"
    if [[ $? -eq 1 ]]; then
        RPC_ADDR=$(cat ${DEPLOY_CONF} | grep "rpc_addr=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    fi
    checkConf "p2p_port"
    if [[ $? -eq 1 ]]; then
        P2P_PORT=$(cat ${DEPLOY_CONF} | grep "p2p_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    fi
}


################################################# Main #################################################
function main() {
    if [[ "${NODE_ID}" == "" ]]; then
        NODE_ID="0"
        NODE_DIR="${DATA_PATH}/node-${NODE_ID}"
        DEPLOY_CONF="${NODE_DIR}/deploy_node-${NODE_ID}.conf"
    fi

    checkEnv
    assignDefault
    readFile

    flag_auto=""
    if [[ "${AUTO}" == "true" ]]; then
        flag_auto=" --auto "
    fi

    ./local/create-account.sh -n ${NODE_ID} -ck ${flag_auto}
    ./local/add-admin-role.sh -n ${NODE_ID} 
    ./add-node.sh -n ${NODE_ID}
    ./update_to_consensus_node.sh -n ${NODE_ID}
}

###########################################################################################################
#################################################  EXECUTE #################################################
###########################################################################################################
while [ ! $# -eq 0 ]; do
    case "$1" in
    --nodeid | -n)
        shiftOption2 $#
        NODE_ID="${2}"
        NODE_DIR="${DATA_PATH}/node-${NODE_ID}"
        DEPLOY_CONF="${NODE_DIR}/deploy_node-${NODE_ID}.conf"
        shift 2
        ;;
    --auto | -a)
        shiftOption2 $#
        AUTO=$2
        shift 2
        ;;
    --help | -h)
        help
        exit
        ;;
    *)
        printLog "error" "COMMAND \"$1\" NOT FOUND"
        help
        exit
        ;;
    esac
done
main
