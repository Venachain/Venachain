#!/bin/bash

###########################################################################################################
################################################# VRIABLES #################################################
###########################################################################################################
SCRIPT_NAME="$(basename ${0})"
DEPLOYMENT_PATH=$(
    cd $(dirname $0)
    cd ../../
    pwd
)
DEPLOYMENT_CONF_PATH="$(cd ${DEPLOYMENT_PATH}/deployment_conf && pwd)"
PROJECT_NAME="test"
PROJECT_CONF_PATH=""

NODE="all"
FIRSTNODE_ID=""
IS_LOCAL=""

DEPLOY_PATH=""
USER_NAME=""
IP_ADDR=""
P2P_PORT=""
DEPLOY_PATH=""
CONF_PATH=""
CONF_PATH=""
SCRIPT_PATH=""

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

################################################# Check Shift Option #################################################
function shiftOption2() {
    if [[ $1 -lt 2 ]]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* MISS OPTION VALUE! PLEASE SET THE VALUE **********"
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
    DEPLOY_PATH=$(cat $file | grep "deploy_path=" | sed -e 's/deploy_path=\(.*\)/\1/g')
    IP_ADDR=$(cat $file | grep "ip_addr=" | sed -e 's/ip_addr=\(.*\)/\1/g')
    USER_NAME=$(cat $file | grep "user_name=" | sed -e 's/user_name=\(.*\)/\1/g')
    P2P_PORT=$(cat $1 | grep "p2p_port=" | sed -e 's/p2p_port=\(.*\)/\1/g')

    if [[ "${DEPLOY_PATH}" == "" ]] || [[ "${IP_ADDR}" == "" ]] || [[ "${USER_NAME}" == "" ]] || [[ "${P2P_PORT}" == "" ]]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* FILE ${file} MISS VALUE **********"
        return 1
    fi

    if [[ "$(ifconfig | grep ${IP_ADDR})" != "" ]]; then
        IS_LOCAL="true"
    else
        IS_LOCAL="false"
    fi

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
    DEPLOY_PATH=""
    CONF_PATH=""
    SCRIPT_PATH=""
}

################################################# Run Node #################################################
function runNode() {
    xcmd "${USER_NAME}@${IP_ADDR}" "lsof -i:${P2P_PORT}" 1>/dev/null
    if [[ $? -ne 0 ]]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "${SCRIPT_PATH}/local-run-node.sh -n ${NODE_ID}"
        xcmd "${USER_NAME}@${IP_ADDR}" "lsof -i:${P2P_PORT}" 1>/dev/null
        if [[ $? -ne 0 ]]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* RUN NODE NODE-${NODE_ID} FAILED **********"
            return 1
        fi
        echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Run node Node-${NODE_ID} completed"
    fi
}

