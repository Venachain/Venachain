#!/bin/bash

###########################################################################################################
################################################# VRIABLES #################################################
###########################################################################################################

## path
CURRENT_PATH="$(cd "$(dirname "$0")";pwd)"
PROJECT_PATH="$(
    cd $(dirname ${0})
    cd ../../
    pwd
)"
BIN_PATH="${PROJECT_PATH}/bin"
DATA_PATH="${PROJECT_PATH}/data"

## global
SCRIPT_NAME="$(basename ${0})"
SCRIPT_ALIAS="$(echo ${CURRENT_PATH}/${SCRIPT_NAME} | sed -e 's/\(.*\)\/scripts\/\(.*\).sh/\2/g')"
NODE_ID=""
AUTO=""

NODE_DIR=""
DEPLOY_CONF=""

#############################################################################################################
################################################# FUNCTIONS #################################################
#############################################################################################################

################################################# Help #################################################
function help() {
    echo
    echo "
USAGE: ${SCRIPT_NAME}  [options] [value]

        OPTIONS:

            --nodeid, -n                the specified node name, must be specified

            --auto                      will read exit node key 

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
        help
        exit 1
    fi
}

################################################# Yes Or No #################################################
function yesOrNo() {
    read -p "" anw
    case "${anw}" in
    [Yy][Ee][Ss] | [yY])
        return 1
        ;;
    [Nn][Oo] | [Nn])
        return 0
        ;;
    esac
    return 0
}

################################################# Check Env #################################################
function checkEnv() {
    NODE_DIR="${DATA_PATH}/node-${NODE_ID}"
    DEPLOY_CONF="${NODE_DIR}/deploy_node-${NODE_ID}.conf"

    if [[ "${NODE_ID}" == "" ]]; then
        printLog "error" "NODE NAME NOT SET"
        exit 1
    fi 

    if [ ! -f "${BIN_PATH}/venakey" ]; then
        printLog "error" "FILE ${BIN_PATH}/venakey NOT FOUND"
        exit 1
    fi
}

################################################# Generate Key #################################################
function generateKey() {
    ## generate node key
    keyinfo="$(${BIN_PATH}/venakey genkeypair | sed s/[[:space:]]//g)"
    address="${keyinfo:10:40}"
    prikey="${keyinfo:62:64}"
    pubkey="${keyinfo:137:128}"
    if [ ${#prikey} -ne 64 ]; then
        printLog "error" "PRIVATE KEY LENGTH INVALID"
        exit 1
    fi
    mkdir -p "${NODE_DIR}"

    ## backup node key
    if [ -f "${NODE_DIR}/node.address" ]; then
        timestamp=$(date '+%Y%m%d%H%M%S')
        mkdir -p "${NODE_DIR}/bak"
        mv "${NODE_DIR}/node.address" "${NODE_DIR}/bak/node.address.bak.${timestamp}"
        if [ -f "${NODE_DIR}/bak/node.address.bak.${timestamp}" ]; then
            printLog "info" "Backup ${NODE_DIR}/node.address completed"
        else
            printLog "error" "BACKUP NODE ADDRESS FAILED"
            exit 1
        fi
    fi
    if [ -f "${NODE_DIR}/node.prikey" ]; then
        timestamp=$(date '+%Y%m%d%H%M%S')
        mkdir -p "${NODE_DIR}/bak"
        mv "${NODE_DIR}/node.prikey" "${NODE_DIR}/bak/node.prikey.bak.${timestamp}"
        if [ -f "${NODE_DIR}/bak/node.prikey.bak.${timestamp}" ]; then
            printLog "info" "Backup ${NODE_DIR}/node.prikey completed"
        else
            printLog "error" "BACKUP NODE PRIVATE KEY FAILED"
            exit 1
        fi
    fi
    if [ -f "${NODE_DIR}/node.pubkey" ]; then
        timestamp=$(date '+%Y%m%d%H%M%S')
        mkdir -p "${NODE_DIR}/bak"
        mv "${NODE_DIR}/node.pubkey" "${NODE_DIR}/bak/node.pubkey.bak.${timestamp}"
        if [ -f "${NODE_DIR}/bak/node.pubkey.bak.${timestamp}" ]; then
            printLog "info" "Backup ${NODE_DIR}/node.pubkey completed"
        else
            printLog "error" "BACKUP NODE PUBLIC KEY FAILED"
            exit 1
        fi
    fi

    ## store node key
    echo "${address}" >"${NODE_DIR}/node.address"
    echo "${prikey}" >"${NODE_DIR}/node.prikey"
    echo "${pubkey}" >"${NODE_DIR}/node.pubkey"
    if [ ! -f "${NODE_DIR}/node.address" ] || [ ! -f "${NODE_DIR}/node.prikey" ] || [ ! -f "${NODE_DIR}/node.pubkey" ]; then
        printLog "error" "STORE KEY INFO FAILED"
        exit 1
    fi
}

################################################# Setup Key #################################################
function setupKey() {
    if [[ "${AUTO}" == "true" ]]; then
        if [ ! -f "${NODE_DIR}/node.pubkey" ] || [ ! -f "${NODE_DIR}/node.prikey" ] || [ ! -f "${NODE_DIR}/node.address" ]; then
            generateKey
        else
            printLog "warn" "Key Already Exists, Will Read Them Automatically"
            exit 0
        fi
    else
        printLog "question" "Do You What To Create a new node key ? Yes or No(y/n):"
        yesOrNo
        if [ $? -eq 1 ]; then
            if [ -f "${NODE_DIR}/node.pubkey" ] || [ -f "${NODE_DIR}/node.prikey" ] || [ -f "${NODE_DIR}/node.address" ]; then
                printLog "question" "Node key already exists, overwrite it? Yes or No(y/n):"
                yesOrNo
                if [ $? -ne 1 ]; then
                    exit 2
                fi
            fi
            generateKey
        else
            exit 2
        fi
    fi
}

################################################# Read Key #################################################
function readKey() {
    address="$(cat "${NODE_DIR}"/node.address)"
    pubkey="$(cat "${NODE_DIR}"/node.pubkey)"
    prikey="$(cat "${NODE_DIR}"/node.prikey)"
    printLog "info" "Key files: ${NODE_DIR}/node.address, ${NODE_DIR}/node.prikey, ${NODE_DIR}/node.pubkey"
    echo "        Node-${NODE_ID}'s address: ${address}"
    echo "        Node-${NODE_ID}'s private key: ${prikey}"
    echo "        Node-${NODE_ID}'s public key: ${pubkey}"
}

################################################# Main #################################################
function main() {
    printLog "info" "## Node-${NODE_ID} generate key Start ##"
    checkEnv

    setupKey    
    readKey
    printLog "success" "Node-${NODE_ID} generate key succeeded"
}

###########################################################################################################
#################################################  EXECUTE #################################################
###########################################################################################################
if [ $# -eq 0 ]; then
    help
    exit 1
fi
while [ ! $# -eq 0 ]; do
    case "${1}" in
    --nodeid | -n)
        shiftOption2 $#
        NODE_ID="${2}"
        shift 2
        ;;
    --auto)
        AUTO="true"
        shift 1
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
