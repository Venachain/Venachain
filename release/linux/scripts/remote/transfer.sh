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
DEPLOYMENT_FILE_PATH="${DEPLOYMENT_PATH}/linux"

## global
SCRIPT_NAME="$(basename ${0})"
SCRIPT_ALIAS="$(echo ${CURRENT_PATH}/${SCRIPT_NAME} | sed -e 's/\(.*\)\/scripts\/\(.*\).sh/\2/g')"
PROJECT_NAME=""
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

            --node, -n                  the specified node name
                                        use \",\" to seperate the name of node

            --all, -a                   transfer files to all nodes

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
    
    if [ ! -d "${PROJECT_CONF_PATH}" ]; then
        printLog "error" "DEIRECTORY ${PROJECT_CONF_PATH} NOT FOUND"
        exit 1
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
    USER_NAME="$(cat ${DEPLOY_CONF} | grep "user_name=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    IP_ADDR="$(cat ${DEPLOY_CONF} | grep "ip_addr=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    P2P_PORT="$(cat ${DEPLOY_CONF} | grep "p2p_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"

    if [[ "${USER_NAME}" == "" ]] || [[ "${IP_ADDR}" == "" ]] || [[ "${P2P_PORT}" == "" ]] || [[ "${DEPLOY_PATH}" == "" ]]; then
        printLog "error" "KEY INFO NOT SET IN ${DEPLOY_CONF}"
        return 1
    fi
}

################################################# Set Up Directory Structure #################################################
function setupDirectoryStructure() {
    if [ ! -d "${PROJECT_CONF_PATH}/global" ] || [ ! -d "${PROJECT_CONF_PATH}/logs" ]; then
        mkdir -p "${PROJECT_CONF_PATH}/global" && mkdir -p "${PROJECT_CONF_PATH}/logs"
        if [ ! -d "${PROJECT_CONF_PATH}/global" ] || [ ! -d "${PROJECT_CONF_PATH}/logs" ]; then
            printLog "error" "SET UP DEIRECTORY STRUCTURE FAILED"
            exit 1
        else
            printLog "info" "Set up directory structure completed"
        fi
    fi
}

