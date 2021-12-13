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
IP_ADDR=""
RPC_PORT=""
ACCOUNT=""
PHRASE=""
AUTO=""
CREATE_KEYFILE=""

#############################################################################################################
################################################# FUNCTIONS #################################################
#############################################################################################################

################################################# Help #################################################
function help() {
    echo
    echo "
USAGE: ${SCRIPT_NAME}  [options] [value]

        OPTIONS:

           --nodeid, -n                   the specified node name. must be specified

           --auto, -a                     create account with default phrase 0

           --create_keyfile, -ck          will create keyfile

           --help, -h                     show help
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
}

################################################# Assign Default #################################################
function assignDefault() {
    IP_ADDR=127.0.0.1
    RPC_PORT=6791
}

################################################# Read File #################################################
function readFile() {
    checkConf "ip_addr"
    if [[ $? -eq 1 ]]; then
        IP_ADDR=$(cat ${DEPLOY_CONF} | grep "ip_addr=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    fi
    checkConf "rpc_addr"
    if [[ $? -eq 1 ]]; then
        RPC_PORT=$(cat ${DEPLOY_CONF} | grep "rpc_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    fi
}

################################################# Create Account #################################################
function createAccount() {

    ## create keystore
    mkdir -p ${NODE_DIR}/keystore
    if [ ! -d "${NODE_DIR}/keystore" ]; then
        printLog "error" "CREATE KEYSTORE DIR FAILED"
        exit
    fi

    ## generate account
    if [[ "${AUTO}" == "true" ]]; then
        printLog "warn" "An account will be created. The default password is 0"
        PHRASE="0"
    else
        printLog "question" "Please input account passphrase."
        read PHRASE
    fi
    ret=$(curl --silent --write-out --output /dev/null -H "Content-Type: application/json" --data "{\"jsonrpc\":\"2.0\",\"method\":\"personal_newAccount\",\"params\":[\"${PHRASE}\"],\"id\":1}" http://${IP_ADDR}:${RPC_PORT}) >/dev/null 2>&1
    if [[ $? -ne 0 ]]; then
        printLog "error" "CREATE ACCOUNT FAILED, CHECK IF NODE HAS STARTED"
        exit
    fi
    substr=${ret##*\"result\":\"}
    if [ ${#substr} -gt 42 ]; then
        ACCOUNT="${substr:0:42}"  
    else
        printLog "error" "CREATE ACCOUNT FAILED"
        exit
    fi

    if [[ "${CREATE_KEYFILE}" == "true" ]]; then
        echo "${PHRASE}" >"${CONF_PATH}/keyfile.phrase"
        echo "${ACCOUNT}" >"${CONF_PATH}/keyfile.account"
        cp ${NODE_DIR}/keystore/UTC* ${CONF_PATH}/keyfile.json

        if [ ! -f "${CONF_PATH}/keyfile.account" ] || [ ! -f "${CONF_PATH}/keyfile.json" ] || [ ! -f "${CONF_PATH}/keyfile.phrase" ]; then
            printLog "error" "CREATE ACCOUNT FAILED"
            exit
        fi
    fi
    
}

################################################# Main #################################################
function main() {
    printLog "info" "## Create account Start ##"
    checkEnv
    assignDefault
    readFile 
    createAccount
    printLog "info" "Account: "
    echo "        New account: ${ACCOUNT}"
    echo "        Passphrase: ${PHRASE}"
    printLog "success" "Create account succeeded"
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
        NODE_ID="${2}"
        NODE_DIR="${DATA_PATH}/node-${NODE_ID}"
        DEPLOY_CONF="${NODE_DIR}/deploy_node-${NODE_ID}.conf"
        shift 2
        ;;
    --auto | -a)
        AUTO="true"
        shift 1
        ;;
    --create_keyfile | -ck)
        CREATE_KEYFILE="true"
        shift 1
        ;;
    *)
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* COMMAND \"$1\" NOT FOUND **********"
        help
        exit
        ;;
    esac
done
main
