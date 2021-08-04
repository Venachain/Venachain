#!/bin/bash

###########################################################################################################
################################################# VRIABLES #################################################
###########################################################################################################
PROJECT_PATH=$(
    cd $(dirname $0)
    cd ../
    pwd
)
DATA_PATH=${PROJECT_PATH}/data
CONF_PATH=${PROJECT_PATH}/conf

NODE_ID=""

DEPLOY_NODE_CONF_PATH=""
NODE_DIR=""
IP_ADDR=""
RPC_PORT=""
ACCOUNT=""
PHRASE="0"

#############################################################################################################
################################################# FUNCTIONS #################################################
#############################################################################################################

################################################# Help #################################################
function help() {
    echo
    echo "
USAGE: local-create-account.sh  [options] [value]

        OPTIONS:

           --node, -n                   the specified node name. must be specified

           --help, -h                   show help
"
}

################################################# Check Shift Option #################################################
function shiftOption2() {
    if [[ $1 -lt 2 ]]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* MISS OPTION VALUE! PLEASE SET THE VALUE **********"
        help
        exit
    fi
}

################################################# Read File #################################################
function readFile() {
    IP_ADDR=$(cat $1 | grep "ip_addr=" | sed -e 's/ip_addr=\(.*\)/\1/g')
    RPC_PORT=$(cat $1 | grep "rpc_port=" | sed -e 's/rpc_port=\(.*\)/\1/g')
    if [[ "${IP_ADDR}" == "" ]] || [[ "${RPC_PORT}" == "" ]]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* FILE ${file} MISS VALUE **********"
        exit
    fi
}

################################################# Create Account #################################################
function createAccount() {

    ## remove keyfile
    rm -rf ${NODE_DIR}/keystore
    rm -rf ${CONF_PATH}/keyfile.json
    rm -rf ${CONF_PATH}/keyfile.phrase
    rm -rf ${CONF_PATH}/keyfile.account
    if [ -d "${NODE_DIR}/keystore" ] || [ -f "${CONF_PATH}/keyfile.json" ] || [ -f "${CONF_PATH}/keyfile.phrase" ] || [ -f "${CONF_PATH}/keyfile.account" ]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* REMOVE KEYFILE FAILED *********"
        exit
    fi

    ## create keystore
    mkdir -p ${NODE_DIR}/keystore
    if [ ! -d "${NODE_DIR}/keystore" ]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')]: ********* CREATE KEYSTORE DIR FAILED *********"
        exit
    fi

    ## generate account
    echo "[WARN] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : !!! An account will be created. The default password is 0 !!!"
    ret=$(curl --silent --write-out --output /dev/null -H "Content-Type: application/json" --data "{\"jsonrpc\":\"2.0\",\"method\":\"personal_newAccount\",\"params\":[\"${PHRASE}\"],\"id\":1}" http://${IP_ADDR}:${RPC_PORT}) >/dev/null 2>&1
    if [[ $? -ne 0 ]]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* CREATE ACCOUNT FAILED *********"
        exit
    fi
    substr=${ret##*\"result\":\"}
    if [ ${#substr} -gt 42 ]; then
        ACCOUNT="${substr:0:42}"
    else
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : *********  CREATE ACCOUNT FAILED *********"
        exit
    fi

    ## generate key related files
    echo "${PHRASE}" >"${CONF_PATH}/keyfile.phrase"
    echo "${ACCOUNT}" >"${CONF_PATH}/keyfile.account"
    cp ${NODE_DIR}/keystore/UTC* ${CONF_PATH}/keyfile.json

    if [ ! -f "${CONF_PATH}/keyfile.account" ] || [ ! -f "${CONF_PATH}/keyfile.json" ] || [ ! -f "${CONF_PATH}/keyfile.phrase" ]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] :  ********* CREATE ACCOUNT FAILED *********"
        exit
    fi
}

################################################# Main #################################################
function main() {
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ## Create account Start ##"
    file="${DEPLOY_NODE_CONF_PATH}/deploy_node-${NODE_ID}.conf"
    readFile $file
    createAccount
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : Account: "
    echo "        New account: ${ACCOUNT}"
    echo "        Passphrase: ${PHRASE}"
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : Create account succeeded"

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
    --node | -n)
        NODE_DIR="${DATA_PATH}/node-$2"
        NODE_ID=$2
        DEPLOY_NODE_CONF_PATH="${DATA_PATH}/node-$2/deploy_conf"

        if [ ! -f "${DEPLOY_NODE_CONF_PATH}/deploy_node-$2.conf" ]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* FILE ${DEPLOY_NODE_CONF_PATH}/deploy_node-$2.conf NOT FOUND **********"
            exit
        fi
        ;;
    *)
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* COMMAND \"$1\" NOT FOUND **********"
        help
        exit
        ;;
    esac
    shiftOption2 $#
    shift 2
done
main
