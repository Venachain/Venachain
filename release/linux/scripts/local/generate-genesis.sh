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
    cd ../../
    pwd
)
BIN_PATH=${PROJECT_PATH}/bin
DATA_PATH=${PROJECT_PATH}/data
CONF_PATH=${PROJECT_PATH}/conf

## global
NODE_ID=""
NODE_DIR=""
IP_ADDR=""
P2P_PORT=""
VALIDATOR_NODES=""
INTERPRETER=""
AUTO=""
NODE_KEY=""
NODE_ADDRESS=""
DEPLOY_CONF=""

## param
ip_addr=""
p2p_pory=""
validator_nodes=""
interpreter=""

#############################################################################################################
################################################# FUNCTIONS #################################################
#############################################################################################################

################################################# Help #################################################
function help() {
    echo
    echo "
USAGE: ${SCRIPT_NAME}  [options] [value]

        OPTIONS:

           --nodeid, -n                 the first node id, must be specified

           --ip                         the first node ip (default: 127.0.0.1)

           --p2p_port, -p               the first node p2p_port (default: 16791)

           --validatorNodes, -v         set the genesis validatorNodes
                                        (default is the first node enode code)

           --interpreter, -i            set the genesis interpreter
                                        (default: all)

           --auto                       will skip ip check
                                        
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

################################################# Check Ip #################################################
function checkIp() {
    ip=$1
    check=$(echo $ip|awk -F. '$1<=255&&$2<=255&&$3<=255&&$4<=255{print "yes"}')
    if echo $ip|grep -E "^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$">/dev/null; then
        if [ ${check:-no} == "yes" ]; then
            return 1
        fi
    fi
    return 0
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
    if [ ! -f "${DEPLOY_CONF}" ]; then
        printLog "error" "FILE ${DEPLOY_CONF} NOT FOUND"
        exit
    fi

    if [ ! -d "${NODE_DIR}" ] || [ ! -f "${NODE_DIR}/node.pubkey" ] || [ ! -f "${NODE_DIR}/node.address" ] || [ ! -f "${NODE_DIR}/node.prikey" ]; then
        printLog "error" "PLEASE GENERATE NODE KEY FIRST"
        exit
    fi
}

################################################# Assign Default #################################################
function assignDefault() { 
    IP_ADDR="127.0.0.1"
    P2P_PORT="16791"
    INTERPRETER="all"
}

################################################# Read File #################################################
function readFile() {
    checkConf "ip_addr"
    if [[ $? -eq 1 ]]; then
        IP_ADDR=$(cat ${DEPLOY_CONF} | grep "ip_addr=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    fi
    checkConf "p2p_port"
    if [[ $? -eq 1 ]]; then
        P2P_PORT=$(cat ${DEPLOY_CONF} | grep "p2p_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    fi
    checkConf "validator_nodes"
    if [[ $? -eq 1 ]]; then
        VALIDATOR_NODES=$(cat ${DEPLOY_CONF} | grep "validator_nodes=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    fi
    checkConf "interpreter"
    if [[ $? -eq 1 ]]; then
        INTERPRETER=$(cat ${DEPLOY_CONF} | grep "interpreter=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    fi

    NODE_KEY=$(cat ${NODE_DIR}/node.pubkey)
    NODE_ADDRESS=$(cat ${NODE_DIR}/node.address)
}

################################################# Read Param #################################################
function readParam() {
    if [[ "${ip_addr}" != "" ]]; then
        IP_ADDR="${ip_addr}"
    fi
    if [[ "${p2p_port}" != "" ]]; then
        P2P_PORT="${p2p_port}"
    fi
    if [[ "${validator_nodes}" != "" ]]; then
        VALIDATOR_NODES="${validator_nodes}"
    fi
    if [[ "${interpreter}" != "" ]]; then
        INTERPRETER="${interpreter}"
    fi
}

################################################# Create Genesis #################################################
function createGenesis() {
    if [[ "${VALIDATOR_NODES}" == "" ]]; then
        VALIDATOR_NODES="enode://${NODE_KEY}@${IP_ADDR}:${P2P_PORT}"
    fi
    now=$(date +%s)


    if [ "${OS}" = "Darwin" ]; then
        sed -i '' "s#__VALIDATOR__#\"${VALIDATOR_NODES}\"#g" ${CONF_PATH}/genesis.temp.json
        sed -i '' "s/DEFAULT-ACCOUNT/${NODE_ADDRESS}/g" "${CONF_PATH}/genesis.temp.json"
        sed -i '' "s/__INTERPRETER__/${INTERPRETER}/g" "${CONF_PATH}/genesis.temp.json"
        sed -i '' "s/TIMESTAMP/${now}/g" "${CONF_PATH}/genesis.temp.json"
    else
        now=$(date +%s)
        sed -i "s#__VALIDATOR__#\"${VALIDATOR_NODES}\"#g" "${CONF_PATH}/genesis.temp.json"
        sed -i "s#DEFAULT-ACCOUNT#${NODE_ADDRESS}#g" "${CONF_PATH}/genesis.temp.json"
        sed -i "s#__INTERPRETER__#${INTERPRETER}#g" "${CONF_PATH}/genesis.temp.json"
        sed -i "s#TIMESTAMP#${now}#g" "${CONF_PATH}/genesis.temp.json"
    fi
}

################################################# Setup Genesis #################################################
function setupGenesis() {    
    ## backup genesis if exists
    if [ -f "${CONF_PATH}/genesis.json" ]; then
        timestamp=$(date '+%Y%m%d%H%M%S')
        mkdir -p "${CONF_PATH}/bak"
        mv "${CONF_PATH}/genesis.json" "${CONF_PATH}/bak/genesis.json.bak.${timestamp}"
        if [ -f "${CONF_PATH}/bak/genesis.json.bak.${timestamp}" ]; then
            printLog "info" "Backup ${CONF_PATH}/genesis.json completed"
        else
            printLog "error" "BACKUP GENESIS FILE FAILED"
            exit
        fi
    fi

    ## create genesis
    cp "${CONF_PATH}/genesis.json.istanbul.template" "${CONF_PATH}/genesis.temp.json"
    if [ ! -f "${CONF_PATH}/genesis.temp.json" ]; then
        printLog "error" "GENERATE GENESIS TEMP FILE FAILED"
        exit
    fi
    createGenesis
    mv "${CONF_PATH}/genesis.temp.json" "${CONF_PATH}/genesis.json"
    if [ ! -f "${CONF_PATH}/genesis.json" ]; then
        printLog "error" "GENERATE GENESIS FILE FAILED"
        exit
    fi
}

################################################# Main #################################################
function main() {
    printLog "info" "## Setup Genesis Start ##"
    assignDefault
    checkEnv
    readFile
    readParam
    
    if [[ "${AUTO}" != "true" ]]; then
        checkIp "${IP_ADDR}"
        if [ $? -ne 1 ]; then
            while true
            do
                printLog "warn" "Invalid Ip!"
                printLog "question" "Please in put a valid ip: "
                read input 
                checkIp "${input}"
                if [ $? -eq 1 ]; then
                    IP_ADDR=${input}
                    break
                fi
            done
        fi
    fi
    saveConf "ip_addr" "${IP_ADDR}"
    saveConf "p2p_port" "${P2P_PORT}"

    setupGenesis 
    printLog "info" "File: ${CONF_PATH}/genesis.json"
    printLog "info" "Genesis: "
    cat ${CONF_PATH}/genesis.json
    printLog "success" "Setup genesis succeeded"
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
    --p2p_port | -p)
        p2p_port=$2
        shift 2
        ;;
    --validatorNodes | -v)
        shiftOption2 $#
        validator_nodes=$2
        shift 2
        ;;
    --interpreter | -i)
        shiftOption2 $#
        interpreter=$2
        shift 2
        ;;
    --auto)
        AUTO="true"
        shift 1
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
