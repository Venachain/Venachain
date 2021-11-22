#!/bin/bash

###########################################################################################################
################################################# VRIABLES #################################################
###########################################################################################################
SCRIPT_NAME="$(basename ${0})"
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
AUTO=""
VALIDATOR_NODES=""
INTERPRETER="all"

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

           --nodeid, -n                 the first node id 

           --validatorNodes, -v         set the genesis validatorNodes
                                        (default is the first node enode code)

           --interpreter, -i            set the genesis interpreter
                                        default='all'

           --auto                   auto=true: will cover exist genesis file 
            　　　　　　 default='false'
                                        
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

################################################# Read File #################################################
function readFile() {
    file=$1
    IP_ADDR=$(cat $1 | grep "ip_addr=" | sed -e 's/ip_addr=\(.*\)/\1/g')
    P2P_PORT=$(cat $1 | grep "p2p_port=" | sed -e 's/p2p_port=\(.*\)/\1/g')
    RPC_PORT=$(cat $1 | grep "rpc_port=" | sed -e 's/rpc_port=\(.*\)/\1/g')

    if [[ "${IP_ADDR}" == "" ]] || [[ "${P2P_PORT}" == "" ]] || [[ "${RPC_PORT}" == "" ]]; then
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

    if [[ $VALIDATOR_NODES != "" ]]; then
        replaceList "__VALIDATOR__" $VALIDATOR_NODES
    else
        replaceList "__VALIDATOR__" $1
    fi
    "${BIN_PATH}/repstr" "${CONF_PATH}/genesis.temp.json" "DEFAULT-ACCOUNT" "$2"
    "${BIN_PATH}/repstr" "${CONF_PATH}/genesis.temp.json" "__INTERPRETER__" "$3"

    now=$(date +%s)
    "${BIN_PATH}/repstr" "${CONF_PATH}/genesis.temp.json" "TIMESTAMP" "${now}"
}

################################################# Create Genesis (MAC OS) #################################################
function createGenesisDarwin() {
    if [[ $VALIDATOR_NODES != "" ]]; then
        sed -i '' "s#__VALIDATOR__#\"${VALIDATOR_NODES}\"#g" ${CONF_PATH}/genesis.temp.json
    else
        sed -i '' "s#__VALIDATOR__#\"${1}\"#g" ${CONF_PATH}/genesis.temp.json
    fi
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
    file="${NODE_DIR}/deploy_node-${NODE_ID}.conf"
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
    cp "${CONF_PATH}/genesis.json.istanbul.template" "${CONF_PATH}/genesis.temp.json"
    if [ ! -f "${CONF_PATH}/genesis.temp.json" ]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* GENERATE GENESIS TEMP FILE FAILED **********"
        exit
    fi
    createGenesis "${default_enode}" "${node_addr}" "${INTERPRETER}"
    mv "${CONF_PATH}/genesis.temp.json" "${CONF_PATH}/genesis.json"
    if [ ! -f "${CONF_PATH}/genesis.json" ]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* GENERATE GENESIS FILE FAILED **********"
        exit
    fi
}

################################################# Save Firstnode Info #################################################
function saveFirstnodeInfo() {
    {
        echo "user_name=$(whoami)"
        echo "node_id=${NODE_ID}"
        echo "ip_addr=${IP_ADDR}"
        echo "rpc_port=${RPC_PORT}"
    } >"${CONF_PATH}/firstnode.info"
    if [[ ! -f "${CONF_PATH}/firstnode.info" ]] || [[ "$(cat ${CONF_PATH}/firstnode.info)" == "" ]]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* SETUP FIRSTNODE INFO FAILED **********"
        exit
    fi
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : Setup firstnode info completed"
}

################################################# Main #################################################
function main() {
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ## Setup Genesis Start ##"
    echo "${CONF_PATH}/genesis.json"
    if [ -f ${CONF_PATH}/genesis.json ] && [[ "${AUTO}" != true ]]; then
        echo "Genesis already exists, re create? Yes or No(y/n):"
        yesOrNo
        if [ $? -ne 0 ]; then
            exit
        fi
    fi
    setupGenesis "${file}"
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : File: ${CONF_PATH}/genesis.json"
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : Genesis:"
    cat ${CONF_PATH}/genesis.json
    saveFirstnodeInfo
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
    --nodeid | -n)
        shiftOption2 $#
        NODE_ID=$2
        NODE_DIR="${DATA_PATH}/node-$2"

        if [ ! -f "${NODE_DIR}/deploy_node-$2.conf" ]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* FILE ${NODE_DIR}/deploy_node-$2.conf NOT FOUND **********"
            exit
        fi
        shift 2
        ;;
    --validatorNodes | -v)
        shiftOption2 $#
        VALIDATOR_NODES=$2
        shift 2
        ;;
    --interpreter | -i)
        shiftOption2 $#
        INTERPRETER=$2
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
