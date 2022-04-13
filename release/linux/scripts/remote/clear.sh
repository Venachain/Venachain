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
MODE=""
ALL=""

PROJECT_CONF_PATH=""

NODE_ID=""
DEPLOY_PATH=""
DEPLOY_CONF=""
BACKUP_PATH=""
BIN_PATH=""
CONF_PATH=""
SCRIPT_PATH=""
DATA_PATH=""
USER_NAME=""
IP_ADDR=""
P2P_PORT=""
RPC_PORT=""
IS_LOCAL=""

FIRSTNODE_USER_NAME=""
FIRSTNODE_ID=""
FIRSTNODE_IP_ADDR=""
FIRSTNODE_RPC_PORT=""

## param
mode=""

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

            --mode, -m                  the specified execute mode
                                        \"deep\", \"delete\", \"stop\", \"restart\", \"clean\" are supported (default: deep)
                                        \"deep\": will do \"delete\" and \"clean\" actions
                                        \"delete\": will delete the node from chain
                                        \"stop\": will stop the node
                                        \"restart\": will restart the node
                                        \"clean\": will stop the node and remove the files, configuration files will be backed up
           
            --all, -a                   clear all nodes

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

################################################# Yes Or No #################################################
function yesOrNo() {
    read -p "" anw
    case "${anw}" in
    [Yy][Ee][Ss] | [yY])
        return 1
        ;;
    [Nn][Oo] | [Nn])
        return 0
        ;;
    esac
    return 0
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
    
    if [ ! -d "${PROJECT_CONF_PATH}" ]; then
        printLog "error" "DIRECTORY ${PROJECT_CONF_PATH} NOT FOUND"
        exit 1
    fi
}

################################################# Assign Default #################################################
function assignDefault() { 
    MODE="deep"
}

################################################# Read Param #################################################
function readParam() {
    if [[ "${mode}" != "" ]]; then
        MODE="${mode}"
    fi
}

################################################# Clear Data #################################################
function clearData() {
    NODE_ID=""
    DEPLOY_PATH=""
    DEPLOY_CONF=""
    BACKUP_PATH=""
    BIN_PATH=""
    CONF_PATH=""
    SCRIPT_PATH=""
    DATA_PATH=""
    USER_NAME=""
    IP_ADDR=""
    P2P_PORT=""
    RPC_PORT=""
    IS_LOCAL=""
}

################################################# Read File #################################################
function readFile() {
    USER_NAME="$(cat ${DEPLOY_CONF} | grep "user_name=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    IP_ADDR="$(cat ${DEPLOY_CONF} | grep "ip_addr=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    P2P_PORT="$(cat ${DEPLOY_CONF} | grep "p2p_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    RPC_PORT="$(cat ${DEPLOY_CONF} | grep "rpc_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    DEPLOY_PATH="$(cat ${DEPLOY_CONF} | grep "deploy_path=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"

    if [[ "${USER_NAME}" == "" ]] || [[ "${IP_ADDR}" == "" ]] || [[ "${P2P_PORT}" == "" ]] || [[ "${DEPLOY_PATH}" == "" ]] || [[ "${RPC_PORT}" == "" ]]; then
        printLog "error" "KEY INFO NOT SET IN ${DEPLOY_CONF}"
        return 1
    fi

    firstnode_info="${PROJECT_CONF_PATH}/global/firstnode.info"
    if [ -f "${firstnode_info}" ]; then
        FIRSTNODE_USER_NAME="$(cat ${firstnode_info} | grep "user_name=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
        FIRSTNODE_ID="$(cat ${firstnode_info} | grep "node_id=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
        FIRSTNODE_IP_ADDR="$(cat ${firstnode_info} | grep "ip_addr=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
        FIRSTNODE_RPC_PORT="$(cat ${firstnode_info} | grep "rpc_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    fi
}

