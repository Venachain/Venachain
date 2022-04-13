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
NODE=""
ALL=""

PROJECT_CONF_PATH=""

DEPLOY_PATH=""
DEPLOY_CONF=""
BIN_PATH=""
CONF_PATH=""
SCRIPT_PATH=""
USER_NAME=""
IP_ADDR=""
RPC_PORT=""
IS_LOCAL=""

FIRSTNODE_ID=""
FIRSTNODE_INFO=""
FIRSTNODE_USERNAME=""
FIRSTNODE_IP_ADDR=""
FIRSTNODE_RPC_PORT=""

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

            --node, -n                  the specified node name
                                        use \",\" to seperate the name of node

            --all, -a                   start all nodes

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

################################################# check Env #################################################
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
    
    if [ ! -f "${PROJECT_CONF_PATH}/global/genesis.json" ]; then
        printLog "error" "FILE ${PROJECT_CONF_PATH}/global/genesis.json NOT FOUND"
        exit 1
    fi
    if [ ! -d "${PROJECT_CONF_PATH}" ]; then
        printLog "error" "DIRECTORY ${PROJECT_CONF_PATH} NOT FOUND"
        exit 1
    fi
}

################################################# Confirm First Node #################################################
function confirmFirstNode() {
    genesis=$(cat ${PROJECT_CONF_PATH}/global/genesis.json)
    cd "${PROJECT_CONF_PATH}/global/data"
    for file in $(ls ./); do
        pubkey="$(cat ${file}/node.pubkey)"
        if [[ "$(echo ${genesis} | grep $(echo ${pubkey}))" != "" ]]; then
            FIRSTNODE_ID=$(echo ${file} | sed -e 's/\(.*\)node-\(.*\)/\2/g')
            break
        fi
    done
}

################################################# Run Node #################################################
function runNode() {
    while [[ 0 -lt 1 ]]; do
        pid_info="$(xcmd "${USER_NAME}@${IP_ADDR}" "lsof -i:${RPC_PORT}")"
        if [[ "${pid_info}" == "" ]]; then
            break
        fi
        pid="$(echo ${pid_info} | awk '{ print $11 }')"
        if [[ $? -ne 0 ]] || [[ "${pid}" == "" ]]; then
            printLog "error" "GET PID OF ${USER_NAME}@${IP_ADDR}:${RPC_PORT} FAILED"
            return 1
        fi
        xcmd "${USER_NAME}@${IP_ADDR}" "kill -9 ${pid}"
    done
    printLog "info" "Stop node-${NODE_ID} completed"

    xcmd "${USER_NAME}@${IP_ADDR}" "${SCRIPT_PATH}/venachainctl.sh start -n ${NODE_ID}"
    pid_info="$(xcmd "${USER_NAME}@${IP_ADDR}" "lsof -i:${RPC_PORT}")"
    if [[ "${pid_info}" == "" ]]; then
        printLog "error" "RUN NODE-${NODE_ID} FAILED"
        return 1
    fi
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Run node Node-${NODE_ID} completed"
}

