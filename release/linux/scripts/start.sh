#!/bin/bash

###########################################################################################################
################################################# VRIABLES #################################################
###########################################################################################################
LOCAL_IP=$(ifconfig | grep inet | grep -v 127.0.0.1 | grep -v inet6 | awk '{print $2}')
if [[ "$(echo ${LOCAL_IP} | grep addr:)" != "" ]]; then
    LOCAL_IP=$(echo ${LOCAL_IP} | tr -s ':' ' ' | awk '{print $2}')
fi
DEPLOYMENT_PATH=$(
    cd $(dirname $0)
    cd ../../../
    pwd
)
DEPLOYMENT_CONF_PATH="$(cd ${DEPLOYMENT_PATH}/deployment_conf && pwd)"
PROJECT_CONF_PATH=""

NODE="all"
FIRSTNODE_ID=""

DEPLOY_PATH=""
USER_NAME=""
IP_ADDR=""
P2P_PORT=""
DEPLOY_PATH=""
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
USAGE: start.sh  [options] [value]

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

    if [[ $(echo "${address}" | grep "${LOCAL_IP}") != "" ]] || [[ $(echo "${address}" | grep "127.0.0.1") != "" ]]; then
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
        xcmd "${USER_NAME}@${IP_ADDR}" "${SCRIPT_PATH}/local-start-node.sh -n ${NODE_ID}"
        xcmd "${USER_NAME}@${IP_ADDR}" "lsof -i:${P2P_PORT}" 1>/dev/null
        if [[ $? -ne 0 ]]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* RUN NODE NODE-${NODE_ID} FAILED **********"
            return 1
        fi
        echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Run node Node-${NODE_ID} completed"
    fi
}

################################################# Add Role #################################################
function addRole() {
    ## create account
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ $(grep $(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g') "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "Create account") == "" ]]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "${SCRIPT_PATH}/local-create-account.sh -n ${NODE_ID}"
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

    ## deploy system contract
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ $(grep $(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g') "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "Deploy system contract") == "" ]]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "${SCRIPT_PATH}/local-deploy-system-contract.sh -n ${NODE_ID}"
        if [[ $? -ne 0 ]]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* DESTROY SYSTEM CONTRACT FAILED **********"
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
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* UPDATE NODE NODE-${NODE_ID} TO CONSENSUS NODEFAILED **********"
            return 1
        fi
        echo "[$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] [node-${NODE_ID}] : Update node to consensus node completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Update node node-${NODE_ID} to consensus node completed"
    fi
}

################################################# Start First Node #################################################
function startFirstNode() {
    ## read file
    clearData
    file=$1
    NODE_ID="${FIRSTNODE_ID}"
    echo
    echo "################ Start first node Node-${NODE_ID} ################"
    readFile "$file"

    ## start node
    runNode
    if [[ $? -ne 0 ]]; then
        return 1
    fi
    addRole
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
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Start firstnode Node-${NODE_ID} completed"
}

################################################# Start Other Node #################################################
function startOtherNode() {
    ## read file
    clearData
    file=$1
    NODE_ID=$(echo ${file} | sed -e 's/\(.*\)deploy_node-\(.*\).conf/\2/g ')
    echo
    echo "################ Start Node-${NODE_ID} ################"
    readFile "$file"

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
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Start node Node-${NODE_ID} completed"
}

################################################# Start All Nodes #################################################
function startAllNodes() {
    ## start first node
    cd "${PROJECT_CONF_PATH}"
    if [ ! -f "${PROJECT_CONF_PATH}/global/firstnode.info" ]; then
        "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* FILE ${PROJECT_CONF_PATH}/global/firstnode.info NOT FOUND **********"
        exit
    fi
    FIRSTNODE_ID=$(cat ${PROJECT_CONF_PATH}/global/firstnode.info | grep "node_id=" | sed -e 's/node_id=\(.*\)/\1/g')
    startFirstNode "${PROJECT_CONF_PATH}/deploy_node-${FIRSTNODE_ID}.conf"
    if [[ $? -ne 0 ]]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* START NODE NODE-${NODE_ID} FAILED **********"
        exit
    fi

    ## start other nodes
    cd "${PROJECT_CONF_PATH}"
    for file in $(ls ./); do
        if [ -f "${file}" ] && [[ $(echo ${file}) != "deploy_node-${FIRSTNODE_ID}.conf" ]]; then
            startOtherNode "${PROJECT_CONF_PATH}/${file}"
            if [[ $? -ne 0 ]]; then
                echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* START NODE NODE-${NODE_ID} FAILED **********"
                exit
            fi
        fi
        cd "${PROJECT_CONF_PATH}"
    done
}

################################################# Main #################################################
function main() {
    if [[ "${NODE}" == "all" ]]; then
        startAllNodes
    else
        if [ ! -f "${PROJECT_CONF_PATH}/global/firstnode.info" ]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* FIRSTNODE NOT INITED **********"
            exit
        fi
        user_name=$(grep user_name ${PROJECT_CONF_PATH}/global/firstnode.info)
        ip_addr=$(grep ip_addr ${PROJECT_CONF_PATH}/global/firstnode.info)
        rpc_port=$(grep rpc_port ${PROJECT_CONF_PATH}/global/firstnode.info)
        xcmd "${user_name}@${ip_addr}" "lsof -i:${rpc_port}" >/dev/null 2>&1
        if [[ $? == "" ]] || [ ! -f "${PROJECT_CONF_PATH}/global/keyfile.json" ] || [ ! -f "${PROJECT_CONF_PATH}/global/keyfile.phrase" ] || [ ! -f "${PROJECT_CONF_PATH}/global/keyfile.account" ]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* FIRSTNODE NOT STARTED **********"
            exit
        fi

        cd "${PROJECT_CONF_PATH}"
        for param in $(echo "${NODE}" | sed 's/,/\n/g'); do
            if [ ! -f "${PROJECT_CONF_PATH}/deploy_node-${param}.conf" ]; then
                echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* FILE deploy_node-${param}.conf NOT EXISTS **********"
                continue
            fi
            startOtherNode "${PROJECT_CONF_PATH}/deploy_node-${param}.conf"
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
        if [ ! -d "${DEPLOYMENT_CONF_PATH}/$2" ]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* ${DEPLOYMENT_CONF_PATH}/$2 HAS NOT BEEN CREATED **********"
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
    *)
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* COMMAND \"$1\" NOT FOUND **********"
        help
        exit
        ;;
    esac
done
main
