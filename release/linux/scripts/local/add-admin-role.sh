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

DEPLOY_PATH=""
NODE_DIR=""
DEPLOY_CONF=""
IP_ADDR=""
RPC_PORT=""
PHRASE=""
ACCOUNT=""

#############################################################################################################
################################################# FUNCTIONS #################################################
#############################################################################################################

################################################# Help #################################################
function help() {
    echo
    echo "
USAGE: ${SCRIPT_NAME}  [options] [value]

        OPTIONS:

            --nodeid, -n                the specified node name, must be specified

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

    if [ ! -f "${DEPLOY_CONF}" ]; then
        printLog "error" "FILE ${DEPLOY_CONF} NOT FOUND"
        exit 1
    fi
    if [ ! -f "${CONF_PATH}/genesis.json" ]; then
        printLog "error" "FILE ${CONF_PATH}/genesis.json NOT FOUND"
        exit 1
    fi
    if [ ! -f "${CONF_PATH}/keyfile.json" ]; then
        printLog "error" "FILE ${CONF_PATH}/keyfile.json NOT FOUND"
        exit
    fi
    if [ ! -f "${CONF_PATH}/keyfile.phrase" ]; then
        printLog "error" "FILE ${CONF_PATH}/keyfile.phrase NOT FOUND"
        exit
    fi
    if [ ! -f "${CONF_PATH}/keyfile.account" ]; then
        printLog "error" "FILE ${CONF_PATH}/keyfile.account NOT FOUND"
        exit
    fi
    if [ ! -f "${BIN_PATH}/vcl" ]; then
        printLog "error" "FILE ${BIN_PATH}/vcl NOT FOUND"
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
}

################################################# Assign Default #################################################
function assignDefault() {
    IP_ADDR="127.0.0.1"
    RPC_PORT="6791"
}

################################################# Read File #################################################
function readFile() {
    checkConf "ip_addr"
    if [[ $? -eq 1 ]]; then
        IP_ADDR="$(cat ${DEPLOY_CONF} | grep "ip_addr=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    fi
    checkConf "rpc_port"
    if [[ $? -eq 1 ]]; then
        RPC_PORT="$(cat ${DEPLOY_CONF} | grep "rpc_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    fi
    checkConf "deploy_path"
    if [[ $? -eq 1 ]]; then
        DEPLOY_PATH="$(cat ${DEPLOY_CONF} | grep "deploy_path=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    fi

    PHRASE="$(cat ${CONF_PATH}/keyfile.phrase)"
    if [[ "${PHRASE}" == "" ]]; then
        printLog "error" "FILE ${CONF_PATH}/keyfile.phrase IS EMPTY"
        exit 1
    fi
    ACCOUNT="$(cat ${CONF_PATH}/keyfile.account)"
    if [[ "${ACCOUNT}" == "" ]]; then
        printLog "error" "FILE ${CONF_PATH}/keyfile.account IS EMPTY"
        exit 1
    fi
}

################################################# Set Super Admin #################################################
function setSuperAdmin() {
    res_super_admin="$(${BIN_PATH}/vcl role hasRole ${ACCOUNT} SUPER_ADMIN --keyfile ${CONF_PATH}/keyfile.json --url ${IP_ADDR}:${RPC_PORT} <${CONF_PATH}/keyfile.phrase)"
    if [[ "$(echo ${res_super_admin} | grep "int32=1")" != "" ]]; then
        printLog "warn" "Node-${NODE_ID} Has Already Been Super Admin"
        return 0
    fi
    
    "${BIN_PATH}/vcl" role setSuperAdmin --keyfile "${CONF_PATH}/keyfile.json" --url "${IP_ADDR}:${RPC_PORT}" <"${CONF_PATH}/keyfile.phrase"
    timer=0
    res_super_admin=""
    while [ ${timer} -lt 10 ]; do
        res_super_admin="$(${BIN_PATH}/vcl role hasRole ${ACCOUNT} SUPER_ADMIN --keyfile ${CONF_PATH}/keyfile.json --url ${IP_ADDR}:${RPC_PORT} <${CONF_PATH}/keyfile.phrase)"
        if [[ "$(echo ${res_super_admin} | grep "int32=1")" != "" ]]; then
            break
        fi
        sleep 1
        let timer++
    done
    if [[ "$(echo ${res_super_admin} | grep "int32=1")" == "" ]]; then
        printLog "error" "SET SUPER ADMIN FAILED"
        exit 1
    else
        printLog "info" "Set super admin completed"
    fi

}

################################################# Set Chain Admin #################################################
function addChainAdmin() {
    res_chain_admin="$(${BIN_PATH}/vcl role hasRole ${ACCOUNT} CHAIN_ADMIN --keyfile ${CONF_PATH}/keyfile.json --url ${IP_ADDR}:${RPC_PORT} <${CONF_PATH}/keyfile.phrase)"
    if [[ $(echo ${res_chain_admin} | grep "int32=1") != "" ]]; then
        printLog "warn" "Node-${NODE_ID} Has Already Been Chain Admin"
        return 0
    fi
    
    "${BIN_PATH}/vcl" role addChainAdmin "${ACCOUNT}" --keyfile "${CONF_PATH}/keyfile.json" --url "${IP_ADDR}:${RPC_PORT}" <"${CONF_PATH}/keyfile.phrase"
    timer=0
    res_chain_admin=""
    while [ ${timer} -lt 10 ]; do
        res_chain_admin="$(${BIN_PATH}/vcl role hasRole ${ACCOUNT} CHAIN_ADMIN --keyfile ${CONF_PATH}/keyfile.json --url ${IP_ADDR}:${RPC_PORT} <${CONF_PATH}/keyfile.phrase)"
        if [[ "$(echo ${res_chain_admin} | grep "int32=1")" != "" ]]; then
            break
        fi
        sleep 1
        let timer++
    done
    if [[ "$(echo ${res_chain_admin} | grep "int32=1")" == "" ]]; then
        printLog "error" "SET CHAIN ADMIN FAILED"
        exit 1
    else
        printLog "info" "Set chain admin completed"
    fi
}

################################################# Save Firstnode Info #################################################
function saveFirstnodeInfo() {
    {
        echo "user_name=$(whoami)"
        echo "node_id=${NODE_ID}"
        echo "deploy_path=${DEPLOY_PATH}"
        echo "ip_addr=${IP_ADDR}"
        echo "rpc_port=${RPC_PORT}"
    } >"${CONF_PATH}/firstnode.info"
    if [[ ! -f "${CONF_PATH}/firstnode.info" ]] || [[ "$(cat ${CONF_PATH}/firstnode.info)" == "" ]]; then
        printLog "error" "SETUP FIRSTNODE INFO FAILED"
        exit 1
    fi
    printLog "info" "File: ${CONF_PATH}/firstnode.info"
    printLog "info" "Firstnode: "
    cat "${CONF_PATH}/firstnode.info"
    printLog "info" "Setup firstnode info completed"
}

################################################# Main #################################################
function main() {
    printLog "info" "## Add Admin Role Start ##"
    checkEnv
    assignDefault
    readFile
    
    setSuperAdmin
    addChainAdmin
    saveFirstnodeInfo
    printLog "success" "Add admin role succeeded"
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