################################################# Deploy System Contract #################################################
function deploySystemContract() {
    ## create account
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ "$(grep "${SCRIPT_ALIAS}" "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "Create account")" == "" ]]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "${SCRIPT_PATH}/venachainctl.sh createacc -n ${NODE_ID} -ck --auto"
        xcmd "${USER_NAME}@${IP_ADDR}" "[ -f ${CONF_PATH}/keyfile.account -a -f ${CONF_PATH}/keyfile.json -a -f ${CONF_PATH}/keyfile.phrase ]"
        if [[ $? -eq 1 ]]; then
            printLog "error" "CREATE ACCOUNT FAILED"
            return 1
        fi
        echo "[${SCRIPT_ALIAS}] [node-${NODE_ID}] : Create account completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        printLog "info" "Create account completed"
    fi

    ## get keyfile
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ "$(grep "${SCRIPT_ALIAS}" "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "Get keyfile")" == "" ]]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "cp -r ${CONF_PATH}/keyfile.json ${PROJECT_CONF_PATH}/global/keyfile.json" "source"
        xcmd "${USER_NAME}@${IP_ADDR}" "cp -r ${CONF_PATH}/keyfile.phrase ${PROJECT_CONF_PATH}/global/keyfile.phrase" "source"
        xcmd "${USER_NAME}@${IP_ADDR}" "cp -r ${CONF_PATH}/keyfile.account ${PROJECT_CONF_PATH}/global/keyfile.account" "source"
        if [[ ! -f "${PROJECT_CONF_PATH}/global/keyfile.json" ]] || [[ ! -f "${PROJECT_CONF_PATH}/global/keyfile.phrase" ]] || [[ ! -f "${PROJECT_CONF_PATH}/global/keyfile.account" ]]; then
            printLog "error" "GET KEYFILE FAILED"
            return 1
        fi
        echo "[${SCRIPT_ALIAS}] [node-${NODE_ID}] : Get keyfile completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        printLog "info" "Get keyfile completed"
    fi

    ## add admin role
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ "$(grep "${SCRIPT_ALIAS}" "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "Add admin role")" == "" ]]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "${SCRIPT_PATH}/venachainctl.sh addadmin -n ${NODE_ID}"
        xcmd "${USER_NAME}@${IP_ADDR}" "[ -f ${CONF_PATH}/firstnode.info ]"
        if [[ $? -eq 1 ]]; then
            printLog "error" "ADD ADMIN ROLE FAILED"
            return 1
        fi
        echo "[${SCRIPT_ALIAS}] [node-${NODE_ID}] : Add admin role completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        printLog "info" "Add admin role completed"
    fi

    ## get firstnode info
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ "$(grep "${SCRIPT_ALIAS}" "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "Get firstnode info")" == "" ]]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "cp -r ${CONF_PATH}/firstnode.info ${PROJECT_CONF_PATH}/global/firstnode.info" "source"
        if [ ! -f "${PROJECT_CONF_PATH}/global/firstnode.info" ]; then
            printLog "error" "GET FIRSTNODE INFO FAILED"
            return 1
        fi
        echo "[${SCRIPT_ALIAS}] [node-${NODE_ID}] : Get firstnode info completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        echo "[${SCRIPT_ALIAS}] [node-${NODE_ID}] : Send firstnode info completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        printLog "info" "Get firstnode info completed"
    fi
}

################################################# Add Node #################################################
function addNode() {
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ "$(grep "${SCRIPT_ALIAS}" "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "node-${NODE_ID}" | grep "Add node")" == "" ]]; then
        if [[ "${FIRSTNODE_ID}" == "${NODE_ID}" ]]; then
            xcmd "${USER_NAME}@${IP_ADDR}" "${DEPLOY_PATH}/scripts/venachainctl.sh addnode -n ${NODE_ID}"
        else
            xcmd "${FIRSTNODE_USERNAME}@${FIRSTNODE_IP_ADDR}" "${FIRSTNODE_DEPLOY_PATH}/scripts/venachainctl.sh addnode -n ${NODE_ID}" 
        fi
        if [[ $? -eq 1 ]]; then
            printLog "error" "ADD NODE-${NODE_ID} FAILED"
            return 1
        fi
        echo "[${SCRIPT_ALIAS}] [node-${NODE_ID}] : Add node completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        printLog "info" "Add node-${NODE_ID} completed"
    fi
}

################################################# Update Node To Consensus #################################################
function updateNodeToConsensus() {
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ "$(grep "${SCRIPT_ALIAS}" "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "node-${NODE_ID}" | grep "Update node")" == "" ]]; then
        if [[ "${FIRSTNODE_ID}" == "${NODE_ID}" ]]; then
            xcmd "${USER_NAME}@${IP_ADDR}" "${SCRIPT_PATH}/venachainctl.sh updatesys -n ${NODE_ID}"
        else
            xcmd "${FIRSTNODE_USERNAME}@${FIRSTNODE_IP_ADDR}" "${FIRSTNODE_DEPLOY_PATH}/scripts/venachainctl.sh updatesys -n ${NODE_ID}" 
        fi
        if [[ $? -eq 1 ]]; then
            printLog "error" "UPDATE NODE-${NODE_ID} TO CONSENSUS FAILED"
            return 1
        fi
        echo "[${SCRIPT_ALIAS}] [node-${NODE_ID}] : Update node to consensus completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        printLog "info" "Update node-${NODE_ID} to consensus completed"
    fi 
}

################################################# Clear Data #################################################
function clearData() {
    DEPLOY_PATH=""
    DEPLOY_CONF=""
    BIN_PATH=""
    CONF_PATH=""
    SCRIPT_PATH=""
    USER_NAME=""
    IP_ADDR=""
    RPC_PORT=""
    IS_LOCAL=""
}

