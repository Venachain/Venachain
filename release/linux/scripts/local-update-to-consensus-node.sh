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
        return 1
    fi
}

################################################# Read File #################################################
function readFile() {
    file="${NODE_DIR}"/deploy_node-"${NODE_ID}".conf
    IP_ADDR=$(cat $file | grep "ip_addr=" | sed -e 's/ip_addr=\(.*\)/\1/g')
    RPC_PORT=$(cat $file | grep "rpc_port=" | sed -e 's/rpc_port=\(.*\)/\1/g')
    if [[ "${IP_ADDR}" == "" ]] || [[ "${RPC_PORT}" == "" ]]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* FILE ${file} MISS VALUE **********"
        exit
    fi

    firstnode_info="${CONF_PATH}/firstnode.info"
    FIRSTNODE_IP_ADDR=$(cat ${firstnode_info} | grep "ip_addr=" | sed -e 's/ip_addr=\(.*\)/\1/g')
    FIRSTNODE_RPC_PORT=$(cat ${firstnode_info} | grep "rpc_port=" | sed -e 's/rpc_port=\(.*\)/\1/g')
    if [[ "${FIRSTNODE_IP_ADDR}" == "" ]] || [[ "${FIRSTNODE_RPC_PORT}" == "" ]]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* FILE ${CONF_PATH}/firstnode.info MISS VALUE **********"
        exit
    fi
}

################################################# Update To Consensus Node #################################################
function updateToConsensusNode() {
    ${BIN_PATH}/platonecli node update "${NODE_ID}" --type "consensus" --keyfile "${CONF_PATH}"/keyfile.json --url "${FIRSTNODE_IP_ADDR}:${FIRSTNODE_RPC_PORT}" <"${CONF_PATH}"/keyfile.phrase >/dev/null 2>&1
    timer=0
    update_node_flag=""
    while [ ${timer} -lt 10 ]; do
        update_node_flag=$(${BIN_PATH}/platonecli node query --type consensus --name ${NODE_ID} --keyfile ${CONF_PATH}/keyfile.json --url "${FIRSTNODE_IP_ADDR}:${FIRSTNODE_RPC_PORT}" <"${CONF_PATH}"/keyfile.phrase) >/dev/null 2>&1
        if [[ $(echo ${update_node_flag} | grep "success") != "" ]]; then
            break
        fi
        sleep 1
        let timer++
    done
    if [[ $(echo ${update_node_flag} | grep "success") == "" ]]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/.\/\(.*\).sh/\1/g')] : ********* UPDATE NODE-${NODE_ID} TO CONSENSUS NODE FAILED**********"
        exit
    fi
}

################################################# Main #################################################
function main() {
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ## Update Node-${NODE_ID} To Consensus Node Start ##"
    readFile
    updateToConsensusNode
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : Update Node-${NODE_ID} to consensus node succeeded"
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
            echo "[ERROR]: ********* FILE ${NODE_DIR}/deploy_node-$2.conf NOT FOUND **********"
            exit
        fi
        shift 2
        ;;
    --auto)
        AUTO="true"
        shift 1
        ;;
    *)
        echo "[ERROR]: ********* COMMAND \"$1\" NOT FOUND **********"
        help
        exit
        ;;
    esac
done
main