################################################# Delete Node #################################################
function deleteNode() {
    ## skip if clean all node
    if [[ "${MODE}" == "deep" ]] && [[ "${ALL}" == "true" ]]; then
        return 0
    fi

    if [[ "${NODE_ID}" == "${FIRSTNODE_ID}" ]]; then
        printLog "warn" "If Delete Firstnode, Many Services Will Not Be Usable"
        printLog "warn" "If You Really Want To Delete It, Please Do It Manually By vcl or venachainctl.sh Tools"
        return 0
    fi

    ## check firstnode's info
    if [[ "${FIRSTNODE_IP_ADDR}" == "" ]] || [[ "${FIRSTNODE_RPC_PORT}" == "" ]] || [[ "${FIRSTNODE_USER_NAME}" == "" ]] || [[ "${FIRSTNODE_ID}" == "" ]]; then
        printLog "error" "FIRSTNODE INFO NOT VALID, PLEASE CHECK ${PROJECT_CONF_PATH}/global/firstnode.info"
        return 1
    fi

    ## check firstnode
    xcmd "${FIRSTNODE_USER_NAME}@${FIRSTNODE_IP_ADDR}" "lsof -i:${FIRSTNODE_RPC_PORT}" 1>/dev/null
    if [[ $? -eq 1 ]]; then
        printLog "error" "DELETE NODE-${NODE_ID} FAILED, FIRSTNODE IS DOWN"
        return 1
    fi

    xcmd "${FIRSTNODE_USER_NAME}@${FIRSTNODE_IP_ADDR}" "${SCRIPT_PATH}/venachainctl.sh delete -n ${NODE_ID}"

    return $?
}

################################################# Stop Node #################################################
function stopNode() {
        
    while [[ 0 -lt 1 ]]; do
        pid_info=$(xcmd "${USER_NAME}@${IP_ADDR}" "lsof -i:${RPC_PORT}")
        if [[ "${pid_info}" == "" ]]; then
            break
        fi
        pid=$(echo ${pid_info} | awk '{ print $11 }')
        if [[ $? -eq 1 ]] || [[ "${pid}" == "" ]]; then
            printLog "error" "GET PID OF ${USER_NAME}@${IP_ADDR}:${RPC_PORT} FAILED"
            return 1
        fi
        xcmd "${USER_NAME}@${IP_ADDR}" "kill -9 ${pid}"
    done

    xcmd "${USER_NAME}@${IP_ADDR}" "lsof -i:${RPC_PORT}"
    if [[ $? -ne 1 ]]; then
        printLog "error" "KILL PID OF ${USER_NAME}@${IP_ADDR}:${RPC_PORT} FAILED"
        return 1
    fi
}

################################################# start Node #################################################
function runNode() {
    pid_info=$(xcmd "${USER_NAME}@${IP_ADDR}" "lsof -i:${RPC_PORT}")
    if [[ "${pid_info}" != "" ]]; then
        printLog "warn" "Node Is Still Running"
        return 0
    fi

    xcmd "${USER_NAME}@${IP_ADDR}" "${SCRIPT_PATH}/venachainctl.sh start -n ${NODE_ID}"
        
    timer=0
    while [ ${timer} -lt 10 ]; do
        pid_info=$(xcmd "${USER_NAME}@${IP_ADDR}" "lsof -i:${RPC_PORT}")
        if [[ "${pid_info}" != "" ]]; then
            break
        fi
        sleep 1
        let timer++
    done
    pid_info=$(xcmd "${USER_NAME}@${IP_ADDR}" "lsof -i:${RPC_PORT}")
    if [[ "${pid_info}" == "" ]]; then
        printLog "error" "RUN NODE-${NODE_ID} FAILED"
        return 1
    fi
    printLog "info" "Run node completed"
}