################################################# Read File #################################################
function readFile() {
    DEPLOY_PATH="$(cat ${DEPLOY_CONF} | grep "deploy_path=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    IP_ADDR="$(cat ${DEPLOY_CONF} | grep "ip_addr=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    USER_NAME="$(cat ${DEPLOY_CONF} | grep "user_name=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    RPC_PORT="$(cat ${DEPLOY_CONF} | grep "rpc_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    FIRSTNODE_DEPLOY_PATH="$(cat ${PROJECT_CONF_PATH}/deploy_node-${FIRSTNODE_ID}.conf | grep "deploy_path=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"

    if [[ "${DEPLOY_PATH}" == "" ]] || [[ "${IP_ADDR}" == "" ]] || [[ "${USER_NAME}" == "" ]] || [[ "${RPC_PORT}" == "" ]]; then
        printLog "error" "KEY INFO NOT SET IN ${DEPLOY_CONF}"
        return 1
    fi
}

################################################# Start First Node #################################################
function startFirstNode() {
    runNode
    if [[ $? -eq 1 ]]; then
        return 1
    fi
    deploySystemContract
    if [[ $? -eq 1 ]]; then
        return 1
    fi
    addNode
    if [[ $? -eq 1 ]]; then
        return 1
    fi
    updateNodeToConsensus
    if [[ $? -eq 1 ]]; then
        return 1
    fi
    printLog "success" "Start firstnode node-${NODE_ID} succeeded"
}

################################################# Start Other Node #################################################
function startOtherNode() {

    FIRSTNODE_INFO="${PROJECT_CONF_PATH}/global/firstnode.info"
    if [[ ! -f "${FIRSTNODE_INFO}" ]]; then
        printLog "error" "FILE ${FIRSTNODE_INFO} NOT FOUND "
        return 1
    fi 
    FIRSTNODE_USERNAME="$(cat ${FIRSTNODE_INFO} | grep "user_name=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    if [[ "${FIRSTNODE_USERNAME}" == "" ]]; then
        printLog "error" "FIRST NODE'S USERNAME NOT SET IN ${FIRSTNODE_INFO}"
        return 1
    fi
    FIRSTNODE_IP_ADDR="$(cat ${FIRSTNODE_INFO} | grep "ip_addr=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    if [[ "${FIRSTNODE_IP_ADDR}" == "" ]]; then
        printLog "error" "FIRST NODE'S IP NOT SET IN ${FIRSTNODE_INFO}"
        return 1
    fi
    FIRSTNODE_RPC_PORT="$(cat ${FIRSTNODE_INFO} | grep "rpc_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    if [[ "${FIRSTNODE_RPC_PORT}" == "" ]]; then
        printLog "error" "FIRST NODE'S RPC PORT NOT SET IN ${FIRSTNODE_INFO}"
        return 1
    fi

    ## send firstnode info
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ "$(grep "${SCRIPT_ALIAS}" "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "node-${NODE_ID}" | grep "Send firstnode info")" == "" ]]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "cp -r ${PROJECT_CONF_PATH}/global/firstnode.info ${CONF_PATH}" "target"
        xcmd "${USER_NAME}@${IP_ADDR}" "[ -f ${CONF_PATH}/firstnode.info ]"
        if [[ $? -eq 1 ]]; then
            printLog "error" "SEND FIRSTNODE INFO TO NODE_${NODE_ID} FAILED"
            return 1
        fi
        echo "[${SCRIPT_ALIAS}] [node-${NODE_ID}] : Send firstnode info to node-${NODE_ID} completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        printLog "info" "Send firstnode info to node-${NODE_ID} completed"
    fi
    

    ## send pubkey
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ "$(grep "${SCRIPT_ALIAS}" "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "node-${NODE_ID}" | grep "Send pubkey")" == "" ]]; then
        xcmd "${FIRSTNODE_USERNAME}@${FIRSTNODE_IP_ADDR}" "cp -r ${PROJECT_CONF_PATH}/global/data/node-${NODE_ID} ${FIRSTNODE_DEPLOY_PATH}/data" "target"
        xcmd "${FIRSTNODE_USERNAME}@${FIRSTNODE_IP_ADDR}" "[ -d ${FIRSTNODE_DEPLOY_PATH}/data/node-${NODE_ID} ]"
        if [[ $? -eq 1 ]]; then
            printLog "error" "SEND PUBKEY TO NODE_${FIRSTNODE_ID} FAILED"
            return 1
        fi
        echo "[${SCRIPT_ALIAS}] [node-${NODE_ID}] : Send pubkey to node-${FIRSTNODE_ID} completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        printLog "info" "Send pubkey to node-${NODE_ID} completed"
    fi

    ## send deploy conf
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ "$(grep "${SCRIPT_ALIAS}" "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "node-${NODE_ID}" | grep "Send deploy conf")" == "" ]]; then
        xcmd "${FIRSTNODE_USERNAME}@${FIRSTNODE_IP_ADDR}" "cp -r ${PROJECT_CONF_PATH}/deploy_node-${NODE_ID}.conf ${FIRSTNODE_DEPLOY_PATH}/data/node-${NODE_ID}/deploy_node-${NODE_ID}.conf" "target"
        xcmd "${FIRSTNODE_USERNAME}@${FIRSTNODE_IP_ADDR}" "[ -f ${FIRSTNODE_DEPLOY_PATH}/data/node-${NODE_ID}/deploy_node-${NODE_ID}.conf ]"
        if [[ $? -eq 1 ]]; then
            printLog "error" "SEND DEPLOY CONF TO NODE_${FIRSTNODE_ID} FAILED"
            return 1
        fi
        echo "[${SCRIPT_ALIAS}] [node-${NODE_ID}] : Send deploy conf info to node-${FIRSTNODE_ID} completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        printLog "info" "Send deploy conf to node-${FIRSTNODE_ID} completed"
    fi

    ## start node
    runNode
    if [[ $? -eq 1 ]]; then
        return 1
    fi
    addNode
    if [[ $? -eq 1 ]]; then
        return 1
    fi
    updateNodeToConsensus
    if [[ $? -eq 1 ]]; then
        return 1
    fi
    printLog "success" "Start node-${NODE_ID} succeeded"
}

