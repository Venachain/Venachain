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
CREATE_KEYFILE=""
AUTO=""

NODE_DIR=""
DEPLOY_CONF=""
IP_ADDR=""
RPC_PORT=""
ACCOUNT=""
PHRASE=""


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

            --create_keyfile, -ck       will create keyfile

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
    checkConf "rpc_addr"
    if [[ $? -eq 1 ]]; then
        RPC_PORT="$(cat ${DEPLOY_CONF} | grep "rpc_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    fi
}

################################################# Unclock Account #################################################
function unlockAccount() {
    http_data="{\"jsonrpc\":\"2.0\",\"method\":\"personal_unlockAccount\",\"params\":[\"${ACCOUNT}\",\"${PHRASE}\"],\"id\":1}"
    res="$(curl -H "Content-Type: application/json" --data "${http_data}" http://${IP_ADDR}:${RPC_PORT})"
    echo "${res}"
    if [[ "$(echo ${res} | grep true )" == "" ]]; then
        printLog "error" "UNLOCK ACCOUNT FAILED"
        exit 1
    fi
    printLog "info" "Unlock account completed"
}

################################################# Create Account #################################################
function createAccount() {
    ## create keystore
    mkdir -p "${NODE_DIR}/keystore"
    if [ ! -d "${NODE_DIR}/keystore" ]; then
        printLog "error" "CREATE KEYSTORE DIR FAILED"
        exit 1
    fi

    ## generate account
    if [[ "${AUTO}" == "true" ]]; then
        printLog "warn" "An account will be created. The default password is 0"
        PHRASE="0"
    else
        printLog "question" "Please input account passphrase."
        read PHRASE
    fi
    ret="$(curl --silent --write-out --output /dev/null -H "Content-Type: application/json" --data "{\"jsonrpc\":\"2.0\",\"method\":\"personal_newAccount\",\"params\":[\"${PHRASE}\"],\"id\":1}" http://${IP_ADDR}:${RPC_PORT})" >/dev/null 2>&1
    if [[ $? -eq 1 ]]; then
        printLog "error" "CREATE ACCOUNT FAILED, CHECK IF NODE HAS STARTED"
        exit 1
    fi
    substr=${ret##*\"result\":\"}
    if [ ${#substr} -gt 42 ]; then
        ACCOUNT="${substr:0:42}"  
    else
        printLog "error" "CREATE ACCOUNT FAILED"
        exit 1
    fi
    unlockAccount


    ## generate keyfile
    if [[ "${CREATE_KEYFILE}" == "true" ]]; then  
        utc_file="$(ls -a ${NODE_DIR}/keystore | tail -n 1)"      
        echo "${PHRASE}" >"${CONF_PATH}/keyfile.phrase"
        echo "${ACCOUNT}" >"${CONF_PATH}/keyfile.account"
        cp "${NODE_DIR}/keystore/${utc_file}" "${CONF_PATH}/keyfile.json"

        if [ ! -f "${CONF_PATH}/keyfile.account" ] || [ ! -f "${CONF_PATH}/keyfile.json" ] || [ ! -f "${CONF_PATH}/keyfile.phrase" ]; then
            printLog "error" "CREATE ACCOUNT FAILED"
            exit 1
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
    exit 1
fi
while [ ! $# -eq 0 ]; do
    case "${1}" in
    --nodeid | -n)
        NODE_ID="${2}"
        shift 2
        ;;
    --create_keyfile | -ck)
        CREATE_KEYFILE="true"
        shift 1
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
