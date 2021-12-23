#!/bin/bash

###########################################################################################################
################################################# VRIABLES #################################################
###########################################################################################################
SCRIPT_NAME="$(basename ${0})"
SCRIPT_ALIAS="$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')"
DEPLOYMENT_PATH=$(
    cd $(dirname $0)
    cd ../../../
    pwd
)

DEPLOYMENT_CONF_PATH="$(cd ${DEPLOYMENT_PATH}/deployment_conf && pwd)"
PROJECT_CONF_PATH=""

NODE="all"
ISLOCAL=""

NODE_ID=""
DEPLOY_PATH=""
USER_NAME=""
IP_ADDR=""
P2P_PORT=""

#############################################################################################################
################################################# FUNCTIONS #################################################
#############################################################################################################

################################################# Show Title #################################################
function showTitle() {
    echo '
###########################################
####             init nodes            ####
###########################################'
}

################################################# Help #################################################
function help() {
    echo
    echo "
USAGE: ${SCRIPT_NAME}  [options] [value]

        OPTIONS:

           --project, -p              the specified project name. must be specified

           --node, -n                 the specified node name. only used in conf mode. 
                                      default='all': deploy all nodes by conf in deployment_conf
                                      use ',' to seperate the name of node

           --help, -h                 show help
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
        exit
    fi
}

