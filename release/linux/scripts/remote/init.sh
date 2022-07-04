#!/bin/bash

###########################################################################################################
################################################# VRIABLES #################################################
###########################################################################################################

## path
CURRENT_PATH="$(cd "$(dirname "$0")";pwd)"
DEPLOYMENT_PATH="$(
    cd $(dirname ${0})
    cd ../../../
    pwd
)"
DEPLOYMENT_CONF_PATH="${DEPLOYMENT_PATH}/deployment_conf"

## global
SCRIPT_NAME="$(basename ${0})"
SCRIPT_ALIAS="$(echo ${CURRENT_PATH}/${SCRIPT_NAME} | sed -e 's/\(.*\)\/scripts\/\(.*\).sh/\2/g')"
PROJECT_NAME=""
INTERPRETER=""
VALIDATOR_NODES=""
NODE=""
ALL=""

PROJECT_CONF_PATH=""

NODE_ID=""
DEPLOY_PATH=""
DEPLOY_CONF=""
USER_NAME=""
IP_ADDR=""
P2P_PORT=""
IS_LOCAL=""

#############################################################################################################
################################################# FUNCTIONS #################################################
#############################################################################################################

################################################# Help #################################################
function help() {
    echo
    echo "
USAGE: ${SCRIPT_NAME}  [options] [value]

        OPTIONS:

            --project, -p               the specified project name, must be specified

            --interpreter, -i           Select virtual machine interpreter, must be specified for new project
                                        \"wasm\", \"evm\" and \"all\" are supported

            --validatorNodes, -v        set the genesis validatorNodes, must be specified for new project
                                        (default: the first node enode code)

            --node, -n                  the specified node name
                                        use \",\" to seperate the name of node

            --all, -a                   init all nodes

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

################################################# Execute Command #################################################
function xcmd() {
    address="${1}"
    cmd="${2}"
    scp_param="${3}"

    if [[ "${IS_LOCAL}" == "true" ]]; then
        eval "${cmd}"
        return $?
    elif [[ "$(echo "${cmd}" | grep "cp")" == "" ]]; then
        ssh "${address}" "${cmd}"
        return $?
    else
        source_path="$(echo ${cmd} | sed -e 's/\(.*\)cp -r \(.*\) \(.*\)/\2/g')"
        target_path="$(echo ${cmd} | sed -e 's/\(.*\)cp -r \(.*\) \(.*\)/\3/g')"
        if [[ "${scp_param}" == "source" ]]; then
            scp -r "${address}:${source_path}" "${target_path}"
        elif [[ "${scp_param}" == "target" ]]; then
            scp -r "${source_path}" "${address}:${target_path}"
        else
            return 1
        fi
        return $?
    fi
}

################################################# Check Remote Access #################################################
function checkRemoteAccess() {
    if [[ "${IP_ADDR}" == "127.0.0.1" ]] || [[ $(ifconfig | grep "\<${IP_ADDR}\>") != "" ]]; then
        IS_LOCAL="true"
        return 0
    fi
    IS_LOCAL="false"

    ## check ip connection
    ping -c 3 -w 3 "${IP_ADDR}" 1>/dev/null 
    if [[ $? -eq 1 ]]; then
        printLog "error" "${IP_ADDR} IS DOWN"
        return 1
    fi
    printLog "info" "Check ip ${IP_ADDR} connection completed"

    ## check ssh connection
    res=$(timeout 3 ssh "${USER_NAME}@${IP_ADDR}" echo "permission")
    if [[ "${res}" != "permission" ]]; then
        printLog "error" "${USER_NAME}@${IP_ADDR} DO NOT SUPPORT PASSWORDLESS ACCCESS"
        return 1
    fi
    printLog "info" "Check ssh ${USER_NAME}@${IP_ADDR} access completed"
}

################################################# Check Env #################################################
function checkEnv() {
    PROJECT_CONF_PATH="${DEPLOYMENT_CONF_PATH}/projects/${PROJECT_NAME}"

    if [[ "${PROJECT_NAME}" == "" ]]; then
        printLog "error" "PROJECT NAME NOT SET"
        exit 1
    fi
    if [[ "${ALL}" != "true" ]] && [[ "${NODE}" == "" ]]; then
        printLog "error" "NODE NAME NOT SET"
        exit 1
    fi
    if [ ! -f "${PROJECT_CONF_PATH}/global/genesis.json" ] && [[ "${INTERPRETER}" == "" ]]; then
        printLog "error" "INTERPRETER NOT SET FOR GENERATING GENESIS"
        exit 1
    fi
    if [ ! -d "${PROJECT_CONF_PATH}" ]; then
        printLog "error" "DIRECTORY ${PROJECT_CONF_PATH} NOT FOUND"
        exit 1
    fi

    if [  -f "${PROJECT_CONF_PATH}/global/genesis.json" ] && [[ "${INTERPRETER}" != "" ]]; then
        printLog "warn" "Interpreter Param Only Firstnode Take Effect"
    fi
    if [  -f "${PROJECT_CONF_PATH}/global/genesis.json" ] && [[ "${VALIDATOR_NODES}" != "" ]]; then
        printLog "warn" "ValidatorNodes Param Only Firstnode Take Effect"
    fi
}

################################################# Clear Data #################################################
function clearData() {
    NODE_ID=""
    DEPLOY_PATH=""
    DEPLOY_CONF=""
    USER_NAME=""
    IP_ADDR=""
    P2P_PORT=""
    IS_LOCAL=""
}

################################################# Read File #################################################
function readFile() {
    DEPLOY_PATH="$(cat ${DEPLOY_CONF} | grep "deploy_path=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    IP_ADDR="$(cat ${DEPLOY_CONF} | grep "ip_addr=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    USER_NAME="$(cat ${DEPLOY_CONF} | grep "user_name=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    P2P_PORT="$(cat ${DEPLOY_CONF} | grep "p2p_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    RPC_PORT="$(cat ${DEPLOY_CONF} | grep "rpc_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"

    if [[ "${DEPLOY_PATH}" == "" ]] || [[ "${IP_ADDR}" == "" ]] || [[ "${USER_NAME}" == "" ]] || [[ "${P2P_PORT}" == "" ]] || [[ "${RPC_PORT}" == "" ]]; then
        printLog "error" "KEY INFO NOT SET IN ${DEPLOY_CONF}"
        return 1
    fi
}

################################################# Key Operation #################################################
function keyOperation() {
    ## generate key
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ "$(grep "${SCRIPT_ALIAS}" "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "node-${NODE_ID}" | grep "Generate key")" == "" ]]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "${script_path}/venachainctl.sh keygen -n ${NODE_ID} --auto"
        xcmd "${USER_NAME}@${IP_ADDR}" "[ -f ${DEPLOY_PATH}/data/node-${NODE_ID}/node.address -a -f ${DEPLOY_PATH}/data/node-${NODE_ID}/node.prikey -a -f ${DEPLOY_PATH}/data/node-${NODE_ID}/node.pubkey ]"
        if [[ $? -eq 1 ]]; then
            printLog "error" "GENERATE KEY FOR NODE-${NODE_ID} FAILED"
            return 1
        fi
        echo "[${SCRIPT_ALIAS}] [node-${NODE_ID}] : Generate key for node-${NODE_ID} completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        printLog "info" "Generate key for node-${NODE_ID} completed"
    fi

    ## get key
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ "$(grep "${SCRIPT_ALIAS}" "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "node-${NODE_ID}" | grep "Get key")" == "" ]]; then
        mkdir -p "${PROJECT_CONF_PATH}/global/data/node-${NODE_ID}"

        xcmd "${USER_NAME}@${IP_ADDR}" "cp -r ${data_path}/node-${NODE_ID}/node.pubkey ${PROJECT_CONF_PATH}/global/data/node-${NODE_ID}/node.pubkey" "source"
        if [[ ! -f "${PROJECT_CONF_PATH}/global/data/node-${NODE_ID}/node.pubkey" ]]; then
            printLog "error" "GET KEY FOR NODE-${NODE_ID} FAILED"
            return 1
        fi
        echo "[${SCRIPT_ALIAS}] [node-${NODE_ID}] : Get key from node-${NODE_ID} completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        printLog "info" "Get key from node-${NODE_ID} completed"
    fi
}

