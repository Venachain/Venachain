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
DESC=""
P2P_PORT=:""
RPC_PORT=""
IP_ADDR=""
PUBKEY=""
TYPE=""
PUBKEY_FILE=""
KEYFILE=""
PHRASE=""
FIRSTNODE_INFO=""
FIRSTNODE_IP_ADDR=""
FIRSTNODE_RPC_PORT=""



## param
desc=""
p2p_port=""
rpc_port=""
ip_addr=""
pubkey=""
type=""

#############################################################################################################
################################################# FUNCTIONS #################################################
#############################################################################################################

################################################# Help #################################################
function help() {
    echo
    echo "
USAGE: ${SCRIPTS_NAME}  [options] [value]

        OPTIONS:

           --nodeid, -n                 the specified node id. must be specified

           --desc, -d                   the specified node desc

           --p2p_port, -p               the specified node p2p_port
                                        If the node specified by nodeid is local,
                                        then you do not need to specify this option.

           --rpc_port, -r               the specified node rpc_port
                                        If the node specified by nodeid is local,
                                        then you do not need to specify this option.

           --ip                         the specified node ip
                                        If the node specified by nodeid is local,
                                        then you do not need to specify this option.

           --pubkey                     the specified node pubkey
                                        If the node specified by nodeid is local,
                                        then you do not need to specify this option.

           --type                       select specified node type in 2 & 3
                                        2 is observer, 3 is lightnode (default: 2)

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
        help
        return 1
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

    PUBKEY_FILE="${NODE_DIR}/node.pubkey"
    if [[ ! -f "${PUBKEY_FILE}" ]]; then
        printLog "error" "FILE ${PUBKEY_FILE} NOT FOUND"
        exit
    fi
    KEYFILE="${CONF_PATH}/keyfile.json"
    if [[ ! -f "${KEYFILE}" ]]; then
        printLog "error" "FILE ${KEYFILE} NOT FOUND"
        exit
    fi
    PHRASE="${CONF_PATH}/keyfile.phrase"
    if [[ ! -f "${PHRASE}" ]]; then
        printLog "error" "FILE ${PHRASE} NOT FOUND"
        exit
    fi
    FIRSTNODE_INFO="${CONF_PATH}/firstnode.info"
    if [[ ! -f "${FIRSTNODE_INFO}" ]]; then
        printLog "error" "FILE ${FIRSTNODE_INFO} NOT FOUND"
        exit
    fi
}

################################################# Assign Default #################################################
function assignDefault() { 
    IP_ADDR="127.0.0.1"
    RPC_PORT="6791"
    P2P_PORT="16791"
    TYPE="2"
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
    checkConf "p2p_port"
    if [[ $? -eq 1 ]]; then
        P2P_PORT=$(cat ${DEPLOY_CONF} | grep "p2p_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    fi

    if [[ "${PUBKEY}" == "" ]]; then
        PUBKEY=$(cat ${PUBKEY_FILE})
        if [[ "${PUBKEY}" == "" ]]; then
            printLog "error" "NODE'S PUBLIC KEY IS EMPTY"
            exit
        fi
    fi

    FIRSTNODE_IP_ADDR=$(cat ${FIRSTNODE_INFO} | grep "ip_addr=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    if [[ "${FIRSTNODE_IP_ADDR}" == "" ]]; then
        printLog "error" "FIRST NODE'S IP NOT FOUND"
        exit
    fi
    FIRSTNODE_RPC_PORT=$(cat ${FIRSTNODE_INFO} | grep "rpc_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    if [[ "${FIRSTNODE_RPC_PORT}" == "" ]]; then
        printLog "error" "FIRST NODE'S RPC PORT NOT FOUND"
        exit
    fi
}

################################################# Read Param #################################################
function readParam() {
    if [[ "${desc}" != "" ]]; then
        DESC="${desc}"
    fi
    if [[ "${ip_addr}" != "" ]]; then
        IP_ADDR="${ip_addr}"
    fi
    if [[ "${rpc_port}" != "" ]]; then
        RPC_PORT="${rpc_port}"
    fi
    if [[ "${p2p_port}" != "" ]]; then
        P2P_PORT="${p2p_port}"
    fi
    if [[ "${pubkey}" != "" ]]; then
        PUBKEY="${pubkey}"
    fi
    if [[ "${type}" != "" ]]; then
        TYPE="${type}"
    fi
}

################################################# Add Node #################################################
function addNode() {
    inter_ip=127.0.0.1
    flag_desc=""
    if [[ "${DESC}" != "" ]]; then
        flag_desc="--desc ${DESC}"
    fi
    ${BIN_PATH}/platonecli node add "${NODE_ID}" "${PUBKEY}" "${IP_ADDR}" "${inter_ip}" "${TYPE}" --p2pPort "${P2P_PORT}" --rpcPort "${RPC_PORT}" ${flag_desc} --keyfile "${KEYFILE}" --url "${FIRSTNODE_IP_ADDR}:${FIRSTNODE_RPC_PORT}" <"${PHRASE}"  
    
    timer=0
    res_add_node=""
    while [ ${timer} -lt 10 ]; do
        res_add_node=$("${BIN_PATH}"/platonecli node query --name "${NODE_ID}" --url "${FIRSTNODE_IP_ADDR}:${FIRSTNODE_RPC_PORT}") >/dev/null 2>&1
        if [[ $(echo ${res_add_node} | grep "success") != "" ]]; then
            break
        fi
        sleep 3
        let timer++
    done
    if [[ $(echo ${res_add_node} | grep "success") == "" ]]; then
        printLog "error" "ADD NODE-${NODE_ID} FAILED"
        exit
    fi
}

################################################# Main #################################################
function main() {
    printLog "info" "## Add Node-${NODE_ID} Start ##"
    checkEnv
    assignDefault
    readFile
    readParam
    addNode
    printLog "success" "Add Node-${NODE_ID} succeeded"
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
    --desc | -d)
        shiftOption2 $#
        desc=$2
        shift 2
        ;;
    --p2p_port | -p)
        shiftOption2 $#
        p2p_port=$2
        shift 2
        ;;
    --rpc_port | -r)
        shiftOption2 $#
        rpc_port=$2
        shift 2
        ;;
    --ip)
        shiftOption2 $#
        ip_addr=$2
        shift 2
        ;;
    --pubkey)
        shiftOption2 $#
        pubkey=$2
        shift 2
        ;;
    --type)
        shiftOption2 $#
        type=$2
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