################################################# Start Node #################################################
function startNode() {
    ## read file
    clearData
    DEPLOY_CONF="${1}"
    NODE_ID="$(echo ${DEPLOY_CONF} | sed -e 's/\(.*\)deploy_node-\(.*\).conf/\2/g ')"
    echo
    echo "################ Start Node-${NODE_ID} ################"

    readFile 
    if [[ $? -eq 1 ]]; then
        printLog "error" "READ FILE ${DEPLOY_CONF} FAILED"
        exit 1
    fi
    checkRemoteAccess 
    if [[ $? -eq 1 ]]; then
        printLog "error" "CHECK REMOTE ACCESS TO NODE-${NODE_ID} FAILED"
        exit 1
    fi
    BIN_PATH="${DEPLOY_PATH}/bin"
    CONF_PATH="${DEPLOY_PATH}/conf"
    SCRIPT_PATH="${DEPLOY_PATH}/scripts"

    if [ "${NODE_ID}" == "${FIRSTNODE_ID}" ]; then
        startFirstNode "${file}"
        return $?
    else
        startOtherNode "${file}"
        return $?
    fi
}

################################################# Start #################################################
function start() {
    if [[ "${ALL}" == "true" ]]; then
        cd "${PROJECT_CONF_PATH}"
        for file in $(ls ./); do
            if [ -f "${file}" ]; then
                startNode "${PROJECT_CONF_PATH}/${file}"
                if [[ $? -eq 1 ]]; then
                    printLog "error" "START NODE-${NODE_ID} FAILED"
                    exit 1
                fi
                cd "${PROJECT_CONF_PATH}"
            fi
        done
    else
        cd "${PROJECT_CONF_PATH}"
        for param in $(echo "${NODE}" | sed 's/,/\n/g'); do
            if [ ! -f "${PROJECT_CONF_PATH}/deploy_node-${param}.conf" ]; then
                printLog "error" "FILE deploy_node-${param}.conf NOT EXISTS"
                exit 1
            fi
            startNode "${PROJECT_CONF_PATH}/deploy_node-${param}.conf"
            if [[ $? -eq 1 ]]; then
                printLog "error" "START NODE-${NODE_ID} FAILED"
                exit 1
            fi
            cd "${PROJECT_CONF_PATH}"
        done
    fi
    printLog "info" "Start completed"
}

################################################# Main #################################################
function main() {
    checkEnv

    confirmFirstNode
    start
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