################################################# Init Node #################################################
function initNode() {
    clearData
    DEPLOY_CONF="${1}"
    NODE_ID="$(echo ${DEPLOY_CONF} | sed -e 's/\(.*\)deploy_node-\(.*\).conf/\2/g')"
    echo
    echo "################ Init Node-${NODE_ID} ################"
    if [ -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] && [[ $(grep $(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g') "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "${NODE_ID}" | grep "Init node") != "" ]]; then
        return 0
    fi

    ## prepare init
    readFile
    if [[ $? -eq 1 ]]; then
        printLog "error" "READ FILE ${DEPLOY_CONF} FAILED"
        return 1
    fi
    checkRemoteAccess 
    if [[ $? -eq 1 ]]; then
        printLog "error" "CHECK REMOTE ACCESS TO NODE-${NODE_ID} FAILED"
        return 1
    fi
    script_path="${DEPLOY_PATH}/scripts"
    conf_path="${DEPLOY_PATH}/conf"
    bin_path="${DEPLOY_PATH}/bin"
    data_path="${DEPLOY_PATH}/data"

    ## key operation for first node
    if [ ! -f "${PROJECT_CONF_PATH}/global/genesis.json" ]; then
        keyOperation
    fi

    ## sync genesis file
    if [ ! -f "${PROJECT_CONF_PATH}/global/genesis.json" ]; then
        ## setup genesis file
        if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ "$(grep "${SCRIPT_ALIAS}" "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "Setup genesis")" == "" ]]; then
            flag_validator_nodes=""
            if [[ "${VALIDATOR_NODES}" != "" ]]; then
                flag_validator_nodes="--validatorNodes ${VALIDATOR_NODES}"
            fi
            xcmd "${USER_NAME}@${IP_ADDR}" "${script_path}/venachainctl.sh gengen --nodeid ${NODE_ID} --interpreter ${INTERPRETER} ${flag_validator_nodes}"
            xcmd "${USER_NAME}@${IP_ADDR}" "[ -f ${DEPLOY_PATH}/conf/genesis.json ]"
            if [[ $? -eq 1 ]]; then
                printLog "error" "SETUP GENESIS FILE FAILED"
                return 1
            fi
            echo "[${SCRIPT_ALIAS}] [node-${NODE_ID}] : Setup genesis file completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
            printLog "info" "Setup genesis file completed"
        fi

        ## get genesis file
        if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ "$(grep "${SCRIPT_ALIAS}" "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "Get genesis file")" == "" ]]; then
            xcmd "${USER_NAME}@${IP_ADDR}" "cp -r ${conf_path}/genesis.json ${PROJECT_CONF_PATH}/global/genesis.json" "source"
            if [ ! -f "${PROJECT_CONF_PATH}/global/genesis.json" ]; then
                printLog "error" "GET GENESIS FILE FAILED"
                return 1
            fi
            echo "[${SCRIPT_ALIAS}] [node-${NODE_ID}] : Get genesis file completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
            echo "[${SCRIPT_ALIAS}] [node-${NODE_ID}] : Send genesis file completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
            printLog "info" "Get genesis file completed"
        fi
    else
        ## send genesis file
        if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ "$(grep "${SCRIPT_ALIAS}" "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "node-${NODE_ID}" | grep "Send genesis file")" == "" ]]; then
            xcmd "${USER_NAME}@${IP_ADDR}" "cp -r ${PROJECT_CONF_PATH}/global/genesis.json ${conf_path}" "target"
            xcmd "${USER_NAME}@${IP_ADDR}" "[ -f ${conf_path}/genesis.json ]"
            if [[ $? -eq 1 ]]; then
                printLog "error" "SEND GENESIS FILE TO NODE_${NODE_ID} FAILED"
                return 1
            fi
            echo "[${SCRIPT_ALIAS}] [node-${NODE_ID}] : Send genesis file to node-${NODE_ID} completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
            printLog "info" "Send genesis file to node-${NODE_ID} completed"
        fi
    fi

    ## key operation for other nodes
    keyOperation

    ## init genesis
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ "$(grep "${SCRIPT_ALIAS}" "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "node-${NODE_ID}" | grep "Init genesis")" == "" ]]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "rm -rf ${data_path}/node-${NODE_ID}/venachain/*"
        if [[ $? -eq 1 ]]; then
            printLog "error" "INIT GENESIS ON NODE_${NODE_ID} FAILED"
            return 1
        fi
        xcmd "${USER_NAME}@${IP_ADDR}" "[ -f ${bin_path}/venachain ]"
        if [[ $? -eq 1 ]]; then
            printLog "error" "FILE ${USER_NAME}@${IP_ADDR}:${bin_path}/venachain NOT FOUND"
        fi

        echo "******************************************************************************************************************************************************************************"
        xcmd "${USER_NAME}@${IP_ADDR}" "${bin_path}/venachain --datadir ${data_path}/node-${NODE_ID} init ${conf_path}/genesis.json"
        xcmd "${USER_NAME}@${IP_ADDR}" "[ -d ${data_path}/node-${NODE_ID}/venachain/chaindata -a -d ${data_path}/node-${NODE_ID}/venachain/lightchaindata ]"
        if [[ $? -eq 1 ]]; then
            printLog "error" "INIT GENESIS ON NODE_${NODE_ID} FAILED"
            return 1
        fi
        echo "******************************************************************************************************************************************************************************"
        echo "[${SCRIPT_ALIAS}] [node-${NODE_ID}] : Init genesis on node-${NODE_ID} completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        printLog "info" "Init genesis on node-${NODE_ID} completed"
    fi
    echo "[${SCRIPT_ALIAS}] [node-${NODE_ID}] : Init node completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
    printLog "success" "Init node Node-${NODE_ID} succeeded"
}

################################################# Init #################################################
function init() {
    if [[ "${ALL}" == "true" ]]; then
        cd "${PROJECT_CONF_PATH}"
        for file in $(ls ./); do
            if [ -f "${file}" ]; then
                initNode "${PROJECT_CONF_PATH}/${file}"
                if [[ $? -eq 1 ]]; then
                    printLog "error" "INIT NODE-${NODE_ID} FAILED"
                    exit 1
                fi
            fi
            cd "${PROJECT_CONF_PATH}"
        done
    else
        for param in $(echo "${NODE}" | sed 's/,/\n/g'); do
            if [ ! -f "${PROJECT_CONF_PATH}/deploy_node-${param}.conf" ]; then
                printLog "error" "FILE ${PROJECT_CONF_PATH}/deploy_node-${param}.conf NOT EXISTS"
                exit 1
            fi
            initNode "${PROJECT_CONF_PATH}/deploy_node-${param}.conf"
            if [[ $? -eq 1 ]]; then
                printLog "error" "INIT NODE NODE-${NODE_ID} FAILED"
                exit 1
            fi
        done
    fi
    printLog "info" "Init node completed"
}

################################################# Main #################################################
function main() {
    checkEnv

    init
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
    --project | -p)
        shiftOption2 $#
        PROJECT_NAME="${2}"
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
    --node | -n)
        shiftOption2 $#
        NODE="${2}"
        shift 2
        ;;
    --all | -a)
        ALL="true"
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
