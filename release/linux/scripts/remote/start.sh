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
IS_LOCAL=""

DEPLOY_PATH=""
USER_NAME=""
IP_ADDR=""
P2P_PORT=""
RPC_PORT=""
DEPLOY_PATH=""
CONF_PATH=""
CONF_PATH=""
SCRIPT_PATH=""
FIRSTNODE_ID=""
FIRSTNODE_INFO=""
FIRSTNODE_USERNAME=""
FIRSTNODE_IP_ADDR=""
FIRSTNODE_RPC_PORT=""

#############################################################################################################
################################################# FUNCTIONS #################################################
#############################################################################################################

################################################# Show Title #################################################
function showTitle() {
    echo '
###########################################
####            start  nodes           ####
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

################################################# Read File #################################################
function readFile() {
    file=$1
    DEPLOY_PATH=$(cat $file | grep "deploy_path=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    IP_ADDR=$(cat $file | grep "ip_addr=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    USER_NAME=$(cat $file | grep "user_name=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    P2P_PORT=$(cat $1 | grep "p2p_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    RPC_PORT=$(cat $1 | grep "rpc_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    FIRSTNODE_DEPLOY_PATH=$(cat ${PROJECT_CONF_PATH}/deploy_node-${FIRSTNODE_ID}.conf | grep "deploy_path=" | sed -e 's/\(.*\)=\(.*\)/\2/g')

    if [[ "${DEPLOY_PATH}" == "" ]] || [[ "${IP_ADDR}" == "" ]] || [[ "${USER_NAME}" == "" ]] || [[ "${P2P_PORT}" == "" ]]; then
        printLog "error" "FILE ${file} MISS VALUE"
        return 1
    fi

    if [[ "$(ifconfig | grep ${IP_ADDR})" != "" ]]; then
        IS_LOCAL="true"
    else
        IS_LOCAL="false"
    fi

    BIN_PATH="${DEPLOY_PATH}/bin"
    CONF_PATH="${DEPLOY_PATH}/conf"
    SCRIPT_PATH="${DEPLOY_PATH}/scripts"
}

################################################# Clear Data #################################################
function clearData() {
    NODE_ID=""
    DEPLOY_PATH=""
    USER_NAME=""
    IP_ADDR=""
    P2P_PORT=""
    RPC_PORT=""
    DEPLOY_PATH=""
    CONF_PATH=""
    SCRIPT_PATH=""
}

################################################# Run Node #################################################
function runNode() {
    xcmd "${USER_NAME}@${IP_ADDR}" "lsof -i:${P2P_PORT}" 1>/dev/null
    if [[ $? -ne 0 ]]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "${SCRIPT_PATH}/start-node.sh -n ${NODE_ID}"
        xcmd "${USER_NAME}@${IP_ADDR}" "lsof -i:${P2P_PORT}" 1>/dev/null
        if [[ $? -ne 0 ]]; then
            printLog "error" "RUN NODE NODE-${NODE_ID} FAILED"
            return 1
        fi
        echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Run node Node-${NODE_ID} completed"
    fi
}

################################################# Add Role #################################################
function deploySystemContract() {
    ## create account
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ $(grep $(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g') "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "Create account") == "" ]]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "${SCRIPT_PATH}/local/create-account.sh -n ${NODE_ID} -ck -a"
        xcmd "${USER_NAME}@${IP_ADDR}" "[ -f ${CONF_PATH}/keyfile.account -a -f ${CONF_PATH}/keyfile.json -a -f ${CONF_PATH}/keyfile.phrase ]"
        if [[ $? -ne 0 ]]; then
            printLog "error" "CREATE ACCOUNT FAILED"
            return 1
        fi
        echo "[$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] [node-${NODE_ID}] : Create account completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        printLog "info" "Create account completed"
    fi

    ## get keyfile
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ $(grep $(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g') "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "Get keyfile") == "" ]]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "cp -r ${CONF_PATH}/keyfile.json ${PROJECT_CONF_PATH}/global/keyfile.json" "source"
        xcmd "${USER_NAME}@${IP_ADDR}" "cp -r ${CONF_PATH}/keyfile.phrase ${PROJECT_CONF_PATH}/global/keyfile.phrase" "source"
        xcmd "${USER_NAME}@${IP_ADDR}" "cp -r ${CONF_PATH}/keyfile.account ${PROJECT_CONF_PATH}/global/keyfile.account" "source"
        if [[ ! -f "${PROJECT_CONF_PATH}/global/keyfile.json" ]] || [[ ! -f "${PROJECT_CONF_PATH}/global/keyfile.phrase" ]] || [[ ! -f "${PROJECT_CONF_PATH}/global/keyfile.account" ]]; then
            printLog "error" "GET KEYFILE FAILED"
            return 1
        fi
        echo "[$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] [node-${NODE_ID}] : Get keyfile completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        printLog "info" "Get keyfile completed"
    fi

    ## add admin role
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ $(grep $(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g') "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "Deploy system contract") == "" ]]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "${SCRIPT_PATH}/local/add-admin-role.sh -n ${NODE_ID}"
        if [[ $? -ne 0 ]]; then
            printLog "error" "DEPLOY SYSTEM CONTRACT FAILED"
            return 1
        fi
        echo "[$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] [node-${NODE_ID}] : Deploy system contract completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        printLog "info" "Deploy system contract completed"
    fi

    ## get firstnode info
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ $(grep $(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g') "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "Get firstnode info") == "" ]]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "cp -r ${CONF_PATH}/firstnode.info ${PROJECT_CONF_PATH}/global/firstnode.info" "source"
        if [ ! -f "${PROJECT_CONF_PATH}/global/firstnode.info" ]; then
            printLog "error" "GET FIRSTNODE INFO FAILED"
            return 1
        fi
        echo "[$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] [node-${NODE_ID}] : Get firstnode info completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        echo "[$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] [node-${NODE_ID}] : Send firstnode info completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        printLog "info" "Get firstnode info completed"
    fi
}

################################################# Add Node #################################################
function addNode() {
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ $(grep $(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g') "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "node-${NODE_ID}" | grep "Add node") == "" ]]; then
        res_add_node=""
        if [[ "${FIRSTNODE_ID}" == "${NODE_ID}" ]]; then
            xcmd "${USER_NAME}@${IP_ADDR}" "${DEPLOY_PATH}/scripts/add-node.sh -n ${NODE_ID}"
            res_add_node=$(xcmd "${USER_NAME}@${IP_ADDR}" "${DEPLOY_PATH}/bin/venachaincli node query --name ${NODE_ID} --url ${IP_ADDR}:${RPC_PORT}") >/dev/null 2>&1
        else
            xcmd "${FIRSTNODE_USERNAME}@${FIRSTNODE_IP_ADDR}" "${FIRSTNODE_DEPLOY_PATH}/scripts/add-node.sh -n ${NODE_ID}" 
            res_add_node=$(xcmd "${USER_NAME}@${IP_ADDR}" "${DEPLOY_PATH}/bin/venachaincli node query --name ${NODE_ID} --url ${FIRSTNODE_IP_ADDR}:${FIRSTNODE_RPC_PORT}") >/dev/null 2>&1
        fi
        if [[ $(echo ${res_add_node} | grep "success") == "" ]]; then
            printLog "error" "ADD NODE NODE-${NODE_ID} FAILED"
            return 1
        fi
        echo "[$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] [node-${NODE_ID}] : Add node completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        printLog "info" "Add node node-${NODE_ID} completed"
    fi
}

################################################# Update Node To Consensus #################################################
function updateNodeToConsensus() {
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ $(grep $(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g') "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "node-${NODE_ID}" | grep "Update node") == "" ]]; then
        res_update_node=""
        if [[ "${FIRSTNODE_ID}" == "${NODE_ID}" ]]; then
            xcmd "${USER_NAME}@${IP_ADDR}" "${SCRIPT_PATH}/update_to_consensus_node.sh -n ${NODE_ID}"
            res_update_node=$(xcmd "${USER_NAME}@${IP_ADDR}" "${DEPLOY_PATH}/bin/venachaincli node query --type consensus --name ${NODE_ID} --url ${IP_ADDR}:${RPC_PORT}") >/dev/null 2>&1
        else
            xcmd "${FIRSTNODE_USERNAME}@${FIRSTNODE_IP_ADDR}" "${FIRSTNODE_DEPLOY_PATH}/scripts/update_to_consensus_node.sh -n ${NODE_ID}" 
            res_update_node=$(xcmd "${USER_NAME}@${IP_ADDR}" "${DEPLOY_PATH}/bin/venachaincli node query --type consensus --name ${NODE_ID} --url ${FIRSTNODE_IP_ADDR}:${FIRSTNODE_RPC_PORT}") >/dev/null 2>&1
        fi
        if [[ $(echo ${res_update_node} | grep "success") == "" ]]; then
            printLog "error" "UPDATE NODE NODE-${NODE_ID} TO CONSENSUS NODE FAILED"
            return 1
        fi
        echo "[$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] [node-${NODE_ID}] : Update node to consensus node completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        printLog "info" "Update node node-${NODE_ID} to consensus node completed"
    fi 
}

################################################# Start First Node #################################################
function startFirstNode() {
    runNode
    if [[ $? -ne 0 ]]; then
        return 1
    fi
    deploySystemContract
    if [[ $? -ne 0 ]]; then
        return 1
    fi
    addNode
    if [[ $? -ne 0 ]]; then
        return 1
    fi
    updateNodeToConsensus
    if [[ $? -ne 0 ]]; then
        return 1
    fi
    echo "[$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] [node-${NODE_ID}] : Start node completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
    printLog "success" "Start firstnode Node-${NODE_ID} succeeded"
}

################################################# Start Other Node #################################################
function startOtherNode() {

    FIRSTNODE_INFO="${PROJECT_CONF_PATH}/global/firstnode.info"
    if [[ ! -f "${FIRSTNODE_INFO}" ]]; then
        printLog "error" "FILE ${FIRSTNODE_INFO} NOT FOUND "
        return 1
    fi 
    FIRSTNODE_USERNAME=$(cat ${FIRSTNODE_INFO} | grep "user_name=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    if [[ "${FIRSTNODE_USERNAME}" == "" ]]; then
        printLog "error" "FIRST NODE'S USERNAME NOT FOUND"
        return 1
    fi
    FIRSTNODE_IP_ADDR=$(cat ${FIRSTNODE_INFO} | grep "ip_addr=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    if [[ "${FIRSTNODE_IP_ADDR}" == "" ]]; then
        printLog "error" "FIRST NODE'S IP NOT FOUND"
        return 1
    fi
    FIRSTNODE_RPC_PORT=$(cat ${FIRSTNODE_INFO} | grep "rpc_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    if [[ "${FIRSTNODE_RPC_PORT}" == "" ]]; then
        printLog "error" "FIRST NODE'S RPC PORT NOT FOUND"
        return 1
    fi

    ## send firstnode info
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ $(grep $(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g') "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "node-${NODE_ID}" | grep "Send firstnode info") == "" ]]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "cp -r ${PROJECT_CONF_PATH}/global/firstnode.info ${CONF_PATH}" "target"
        xcmd "${USER_NAME}@${IP_ADDR}" "[ -f ${CONF_PATH}/firstnode.info ]"
        if [[ $? -ne 0 ]]; then
            printLog "error" "SEND FIRSTNODE INFO TO NODE_${NODE_ID} FAILED"
            return 1
        fi
        echo "[$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] [node-${NODE_ID}] : Send firstnode info to node-${NODE_ID} completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        printLog "info" "Send firstnode info to node-${NODE_ID} completed"
    fi
    

    ## send pubkey
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ $(grep $(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g') "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "node-${NODE_ID}" | grep "Send pubkey") == "" ]]; then
        xcmd "${FIRSTNODE_USERNAME}@${FIRSTNODE_IP_ADDR}" "cp -r ${PROJECT_CONF_PATH}/global/data/node-${NODE_ID} ${FIRSTNODE_DEPLOY_PATH}/data" "target"
        xcmd "${FIRSTNODE_USERNAME}@${FIRSTNODE_IP_ADDR}" "[ -d ${FIRSTNODE_DEPLOY_PATH}/data/node-${NODE_ID} ]"
        if [[ $? -ne 0 ]]; then
            printLog "error" "SEND PUBKEY TO NODE_${FIRSTNODE_ID} FAILED"
            return 1
        fi
        echo "[$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] [node-${NODE_ID}] : Send pubkey to node-${FIRSTNODE_ID} completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        printLog "info" "Send pubkey to node-${NODE_ID} completed"
    fi

    ## send deploy conf
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ $(grep $(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g') "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "node-${NODE_ID}" | grep "Send deploy conf") == "" ]]; then
        xcmd "${FIRSTNODE_USERNAME}@${FIRSTNODE_IP_ADDR}" "cp -r ${PROJECT_CONF_PATH}/deploy_node-${NODE_ID}.conf ${FIRSTNODE_DEPLOY_PATH}/data/node-${NODE_ID}/deploy_node-${NODE_ID}.conf" "target"
        xcmd "${FIRSTNODE_USERNAME}@${FIRSTNODE_IP_ADDR}" "[ -f ${FIRSTNODE_DEPLOY_PATH}/data/node-${NODE_ID}/deploy_node-${NODE_ID}.conf ]"
        if [[ $? -ne 0 ]]; then
            printLog "error" "SEND DEPLOY CONF TO NODE_${FIRSTNODE_ID} FAILED"
            return 1
        fi
        echo "[$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] [node-${NODE_ID}] : Send deploy conf info to node-${FIRSTNODE_ID} completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        printLog "info" "Send deploy conf to node-${FIRSTNODE_ID} completed"
    fi

    ## start node
    runNode
    if [[ $? -ne 0 ]]; then
        return 1
    fi
    addNode
    if [[ $? -ne 0 ]]; then
        return 1
    fi
    updateNodeToConsensus
    if [[ $? -ne 0 ]]; then
        return 1
    fi
    printLog "success" "Start node Node-${NODE_ID} succeeded"
}

################################################# Start #################################################
function start() {
    file=$1

    ## read file
    clearData
    file=$1
    NODE_ID=$(echo ${file} | sed -e 's/\(.*\)deploy_node-\(.*\).conf/\2/g ')
    echo
    echo "################ Start Node-${NODE_ID} ################"
    readFile "$file"

    if [ "${NODE_ID}" == "${FIRSTNODE_ID}" ]; then
        startFirstNode "${file}"
        if [[ $? -ne 0 ]]; then
            printLog "error" "START FIRSTNODE NODE-${NODE_ID} FAILED"
            exit
        fi
    else
        startOtherNode "${file}"
        if [[ $? -ne 0 ]]; then
            printLog "error" "START NODE NODE-${NODE_ID} FAILED"
            exit
        fi
    fi
}

################################################# Start #################################################
function confirmFirstNode() {
    cd ${PROJECT_CONF_PATH}/global
    genesis=$(cat genesis.json)
    cd ${PROJECT_CONF_PATH}/global/data
    for file in $(ls ./); do
        pubkey=$(cat ${file}/node.pubkey)
        if [[ $(echo ${genesis} | grep $(echo ${pubkey})) != "" ]]; then
            FIRSTNODE_ID=$(echo ${file} | sed -e 's/\(.*\)node-\(.*\)/\2/g')
            break
        fi
    done
}

################################################# Main #################################################
function main() {
    showTitle
    if [ ! -f "${PROJECT_CONF_PATH}/global/genesis.json" ]; then
        printLog "error" "GENESIS FILE NOT FOUND"
        exit
    fi
    confirmFirstNode

    if [[ "${NODE}" == "all" ]]; then
        cd "${PROJECT_CONF_PATH}"
        for file in $(ls ./); do
            if [ -f "${file}" ]; then
                start "${PROJECT_CONF_PATH}/${file}"
                cd "${PROJECT_CONF_PATH}"
            fi
        done
    else
        cd "${PROJECT_CONF_PATH}"
        for param in $(echo "${NODE}" | sed 's/,/\n/g'); do
            if [ ! -f "${PROJECT_CONF_PATH}/deploy_node-${param}.conf" ]; then
                printLog "error" "FILE deploy_node-${param}.conf NOT EXISTS"
                continue
            fi
            start "${PROJECT_CONF_PATH}/deploy_node-${param}.conf"
            cd "${PROJECT_CONF_PATH}"
        done
    fi
    echo
    printLog "info" "Start completed"
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
