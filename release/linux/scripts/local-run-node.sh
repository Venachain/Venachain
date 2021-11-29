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

NODE_DIR=""
IP_ADDR=""
P2P_PORT=""
RPC_ADDR=""
RPC_PORT=""
RPC_API=""
WS_ADDR=""
WS_PORT=""
LOG_SIZE=""
LOG_DIR=""
GCMODE=""
TX_COUNT=""
TX_GLOBAL_SLOTS=""
LIGHTMODE=""
BOOTNODES=""
EXTRA_OPTIONS=""
PPROF_ADDR=""

# TODO
DBTYPE="leveldb"
#DBTYPE="pebbledb"

#############################################################################################################
################################################# FUNCTIONS #################################################
#############################################################################################################

################################################# Help #################################################
function help() {
    echo
    echo "
USAGE: ${SCRIPT_NAME}  [options] [value]

        OPTIONS:

           --node, -n                   the specified node name. must be specified

           --lightmode                  light node mode
                                        option: lightnode, lightserver or ''

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

    file=${NODE_DIR}/deploy_node-"${NODE_ID}".conf
    IP_ADDR=$(cat $file | grep "ip_addr=" | sed -e 's/ip_addr=\(.*\)/\1/g')
    P2P_PORT=$(cat $file | grep "p2p_port=" | sed -e 's/p2p_port=\(.*\)/\1/g')
    RPC_ADDR=$(cat $file | grep "rpc_addr=" | sed -e 's/rpc_addr=\(.*\)/\1/g')
    RPC_PORT=$(cat $file | grep "rpc_port=" | sed -e 's/rpc_port=\(.*\)/\1/g')
    RPC_API=$(cat $file | grep "rpc_api=" | sed -e 's/rpc_api=\(.*\)/\1/g')
    WS_ADDR=$(cat $file | grep "ws_addr=" | sed -e 's/ws_addr=\(.*\)/\1/g')
    WS_PORT=$(cat $file | grep "ws_port=" | sed -e 's/ws_port=\(.*\)/\1/g')
    LOG_SIZE=$(cat $file | grep "log_size=" | sed -e 's/log_size=\(.*\)/\1/g')
    LOG_DIR=$(cat $file | grep "log_dir=" | sed -e 's/log_dir=\(.*\)/\1/g')
    GCMODE=$(cat $file | grep "gcmode=" | sed -e 's/gcmode=\(.*\)/\1/g')
    TX_COUNT=$(cat $file | grep "txcount=" | sed -e 's/txcount=\(.*\)/\1/g')
    TX_GLOBAL_SLOTS=$(cat $file | grep "tx_global_slots=" | sed -e 's/tx_global_slots=\(.*\)/\1/g')

    BOOTNODES=$(cat $file | grep "bootnodes=" | sed -e 's/bootnodes=\(.*\)/\1/g')
    if [[ "${BOOTNODES}" == "" ]]; then
        if [[ -f ${CONF_PATH}/genesis.json ]]; then
            BOOTNODES=$(cat ${CONF_PATH}/genesis.json | sed -n '9p' | sed 's/^.*"\(firstValidatorNode\)": "\(.*\)"/\2/g')
        else
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ************** FILE ${CONF_PATH}/genesis.json NOT FOUND ***************"
        fi
    fi
    EXTRA_OPTIONS=$(cat $file | grep "extra_options=" | sed -e 's/extra_options=\(.*\)/\1/g')
    PPROF_ADDR=$(cat $file | grep "pprof_addr=" | sed -e 's/pprof_addr=\(.*\)/\1/g')

    if [[ "${IP_ADDR}" == "" ]] || [[ "${P2P_PORT}" == "" ]] || [[ "${RPC_ADDR}" == "" ]] || [[ "${RPC_PORT}" == "" ]] || [[ "${RPC_API}" == "" ]] || [[ "${WS_ADDR}" == "" ]] || [[ "${WS_PORT}" == "" ]] || [[ "${LOG_SIZE}" == "" ]] || [[ "${LOG_DIR}" == "" ]] || [[ "${GCMODE}" == "" ]] || [ "${BOOTNODES}" == "" ]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* FILE ${file} MISS VALUE **********"
        exit
    fi
}

