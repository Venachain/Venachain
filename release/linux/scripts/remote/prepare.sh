#!/bin/bash

###########################################################################################################
################################################# VRIABLES #################################################
###########################################################################################################
SCRIPT_NAME="$(basename ${0})"
SCRIPT_ALIAS="$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')"
OS=$(uname)
DEPLOYMENT_PATH=$(
    cd $(dirname $0)
    cd ../../../
    pwd
)
CONF_PATH=$(
    cd $(dirname $0)
    cd ../../conf
    pwd
)
SCRIPT_PATH=$(
    cd $(dirname $0)
    cd ../../scripts
    pwd
)
DEPLOYMENT_CONF_PATH="${DEPLOYMENT_PATH}/deployment_conf"
if [ ! -d "${DEPLOYMENT_CONF_PATH}" ]; then
    mkdir -p ${DEPLOYMENT_CONF_PATH}
fi
PROJECT_NAME=""
PROJECT_CONF_PATH=""

REMOTE_ADDRS=""
COVER=""
IS_LOCAL=""

IP_ADDR=""

#############################################################################################################
################################################# FUNCTIONS #################################################
#############################################################################################################

################################################# Show Title #################################################
function showTitle() {
    echo '
###########################################
####       prepare default files       ####
###########################################'
}

################################################# Help #################################################
function help() {
    echo
    echo "
USAGE: ${SCRIPT_NAME}  [options] [value]

        OPTIONS:

           --project, -p             the specified project name. must be specified

           --address, -a             nodes' addresses. must be specified

           --cover                   will backup the project directory if exists

           --help, -h                show help
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

################################################# Yes Or No #################################################
function yesOrNo() {
    read -p "" anw
    case $anw in
    [Yy][Ee][Ss] | [yY])
        return 1
        ;;
    [Nn][Oo] | [Nn])
        return 0
        ;;
    esac
    return 0
}

