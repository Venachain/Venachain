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

NODE_ID=""
AUTO=""

NODE_DIR=""
NODE_DIR=""

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

           --auto                   auto=true: will no prompt to create the node key 
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

################################################# Yes Or No #################################################
function yesOrNo() {
    read -p "" anw
    case $anw in
    [Yy][Ee][Ss] | [yY])
        return 0
        ;;
    [Nn][Oo] | [Nn])
        return 1
        ;;
    esac
    return 1
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
}

################################################# Read Key #################################################
function readKey() {
    address=$(cat "${NODE_DIR}"/node.address)
    pubkey=$(cat "${NODE_DIR}"/node.pubkey)
    prikey=$(cat "${NODE_DIR}"/node.prikey)
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : Key files: ${NODE_DIR}/node.address, ${NODE_DIR}/node.prikey, ${NODE_DIR}/node.pubkey"
    echo "        Node-${NODE_ID}'s address: ${address}"
    echo "        Node-${NODE_ID}'s private key: ${prikey}"
    echo "        Node-${NODE_ID}'s public key: ${pubkey}"
}

################################################# Main #################################################
function main() {
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ## Node-${NODE_ID} Keygen Start ##"
    if [[ "${AUTO}" == "true" ]]; then
        if [ ! -f "${NODE_DIR}"/node.pubkey ] || [ ! -f "${NODE_DIR}"/node.prikey ] || [ ! -f "${NODE_DIR}"/node.address ]; then
            generateKey
        fi
    else
        echo
        echo "Do You What To Create a new node key ? (Please do not recreate the first node node.key) Yes or No(y/n):"
        yesOrNo
        if [ $? -eq 0 ]; then
            if [ -f "${NODE_DIR}"/node.pubkey ] || [ -f "${NODE_DIR}"/node.prikey ] || [ -f "${NODE_DIR}"/node.address ]; then
                echo "Node key already exists, re create? (Please do not recreate the first node node.key) Yes or No(y/n):"
                yesOrNo
                if [ $? -eq 0 ]; then
                    generateKey
                fi
            else
                generateKey
            fi
        else
            if [ ! -f "${NODE_DIR}"/node.pubkey ] || [ ! -f "${NODE_DIR}"/node.prikey ] || [ ! -f "${NODE_DIR}"/node.address ]; then
                echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* Please Put Your Node's key file \"node.pubkey\" to the directory ${NODE_DIR} *********"
                exit
            fi
        fi
    fi
    readKey
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
