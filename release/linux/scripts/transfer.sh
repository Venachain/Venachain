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
DEPLOYMENT_CONF_PATH=$(cd ${DEPLOYMENT_PATH}/deployment_conf && pwd)
DEPLOYMENT_FILE_PATH="$(cd ${DEPLOYMENT_PATH}/release/linux && pwd)"
PROJECT_CONF_PATH=""

NODE="all"

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
####       transfer file to nodes      ####
###########################################'
}

################################################# Help #################################################
function help() {
    echo
    echo "
USAGE: transfer.sh  [options] [value]

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

################################################# Clear Data #################################################
function clearData() {
    NODE_ID=""
    DEPLOY_PATH=""
    USER_NAME=""
    IP_ADDR=""
    P2P_PORT=""
}

################################################# Read File #################################################
function readFile() {
    file=$1
    DEPLOY_PATH=$(cat ${file} | grep "deploy_path=" | sed -e 's/deploy_path=\(.*\)/\1/g')
    USER_NAME=$(cat ${file} | grep "user_name=" | sed -e 's/user_name=\(.*\)/\1/g')
    IP_ADDR=$(cat ${file} | grep "ip_addr=" | sed -e 's/ip_addr=\(.*\)/\1/g')
    P2P_PORT=$(cat ${file} | grep "p2p_port=" | sed -e 's/p2p_port=\(.*\)/\1/g')

    if [[ "${USER_NAME}" == "" ]] || [[ "${IP_ADDR}" == "" ]] || [[ "${P2P_PORT}" == "" ]] || [[ "${DEPLOY_PATH}" == "" ]]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* FILE ${file} MISS VALUE **********"
        return 1
    fi
}

################################################# Transfer #################################################
function transfer() {
    ## read file
    clearData
    file=$1
    NODE_ID=$(echo "${file}" | sed -e 's/\(.*\)deploy_node-\(.*\).conf/\2/g')
    if [ -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] && [[ $(grep $(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g') "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "${NODE_ID}" | grep "Transfer files") != "" ]]; then
        return 0
    fi
    readFile "${file}"
    echo
    echo "################ Transfer file to Node-${NODE_ID} ################"

    path_id=$(echo ${USER_NAME}@${IP_ADDR}:${DEPLOY_PATH} | sed 's/\//#/g')

    ## create directories
    # create conf directory
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ $(grep $(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g') "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "${path_id}" | grep "Create conf directory") == "" ]]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "mkdir -p ${DEPLOY_PATH}/conf"
        xcmd "${USER_NAME}@${IP_ADDR}" "[ -d ${DEPLOY_PATH}/conf ]"
        if [[ $? -ne 0 ]]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* CREATE ${DEPLOY_PATH}/conf FAILED **********"
            return 1
        fi
        echo "[$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] [${path_id}] : Create conf directory completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Create ${DEPLOY_PATH}/conf completed"
    fi
    # create script directory
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ $(grep $(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g') "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "${path_id}" | grep "Create scripts directory") == "" ]]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "mkdir -p ${DEPLOY_PATH}/scripts"
        xcmd "${USER_NAME}@${IP_ADDR}" "[ -d ${DEPLOY_PATH}/scripts ]"
        if [[ $? -ne 0 ]]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* CREATE ${DEPLOY_PATH}/scripts FAILED **********"
            return 1
        fi
        echo "[$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] [${path_id}] : Create scripts directory completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Create ${DEPLOY_PATH}/scripts completed"
    fi
    # create bin directory
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ $(grep $(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g') "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "${path_id}" | grep "Create bin directory") == "" ]]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "mkdir -p ${DEPLOY_PATH}/bin"
        xcmd "${USER_NAME}@${IP_ADDR}" "[ -d ${DEPLOY_PATH}/bin ]"
        if [[ $? -ne 0 ]]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* CREATE ${DEPLOY_PATH}/bin FAILED **********"
            return 1
        fi
        echo "[$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] [${path_id}] : Create bin directory completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Create ${DEPLOY_PATH}/bin completed"
    fi

    ## transfer files
    # transfer conf file
    cd ${DEPLOYMENT_FILE_PATH}/conf
    for f in $(ls ./); do
        if [ ! -f "${f}" ]; then
            continue
        fi
        if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ $(grep $(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g') "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "${path_id}" | grep "Transfer ${f}") == "" ]]; then
            xcmd "${USER_NAME}@${IP_ADDR}" "cp -r ${f} ${DEPLOY_PATH}/conf" "target"
            xcmd "${USER_NAME}@${IP_ADDR}" "[ -f ${DEPLOY_PATH}/conf/${f} ]"
            if [[ $? -ne 0 ]]; then
                echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* TRANSFER ${DEPLOY_PATH}/conf/${f} FAILED **********"
                return 1
            else
                echo "[$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] [${path_id}] : Transfer ${f} completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
                echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Transfer ${DEPLOY_PATH}/conf/${f} completed"
            fi
        fi
        cd ${DEPLOYMENT_FILE_PATH}/conf
    done
    # transfer script files
    cd ${DEPLOYMENT_FILE_PATH}/scripts
    for f in $(ls ./); do
        if [ ! -f "${f}" ]; then
            continue
        fi
        if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ $(grep $(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g') "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "${path_id}" | grep "Transfer ${f}") == "" ]]; then
            xcmd "${USER_NAME}@${IP_ADDR}" "cp -r ${f} ${DEPLOY_PATH}/scripts" "target"
            xcmd "${USER_NAME}@${IP_ADDR}" "[ -f ${DEPLOY_PATH}/scripts/${f} ]"
            if [[ $? -ne 0 ]]; then
                echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* TRANSFER ${DEPLOY_PATH}/scripts/${f} FAILED **********"
                return 1
            else
                echo "[$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] [${path_id}] : Transfer ${f} completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
                echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Transfer ${DEPLOY_PATH}/scripts/${f} completed"
            fi
        fi
        cd ${DEPLOYMENT_FILE_PATH}/scripts
    done
    # transfer bin file
    cd ${DEPLOYMENT_FILE_PATH}/bin
    for f in $(ls ./); do
        if [ ! -f "${f}" ]; then
            continue
        fi
        if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ $(grep $(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g') "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "${path_id}" | grep "Transfer ${f}") == "" ]]; then
            xcmd "${USER_NAME}@${IP_ADDR}" "cp -r ${f} ${DEPLOY_PATH}/bin" "target"
            xcmd "${USER_NAME}@${IP_ADDR}" "[ -f ${DEPLOY_PATH}/bin/${f} ]"
            if [[ $? -ne 0 ]]; then
                echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* TRANSFER ${DEPLOY_PATH}/bin/${f} FAILED **********"
                return 1
            else
                echo "[$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] [${path_id}] : Transfer ${f} completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
                echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Transfer ${DEPLOY_PATH}/bin/${f} completed"
            fi
        fi
        cd ${DEPLOYMENT_FILE_PATH}/bin
    done

    ## create deploy conf directory
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ $(grep $(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g') "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "node-${NODE_ID}" | grep "Create deploy conf directory") == "" ]]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "mkdir -p ${DEPLOY_PATH}/data/node-${NODE_ID}/deploy_conf"
        xcmd "${USER_NAME}@${IP_ADDR}" "[ -d ${DEPLOY_PATH}/data/node-${NODE_ID}/deploy_conf ]"
        if [[ $? -ne 0 ]]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* CREATE ${DEPLOY_PATH}/data/node-${NODE_ID}/deploy_conf FAILED **********"
            return 1
        fi
        echo "[$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] [node-${NODE_ID}] : Create deploy conf directory completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
        echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Create ${DEPLOY_PATH}/data/node-${NODE_ID}/deploy_conf completed"
    fi

    # transfer deploy conf file
    if [ ! -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] || [[ $(grep $(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g') "${PROJECT_CONF_PATH}/logs/deploy_log.txt" | grep "node-${NODE_ID}" | grep "Transfer deploy conf") == "" ]]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "cp -r ${file} ${DEPLOY_PATH}/data/node-${NODE_ID}/deploy_conf" "target"
        xcmd "${USER_NAME}@${IP_ADDR}" "[ -f ${DEPLOY_PATH}/data/node-${NODE_ID}/deploy_conf/deploy_node-${NODE_ID}.conf ]"
        if [[ $? -ne 0 ]]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* TRANSFER ${DEPLOY_PATH}/data/node-${NODE_ID}/deploy_conf/deploy_node-${NODE_ID}.conf  FAILED **********"
            return 1
        else
            echo "[$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] [node-${NODE_ID}] : Transfer deploy conf completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
            echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Transfer ${DEPLOY_PATH}/data/node-${NODE_ID}/deploy_conf/deploy_node-${NODE_ID}.conf completed"
        fi
    fi
    echo "[$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] [node-${NODE_ID}] : Transfer files completed" >>"${PROJECT_CONF_PATH}/logs/deploy_log.txt"
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Transfer files to Node-${NODE_ID} completed"
}

################################################# Main #################################################
function main() {
    mkdir -p ${PROJECT_CONF_PATH}/global && mkdir -p ${PROJECT_CONF_PATH}/logs
    if [ ! -d ${PROJECT_CONF_PATH}/global ] || [ ! -d ${PROJECT_CONF_PATH}/logs ]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* GLOBAL OR LOGS DIRECTORY UNDER PROJECT DEPLOYMENT PATH NOT EXIST **********"
        exit
    fi

    if [[ ${NODE} == "all" ]]; then
        cd ${PROJECT_CONF_PATH}
        for file in $(ls ./); do
            if [ -f "$file" ]; then
                transfer "${PROJECT_CONF_PATH}/$file"
                if [[ $? -ne 0 ]]; then
                    echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* TRANSFER FILES TO NODE-${NODE_ID} FAILED **********"
                    exit
                fi
            fi
            cd ${PROJECT_CONF_PATH}
        done

    else
        cd ${PROJECT_CONF_PATH}
        for param in $(echo "${NODE}" | sed 's/,/\n/g'); do
            if [ -f "${PROJECT_CONF_PATH}/deploy_node-${param}.conf" ]; then
                transfer "${PROJECT_CONF_PATH}/deploy_node-${param}.conf"
                if [[ $? -ne 0 ]]; then
                    echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* TRANSFER FILES TO NODE-${NODE_ID} FAILED **********"
                    exit
                fi
            else
                echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* FILE deploy_node-${param}.conf NOT EXISTS **********"
            fi
            cd ${PROJECT_CONF_PATH}
        done
    fi
    echo
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Transfer completed"
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
        if [ ! -d ${DEPLOYMENT_CONF_PATH}/$2 ]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* ${DEPLOYMENT_CONF_PATH}/$2 have not been created **********"
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