################################################# Remove Node #################################################
function removeNode() {

    ## clean node data
    xcmd "${USER_NAME}@${IP_ADDR}" "[ -d ${DEPLOY_PATH}/data/node-${NODE_ID} ]"
    if [ $? -ne 0 ]; then
        printLog "warn" "File ${USER_NAME}@${IP_ADDR}:${DEPLOY_PATH}/data/node-${NODE_ID} Not Found, May Be Has Already Been Cleaned"
    else
        # backup deployment conf
        xcmd "${USER_NAME}@${IP_ADDR}" "[ -f ${DEPLOY_PATH}/data/node-${NODE_ID}/deploy_node-${NODE_ID}.conf ]"
        if [ $? -ne 1 ]; then
            timestamp=$(date '+%Y%m%d%H%sM%S')
            xcmd "${USER_NAME}@${IP_ADDR}" "mkdir -p ${BACKUP_PATH}"
            xcmd "${USER_NAME}@${IP_ADDR}" "mv ${DEPLOY_PATH}/data/node-${NODE_ID}/deploy_node-${NODE_ID}.conf ${BACKUP_PATH}/deploy_node-${NODE_ID}.conf.bak.${timestamp}"
            xcmd "${USER_NAME}@${IP_ADDR}" "[ -f ${BACKUP_PATH}/deploy_node-${NODE_ID}.conf.bak.${timestamp} ]"
            if [[ $? -eq 1 ]]; then
                printLog "error" "BACKUP ${USER_NAME}@${IP_ADDR}:${DEPLOY_PATH}/data/node-${NODE_ID}/deploy_node-${NODE_ID}.conf FAILED"
                return 1
            fi
            printLog "info" "Backup ${USER_NAME}@${IP_ADDR}:${DEPLOY_PATH}/data/node-${NODE_ID}/deploy_node-${NODE_ID}.conf to ${USER_NAME}@${IP_ADDR}:${BACKUP_PATH}/deploy_node-${NODE_ID}.conf.bak.${timestamp} completed"
        fi

        # remove node dir
        xcmd "${USER_NAME}@${IP_ADDR}" "rm -rf ${DEPLOY_PATH}/data/node-${NODE_ID}"
        xcmd "${USER_NAME}@${IP_ADDR}" "[ -d ${DEPLOY_PATH}/data/node-${NODE_ID} ]"
        if [ $? -ne 1 ]; then
            printLog "error" "REMOVE ${USER_NAME}@${IP_ADDR}:${DEPLOY_PATH}/data/node-${NODE_ID} FAILED"
            return 1
        elif [ -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ]; then
            if [[ "${OS}" == "Darwin" ]]; then
                sed -i '' "/\[*\] \[node-${NODE_ID}\] : */d" "${PROJECT_CONF_PATH}/logs/deploy_log.txt"
            else
                sed -i "/\[*\] \[node-${NODE_ID}\] : */d" "${PROJECT_CONF_PATH}/logs/deploy_log.txt"
            fi
            printLog "info" "Remove ${USER_NAME}@${IP_ADDR}:${DEPLOY_PATH}/data/node-${NODE_ID} completed"
        fi
    fi

    ## check project
    cnt=0
    cd "${PROJECT_CONF_PATH}"
    for file in $(ls ./); do
        if [ ! -f "${file}" ]; then
            continue
        fi
        node_id="$(echo ${file} | sed -e 's/\(.*\)deploy_node-\(.*\).conf/\2/g')"
        xcmd "${USER_NAME}@${IP_ADDR}" "[ -f ${DEPLOY_PATH}/data/node-${node_id}/node.prikey ]"
        if [[ $? -ne 1 ]]; then
            cnt=$(expr ${cnt} + 1)
        fi
        cd "${PROJECT_CONF_PATH}"
    done
    if [[ ${cnt} -ne 0 ]]; then
        return 0
    fi

    ## remove project
    # remove scripts dir
    xcmd "${USER_NAME}@${IP_ADDR}" "[ -d ${DEPLOY_PATH}/scripts ]"
    if [ $? -ne 1 ]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "rm -rf ${DEPLOY_PATH}/scripts"
        xcmd "${USER_NAME}@${IP_ADDR}" "[ -d ${DEPLOY_PATH}/scripts ]"
        if [ $? -ne 1 ]; then
            printLog "error" "REMOVE ${USER_NAME}@${IP_ADDR}:${DEPLOY_PATH}/scripts FAILED"
            return 1
        else
            printLog "info" "Remove ${USER_NAME}@${IP_ADDR}:${DEPLOY_PATH}/scripts completed"
        fi
    fi

    # remove data dir
    xcmd "${USER_NAME}@${IP_ADDR}" "[ -d ${DEPLOY_PATH}/data ]"
    if [[ $? -ne 1 ]]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "rm -rf ${DEPLOY_PATH}/data"
        xcmd "${USER_NAME}@${IP_ADDR}" "[ -d ${DEPLOY_PATH}/data ]"
        if [[ $? -ne 1 ]]; then
            printLog "error" "REMOVE ${USER_NAME}@${IP_ADDR}:${DEPLOY_PATH}/data FAILED"
            return 1
        else
            printLog "info" "Remove ${USER_NAME}@${IP_ADDR}:${DEPLOY_PATH}/data completed"
        fi
    fi

    # remove bin dir
    xcmd "${USER_NAME}@${IP_ADDR}" "[ -d ${DEPLOY_PATH}/bin ]"
    if [[ $? -ne 1 ]]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "rm -rf ${DEPLOY_PATH}/bin"
        xcmd "${USER_NAME}@${IP_ADDR}" "[ -d ${DEPLOY_PATH}/bin ]"
        if [[ $? -ne 1 ]]; then
            printLog "error" "REMOVE ${USER_NAME}@${IP_ADDR}:${DEPLOY_PATH}/bin FAILED"
            return 1
        else
            printLog "info" "Remove ${USER_NAME}@${IP_ADDR}:${DEPLOY_PATH}/bin completed"
        fi
    fi

    # backup conf dir
    xcmd "${USER_NAME}@${IP_ADDR}" "[ -d ${DEPLOY_PATH}/conf ]"
    if [ $? -ne 1 ]; then
        timestamp=$(date '+%Y%m%d%H%sM%S')
        xcmd "${USER_NAME}@${IP_ADDR}" "mkdir -p ${BACKUP_PATH}"
        xcmd "${USER_NAME}@${IP_ADDR}" "mv ${DEPLOY_PATH}/conf ${BACKUP_PATH}/conf"
        xcmd "${USER_NAME}@${IP_ADDR}" "mv ${BACKUP_PATH} ${BACKUP_PATH}.bak.${timestamp}"
        xcmd "${USER_NAME}@${IP_ADDR}" "[ -d ${BACKUP_PATH}.bak.${timestamp} ]"
        if [[ $? -eq 1 ]]; then
            printLog "error" "BACKUP ${USER_NAME}@${IP_ADDR}:${DEPLOY_PATH}/conf FAILED"
            return 1
        else
            xcmd "${USER_NAME}@${IP_ADDR}" "rm -rf ${DEPLOY_PATH}"
            printLog "info" "Backup ${USER_NAME}@${IP_ADDR}:${DEPLOY_PATH}/conf completed"
        fi
    fi

    # clear all nodes' deploy log
    cd "${PROJECT_CONF_PATH}"
    for file in $(ls ./); do
        if [ ! -f "${file}" ]; then
            continue
        fi
        node_id="$(echo ${file} | sed -e 's/\(.*\)deploy_node-\(.*\).conf/\2/g')"
        deploy_conf="${PROJECT_CONF_PATH}/${file}"
        deploy_path="$(cat ${deploy_conf} | grep "deploy_path=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
        user_name="$(cat ${deploy_conf} | grep "user_name=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
        ip_addr="$(cat ${deploy_conf} | grep "ip_addr=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
        xcmd "${user_name}@${ip_addr}" "[ -d ${deploy_path}/data/node-${node_id} ]"
        if [[ $? -eq 1 ]] && [ -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ]; then
            if [[ "${OS}" == "Darwin" ]]; then
                sed -i '' "/\[*\] \[node-${node_id}\] : */d" "${PROJECT_CONF_PATH}/logs/deploy_log.txt"
            else
                sed -i "/\[*\] \[node-${node_id}\] : */d" "${PROJECT_CONF_PATH}/logs/deploy_log.txt"
            fi
        fi
        cd "${PROJECT_CONF_PATH}"
    done

    # clear logs
    if [ -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ]; then
        path_id=$(echo ${USER_NAME}@${IP_ADDR}:${DEPLOY_PATH} | sed 's/\//#/g')
        if [[ "${OS}" == "Darwin" ]]; then
            sed -i '' "/\[*\] \[${path_id}\] : */d" "${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        else
            sed -i "/\[*\] \[${path_id}\] : */d" "${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        fi
    fi

    # clear global and log data
    if [ -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] && [[ "$(cat ${PROJECT_CONF_PATH}/logs/deploy_log.txt)" == "" ]]; then
        rm -rf "${PROJECT_CONF_PATH}/global"
        rm -rf "${PROJECT_CONF_PATH}/logs"
        if [ -d "${PROJECT_CONF_PATH}/global" ] || [ -d "${PROJECT_CONF_PATH}/logs" ]; then
            printLog "error" "REMOVE ${PROJECT_CONF_PATH}/global AND ${PROJECT_CONF_PATH}/logs FAILED"
            printLog "warn" "You Can Remove Them Manually"
            return 1
        else
            printLog "info" "Remove ${PROJECT_CONF_PATH}/global and ${PROJECT_CONF_PATH}/logs completed"
        fi
    fi
}

