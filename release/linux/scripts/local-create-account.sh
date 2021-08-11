#!/bin/bash

###########################################################################################################
################################################# VRIABLES #################################################
###########################################################################################################
SCRIPT_NAME="$(basename ${0})"
PROJECT_PATH=$(
    cd $(dirname $0)
    cd ../
    pwd
)
DATA_PATH=${PROJECT_PATH}/data
CONF_PATH=${PROJECT_PATH}/conf

NODE_ID=""
AUTO=""
ADMIN=""

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
USAGE: ${SCRIPT_NAME}  [options] [value]

        OPTIONS:

           --nodeid, -n                   the specified node name. must be specified

           --auto                     'true': will no prompt to create the node account
            　　　　　　　　　　　　　　　　　default='false'

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

    ## create keystore
    mkdir -p ${NODE_DIR}/keystore
    if [ ! -d "${NODE_DIR}/keystore" ]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')]: ********* CREATE KEYSTORE DIR FAILED *********"
        exit
    fi

    ## generate account
    if [[ "${AUTO}" == "true" ]]; then
        echo "[WARN] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : !!! An account will be created. The default password is 0 !!!"
    else
        echo "Please input account passphrase."
        read -p "passphrase: " PHRASE
    fi
    ret=$(curl --silent --write-out --output /dev/null -H "Content-Type: application/json" --data "{\"jsonrpc\":\"2.0\",\"method\":\"personal_newAccount\",\"params\":[\"${PHRASE}\"],\"id\":1}" http://${IP_ADDR}:${RPC_PORT}) >/dev/null 2>&1
    if [[ $? -ne 0 ]]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* CREATE ACCOUNT FAILED,  CHECK IF NODE HAS STARTED *********"
        exit
    fi
    substr=${ret##*\"result\":\"}
    if [ ${#substr} -gt 42 ]; then
        ACCOUNT="${substr:0:42}"
    else
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : *********  CREATE ACCOUNT FAILED *********"
        exit
    fi

    if [[ "${ADMIN}" == "true" ]]; then
        echo "${PHRASE}" >"${CONF_PATH}/keyfile.phrase"
        echo "${ACCOUNT}" >"${CONF_PATH}/keyfile.account"
        account=$(echo ${ACCOUNT} | grep "0x" | sed -e 's/0x\(.*\)/\1/g')

        cd ${NODE_DIR}/keystore
        for acc in $(ls ./); do
            if [[ $(echo ${acc} | grep ${account}) != "" ]]; then
                cp ${acc} ${CONF_PATH}/keyfile.json
            fi
            cd ${NODE_DIR}/keystore
        done

        if [ ! -f "${CONF_PATH}/keyfile.account" ] || [ ! -f "${CONF_PATH}/keyfile.json" ] || [ ! -f "${CONF_PATH}/keyfile.phrase" ]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] :  ********* CREATE ACCOUNT FAILED *********"
            exit
        fi
    fi
}

################################################# Main #################################################
function main() {
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ## Create account Start ##"
    file="${NODE_DIR}/deploy_node-${NODE_ID}.conf"
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
    --nodeid | -n)
        shiftOption2 $#
        NODE_DIR="${DATA_PATH}/node-$2"
        NODE_ID=$2

        if [ ! -f "${NODE_DIR}/deploy_node-$2.conf" ]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* FILE ${NODE_DIR}/deploy_node-$2.conf NOT FOUND **********"
            exit
        fi
        shift 2
        ;;
    --auto)
        AUTO="true"
        shift 1
        ;;
    --admin)
        ADMIN="true"
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