################################################# Transfer Files #################################################
function transferFiles() {
    clearData
    DEPLOY_CONF="${1}"
    NODE_ID="$(echo "${DEPLOY_CONF}" | sed -e 's/\(.*\)deploy_node-\(.*\).conf/\2/g')"
    echo
    echo "################ Transfer file to Node-${NODE_ID} ################"
    if [ -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] && [[ "$(grep "${SCRIPT_ALIAS}" "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "${NODE_ID}" | grep "Transfer files")" != "" ]]; then
        return 0
    fi

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
    path_id="$(echo ${USER_NAME}@${IP_ADDR}:${DEPLOY_PATH} | sed 's/\//#/g')"

    # create bin directory
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ "$(grep "${SCRIPT_ALIAS}" "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "${path_id}" | grep "Create bin directory")" == "" ]]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "mkdir -p ${DEPLOY_PATH}/bin"
        xcmd "${USER_NAME}@${IP_ADDR}" "[ -d ${DEPLOY_PATH}/bin ]"
        if [[ $? -eq 1 ]]; then
            printLog "error" "CREATE ${DEPLOY_PATH}/bin FAILED"
            return 1
        fi
        echo "[${SCRIPT_ALIAS}] [${path_id}] : Create bin directory completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        printLog "info" "Create ${DEPLOY_PATH}/bin completed"
    fi

    ## transfer files
    # transfer conf files
    cd "${DEPLOYMENT_FILE_PATH}/conf"
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ "$(grep "${SCRIPT_ALIAS}" "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "${path_id}" | grep "Transfer conf file")" == "" ]]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "cp -r ${DEPLOYMENT_FILE_PATH}/conf ${DEPLOY_PATH}" "target"
        file_num=$(xcmd "${USER_NAME}@${IP_ADDR}" "cd ${DEPLOY_PATH}/conf && ls -lR | grep "^-" | wc -l")
        if [[ "$(cd ${DEPLOYMENT_FILE_PATH}/conf && ls -lR | grep "^-" | wc -l)" != "${file_num}" ]]; then
            printLog "error" "TRANSFER ${DEPLOYMENT_FILE_PATH}/conf FAILED"
            return 1
        else
            echo "[${SCRIPT_ALIAS}] [${path_id}] : Transfer conf file completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
            printLog "info" "Transfer ${DEPLOYMENT_FILE_PATH}/conf completed"
        fi
    fi
    # transfer script files
    cd "${DEPLOYMENT_FILE_PATH}/scripts"
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ "$(grep "${SCRIPT_ALIAS}" "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "${path_id}" | grep "Transfer scripts file")" == "" ]]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "cp -r ${DEPLOYMENT_FILE_PATH}/scripts ${DEPLOY_PATH}" "target"
        file_num=$(xcmd "${USER_NAME}@${IP_ADDR}" "cd ${DEPLOY_PATH}/scripts && ls -lR | grep "^-" | wc -l")
        if [[ "$(cd ${DEPLOYMENT_FILE_PATH}/scripts && ls -lR | grep "^-" | wc -l)" != "${file_num}" ]]; then
            printLog "error" "TRANSFER ${DEPLOYMENT_FILE_PATH}/scripts FAILED"
            return 1
        else
            echo "[${SCRIPT_ALIAS}] [${path_id}] : Transfer scripts file completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
            printLog "info" "Transfer ${DEPLOYMENT_FILE_PATH}/scripts completed"
        fi
    fi
    # transfer bin files
    cd "${DEPLOYMENT_FILE_PATH}/bin"
    for file in $(ls ./); do
        if [ ! -f "${file}" ]; then
            continue
        fi
        if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ "$(grep "${SCRIPT_ALIAS}" "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "${path_id}" | grep "Transfer ${file} completed")" == "" ]]; then
            xcmd "${USER_NAME}@${IP_ADDR}" "cp -r ${file} ${DEPLOY_PATH}/bin" "target"
            xcmd "${USER_NAME}@${IP_ADDR}" "[ -f ${DEPLOY_PATH}/bin/${file} ]"
            if [[ $? -eq 1 ]]; then
                printLog "error" "TRANSFER ${DEPLOYMENT_FILE_PATH}/bin/${file} FAILED"
                return 1
            else
                echo "[${SCRIPT_ALIAS}] [${path_id}] : Transfer ${file} completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
                printLog "info" "Transfer ${DEPLOYMENT_FILE_PATH}/bin/${file} completed"
            fi
        fi
        cd "${DEPLOYMENT_FILE_PATH}/bin"
    done

    # create node directory
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ "$(grep "${SCRIPT_ALIAS}" "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "node-${NODE_ID}" | grep "Create node directory")" == "" ]]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "mkdir -p ${DEPLOY_PATH}/data/node-${NODE_ID}"
        xcmd "${USER_NAME}@${IP_ADDR}" "[ -d ${DEPLOY_PATH}/data/node-${NODE_ID} ]"
        if [[ $? -eq 1 ]]; then
            printLog "error" "CREATE ${DEPLOY_PATH}/data/node-${NODE_ID} FAILED"
            return 1
        fi
        echo "[${SCRIPT_ALIAS}] [node-${NODE_ID}] : Create node directory completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        printLog "info" "Create ${DEPLOY_PATH}/data/node-${NODE_ID} completed"
    fi

    # transfer deploy conf file
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ "$(grep "${SCRIPT_ALIAS}" "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "node-${NODE_ID}" | grep "Transfer deploy conf")" == "" ]]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "cp -r ${DEPLOY_CONF} ${DEPLOY_PATH}/data/node-${NODE_ID}" "target"
        xcmd "${USER_NAME}@${IP_ADDR}" "[ -f ${DEPLOY_PATH}/data/node-${NODE_ID}/deploy_node-${NODE_ID}.conf ]"
        if [[ $? -eq 1 ]]; then
            printLog "error" "TRANSFER ${DEPLOY_PATH}/data/node-${NODE_ID}/deploy_node-${NODE_ID}.conf FAILED"
            return 1
        else
            echo "[${SCRIPT_ALIAS}] [node-${NODE_ID}] : Transfer deploy conf completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
            printLog "info" "Transfer ${DEPLOY_PATH}/data/node-${NODE_ID}/deploy_node-${NODE_ID}.conf completed"
        fi
    fi

    echo "[${SCRIPT_ALIAS}] [node-${NODE_ID}] : Transfer files completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
    printLog "success" "Transfer files to Node-${NODE_ID} succeeded"
}

################################################# Transfer #################################################
function transfer() {
    if [[ "${ALL}" == "true" ]]; then
        cd "${PROJECT_CONF_PATH}"
        for file in $(ls ./); do
            if [ -f "${file}" ]; then
                transferFiles "${PROJECT_CONF_PATH}/${file}"
                if [[ $? -eq 1 ]]; then
                    printLog "error" "TRANSFER FILES TO NODE-${NODE_ID} FAILED"
                    exit 1
                fi
            fi
            cd "${PROJECT_CONF_PATH}"
        done
    else
        cd "${PROJECT_CONF_PATH}"
        for node_id in $(echo "${NODE}" | sed 's/,/\n/g'); do
            if [ ! -f "${PROJECT_CONF_PATH}/deploy_node-${node_id}.conf" ]; then
                printLog "error" "FILE deploy_node-${node_id}.conf NOT FOUND"
                exit 1
            fi              
            transferFiles "${PROJECT_CONF_PATH}/deploy_node-${node_id}.conf"
            if [[ $? -eq 1 ]]; then
                printLog "error" "TRANSFER FILES TO NODE-${NODE_ID} FAILED"
                exit 1
            fi
            cd "${PROJECT_CONF_PATH}"
        done
    fi
    printLog "info" "Transfer completed"
}

################################################# Main #################################################
function main() {
    checkEnv
    
    setupDirectoryStructure
    transfer
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