################################################# Execute Command #################################################
function xcmd() {
    address=$1
    cmd=$2
    scp_param=$3

    if [[ "${IS_LOCAL}" == "true" ]]; then
        eval ${cmd}
        return $?
    elif [[ $(echo "${cmd}" | grep "cp") == "" ]]; then
        ssh "${address}" "${cmd}"
        return $?
    else
        source_path=$(echo ${cmd} | sed -e 's/\(.*\)cp -r \(.*\) \(.*\)/\2/g')
        target_path=$(echo ${cmd} | sed -e 's/\(.*\)cp -r \(.*\) \(.*\)/\3/g')
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

################################################# Clear Data #################################################
function clearData() {
    NODE_ID=""
    DEPLOY_PATH=""
    USER_NAME=""
    IP_ADDR=""
    P2P_PORT=""
    RPC_PORT=""
}

################################################# Read File #################################################
function readFile() {
    file=$1
    DEPLOY_PATH=$(cat $file | grep "deploy_path=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    IP_ADDR=$(cat $file | grep "ip_addr=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    USER_NAME=$(cat $file | grep "user_name=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    P2P_PORT=$(cat $1 | grep "p2p_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    RPC_PORT=$(cat $1 | grep "rpc_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')

    if [[ "${DEPLOY_PATH}" == "" ]] || [[ "${IP_ADDR}" == "" ]] || [[ "${USER_NAME}" == "" ]] || [[ "${P2P_PORT}" == "" ]] || [[ "${RPC_PORT}" == "" ]]; then
        printLog "error" "FILE ${file} MISS VALUE"
        return 1
    fi

    if [[ "$(ifconfig | grep ${IP_ADDR})" != "" ]]; then
        IS_LOCAL="true"
    else
        IS_LOCAL="false"
    fi

}

################################################# Init #################################################
function init() {
    ## read file
    clearData
    file=$1
    NODE_ID=$(echo ${file} | sed -e 's/\(.*\)deploy_node-\(.*\).conf/\2/g')
    if [ -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] && [[ $(grep $(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g') "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "${NODE_ID}" | grep "Init node") != "" ]]; then
        return 0
    fi
    readFile "${file}"
    script_path="${DEPLOY_PATH}"/scripts
    conf_path="${DEPLOY_PATH}"/conf
    bin_path="${DEPLOY_PATH}"/bin
    data_path="${DEPLOY_PATH}"/data
    echo
    echo "################ Init Node-${NODE_ID} ################"

    ## generate key
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ $(grep $(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g') "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "node-${NODE_ID}" | grep "Generate key") == "" ]]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "${script_path}/local/generate-key.sh -n ${NODE_ID} --auto"
        xcmd "${USER_NAME}@${IP_ADDR}" "[ -f ${DEPLOY_PATH}/data/node-${NODE_ID}/node.address -a -f ${DEPLOY_PATH}/data/node-${NODE_ID}/node.prikey -a -f ${DEPLOY_PATH}/data/node-${NODE_ID}/node.pubkey ]"
        if [[ $? -ne 0 ]]; then
            printLog "error" "GENERATE KEY FOR NODE-${NODE_ID} FAILED"
            return 1
        fi
        echo "[$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] [node-${NODE_ID}] : Generate key for node-${NODE_ID} completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        printLog "info" "Generate key for node-${NODE_ID} completed"
    fi

    ## get key
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ $(grep $(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g') "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "node-${NODE_ID}" | grep "Get key") == "" ]]; then
        mkdir -p ${PROJECT_CONF_PATH}/global/data/node-${NODE_ID}

        xcmd "${USER_NAME}@${IP_ADDR}" "cp -r ${data_path}/node-${NODE_ID}/node.pubkey ${PROJECT_CONF_PATH}/global/data/node-${NODE_ID}/node.pubkey" "source"
        if [[ ! -f "${PROJECT_CONF_PATH}/global/data/node-${NODE_ID}/node.pubkey" ]]; then
            printLog "error" "GET KEY FOR NODE-${NODE_ID} FAILED"
            return 1
        fi
        echo "[$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] [node-${NODE_ID}] : Get key from node-${NODE_ID} completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        printLog "info" "Get key from node-${NODE_ID} completed"
    fi

    ## sync genesis file
    if [ ! -f "${PROJECT_CONF_PATH}/global/genesis.json" ]; then
        ## setup genesis file
        if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ $(grep $(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g') "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "Setup genesis") == "" ]]; then
            xcmd "${USER_NAME}@${IP_ADDR}" "${script_path}/local/generate-genesis.sh -n ${NODE_ID} --auto"
            xcmd "${USER_NAME}@${IP_ADDR}" "[ -f ${DEPLOY_PATH}/conf/genesis.json ]"
            if [[ $? -ne 0 ]]; then
                printLog "error" "SETUP GENESIS FILE FAILED"
                return 1
            fi
            echo "[$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] [node-${NODE_ID}] : Setup genesis file completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
            printLog "info" "Setup genesis file completed"
        fi

        ## get genesis file
        if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ $(grep $(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g') "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "Get genesis file") == "" ]]; then
            xcmd "${USER_NAME}@${IP_ADDR}" "cp -r ${conf_path}/genesis.json ${PROJECT_CONF_PATH}/global/genesis.json" "source"
            if [ ! -f "${PROJECT_CONF_PATH}/global/genesis.json" ]; then
                printLog "error" "GET GENESIS FILE FAILED"
                return 1
            fi
            echo "[$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] [node-${NODE_ID}] : Get genesis file completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
            echo "[$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] [node-${NODE_ID}] : Send genesis file completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
            printLog "info" "Get genesis file completed"
        fi
    else
        ## send genesis file
        if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ $(grep $(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g') "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "node-${NODE_ID}" | grep "Send genesis file") == "" ]]; then
            xcmd "${USER_NAME}@${IP_ADDR}" "cp -r ${PROJECT_CONF_PATH}/global/genesis.json ${conf_path}" "target"
            xcmd "${USER_NAME}@${IP_ADDR}" "[ -f ${conf_path}/genesis.json ]"
            if [[ $? -ne 0 ]]; then
                printLog "error" "SEND GENESIS FILE TO NODE_${NODE_ID} FAILED"
                return 1
            fi
            echo "[$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] [node-${NODE_ID}] : Send genesis file to node-${NODE_ID} completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
            printLog "info" "Send genesis file to node-${NODE_ID} completed"
        fi
    fi

    ## init genesis
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ $(grep $(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g') "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "node-${NODE_ID}" | grep "Init genesis") == "" ]]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "rm -rf ${data_path}/node-${NODE_ID}/venachain/*"
        if [ $? -ne 0 ]; then
            printLog "error" "INIT GENESIS ON NODE_${NODE_ID} FAILED"
            return 1
        fi
        echo "******************************************************************************************************************************************************************************"
        xcmd "${USER_NAME}@${IP_ADDR}" "${bin_path}/venachain --datadir ${data_path}/node-${NODE_ID} init ${conf_path}/genesis.json"
        xcmd "${USER_NAME}@${IP_ADDR}" "[ -d ${data_path}/node-${NODE_ID}/venachain/chaindata -a -d ${data_path}/node-${NODE_ID}/venachain/lightchaindata ]"
        if [ $? -ne 0 ]; then
            printLog "error" "INIT GENESIS ON NODE_${NODE_ID} FAILED"
            return 1
        fi
        echo "******************************************************************************************************************************************************************************"
        echo "[$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] [node-${NODE_ID}] : Init genesis on node-${NODE_ID} completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        printLog "info" "Init genesis on node-${NODE_ID} completed"
    fi
    echo "[$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] [node-${NODE_ID}] : Init node completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
    printLog "success" "Init node Node-${NODE_ID} succeeded"
}

################################################# Main #################################################
function main() {
    showTitle
    if [[ "${NODE}" == "all" ]]; then
        # backupFile
        cd "${PROJECT_CONF_PATH}"
        for file in $(ls ./); do
            if [ -f "${file}" ]; then
                init "${PROJECT_CONF_PATH}/${file}"
                if [[ $? -ne 0 ]]; then
                    printLog "error" "INIT NODE NODE-${NODE_ID} "
                    exit
                fi
            fi
            cd "${PROJECT_CONF_PATH}"
        done
    else
        cd "${PROJECT_CONF_PATH}"
        for param in $(echo "${NODE}" | sed 's/,/\n/g'); do
            if [ ! -f "deploy_node-${param}.conf" ]; then
                printLog "error" "FILE deploy_node-${param}.conf NOT EXISTS"
                continue
            fi
            init "${PROJECT_CONF_PATH}/deploy_node-${param}.conf"
            if [[ $? -ne 0 ]]; then
                printLog "error" "INIT NODE NODE-${NODE_ID} FAILED"
                exit
            fi
            cd "${PROJECT_CONF_PATH}"
        done
    fi
    echo
    printLog "info" "Init node completed"

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
    --project | -p)
        shiftOption2 $#
        if [ ! -d "${DEPLOYMENT_CONF_PATH}/$2" ]; then
            printLog "error" "${DEPLOYMENT_CONF_PATH}/$2 HAS NOT BEEN CREATED"
            exit
        fi
        PROJECT_CONF_PATH="${DEPLOYMENT_CONF_PATH}/$2"
        shift 2
        ;;
    --node | -n)
        shiftOption2 $#
        NODE=$2
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
