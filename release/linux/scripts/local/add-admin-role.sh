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
BIN_PATH=${PROJECT_PATH}/bin
DATA_PATH=${PROJECT_PATH}/data
CONF_PATH=${PROJECT_PATH}/conf

NODE_ID=""
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

           --nodeid, -n                   the specified node name. must be specified

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
    checkConf "rpc_port"
    if [[ $? -eq 1 ]]; then
        RPC_PORT=$(cat ${DEPLOY_CONF} | grep "rpc_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    fi

    PHRASE=$(cat ${CONF_PATH}/keyfile.phrase)
    if [[ "${PHRASE}" == "" ]]; then
        printLog "error" "READ PHRASE FAILED"
        exit
    fi
    ACCOUNT=$(cat ${CONF_PATH}/keyfile.account)
    if [[ "${ACCOUNT}" == "" ]]; then
        printLog "error" "READ ACCOUNT FAILED"
        exit
    fi
}

################################################# Unclock Account #################################################
function unlockAccount() {
    http_data="{\"jsonrpc\":\"2.0\",\"method\":\"personal_unlockAccount\",\"params\":[\"${ACCOUNT}\",\"${PHRASE}\",0],\"id\":1}"
    res=$(curl -H "Content-Type: application/json" --data "${http_data}" http://${IP_ADDR}:${RPC_PORT})
    echo "${res}"
    if [[ $(echo ${res} | grep error ) != "" ]]; then
        printLog "error" "UNLOCK ACCOUNT FAILED"
        exit
    fi
    printLog "info" "Unlock account completed"
}

################################################# Set Super Admin #################################################
function setSuperAdmin() {
    ${BIN_PATH}/platonecli role setSuperAdmin --keyfile ${CONF_PATH}/keyfile.json --url ${IP_ADDR}:${RPC_PORT} <${CONF_PATH}/keyfile.phrase
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
        printLog "error" "SET SUPER ADMIN FAILED"
        exit
    else
        printLog "info" "Set super admin completed"
    fi

}

################################################# Set Chain Admin #################################################
function addChainAdmin() {
    ${BIN_PATH}/platonecli role addChainAdmin ${ACCOUNT} --keyfile ${CONF_PATH}/keyfile.json --url ${IP_ADDR}:${RPC_PORT} <${CONF_PATH}/keyfile.phrase
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
        printLog "error" "SET CHAIN ADMIN FAILED"
        exit
    else
        printLog "info" "Set chain admin completed"
    fi
}

################################################# Save Firstnode Info #################################################
function saveFirstnodeInfo() {
    {
        echo "user_name=$(whoami)"
        echo "node_id=${NODE_ID}"
        echo "ip_addr=${IP_ADDR}"
        echo "rpc_port=${RPC_PORT}"
    } >"${CONF_PATH}/firstnode.info"
    if [[ ! -f "${CONF_PATH}/firstnode.info" ]] || [[ "$(cat ${CONF_PATH}/firstnode.info)" == "" ]]; then
        printLog "error" "SETUP FIRSTNODE INFO FAILED"
        exit
    fi
    printLog "info" "File: ${CONF_PATH}/firstnode.info"
    printLog "info" "Firstnode: "
    cat ${CONF_PATH}/firstnode.info
    printLog "info" "Setup firstnode info completed"
}

################################################# Main #################################################
function main() {
    printLog "info" "## Deploy System Contract Start ##"
    checkEnv
    assignDefault
    readFile
    unlockAccount
    setSuperAdmin
    addChainAdmin
    saveFirstnodeInfo
    printLog "success" "Deploy system contract succeeded"
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
    --auto)
        AUTO="true"
        shift 1
        ;;
    *)
        printLog "error" "COMMAND \"$1\" NOT FOUND"
        help
        exit
        ;;
    esac

done
main
