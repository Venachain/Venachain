#!/bin/bash

###########################################################################################################
################################################# VRIABLES #################################################
###########################################################################################################

## path
SCRIPT_NAME="$(basename ${0})"
SCRIPT_ALIAS="$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')"
OS=$(uname)
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
IP_ADDR=""
RPC_PORT=""
P2P_PORT=""
WS_PORT=""

## param
ip_addr=""
rpc_port=""
p2p_port=""
ws_port=""
auto=""

#############################################################################################################
################################################# FUNCTIONS #################################################
#############################################################################################################

################################################# Help #################################################
function help() {
    echo
    echo "
USAGE: ${SCRIPT_NAME}  [options] [value]

        OPTIONS:

           --nodeid, -n                 the node id (default: 0)

           --ip                         the node ip (default: 127.0.0.1)

           --rpc_port, -r               the node rpc_port (default: 6791)

           --p2p_port, -p               the node p2p_port (default: 16791)

           --ws_port, -w                the node ws_port (default: 26791)

           --auto, -a                   will no prompt to create the node key and init
                                        
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

################################################# Save Config #################################################
function saveConf() {
    conf="${NODE_DIR}/deploy_node-${NODE_ID}.conf"
    conf_tmp="${NODE_DIR}/deploy_node-${NODE_ID}.tmp.conf"
    if [[ "${2}" == "" ]]; then
        return
    fi
    cat "${conf}" | sed "s#${1}=.*#${1}=${2}#g" | cat >"${conf_tmp}"
    mv "${conf_tmp}" "${conf}"
}

################################################# Check Env #################################################
function checkEnv() {
    if [ ! -f "${CONF_PATH}/genesis.json" ]; then
        printLog "error" "FILE ${CONF_PATH}/genesis.json NOT FOUND"
        exit
    fi
}

################################################# Assign Default #################################################
function assignDefault() { 
    IP_ADDR="127.0.0.1"
    RPC_PORT="6791"
    P2P_PORT="16791"
    WS_PORT="26791"
}

################################################# Read Param #################################################
function readParam() {
    if [[ "${ip_addr}" != "" ]]; then
        IP_ADDR="${ip_addr}"
    fi
    if [[ "${rpc_port}" != "" ]]; then
        RPC_PORT="${rpc_port}"
    fi
    if [[ "${p2p_port}" != "" ]]; then
        P2P_PORT="${p2p_port}"
    fi
    if [[ "${ws_port}" != "" ]]; then
        WS_PORT="${ws_port}"
    fi
}

################################################# Main #################################################
function main() {
    if [[ "${NODE_ID}" == "" ]]; then
        NODE_ID="0"
    fi

    flag_auto=""
    if [[ "${auto}" == "true" ]]; then
        flag_auto=" --auto "
    fi

    printLog "info" "## Init Node-${NODE_ID} Start ##"
    checkEnv
    assignDefault
    readParam

    ./local/generate-deployconf.sh -n ${NODE_ID} ${flag_auto}
    saveConf "ip_addr" "${IP_ADDR}"
    saveConf "rpc_port" "${RPC_PORT}"
    saveConf "p2p_port" "${P2P_PORT}"
    saveConf "ws_port" "${WS_PORT}"

    ./local/generate-key.sh -n ${NODE_ID} ${flag_auto}
    ${BIN_PATH}/platone --datadir ${DATA_PATH}/node-${NODE_ID} init ${CONF_PATH}/genesis.json
    printLog "success" "Init Node-${NODE_ID} succeeded"
}

###########################################################################################################
#################################################  EXECUTE #################################################
###########################################################################################################
while [ ! $# -eq 0 ]; do
    case $1 in
    --nodeid | -n)
        shiftOption2 $#
        NODE_ID=$2
        NODE_DIR="${DATA_PATH}/node-${NODE_ID}"
        DEPLOY_CONF="${NODE_DIR}/deploy_node-${NODE_ID}.conf"
        shift 2
        ;;
    --ip)
        shiftOption2 $#
        ip_addr=$2
        shift 2
        ;;
    --rpc_port | -r)
        shiftOption2 $#
        rpc_port=$2
        shift 2
        ;;
    --p2p_port | -p)
        shiftOption2 $#
        p2p_port=$2
        shift 2
        ;;
    --ws_port | -w)
        shiftOption2 $#
        ws_port=$2
        shift 2
        ;;
    --auto | -a)
        shiftOption2 $#
        auto=$2
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