################################################# Check Exist File #################################################
function checkExistFile() {
    cd "${PROJECT_CONF_PATH}"
    for file in $(ls ./); do
        if [ -d "${file}" ]; then
            continue
        fi

        node_id=$(echo $file | sed -e 's/\(.*\)deploy_node-\(.*\).conf/\2/g')
        ip_addr=$(cat $file | grep "ip_addr=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
        p2p_port=$(cat $file | grep "p2p_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
        rpc_port=$(cat $file | grep "rpc_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
        ws_port=$(cat $file | grep "ws_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')

        if [[ "${ip_addr}" == "" ]] || [[ "${p2p_port}" == "" ]] || [[ "${rpc_port}" == "" ]] || [[ "${ws_port}" == "" ]]; then
            printLog "error" "FILE ${file} MISS VALUE"
            exit
        fi
        echo "[${PROJECT_NAME}] [node-${node_id}] ip_addr:${ip_addr} p2p_port:${p2p_port} rpc_port:${rpc_port} ws_port:${ws_port}" >>"${DEPLOYMENT_CONF_PATH}/logs/prepare_log.txt"
    done
    printLog "info" "Check exist configuration file completed"
}

################################################# Check Execute Mode #################################################
function checkProjectExistence() {
    if [ ! -d "${PROJECT_CONF_PATH}" ]; then
        return 0
    fi

    ## only cover mode can execute cover mode automatically
    if [[ "${COVER}" != "true" ]]; then
        printLog "question" "${PROJECT_CONF_PATH} already exists, do you want to cover it? Yes or No(y/n): "
        yesOrNo
        if [ $? -ne 1 ]; then
            printLog "question" "Do you mean you want to create new conf file in exist path? Yes or No(y/n): "
            yesOrNo
            if [ $? -ne 1 ]; then
              exit
            fi
            printLog "warn" "New Conf Files Will Be Generated In Exist Path"
            return 0
        fi
        checkExistFile
    fi

    timestamp=$(date '+%Y%m%d%H%M%S')
    mkdir -p "${DEPLOYMENT_CONF_PATH}/bak"
    cp -r "${PROJECT_CONF_PATH}" "${DEPLOYMENT_CONF_PATH}/bak/${PROJECT_NAME}.bak.${timestamp}"
    if [ ! -d "${DEPLOYMENT_CONF_PATH}/bak/${PROJECT_NAME}.bak.${timestamp}" ]; then
        printLog "error" "BACKUP ${PROJECT_CONF_PATH} FAILED"
        exit
    fi
    if [[ -f "${DEPLOYMENT_CONF_PATH}/logs/prepare_log.txt" ]]; then
        if [[ "${OS}" == "Darwin" ]]; then
            sed -i '' "/\[${PROJECT_NAME}\]*/d" ${DEPLOYMENT_CONF_PATH}/logs/prepare_log.txt
        else
            sed -i "/\[${PROJECT_NAME}\]*/d" ${DEPLOYMENT_CONF_PATH}/logs/prepare_log.txt
        fi
    fi
    printLog "info" "Backup ${PROJECT_CONF_PATH} to ${DEPLOYMENT_CONF_PATH}/bak/${PROJECT_NAME}.bak.${timestamp} completed"

    ${SCRIPT_PATH}/remote/clear.sh -p ${PROJECT_NAME} --skip
    if [ $? -ne 0 ]; then
        printLog "error" "CLEAR PROJECT ${PROJECT_NAME} FAILED"
        exit
    fi

    rm -rf "${PROJECT_CONF_PATH}"
    if [ -d "${PROJECT_CONF_PATH}" ]; then
        printLog "error" "CLEAR PROJECT CONF PATH ${PROJECT_CONF_PATH} FAILED"
        exit
    fi
}

################################################# Set Up Directory Structure #################################################
function setupDirectoryStructure() {
    if [ ! -d "${PROJECT_CONF_PATH}/global" ] || [ ! -d "${PROJECT_CONF_PATH}/logs" ] || [ ! -d "${DEPLOYMENT_CONF_PATH}/logs" ]; then
        mkdir -p "${PROJECT_CONF_PATH}/global" && mkdir -p "${PROJECT_CONF_PATH}/logs" && mkdir -p "${DEPLOYMENT_CONF_PATH}/logs"
        if [ ! -d "${PROJECT_CONF_PATH}/global" ] || [ ! -d "${PROJECT_CONF_PATH}/logs" ] || [ ! -d "${DEPLOYMENT_CONF_PATH}/logs" ]; then
            printLog "error" "SET UP DEIRECTORY STRUCTURE FAILED"
            exit
        else
            printLog "info" "Set up directory structure completed"
        fi
    fi
}

################################################# Check Remote Access #################################################
function checkRemoteAccess() {
    if [[ $(ifconfig | grep ${IP_ADDR}) != "" ]]; then
      IS_LOCAL="true"
        return 0
    fi
    IS_LOCAL="false"

    ## check ip connection
    ping -c 3 -w 3 "$2" 1>/dev/null
    if [[ $? -ne 0 ]]; then
        printLog "error" "IP $2 IS DOWN"
        return 1
    fi
    printLog "info" "Check ip $2 connection completed"

    ## check ssh connection
    timeout 3 ssh "$1@$2" echo "permission" 1>/dev/null
    if [[ $? -ne 0 ]]; then
        printLog "error" "REMOTE $1@$2 DO NOT SUPPORT PASSWORDLESS ACCCESS"
        return 1
    fi
    printLog "info" "Check ssh $1@$2 access completed"
}

################################################# Save Conf #################################################
function saveConf() {
    node_conf=${PROJECT_CONF_PATH}/deploy_node-${1}.conf
    node_conf_tmp=${PROJECT_CONF_PATH}/deploy_node-${1}.temp.conf
    if [[ $3 == "" ]]; then
        return
    fi
    if ! [[ -f "${node_conf}" ]]; then
        printLog "error" "FILE ${node_conf} NOT FOUND"
        return
    fi
    cat $node_conf | sed "s#${2}=.*#${2}=${3}#g" | cat >$node_conf_tmp
    mv $node_conf_tmp $node_conf
}

################################################# Generate Configuration File #################################################
function generateConfFile() {
    node_id="0"

    for remote_addr in $(echo "${REMOTE_ADDRS}" | sed 's/,/\n/g'); do
        IP_ADDR=""
        echo
        echo "################ Generate Configuration File For ${remote_addr} Start ################"
        user_name=$(echo "${remote_addr}" | sed -e 's/\(.*\)@\(.*\)/\1/g')
        IP_ADDR=$(echo "${remote_addr}" | sed -e 's/\(.*\)@\(.*\)/\2/g')
        p2p_port="16791"
        rpc_port="6791"
        ws_port="26791"

        ## check remote access
        check_port_flag="true"
        checkRemoteAccess "${user_name}" "${IP_ADDR}"
        if [[ $? -ne 0 ]]; then
            printLog "question" "${remote_addr} cannot be accessed, do you still want to generate a configuration file for it? Yes or No(y/n): "
            yesOrNo
            if [[ $? -ne 1 ]]; then
                printLog "question"  "Do you want to continue to generate configuration files for other address? Yes or No(y/n): "
                yesOrNo
                if [[ $? -ne 1 ]]; then
                    exit
                fi
                continue
            fi
            check_port_flag="false"
        fi

        ## check node id
        while [ -f "${DEPLOYMENT_CONF_PATH}/logs/prepare_log.txt" ] && [[ $(grep "${PROJECT_NAME}" "${DEPLOYMENT_CONF_PATH}/logs/prepare_log.txt" | grep "node-${node_id}") != "" ]]; do
            node_id=$(expr ${node_id} + 1)
        done

        ## check p2p port
        while [[ 0 -lt 1 ]]; do
            if [ -f "${DEPLOYMENT_CONF_PATH}/logs/prepare_log.txt" ] && [[ $(grep "ip_addr:${IP_ADDR}" "${DEPLOYMENT_CONF_PATH}/logs/prepare_log.txt" | grep "p2p_port:${p2p_port}") != "" ]]; then
                p2p_port=$(expr ${p2p_port} + 1)
                continue
            fi
            if [[ "${check_port_flag}" == "true" ]]; then
                if [[ $(xcmd "${remote_addr}" "lsof -i:${p2p_port}") == "" ]]; then
                    break
                else
                    p2p_port=$(expr ${p2p_port} + 1)
                fi
            else
                break
            fi
        done

        ## check rpc port
        while [[ 0 -lt 1 ]]; do
            if [ -f "${DEPLOYMENT_CONF_PATH}/logs/prepare_log.txt" ] && [[ $(grep "ip_addr:${IP_ADDR}" "${DEPLOYMENT_CONF_PATH}/logs/prepare_log.txt" | grep "rpc_port:${rpc_port}") != "" ]]; then
                rpc_port=$(expr ${rpc_port} + 1)
                continue
            fi
            if [[ "${check_port_flag}" == "true" ]]; then
                if [[ $(xcmd "${remote_addr}" "lsof -i:${rpc_port}") == "" ]]; then
                    break
                else
                    rpc_port=$(expr ${rpc_port} + 1)
                fi
            else
                break
            fi
        done

        ## check ws port
        while [[ 0 -lt 1 ]]; do
            if [ -f "${DEPLOYMENT_CONF_PATH}/logs/prepare_log.txt" ] && [[ $(grep "ip_addr:${IP_ADDR}" "${DEPLOYMENT_CONF_PATH}/logs/prepare_log.txt" | grep "ws_port:${ws_port}") != "" ]]; then
                ws_port=$(expr ${ws_port} + 1)
                continue
            fi
            if [[ "${check_port_flag}" == "true" ]]; then
                if [[ $(xcmd "${remote_addr}" "lsof -i:${ws_port}") == "" ]]; then
                    break
                else
                    ws_port=$(expr ${ws_port} + 1)
                fi
            else
                break
            fi
        done

        ## write conf file
        home="home"
        os=$(xcmd "${user_name}@${IP_ADDR}" "uname")
        if [[ "${os}" == "Darwin" ]]; then
            home="Users"
        fi
        deploy_path="/${home}/${user_name}/PlatONE/${PROJECT_NAME}"

        cp ${CONF_PATH}/deploy.conf.template ${PROJECT_CONF_PATH}/deploy_node-${node_id}.conf
        saveConf "${node_id}" "deploy_path" "${deploy_path}"
        saveConf "${node_id}" "user_name" "${user_name}"
        saveConf "${node_id}" "ip_addr" "${IP_ADDR}"
        saveConf "${node_id}" "p2p_port" "${p2p_port}"
        saveConf "${node_id}" "rpc_port" "${rpc_port}"
        saveConf "${node_id}" "ws_port" "${ws_port}"
        saveConf "${node_id}" "log_dir" "${deploy_path}/data/node-${node_id}/logs"

        if [[ ! -f "${PROJECT_CONF_PATH}/deploy_node-${node_id}.conf" ]]; then
            printLog "error" "GENERATE ${PROJECT_CONF_PATH}/deploy_node-${node_id}.conf for ${remote_addr} FAILED"
            printLog "question" "Do you want to continue to generate configuration files for other address? Yes or No(y/n): "
            yesOrNo
            if [[ $? -ne 1 ]]; then
                exit
            fi
            continue
        fi
        echo "[${PROJECT_NAME}] [node-${node_id}] ip_addr:${IP_ADDR} p2p_port:${p2p_port} rpc_port:${rpc_port} ws_port:${ws_port}" >>"${DEPLOYMENT_CONF_PATH}/logs/prepare_log.txt"
        printLog "success" "Generate ${PROJECT_CONF_PATH}/deploy_node-${node_id}.conf for ${remote_addr} succeeded"
    done
}

################################################# Main #################################################
function main() {
    showTitle
    checkProjectExistence
    setupDirectoryStructure
    generateConfFile
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
        PROJECT_CONF_PATH=${DEPLOYMENT_CONF_PATH}/$2
        PROJECT_NAME=$2
        shift 2
        ;;
    --address | -a)
        shiftOption2 $#
        REMOTE_ADDRS=$2
        shift 2
        ;;
    --cover)
        COVER="true"
        shift 1
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
