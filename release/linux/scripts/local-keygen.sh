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

NODE_ID=""

DEPLOYMENT_CONF_PATH=""
NODE_DIR=""

#############################################################################################################
################################################# FUNCTIONS #################################################
#############################################################################################################

################################################# Help #################################################
function help() {
    echo
    echo "
USAGE: local-keygen.sh  [options] [value]

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

################################################# Generate Key #################################################
function generateKey() {
    ## generate node key
    keyinfo=$(${BIN_PATH}/ethkey genkeypair | sed s/[[:space:]]//g)
    address="${keyinfo:10:40}"
    prikey="${keyinfo:62:64}"
    pubkey="${keyinfo:137:128}"
    if [ ${#prikey} -ne 64 ]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* PRIVATE KEY LENGTH INVALID **********"
        exit
    fi

    ## backup node key
    if [ -f "${NODE_DIR}/node.address" ]; then
        timestamp=$(date '+%Y%m%d%H%M%S')
        mkdir -p "${NODE_DIR}/bak"
        mv "${NODE_DIR}/node.address" "${NODE_DIR}/bak/node.address.bak.${timestamp}"
        if [ -f "${NODE_DIR}/bak/node.address.bak.${timestamp}" ]; then
            echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : Backup ${NODE_DIR}/node.address completed"
        else
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* BACKUP NODE ADDRESS FAILED **********"
            exit
        fi
    fi
    if [ -f "${NODE_DIR}/node.prikey" ]; then
        timestamp=$(date '+%Y%m%d%H%M%S')
        mkdir -p "${NODE_DIR}/bak"
        mv "${NODE_DIR}/node.prikey" "${NODE_DIR}/bak/node.prikey.bak.${timestamp}"
        if [ -f "${NODE_DIR}/bak/node.prikey.bak.${timestamp}" ]; then
            echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : Backup ${NODE_DIR}/node.prikey succ"
        else
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* BACKUP NODE PRIVATE KEY FAILED **********"
            exit
        fi
    fi
    if [ -f "${NODE_DIR}/node.pubkey" ]; then
        timestamp=$(date '+%Y%m%d%H%M%S')
        mkdir -p "${NODE_DIR}/bak"
        mv "${NODE_DIR}/node.pubkey" "${NODE_DIR}/bak/node.pubkey.bak.${timestamp}"
        if [ -f "${NODE_DIR}/bak/node.pubkey.bak.${timestamp}" ]; then
            echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : Backup ${NODE_DIR}/node.pubkey succ"
        else
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* BACKUP NODE PUBLIC KEY FAILED **********"
            exit
        fi
    fi

    ## store node key
    echo "${address}" >"${NODE_DIR}/node.address"
    echo "${prikey}" >"${NODE_DIR}/node.prikey"
    echo "${pubkey}" >"${NODE_DIR}/node.pubkey"
    if [ ! -f "${NODE_DIR}/node.address" ] || [ ! -f "${NODE_DIR}/node.prikey" ] || [ ! -f "${NODE_DIR}/node.pubkey" ]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* STORE KEY INFO FAILED **********"
        exit
    fi
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : Files: ${NODE_DIR}/node.address, ${NODE_DIR}/node.prikey, ${NODE_DIR}/node.pubkey"
    echo "        Node-${NODE_ID}'s address: ${address}"
    echo "        Node-${NODE_ID}'s private key: ${prikey}"
    echo "        Node-${NODE_ID}'s public key: ${pubkey}"
}

################################################# Main #################################################
function main() {
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ## Node-${NODE_ID} Keygen Start ##"
    generateKey
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : Node-${NODE_ID} keygen succeeded"
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
        NODE_ID=$2
        NODE_DIR="${DATA_PATH}/node-$2"
        DEPLOYMENT_CONF_PATH="${DATA_PATH}/node-$2/deploy_conf"

        if [ ! -f "${DEPLOYMENT_CONF_PATH}/deploy_node-$2.conf" ]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* ${DEPLOYMENT_CONF_PATH}/deploy_node-$2.conf NOT FOUND **********"
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
