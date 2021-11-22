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
BIN_PATH=${PROJECT_PATH}/bin
DATA_PATH=${PROJECT_PATH}/data
CONF_PATH=${PROJECT_PATH}/conf

NODE_ID=""
AUTO=""

NODE_DIR=""
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

           --nodeid, -n                   the specified node name. must be specified

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

    PHRASE=$(cat ${CONF_PATH}/keyfile.phrase)
    ACCOUNT=$(cat ${CONF_PATH}/keyfile.account)
    if [[ "${PHRASE}" == "" ]] || [[ "${ACCOUNT}" == "" ]]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* READ KEYFILE FAILED **********"
        exit
    fi
}

################################################# Unclock Account #################################################
function unlockAccount() {
    http_data="{\"jsonrpc\":\"2.0\",\"method\":\"personal_unlockAccount\",\"params\":[\"${ACCOUNT}\",\"${PHRASE}/\",0],\"id\":1}"
    curl -H "Content-Type: application/json" --data "${http_data}" "http://${IP_ADDR}:${RPC_PORT}" >/dev/null 2>&1
    if [[ $? -ne 0 ]]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : UNLOCK ACCOUNT FAILED!!!"
        exit
    fi
}

################################################# Set Super Admin #################################################
function setSuperAdmin() {
    ${BIN_PATH}/platonecli role setSuperAdmin --keyfile ${CONF_PATH}/keyfile.json --url ${IP_ADDR}:${RPC_PORT} <${CONF_PATH}/keyfile.phrase 1>/dev/null
    timer=0
    super_admin_flag=""
    while [ ${timer} -lt 10 ]; do
        super_admin_flag=$(${BIN_PATH}/platonecli role hasRole ${ACCOUNT} SUPER_ADMIN --keyfile ${CONF_PATH}/keyfile.json --url ${IP_ADDR}:${RPC_PORT} <${CONF_PATH}/keyfile.phrase)
        if [[ $(echo ${super_admin_flag} | grep "int32=1") != "" ]]; then
            break
        fi
        sleep 1
        let timer++
    done
    if [[ $(echo ${super_admin_flag} | grep "int32=1") == "" ]]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* SET SUPER ADMIN FAILED **********"
        exit
    else
        echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : Set super admin completed"
    fi

}

################################################# Set Chain Admin #################################################
function addChainAdmin() {
    ${BIN_PATH}/platonecli role addChainAdmin ${ACCOUNT} --keyfile ${CONF_PATH}/keyfile.json --url ${IP_ADDR}:${RPC_PORT} <${CONF_PATH}/keyfile.phrase 1>/dev/null
    timer=0
    chain_admin_flag=""
    while [ ${timer} -lt 10 ]; do
        chain_admin_flag=$(${BIN_PATH}/platonecli role hasRole ${ACCOUNT} CHAIN_ADMIN --keyfile ${CONF_PATH}/keyfile.json --url ${IP_ADDR}:${RPC_PORT} <${CONF_PATH}/keyfile.phrase)
        if [[ $(echo ${chain_admin_flag} | grep "int32=1") != "" ]]; then
            break
        fi
        sleep 1
        let timer++
    done
    if [[ $(echo ${chain_admin_flag} | grep "int32=1") == "" ]]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* SET CHAIN ADMIN FAILED **********"
        exit
    else
        echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : Set chain admin completed"
    fi
}

################################################# Main #################################################
function main() {
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ## Deploy System Contract Start ##"
    file="${NODE_DIR}/deploy_node-${NODE_ID}.conf"
    readFile $file
    setSuperAdmin
    addChainAdmin
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : Deploy system contract succeeded"
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
        NODE_ID=$2
        NODE_DIR="${DATA_PATH}/node-$2"

        if [ ! -f "${NODE_DIR}/deploy_node-$2.conf" ]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* ${NODE_DIR}/deploy_node-$2.conf NOT FOUND **********"
            exit
        fi
        shift 2
        ;;
    --auto)
        AUTO="true"
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