################################################# Start Command #################################################
function startCmd() {
    ## generate command segments
    flag_datadir="--datadir ${NODE_DIR}"
    flag_nodekey="--nodekey ${NODE_DIR}/node.prikey"
    flag_rpc="--rpc --rpcaddr ${RPC_ADDR} --rpcport ${RPC_PORT}  --rpcapi ${RPC_API} "
    flag_ws="--ws --wsaddr ${WS_ADDR} --wsport ${WS_PORT} "
    flag_logs="--wasmlog  ${LOG_DIR}/wasm_log --wasmlogsize ${LOG_SIZE} "
    flag_ipc="--ipcpath ${NODE_DIR}/node-${NODE_ID}.ipc "
    flag_gcmode="--gcmode  ${GCMODE} "
    flag_txcount=" --txpool.globaltxcount ${TX_COUNT} "
    flag_tx_global_slots=" --txpool.globalslots ${TX_GLOBAL_SLOTS}"

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

    ## backup node's log if exists
    timestamp=$(date '+%Y%m%d%H%M%S')
    if [ -f "${LOG_DIR}/node-${NODE_ID}.log" ]; then
        mkdir -p "${LOG_DIR}/bak"
        mv "${LOG_DIR}/node-${NODE_ID}.log" "${LOG_DIR}/bak/node-${NODE_ID}.log.bak.${timestamp}"
        if [ ! -f "${LOG_DIR}/bak/node-${NODE_ID}.log.bak.${timestamp}" ]; then
            echo "[ERROR] $(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g') : ********* BACKUP LOG DIR FAILED **********"
            exit
        fi
    fi

    ## create log dir
    mkdir -p "${LOG_DIR}"
    if [ ! -d "${LOG_DIR}" ]; then
        echo "[ERROR] $(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g') : ********* CREATE LOG DIR FAILED **********"
        exit
    fi

    ## execute command
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : Exec command: "
    echo "
        nohup ${BIN_PATH}/platone --identity platone ${flag_datadir} --nodiscover 
            --port ${P2P_PORT} ${flag_nodekey} 
            ${flag_rpc} --rpccorsdomain \""*"\" 
            ${flag_ws} --wsorigins \""*"\" 
            ${flag_logs} ${flag_ipc} 
            --bootnodes ${BOOTNODES} 
            --verbosity 3 
            --dbtype ${DBTYPE} 
            --moduleLogParams '{\"platone_log\": [\"/\"], \"__dir__\": [\"'${LOG_DIR}'\"], \"__size__\": [\"'${LOG_SIZE}'\"]}' ${flag_gcmode} ${EXTRA_OPTIONS} 
            ${flag_lightmode} ${flag_txcount} ${flag_tx_global_slots}
            1>/dev/null 2>${LOG_DIR}/platone_error.log &
    "
    nohup ${BIN_PATH}/platone --identity platone ${flag_datadir} --nodiscover \
        --port ${P2P_PORT} ${flag_nodekey} \
        ${flag_rpc} --rpccorsdomain "*" \
        ${flag_ws} --wsorigins "*" \
        ${flag_logs} ${flag_ipc} \
        --bootnodes ${BOOTNODES} \
        --verbosity 3 \
        --dbtype ${DBTYPE} \
        --moduleLogParams '{"platone_log": ["/"], "__dir__": ["'${LOG_DIR}'"], "__size__": ["'${LOG_SIZE}'"]}'  ${flag_gcmode}  ${EXTRA_OPTIONS} \
        ${flag_lightmode} ${flag_txcount} ${flag_tx_global_slots} \
        ${flag_pprof} \
        1>/dev/null 2>${LOG_DIR}/platone_error.log &

    timer=0
    start_succ_flag=""
    while [ ${timer} -lt 10 ]; do
        start_succ_flag=$(lsof -i:${P2P_PORT})
        if [[ "${start_succ_flag}" != "" ]]; then
            break
        fi
        sleep 1
        let timer++
    done
    if [[ "${start_succ_flag}" == "" ]]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* RUN NODE NODE-${NODE_ID} FAILED **********"
        exit
    fi
}

################################################# Main #################################################
function main() {
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ## Run node-${NODE_ID} ##"
    readFile
    if [[ "$(lsof -i:${P2P_PORT})" != "" ]]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* PORT ${P2P_PORT} IS IN USAGE **********"
        exit
    fi
    startCmd
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : Node's url: ${IP_ADDR}:${RPC_PORT}"
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : Run node-${NODE_ID} succeeded"
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
        NODE_DIR=${DATA_PATH}/node-$2
        NODE_ID=$2
        NODE_DIR="${DATA_PATH}/node-$2"

        if [ ! -f "${NODE_DIR}/deploy_node-$2.conf" ]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/local-\(.*\).sh/\2/g')] : ********* ${NODE_DIR}/deploy_node-$2.conf NOT FOUND **********"
            exit
        fi
        ;;
    --lightmode)
        LIGHTMODE=$2
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
