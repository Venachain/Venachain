#!/bin/bash

###########################################################################################################
################################################# VRIABLES #################################################
###########################################################################################################
OS=$(uname)
PROJECT_PATH=$(
    cd $(dirname $0)
    cd ../
    pwd
)
BIN_PATH=${PROJECT_PATH}/bin
DATA_PATH=${PROJECT_PATH}/data
CONF_PATH=${PROJECT_PATH}/conf

NODE_ID=""

DEPLOYMENT_CONF_PATH=""
NODE_DIR=""
IP_ADDR=""
P2P_PORT=""

#############################################################################################################
################################################# FUNCTIONS #################################################
#############################################################################################################

################################################# Help #################################################
function help() {
    echo
    echo "
USAGE: local-setup-genesis.sh  [options] [value]

        OPTIONS:

           --node, -n                   the specified node name. must be specified

           --help, -h                   show help
"
}

################################################# Check Shift Option #################################################
function shiftOption2() {
    if [[ $1 -lt 2 ]]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* MISS OPTION VALUE! PLEASE SET THE VALUE **********"
        exit
    fi
}

################################################# Read File #################################################
function readFile() {
    file=$1
    IP_ADDR=$(cat $file | grep "ip_addr=" | sed -e 's/ip_addr=\(.*\)/\1/g')
    P2P_PORT=$(cat $1 | grep "p2p_port=" | sed -e 's/p2p_port=\(.*\)/\1/g')

    if [[ "${IP_ADDR}" == "" ]] || [[ "${P2P_PORT}" == "" ]]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* FILE ${file} MISS VALUE **********"
        exit
    fi
}

################################################# Replace List #################################################
function replaceList() {
    res=$(echo $2 | sed "s/,/\",\"/g")
    "${BIN_PATH}"/repstr "${CONF_PATH}"/genesis.temp.json $1 $(echo \"${res}\")
}

################################################# Create Genesis #################################################
function createGenesis() {
    if [ "${OS}" = "Darwin" ]; then
        createGenesisDarwin "$@"
        return
    fi

    replaceList "__VALIDATOR__" $1
    "${BIN_PATH}/repstr" "${CONF_PATH}/genesis.temp.json" "DEFAULT-ACCOUNT" "$2"
    "${BIN_PATH}/repstr" "${CONF_PATH}/genesis.temp.json" "__INTERPRETER__" "$3"

    now=$(date +%s)
    "${BIN_PATH}/repstr" "${CONF_PATH}/genesis.temp.json" "TIMESTAMP" "${now}"
}

################################################# Create Genesis (MAC OS) #################################################
function createGenesisDarwin() {
    sed -i '' "s#__VALIDATOR__#\"$1\"#g" "${CONF_PATH}/genesis.temp.json"
    sed -i '' "s/DEFAULT-ACCOUNT/$2/g" "${CONF_PATH}/genesis.temp.json"
    sed -i '' "s/__INTERPRETER__/$3/g" "${CONF_PATH}/genesis.temp.json"

    now=$(date +%s)
    sed -i '' "s/TIMESTAMP/${now}/g" "${CONF_PATH}/genesis.temp.json"
}

################################################# Setup Genesis #################################################
function setupGenesis() {
    ## ckeck node's key and address
    if [ ! -d "${NODE_DIR}" ] || [ ! -f "${NODE_DIR}/node.pubkey" ] || [ ! -f "${NODE_DIR}/node.address" ] || [ ! -f "${NODE_DIR}/node.prikey" ]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* PLEASE GENERATE NODE KEY FIRST **********"
        exit
    fi

    ## read file
    file="${DEPLOYMENT_CONF_PATH}/deploy_node-${NODE_ID}.conf"
    readFile "$file"

    ## backup genesis if exists
    if [ -f "${CONF_PATH}/genesis.json" ]; then
        timestamp=$(date '+%Y%m%d%H%M%S')
        mkdir -p "${CONF_PATH}/bak"
        mv "${CONF_PATH}/genesis.json" "${CONF_PATH}/bak/genesis.json.bak.${timestamp}"
        if [ -f "${CONF_PATH}/bak/genesis.json.bak.${timestamp}" ]; then
            echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : Backup ${CONF_PATH}/genesis.json completed"
        else
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* BACKUP GENESIS FILE FAILED **********"
            exit
        fi
    fi

    ## create genesis
    node_key=$(cat ${NODE_DIR}/node.pubkey)
    node_addr=$(cat ${NODE_DIR}/node.address)
    default_enode="enode://${node_key}@${IP_ADDR}:${P2P_PORT}"
    interpreter="all"
    cp "${CONF_PATH}/genesis.json.istanbul.template" "${CONF_PATH}/genesis.temp.json"
    if [ ! -f "${CONF_PATH}/genesis.temp.json" ]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* GENERATE GENESIS TEMP FILE FAILED **********"
        exit
    fi
    createGenesis "${default_enode}" "${node_addr}" "${interpreter}"
    mv "${CONF_PATH}/genesis.temp.json" "${CONF_PATH}/genesis.json"
    if [ ! -f "${CONF_PATH}/genesis.json" ]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* GENERATE GENESIS FILE FAILED **********"
        exit
    fi
}

################################################# Main #################################################
function main() {
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ## Setup Genesis Start ##"
    setupGenesis "${file}"
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : File: ${CONF_PATH}/genesis.json"
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : Genesis:"
    cat ${CONF_PATH}/genesis.json
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : Setup genesis succeeded"

}

###########################################################################################################
#################################################  EXECUTE #################################################
###########################################################################################################
if [ $# -eq 0 ]; then
    help
    exit
fi
while [ ! $# -eq 0 ]; do
    case $1 in
    --node | -n)
        NODE_ID=$2
        NODE_DIR="${DATA_PATH}/node-$2"
        DEPLOYMENT_CONF_PATH="${DATA_PATH}/node-$2/deploy_conf"

        if [ ! -f "${DEPLOYMENT_CONF_PATH}/deploy_node-$2.conf" ]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* FILE ${DEPLOYMENT_CONF_PATH}/deploy_node-$2.conf NOT FOUND **********"
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
