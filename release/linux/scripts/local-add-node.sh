#!/bin/bash

###########################################################################################################
################################################# VRIABLES #################################################
###########################################################################################################
PROJECT_PATH=$(
    cd $(dirname $0)
    cd ../
    pwd
)
BIN_PATH=${PROJECT_PATH}/bin
DATA_PATH=${PROJECT_PATH}/data
CONF_PATH=${PROJECT_PATH}/conf

NODE_ID=""

DEPLOY_NODE_CONF_PATH=""
NODE_DIR=""
P2P_PORT=""
RPC_PORT=""
IP_ADDR=""
FIRSTNODE_IP_ADDR=""
FIRSTNODE_RPC_PORT=""

#############################################################################################################
################################################# FUNCTIONS #################################################
#############################################################################################################

################################################# Help #################################################
function help() {
    echo
    echo "
USAGE: local-add-node.sh  [options] [value]

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
        return 1
    fi
}

################################################# Read File #################################################
function readFile() {
    file="${DEPLOY_NODE_CONF_PATH}"/deploy_node-"${NODE_ID}".conf
    PUB_KEY=$(cat ${NODE_DIR}/node.pubkey)
    IP_ADDR=$(cat $file | grep "ip_addr=" | sed -e 's/ip_addr=\(.*\)/\1/g')
    P2P_PORT=$(cat $file | grep "p2p_port=" | sed -e 's/p2p_port=\(.*\)/\1/g')
    RPC_PORT=$(cat $file | grep "rpc_port=" | sed -e 's/rpc_port=\(.*\)/\1/g')
    if [[ "${PUB_KEY}" == "" ]] || [[ "${IP_ADDR}" == "" ]] || [[ "${P2P_PORT}" == "" ]] || [[ "${RPC_PORT}" == "" ]]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* FILE ${file} MISS VALUE **********"
        exit
    fi

    firstnode_info="${CONF_PATH}/firstnode.info"
    FIRSTNODE_IP_ADDR=$(cat ${firstnode_info} | grep "ip_addr=" | sed -e 's/ip_addr=\(.*\)/\1/g')
    FIRSTNODE_RPC_PORT=$(cat ${firstnode_info} | grep "rpc_port=" | sed -e 's/rpc_port=\(.*\)/\1/g')
    if [[ "${FIRSTNODE_IP_ADDR}" == "" ]] || [[ "${FIRSTNODE_RPC_PORT}" == "" ]]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* FILE ${file} MISS VALUE **********"
        exit
    fi
}

################################################# Add Node #################################################
function addNode() {
    inter_ip=127.0.0.1
    ${BIN_PATH}/platonecli node add "${NODE_ID}" "${PUB_KEY}" "${IP_ADDR}" "${inter_ip}" --p2pPort "${P2P_PORT}" --rpcPort "${RPC_PORT}" --keyfile "${CONF_PATH}"/keyfile.json --url "${FIRSTNODE_IP_ADDR}:${FIRSTNODE_RPC_PORT}" <"${CONF_PATH}"/keyfile.phrase >/dev/null 2>&1
    timer=0
    add_node_flag=""
    while [ ${timer} -lt 10 ]; do
        add_node_flag=$("${BIN_PATH}"/platonecli node query --name "${NODE_ID}" --keyfile "${CONF_PATH}"/keyfile.json --url "${FIRSTNODE_IP_ADDR}:${FIRSTNODE_RPC_PORT}" <"${CONF_PATH}"/keyfile.phrase) >/dev/null 2>&1
        if [[ $(echo ${add_node_flag} | grep "success") != "" ]]; then
            break
        fi
        sleep 2
        let timer++
    done
    if [[ $(echo ${add_node_flag} | grep "success") == "" ]]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* ADD NODE-${NODE_ID} FAILED **********"
        exit
    fi
}

################################################# Main #################################################
function main() {
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ## Add Node-${NODE_ID} Start ##"
    readFile
    addNode
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : Add Node-${NODE_ID} succeeded"
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
    --node_name | -n)
        NODE_DIR="${DATA_PATH}"/node-$2
        NODE_ID=$2
        DEPLOY_NODE_CONF_PATH=${DATA_PATH}/node-$2/deploy_conf

        if [ ! -f ${DEPLOY_NODE_CONF_PATH}/deploy_node-$2.conf ]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* ${DEPLOY_NODE_CONF_PATH}/deploy_node-$2.conf NOT FOUND **********"
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