################################################# Add Role #################################################
function deploySystemContract() {
    ## create account
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ $(grep $(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g') "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "Create account") == "" ]]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "${SCRIPT_PATH}/local-create-account.sh -n ${NODE_ID} --auto --admin"
        xcmd "${USER_NAME}@${IP_ADDR}" "[ -f ${CONF_PATH}/keyfile.account -a -f ${CONF_PATH}/keyfile.json -a -f ${CONF_PATH}/keyfile.phrase ]"
        if [[ $? -ne 0 ]]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* CREATE ACCOUNT FAILED **********"
            return 1
        fi
        echo "[$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] [node-${NODE_ID}] : Create account completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Create account completed"
    fi

    ## get keyfile
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ $(grep $(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g') "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "Get keyfile") == "" ]]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "cp -r ${CONF_PATH}/keyfile.json ${PROJECT_CONF_PATH}/global/keyfile.json" "source"
        xcmd "${USER_NAME}@${IP_ADDR}" "cp -r ${CONF_PATH}/keyfile.phrase ${PROJECT_CONF_PATH}/global/keyfile.phrase" "source"
        xcmd "${USER_NAME}@${IP_ADDR}" "cp -r ${CONF_PATH}/keyfile.account ${PROJECT_CONF_PATH}/global/keyfile.account" "source"
        xcmd "${USER_NAME}@${IP_ADDR}" "[ -f ${PROJECT_CONF_PATH}/global/keyfile.json -a -f ${PROJECT_CONF_PATH}/global/keyfile.phrase -a -f ${PROJECT_CONF_PATH}/global/keyfile.account ]"
        if [[ $? -ne 0 ]]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* GET KEYFILE FAILED **********"
            return 1
        fi
        echo "[$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] [node-${NODE_ID}] : Get keyfile completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Get keyfile completed"
    fi

    ## add admin role
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ $(grep $(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g') "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "Deploy system contract") == "" ]]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "${SCRIPT_PATH}/local-add-admin-role.sh -n ${NODE_ID}"
        if [[ $? -ne 0 ]]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* DEPLOY SYSTEM CONTRACT FAILED **********"
            return 1
        fi
        echo "[$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] [node-${NODE_ID}] : Deploy system contract completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Deploy system contract completed"
    fi
}

################################################# Add Node #################################################
function addNode() {
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ $(grep $(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g') "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "node-${NODE_ID}" | grep "Add node") == "" ]]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "${SCRIPT_PATH}/local-add-node.sh -n ${NODE_ID}"
        if [[ $? -ne 0 ]]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* ADD NODE NODE-${NODE_ID} FAILED **********"
            return 1
        fi
        echo "[$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] [node-${NODE_ID}] : Add node completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Add node node-${NODE_ID} completed"
    fi
}

################################################# Update Node To Consensus #################################################
function updateNodeToConsensus() {
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ $(grep $(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g') "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "node-${NODE_ID}" | grep "Update node") == "" ]]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "${SCRIPT_PATH}/local-update-to-consensus-node.sh -n ${NODE_ID}"
        if [[ $? -ne 0 ]]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* UPDATE NODE NODE-${NODE_ID} TO CONSENSUS NODE FAILED **********"
            return 1
        fi
        echo "[$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] [node-${NODE_ID}] : Update node to consensus node completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Update node node-${NODE_ID} to consensus node completed"
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
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Start firstnode Node-${NODE_ID} completed"
}

################################################# Start Other Node #################################################
function startOtherNode() {

    ## send keyfile
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ $(grep $(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g') "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "node-${NODE_ID}" | grep "Send keyfile") == "" ]]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "cp -r ${PROJECT_CONF_PATH}/global/keyfile.json ${CONF_PATH}" "target"
        xcmd "${USER_NAME}@${IP_ADDR}" "cp -r ${PROJECT_CONF_PATH}/global/keyfile.phrase ${CONF_PATH}" "target"
        xcmd "${USER_NAME}@${IP_ADDR}" "[ -f ${CONF_PATH}/keyfile.json -a -f ${CONF_PATH}/keyfile.phrase ]"
        if [[ $? -ne 0 ]]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* SEND KEYFILE TO NODE_${NODE_ID} FAILED **********"
            return 1
        fi
        echo "[$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] [node-${NODE_ID}] : Send keyfile to Node-${NODE_ID} completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Send keyfile to Node-${NODE_ID} completed"
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
    echo "[$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] [node-${NODE_ID}] : Start node completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Start node Node-${NODE_ID} completed"
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
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* START FIRSTNODE NODE_${NODE_ID} FAILED **********"
            exit
        fi
    else
        startOtherNode "${file}"
        if [[ $? -ne 0 ]]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* START NODE NODE_${NODE_ID} FAILED **********"
            exit
        fi
    fi
}

################################################# Main #################################################
function main() {
    if [ ! -f "${PROJECT_CONF_PATH}/global/firstnode.info" ] || [ ! -f "${PROJECT_CONF_PATH}/global/genesis.json" ]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* FIRSTNODE NOT INITED **********"
        exit
    fi
    FIRSTNODE_ID=$(cat ${PROJECT_CONF_PATH}/global/firstnode.info | grep "node_id=" | sed -e 's/node_id=\(.*\)/\1/g')

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
                echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* FILE deploy_node-${param}.conf NOT EXISTS **********"
                continue
            fi
            start "${PROJECT_CONF_PATH}/deploy_node-${param}.conf"
            cd "${PROJECT_CONF_PATH}"
        done
    fi
    echo
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Start completed"
}

###########################################################################################################
#################################################  EXECUTE #################################################
###########################################################################################################
showTitle
if [ $# -eq 0 ]; then
    help
    exit
fi
while [ ! $# -eq 0 ]; do
    case "$1" in
    --project | -p)
        shiftOption2 $#
        if [[ "$2" != "" ]]; then
            PROJECT_NAME=$2
        fi
        PROJECT_CONF_PATH="${DEPLOYMENT_CONF_PATH}/${PROJECT_NAME}"

        if [ ! -d "${DEPLOYMENT_CONF_PATH}/${PROJECT_NAME}" ]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* ${DEPLOYMENT_CONF_PATH}/${PROJECT_NAME} HAS NOT BEEN CREATED **********"
            exit
        fi
        shift 2
        ;;
    --node | -n)
        shiftOption2 $#
        NODE=$2
        shift 2
        ;;
    *)
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* COMMAND \"$1\" NOT FOUND **********"
        help
        exit
        ;;
    esac
done
main
