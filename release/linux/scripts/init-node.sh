#!/bin/bash

###########################################################################################################
################################################# VRIABLES #################################################
###########################################################################################################

## path
CURRENT_PATH="$(cd "$(dirname "$0")";pwd)"
PROJECT_PATH="$(
    cd $(dirname ${0})
    cd ../
    pwd
)"
BIN_PATH="${PROJECT_PATH}/bin"
DATA_PATH="${PROJECT_PATH}/data"
CONF_PATH="${PROJECT_PATH}/conf"
SCRIPT_PATH="${PROJECT_PATH}/scripts"

## global
OS=$(uname)
SCRIPT_NAME="$(basename ${0})"
SCRIPT_ALIAS="$(echo ${CURRENT_PATH}/${SCRIPT_NAME} | sed -e 's/\(.*\)\/scripts\/\(.*\).sh/\2/g')"
NODE_ID=""
IP_ADDR=""
RPC_PORT=""
P2P_PORT=""
WS_PORT=""
AUTO=""

NODE_DIR=""
DEPLOY_CONF=""

## param
ip_addr=""
rpc_port=""
p2p_port=""
ws_port=""

#############################################################################################################
################################################# FUNCTIONS #################################################
#############################################################################################################

################################################# Help #################################################
function help() {
    echo
    echo "
USAGE: ${SCRIPT_NAME}  [options] [value]

        OPTIONS:

            --nodeid, -n                the node id (default: 0)

            --ip                        the node ip (default: 127.0.0.1)

            --rpc_port, -rpc            the node rpc_port (default: 6791)

            --p2p_port, -p2p            the node p2p_port (default: 16791)

            --ws_port, -ws              the node ws_port (default: 26791)

            --auto                      will read exit the deploy conf and node key 
                                        
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
        exit 1
    fi
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
    if [[ "${NODE_ID}" == "" ]]; then
        NODE_ID="0"
    fi
    NODE_DIR="${DATA_PATH}/node-${NODE_ID}"
    DEPLOY_CONF="${NODE_DIR}/deploy_node-${NODE_ID}.conf"
    
    if [ ! -f "${CONF_PATH}/genesis.json" ]; then
        printLog "error" "FILE ${CONF_PATH}/genesis.json NOT FOUND"
        exit 1
    fi
    if [ ! -f "${BIN_PATH}/venachain" ]; then 
        printLog "error" "FILE ${BIN_PATH}/venachain NOT FOUND"
        exit 1
    fi
}

################################################# Assign Default #################################################
function assignDefault() { 
    IP_ADDR="127.0.0.1"
    RPC_PORT="6791"
    P2P_PORT="16791"
    WS_PORT="26791"
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
    checkConf "p2p_port"
    if [[ $? -eq 1 ]]; then
        P2P_PORT="$(cat ${DEPLOY_CONF} | grep "p2p_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    fi
    checkConf "ws_port"
    if [[ $? -eq 1 ]]; then
        WS_PORT="$(cat ${DEPLOY_CONF} | grep "ws_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    fi
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

################################################# Init Node #################################################
function initNode() {
    ## setup flag
    flag_auto=""
    if [[ "${AUTO}" == "true" ]]; then
        flag_auto=" --auto "
    fi

    ## generate deploy conf
    "${SCRIPT_PATH}"/venachainctl.sh dcgen -n "${NODE_ID}" ${flag_auto}
    res=$?
    if [[ ${res} -eq 1 ]]; then
        exit 1
    fi
    if [[ ${res} -eq 2 ]]; then
        if [[ -f "${NODE_DIR}/deploy_node-${NODE_ID}.conf" ]]; then
            printLog "warn" "Deploy conf Exists, Will Read Them Automatically"
        else 
            printLog "warn" "Please Put Your Deploy Conf File to the directory ${NODE_DIR}"
            exit 1
        fi
    fi

    assignDefault
    readFile
    readParam
    saveConf "ip_addr" "${IP_ADDR}"
    saveConf "rpc_port" "${RPC_PORT}"
    saveConf "p2p_port" "${P2P_PORT}"
    saveConf "ws_port" "${WS_PORT}"

    ## generate key
    "${SCRIPT_PATH}"/venachainctl.sh keygen -n "${NODE_ID}" ${flag_auto}
    res=$?
    if [[ ${res} -eq 1 ]]; then
        exit 1
    fi
    if [[ ${res} -eq 2 ]]; then
        if [[ -f "${DATA_PATH}/node-${NODE_ID}/node.prikey" ]] && [[ -f "${DATA_PATH}/node-${NODE_ID}/node.pubkey" ]] && [[ -f "${DATA_PATH}/node-${NODE_ID}/node.address" ]]; then
            printLog "warn" "Key Already Exists, Will Read Them Automatically"
        else 
            printLog "warn" "Please Put Your Node keys \"node.prikey\",\"node.pubkey\",\"node.address\" to the directory ${NODE_DIR}"
            exit 1
        fi
    fi

    ## init node
    "${BIN_PATH}"/venachain --datadir "${DATA_PATH}/node-${NODE_ID}" init "${CONF_PATH}/genesis.json"
    if [[ $? -eq 1 ]]; then
        printLog "error" "INIT NODE-${NODE_ID} FAILED"
        exit 1
    fi
    printLog "success" "Init Node-${NODE_ID} succeeded"
}

################################################# Main #################################################
function main() {
    checkEnv  
    assignDefault
    readParam

    initNode
}

###########################################################################################################
#################################################  EXECUTE #################################################
###########################################################################################################
while [ ! $# -eq 0 ]; do
    case "${1}" in
    --nodeid | -n)
        shiftOption2 $#
        NODE_ID="${2}"
        shift 2
        ;;
    --ip)
        shiftOption2 $#
        ip_addr="${2}"
        shift 2
        ;;
    --rpc_port | -rpc)
        shiftOption2 $#
        rpc_port="${2}"
        shift 2
        ;;
    --p2p_port | -p2p)
        shiftOption2 $#
        p2p_port="${2}"
        shift 2
        ;;
    --ws_port | -ws)
        shiftOption2 $#
        ws_port="${2}"
        shift 2
        ;;
    --auto)
        shiftOption2 $#
        AUTO="${2}"
        shift 2
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