################################################# Clear Node #################################################
function clearNode() {
    clearData
    DEPLOY_CONF="${1}"
    NODE_ID=$(echo "${DEPLOY_CONF}" | sed -e 's/\(.*\)deploy_node-\(.*\).conf/\2/g')
    echo
    echo "################ Start to clear Node-${NODE_ID} ################"

    readFile "${DEPLOY_CONF}"
    if [[ $? -eq 1 ]]; then
        printLog "error" "READ FILE ${DEPLOY_CONF} FAILED"
        exit 1
    fi
    checkRemoteAccess 
    if [[ $? -eq 1 ]]; then
        printLog "error" "CHECK REMOTE ACCESS TO NODE-${NODE_ID} FAILED"
        exit 1
    fi
    BACKUP_PATH="${DEPLOY_PATH}/../bak/${PROJECT_NAME}"
    BIN_PATH="${DEPLOY_PATH}/bin"
    CONF_PATH="${DEPLOY_PATH}/conf"
    SCRIPT_PATH="${DEPLOY_PATH}/scripts"
    DATA_PATH="${DEPLOY_PATH}/data"

    ## delete mode
    if [[ "${MODE}" == "delete" ]]; then
        deleteNode 
        if [[ $? -eq 1 ]]; then
            printLog "error" "DELETE NODE-${NODE_ID} FAILED"
            return 1
        fi
        printLog "info" "Delete node-${NODE_ID} end"

    ## stop mode
    elif [[ "${MODE}" == "stop" ]]; then
        stopNode 
        if [[ $? -eq 1 ]]; then
            printLog "error" "STOP NODE-${NODE_ID} FAILED"
            return 1
        fi
        printLog "info" "Stop node-${NODE_ID} end"

    ## restart mode
    elif [[ "${MODE}" == "restart" ]]; then
        stopNode 
        if [[ $? -eq 1 ]]; then
            printLog "error" "RESTART NODE-${NODE_ID} FAILED"
            return 1
        fi
        printLog "info" "Stop node-${NODE_ID} end"
        runNode
        if [[ $? -eq 1 ]]; then
            printLog "error" "RESTART NODE-${NODE_ID} FAILED"
            return 1
        fi
        printLog "info" "Restart node-${NODE_ID} end"

    ## clean mode
    elif [[ "${MODE}" == "clean" ]]; then
        stopNode 
        if [[ $? -eq 1 ]]; then
            printLog "error" "CLEAN NODE-${NODE_ID} FAILED"
            return 1
        fi
        printLog "info" "Stop node-${NODE_ID} end"
        removeNode 
        if [[ $? -eq 1 ]]; then
            printLog "error" "CLEAN NODE-${NODE_ID} FAILED"
            return 1
        fi
        printLog "info" "Clean node-${NODE_ID} end"

    ## deep mode
    elif [[ "${MODE}" == "deep" ]]; then
        deleteNode 
        if [[ $? -eq 1 ]]; then
            printLog "error" "DEEP CLEAN NODE-${NODE_ID} FAILED"
            return 1
        fi
        printLog "info" "Delete node-${NODE_ID} end"

        stopNode 
        if [[ $? -eq 1 ]]; then
            printLog "error" "DEEP CLEAN NODE-${NODE_ID} FAILED"
            return 1
        fi
        printLog "info" "Stop node-${NODE_ID} end"

        removeNode
        if [[ $? -eq 1 ]]; then
            printLog "error" "DEEP CLEAN NODE-${NODE_ID} FAILED"
            return 1
        fi
        printLog "info" "Deep clean node-${NODE_ID} end"
    else
        printLog "error" "MODE ${MODE} NOT FOUND"
        return 1
    fi
}

