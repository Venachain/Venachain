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
CONF_PATH="${PROJECT_PATH}/conf"

## global
OS=$(uname)
SCRIPT_NAME="$(basename ${0})"
SCRIPT_ALIAS="$(echo ${CURRENT_PATH}/${SCRIPT_NAME} | sed -e 's/\(.*\)\/scripts\/\(.*\).sh/\2/g')"
NODE_ID=""
VALIDATOR_NODES=""
INTERPRETER=""
IP_ADDR=""
P2P_PORT=""

NODE_DIR=""
DEPLOY_CONF=""
NODE_KEY=""
NODE_ADDRESS=""

## param
ip_addr=""
p2p_port=""

#############################################################################################################
################################################# FUNCTIONS #################################################
#############################################################################################################

################################################# Help #################################################
function help() {
    echo
    echo "
USAGE: ${SCRIPT_NAME}  [options] [value]

        OPTIONS:

            --nodeid, -n                the first node id, must be specified

            --interpreter, -i           Select virtual machine interpreter, must be specified
                                        \"wasm\", \"evm\" and \"all\" are supported

            --validatorNodes, -v        set the genesis validatorNodes
                                        (default: the first node enode code)

            --ip                        the first node ip 

            --p2p_port, -p2p            the first node p2p_port 
                                        
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

################################################# Yes Or No #################################################
function yesOrNo() {
    read -p "" anw
    case "${anw}" in
    [Yy][Ee][Ss] | [yY])
        return 0
        ;;
    [Nn][Oo] | [Nn])
        return 1
        ;;
    esac
    return 1
}

################################################# Check Conf #################################################
function checkConf() {
    ref=$(cat "${DEPLOY_CONF}" | grep "${1}=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    if [[ "${ref}" != "" ]]; then
        return 1
    fi
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
    if [[ "${INTERPRETER}" == "" ]]; then
        printLog "error" "INTERPRETER METHOD NOT SET"
        exit 1
    fi 

    if [ ! -f "${CONF_PATH}/genesis.json.istanbul.template" ]; then
        printLog "error" "File ${CONF_PATH}/genesis.json.istanbul.template NOT FOUND"
        exit 1
    fi
    if [ ! -f "${DEPLOY_CONF}" ]; then
        printLog "error" "FILE ${DEPLOY_CONF} NOT FOUND"
        exit 1
    fi
    if [ ! -d "${NODE_DIR}" ] || [ ! -f "${NODE_DIR}/node.pubkey" ] || [ ! -f "${NODE_DIR}/node.address" ] || [ ! -f "${NODE_DIR}/node.prikey" ]; then
        printLog "error" "PLEASE GENERATE NODE KEY FIRST"
        exit 1
    fi
}

################################################# Read File #################################################
function readFile() {
    checkConf "ip_addr"
    if [[ $? -eq 1 ]]; then
        IP_ADDR="$(cat ${DEPLOY_CONF} | grep "ip_addr=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    fi
    checkConf "p2p_port"
    if [[ $? -eq 1 ]]; then
        P2P_PORT="$(cat ${DEPLOY_CONF} | grep "p2p_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    fi

    NODE_ADDRESS="$(cat ${NODE_DIR}/node.address)"
}

################################################# Read Param #################################################
function readParam() {
    if [[ "${ip_addr}" != "" ]]; then
        IP_ADDR="${ip_addr}"
    fi
    if [[ "${p2p_port}" != "" ]]; then
        P2P_PORT="${p2p_port}"
    fi
}

################################################# Create Genesis #################################################
function createGenesis() {

    if [[ "${VALIDATOR_NODES}" == "" ]]; then
        node_key="$(cat ${NODE_DIR}/node.pubkey)"
        VALIDATOR_NODES="enode://${node_key}@${IP_ADDR}:${P2P_PORT}"
    fi

    now=$(date +%s)
    if [ "${OS}" = "Darwin" ]; then
        sed -i '' "s#__VALIDATOR__#${VALIDATOR_NODES}#g" "${CONF_PATH}/genesis.temp.json"
        sed -i '' "s#DEFAULT-ACCOUNT#${NODE_ADDRESS}#g" "${CONF_PATH}/genesis.temp.json"
        sed -i '' "s#__INTERPRETER__#${INTERPRETER}#g" "${CONF_PATH}/genesis.temp.json"
        sed -i '' "s#TIMESTAMP#${now}#g" "${CONF_PATH}/genesis.temp.json"
    else
        sed -i "s#__VALIDATOR__#${VALIDATOR_NODES}#g" "${CONF_PATH}/genesis.temp.json"
        sed -i "s#DEFAULT-ACCOUNT#${NODE_ADDRESS}#g" "${CONF_PATH}/genesis.temp.json"
        sed -i "s#__INTERPRETER__#${INTERPRETER}#g" "${CONF_PATH}/genesis.temp.json"
        sed -i "s#TIMESTAMP#${now}#g" "${CONF_PATH}/genesis.temp.json"  
    fi
}

################################################# Setup Genesis #################################################
function generateGenesis() {    
    ## backup genesis if exists
    if [ -f "${CONF_PATH}/genesis.json" ]; then
        timestamp=$(date '+%Y%m%d%H%M%S')
        mkdir -p "${CONF_PATH}/bak"
        mv "${CONF_PATH}/genesis.json" "${CONF_PATH}/bak/genesis.json.bak.${timestamp}"
        if [ -f "${CONF_PATH}/bak/genesis.json.bak.${timestamp}" ]; then
            printLog "info" "Backup ${CONF_PATH}/genesis.json completed"
        else
            printLog "error" "BACKUP GENESIS FILE FAILED"
            exit 1
        fi
    fi

    ## create genesis
    cp "${CONF_PATH}/genesis.json.istanbul.template" "${CONF_PATH}/genesis.temp.json"
    if [ ! -f "${CONF_PATH}/genesis.temp.json" ]; then
        printLog "error" "GENERATE GENESIS TEMP FILE FAILED"
        exit 1
    fi
    createGenesis
    mv "${CONF_PATH}/genesis.temp.json" "${CONF_PATH}/genesis.json"
    if [ ! -f "${CONF_PATH}/genesis.json" ]; then
        printLog "error" "GENERATE GENESIS FILE FAILED"
        exit 1
    fi

    printLog "info" "File: ${CONF_PATH}/genesis.json"
    printLog "info" "Genesis: "
    cat "${CONF_PATH}/genesis.json"
}

################################################# Main #################################################
function main() {
    printLog "info" "## Generate Genesis Start ##"
    checkEnv
    readFile
    readParam

    generateGenesis 
    printLog "success" "Generate genesis succeeded"
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
    --interpreter | -i)
        shiftOption2 $#
        INTERPRETER="${2}"
        shift 2
        ;;
    --validatorNodes | -v)
        shiftOption2 $#
        VALIDATOR_NODES="${2}"
        shift 2
        ;;
    --ip)
        shiftOption2 $#
        ip_addr="${2}"
        shift 2
        ;;
    --p2p_port | -p2p)
        p2p_port="${2}"
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
