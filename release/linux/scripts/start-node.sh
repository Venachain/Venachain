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
IP_ADDR=""
P2P_PORT=""
BOOTNODES=""
RPC_ADDR=""
RPC_PORT=""
RPC_API=""
WS_ADDR=""
WS_PORT=""
LOG_SIZE=""
LOG_DIR=""
GCMODE=""
LIGHTMODE=""
DBTYPE=""
TX_COUNT=""
TX_GLOBAL_SLOTS=""
EXTRA_OPTIONS=""
PPROF_ADDR=""

#############################################################################################################
################################################# FUNCTIONS #################################################
#############################################################################################################

################################################# Help #################################################
function help() {
    echo
    echo "
USAGE: ${SCRIPT_NAME}  [options] [value]

        OPTIONS:
 
        --nodeid, -n                 start the specified node (default: 0)

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
    checkConf "ws_port"
    if [[ $? -ne 1 ]]; then
        printLog "error" "NODE'S WEBSOCKET PORT HAVE NOT BEEN SET"
        exit
    fi
}

################################################# Assign Default #################################################
function assignDefault() {
        IP_ADDR=127.0.0.1
        P2P_PORT=16791
   
        RPC_ADDR=0.0.0.0
        RPC_API=db,eth,platone,net,web3,admin,personal,txpool,istanbul
        RPC_PORT=6791

        WS_ADDR=0.0.0.0
        WS_PORT=26791
 
        LOG_SIZE=67108864
        LOG_DIR=${NODE_DIR}/logs

        TX_COUNT=1000
        TX_GLOBAL_SLOTS=4096
  
        GCMODE=archive
        DBTYPE=leveldb
        EXTRA_OPTIONS=--debug
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
    checkConf "bootnodes"
    if [[ $? -eq 1 ]]; then
        BOOTNODES=$(cat ${DEPLOY_CONF} | grep "bootnodes=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    fi

    checkConf "rpc_addr"
    if [[ $? -eq 1 ]]; then
        RPC_ADDR=$(cat ${DEPLOY_CONF} | grep "rpc_addr=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    fi
    checkConf "rpc_api"
    if [[ $? -eq 1 ]]; then
        RPC_API=$(cat ${DEPLOY_CONF} | grep "rpc_api=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    fi
    checkConf "rpc_port"
    if [[ $? -eq 1 ]]; then
        RPC_PORT=$(cat ${DEPLOY_CONF} | grep "rpc_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    fi

    checkConf "ws_addr"
    if [[ $? -eq 1 ]]; then
        WS_ADDR=$(cat ${DEPLOY_CONF} | grep "ws_addr=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    fi
    checkConf "ws_port"
    if [[ $? -eq 1 ]]; then
        WS_PORT=$(cat ${DEPLOY_CONF} | grep "ws_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    fi

    checkConf "log_size"
    if [[ $? -eq 1 ]]; then
        LOG_SIZE=$(cat ${DEPLOY_CONF} | grep "log_size=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    fi
    checkConf "log_dir"
    if [[ $? -eq 1 ]]; then
        LOG_DIR=$(cat ${DEPLOY_CONF} | grep "log_dir=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    fi

    checkConf "tx_count"
    if [[ $? -eq 1 ]]; then
        TX_COUNT=$(cat ${DEPLOY_CONF} | grep "tx_count=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    fi
    checkConf "tx_global_slots"
    if [[ $? -eq 1 ]]; then
        TX_GLOBAL_SLOTS=$(cat ${DEPLOY_CONF} | grep "tx_global_slots=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    fi
    
    checkConf "gcmode"
    if [[ $? -eq 1 ]]; then
        GCMODE=$(cat ${DEPLOY_CONF} | grep "gcmode=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    fi
    checkConf "lightmode"
    if [[ $? -eq 1 ]]; then
        LIGHTMODE=$(cat ${DEPLOY_CONF} | grep "lightmode=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    fi
    checkConf "dbtype"
    if [[ $? -eq 1 ]]; then
        DBTYPE=$(cat ${DEPLOY_CONF} | grep "dbtype=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    fi
    checkConf "extra_options"
    if [[ $? -eq 1 ]]; then
        EXTRA_OPTIONS=$(cat ${DEPLOY_CONF} | grep "extra_options=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    fi
    checkConf "pprof_addr"
    if [[ $? -eq 1 ]]; then
        PPROF_ADDR=$(cat ${DEPLOY_CONF} | grep "pprof_addr=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    fi
}

################################################# Start Command #################################################
function startCmd() {
    if [[ "${BOOTNODES}" == "" ]]; then
        if [[ -f ${CONF_PATH}/genesis.json ]]; then
            BOOTNODES=$(cat ${CONF_PATH}/genesis.json | sed -n '9p' | sed 's/^.*"\(firstValidatorNode\)": "\(.*\)"/\2/g')
        else
            printLog "error" "FILE ${CONF_PATH}/genesis.json NOT FOUND"
        fi
    fi

    ## check p2p port
    if [[ "$(lsof -i:${P2P_PORT})" != "" ]]; then
        printLog "error" "PORT ${P2P_PORT} IS IN USAGE"
        exit
    fi

    ## create log dir
    mkdir -p "${LOG_DIR}"
    if [ ! -d "${LOG_DIR}" ]; then
        printLog "error" "CREATE ${LOG_DIR} FAILED"
        exit
    fi

    ## backup node's log if exists
    timestamp=$(date '+%Y%m%d%H%M%S')
    if [ -f "${LOG_DIR}/node-${NODE_ID}.log" ]; then
        mv "${LOG_DIR}/node-${NODE_ID}.log" "${LOG_DIR}/node-${NODE_ID}.log.bak.${timestamp}"
        if [ ! -f "${LOG_DIR}/node-${NODE_ID}.log.bak.${timestamp}" ]; then
            printLog "error" "BACKUP {LOG_DIR}/node-${NODE_ID}.log FAILED"
            exit
        fi
    fi

    ## generate command segments
    flag_node=" --identity platone --datadir ${NODE_DIR}"
    flag_discov=" --nodiscover --port ${P2P_PORT} --nodekey ${NODE_DIR}/node.prikey"
    flag_bootnodes=" --bootnodes ${BOOTNODES} "
    flag_rpc=" --rpc --rpcaddr ${RPC_ADDR} --rpcport ${RPC_PORT} --rpcapi ${RPC_API} --rpccorsdomain \"*\" "
    flag_ws=" --ws --wsaddr ${WS_ADDR} --wsport ${WS_PORT} --wsorigins \"*\" "
    flag_logs=" --wasmlog  ${LOG_DIR}/wasm_log --wasmlogsize ${LOG_SIZE} "
    flag_ipc=" --ipcpath ${NODE_DIR}/node-${NODE_ID}.ipc "
    flag_gcmode=" --gcmode ${GCMODE} "
    flag_dbtype=" --dbtype ${DBTYPE} "
    flag_tx=" --txpool.globaltxcount ${TX_COUNT} --txpool.globalslots ${TX_GLOBAL_SLOTS} "

    # lightnode mode
    flag_lightmode=""
    if [[ "${LIGHTMODE}" == "lightnode" ]]; then
        flag_lightmode="--syncmode light"
    elif [[ "${LIGHTMODE}" == "lightserver" ]]; then
        flag_lightmode="--lightserv 70"
    fi

    # include pprof if setted
    flag_pprof=""
    if [[ "${PPROF_ADDR}" != "" ]]; then
        flag_pprof=" --pprof --pprofaddr ${PPROF_ADDR} "
    fi

    ## execute command
    printLog "info" "Exec command: "
    echo "
        nohup ${BIN_PATH}/platone ${flag_node} 
            ${flag_discov}
            ${flag_bootnodes} 
            ${flag_rpc}
            ${flag_ws}
            ${flag_logs} 
            ${flag_ipc} ${flag_gcmode} ${flag_dbtype}
            ${flag_tx} ${flag_lightmode} ${flag_pprof}
             --moduleLogParams '{\"platone_log\": [\"/\"], \"__dir__\": [\"'${LOG_DIR}'\"], \"__size__\": [\"'${LOG_SIZE}'\"]}'
             ${EXTRA_OPTIONS}
             1>/dev/null 2>${LOG_DIR}/platone_error.log &
    "
    nohup ${BIN_PATH}/platone ${flag_node} \
            ${flag_discov} \
            ${flag_bootnodes} \
            ${flag_rpc} \
            ${flag_ws} \
            ${flag_logs} \
            ${flag_ipc} ${flag_gcmode} ${flag_dbtype} \
            ${flag_tx} ${flag_lightmode} ${flag_pprof} \
             --moduleLogParams '{"platone_log": ["/"], "__dir__": ["'${LOG_DIR}'"], "__size__": ["'${LOG_SIZE}'"]}' \
             ${EXTRA_OPTIONS} \
             1>/dev/null 2>${LOG_DIR}/platone_error.log &

    timer=0
    res_start=""
    while [ ${timer} -lt 10 ]; do
        res_start=$(lsof -i:${P2P_PORT})
        if [[ "${res_start}" != "" ]]; then
            break
        fi
        sleep 1
        let timer++
    done
    if [[ "${res_start}" == "" ]]; then
        printLog "error" "RUN NODE NODE-${NODE_ID} FAILED"
        exit
    fi
}

################################################# Main #################################################
function main() {
    if [[ "${NODE_ID}" == "" ]]; then
        NODE_ID="0"
        NODE_DIR="${DATA_PATH}/node-${NODE_ID}"
        DEPLOY_CONF="${NODE_DIR}/deploy_node-${NODE_ID}.conf"
    fi

    printLog "info" "## Run node-${NODE_ID} ##"
    checkEnv
    assignDefault
    readFile
    startCmd
    printLog "info" "Node's url: ${IP_ADDR}:${RPC_PORT}"
    printLog "success" "Run node-${NODE_ID} succeeded"
}

###########################################################################################################
#################################################  EXECUTE #################################################
###########################################################################################################
while [ ! $# -eq 0 ]; do
    case "$1" in
    --nodeid | -n)
        shiftOption2 $#
        NODE_ID="${2}"
        NODE_DIR="${DATA_PATH}/node-${NODE_ID}"
        DEPLOY_CONF="${NODE_DIR}/deploy_node-${NODE_ID}.conf"
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