################################################# Clear All Node #################################################
function clearAllNode() {
    cd "${PROJECT_CONF_PATH}"
    for file in $(ls ./); do
        if [ ! -f "${file}" ]; then
            continue
        fi

        clearNode "${PROJECT_CONF_PATH}/${file}"
        if [[ $? -eq 1 ]]; then
            printLog "error" "CLEAR NODE-${NODE_ID} FAILED"
            exit 1
        fi
        cd "${PROJECT_CONF_PATH}"
    done
}

################################################# Clear Specified Node #################################################
function clearSpecifiedNode() {
    if [[ "${NODE}" == "" ]]; then
        printLog "error" "NODE IS EMPTY"
        exit 1
    fi

    cd "${PROJECT_CONF_PATH}"
    for node_id in $(echo "${NODE}" | sed 's/,/\n/g'); do
        if [ ! -f "${PROJECT_CONF_PATH}/deploy_node-${node_id}.conf" ]; then
            printLog "error" "FILE ${PROJECT_CONF_PATH}/deploy_node-${node_id}.conf NOT EXISTS"
            exit 1
        fi
        if [[ "${node_id}" == "${FIRSTNODE_ID}" ]]; then
            printLog "warn" "If Clear Firstnode, Many Services Will Not Be Usable"
            printLog "question" "Are you sure to clear firstnode node-${NODE_ID}? Yes or No(y/n):"
            yesOrNo
            if [ $? -ne 1 ]; then
                exit 2
            fi
        fi

        clearNode "${PROJECT_CONF_PATH}/deploy_node-${node_id}.conf"
        if [[ $? -eq 1 ]]; then
            printLog "error" "CLEAR NODE-${NODE_ID} FAILED"
            exit 1
        fi
        cd "${PROJECT_CONF_PATH}"
    done
}

################################################# Clear #################################################
function clear() {
    if [[ "${ALL}" == "true" ]]; then
        clearAllNode
    else
        clearSpecifiedNode
    fi
    printLog "info" "Clear action end"
}

################################################# Main #################################################
function main() {
    checkEnv
    assignDefault
    readParam

    clear
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
    --mode | -m)
        shiftOption2 $#
        mode="${2}"
        shift 2
        ;;
    --all | -a)
        ALL="true"
        shift 1
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
