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
CONF_PATH="${DEPLOYMENT_PATH}/linux/conf"
SCRIPT_PATH="${DEPLOYMENT_PATH}/linux/scripts"

## global
OS=$(uname)
SCRIPT_NAME="$(basename ${0})"
SCRIPT_ALIAS="$(echo ${CURRENT_PATH}/${SCRIPT_NAME} | sed -e 's/\(.*\)\/scripts\/\(.*\).sh/\2/g')"
PROJECT_NAME=""
ADDRESS=""
COVER=""

PROJECT_CONF_PATH=""
USER_NAME=""
IP_ADDR=""
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

            --address, -addr            nodes' addresses, must be specified

            --cover                     backup the project directory if exists

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

################################################# Save Conf #################################################
function saveConf() {
    deploy_conf="${PROJECT_CONF_PATH}/deploy_node-${1}.conf"
    deploy_conf_tmp="${PROJECT_CONF_PATH}/deploy_node-${1}.temp.conf"

    if ! [ -f "${deploy_conf}" ]; then
        printLog "error" "FILE ${deploy_conf} NOT FOUND"
        exit 1
    fi
    if [[ "${3}" == "" ]]; then
        return 1
    fi
    cat "${deploy_conf}" | sed "s#${2}=.*#${2}=${3}#g" | cat >"${deploy_conf_tmp}"
    mv "${deploy_conf_tmp}" "${deploy_conf}"
}

################################################# Check Env #################################################
function checkEnv() {
    PROJECT_CONF_PATH="${DEPLOYMENT_CONF_PATH}/projects/${PROJECT_NAME}"

    if [[ "${PROJECT_NAME}" == "" ]]; then
        printLog "error" "PROJECT NAME NOT SET"
        exit 1
    fi
    if [[ "${ADDRESS}" == "" ]]; then
        printLog "error" "NODE'S ADDRESS NOT SET"
        exit 1
    fi

    if [ ! -f "${CONF_PATH}/deploy.conf.template" ]; then
        printLog "error" "FILE ${CONF_PATH}/deploy.conf.template NOT FOUND"
        exit 1
    fi

    if [ ! -d "${DEPLOYMENT_CONF_PATH}" ]; then
        mkdir -p "${DEPLOYMENT_CONF_PATH}"
    fi
}

################################################# Clear Data #################################################
function clearData() {
    USER_NAME=""
    IP_ADDR=""
    IS_LOCAL=""
}

################################################# Check Exist File #################################################
function checkExistFile() {
    cd "${PROJECT_CONF_PATH}"

    if [ -f "${DEPLOYMENT_CONF_PATH}/logs/prepare_log.txt" ]; then
        if [[ "${OS}" == "Darwin" ]]; then
            sed -i '' "/\[${PROJECT_NAME}\] \[.*\] */d" "${DEPLOYMENT_CONF_PATH}/logs/prepare_log.txt"
        else
            sed -i "/\[${PROJECT_NAME}\] \[.*\] */d" "${DEPLOYMENT_CONF_PATH}/logs/prepare_log.txt"
        fi
    fi

    for file in $(ls ./); do
        if [ -d "${file}" ]; then
            continue
        fi

        node_id="$(echo $file | sed -e 's/\(.*\)deploy_node-\(.*\).conf/\2/g')"
        ip_addr="$(cat $file | grep "ip_addr=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
        p2p_port="$(cat $file | grep "p2p_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
        rpc_port="$(cat $file | grep "rpc_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
        ws_port="$(cat $file | grep "ws_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"

        if [[ "${ip_addr}" == "" ]] || [[ "${p2p_port}" == "" ]] || [[ "${rpc_port}" == "" ]] || [[ "${ws_port}" == "" ]]; then
            printLog "error" "FILE ${file} MISS VALUE"
            exit 1
        fi
        echo "[${PROJECT_NAME}] [node-${node_id}] ip_addr:${ip_addr} p2p_port:${p2p_port} rpc_port:${rpc_port} ws_port:${ws_port}" >>"${DEPLOYMENT_CONF_PATH}/logs/prepare_log.txt"
    done
    printLog "info" "Check exist configuration file completed"
}

################################################# Check Project Existence #################################################
function checkProjectExistence() {
    if [ ! -d "${PROJECT_CONF_PATH}" ]; then
        return 0
    fi
    checkExistFile

    ## only cover mode can cover existing files automatically
    if [[ "${COVER}" != "true" ]]; then
        printLog "question" "${PROJECT_CONF_PATH} already exists, do you want to cover it? Yes or No(y/n): "
        yesOrNo
        if [ $? -ne 1 ]; then
            printLog "question" "Do you mean you want to create new conf file in exist path? Yes or No(y/n): "
            yesOrNo
            if [ $? -ne 1 ]; then
              exit 2
            fi
            printLog "warn" "New Conf Files Will Be Generated In Exist Path"
            return 0
        fi
    fi

    "${SCRIPT_PATH}"/venachainctl.sh remote clear -p "${PROJECT_NAME}" --all
    if [ $? -eq 1 ]; then
        printLog "error" "CLEAR PROJECT ${PROJECT_NAME} FAILED"
        exit 1
    fi

    timestamp=$(date '+%Y%m%d%H%M%S')
    mkdir -p "${DEPLOYMENT_CONF_PATH}/bak"
    mv "${PROJECT_CONF_PATH}" "${DEPLOYMENT_CONF_PATH}/bak/${PROJECT_NAME}.bak.${timestamp}"
    if [ ! -d "${DEPLOYMENT_CONF_PATH}/bak/${PROJECT_NAME}.bak.${timestamp}" ]; then
        printLog "error" "BACKUP ${PROJECT_CONF_PATH} FAILED"
        exit 1
    fi
    if [ -d "${PROJECT_CONF_PATH}" ]; then
        printLog "error" "CLEAR PROJECT CONF PATH ${PROJECT_CONF_PATH} FAILED"
        exit 1
    fi
    if [ -f "${DEPLOYMENT_CONF_PATH}/logs/prepare_log.txt" ]; then
        if [[ "${OS}" == "Darwin" ]]; then
            sed -i '' "/\[${PROJECT_NAME}\] \[.*\] */d" "${DEPLOYMENT_CONF_PATH}/logs/prepare_log.txt"
        else
            sed -i "/\[${PROJECT_NAME}\] \[.*\] */d" "${DEPLOYMENT_CONF_PATH}/logs/prepare_log.txt"
        fi
    fi
    printLog "info" "Backup ${PROJECT_CONF_PATH} to ${DEPLOYMENT_CONF_PATH}/bak/${PROJECT_NAME}.bak.${timestamp} completed"
}

################################################# Set Up Directory Structure #################################################
function setupDirectoryStructure() {
    if [ ! -d "${DEPLOYMENT_CONF_PATH}/logs" ] || [ ! -d "${PROJECT_CONF_PATH}" ]; then
        mkdir -p "${DEPLOYMENT_CONF_PATH}/logs" && mkdir -p "${PROJECT_CONF_PATH}"
        if [ ! -d "${DEPLOYMENT_CONF_PATH}/logs" ] || [ ! -d "${PROJECT_CONF_PATH}" ]; then
            printLog "error" "SET UP DEIRECTORY STRUCTURE FAILED"
            exit 1
        else
            printLog "info" "Set up directory structure completed"
        fi
    fi
}

################################################# Generate Configuration File #################################################
function generateConfFile() {
    node_id="0"

    for remote_addr in $(echo "${ADDRESS}" | sed 's/,/\n/g'); do
        clearData
        echo
        echo "################ Generate Configuration File For ${remote_addr} Start ################"

        USER_NAME="$(echo "${remote_addr}" | sed -e 's/\(.*\)@\(.*\)/\1/g')"
        IP_ADDR="$(echo "${remote_addr}" | sed -e 's/\(.*\)@\(.*\)/\2/g')"
        p2p_port="16791"
        rpc_port="6791"
        ws_port="26791"

        ## check remote access
        res_check_port="true"
        checkRemoteAccess 
        if [[ $? -ne 0 ]]; then
            printLog "error" "CHECK REMOTE ACCESS TO ${IP_ADDR} FAILED"
            exit 1
        fi

        ## check node id
        while [ -f "${DEPLOYMENT_CONF_PATH}/logs/prepare_log.txt" ] && [[ $(grep "\[${PROJECT_NAME}\]" "${DEPLOYMENT_CONF_PATH}/logs/prepare_log.txt" | grep "node-${node_id}") != "" ]]; do
            node_id="$(expr ${node_id} + 1)"
        done

        ## check p2p port
        while [[ 0 -lt 1 ]]; do
            if [ -f "${DEPLOYMENT_CONF_PATH}/logs/prepare_log.txt" ] && [[ $(grep "ip_addr:${IP_ADDR}" "${DEPLOYMENT_CONF_PATH}/logs/prepare_log.txt" | grep "p2p_port:${p2p_port}") != "" ]]; then
                p2p_port="$(expr ${p2p_port} + 1)"
                continue
            fi
            if [[ $(xcmd "${remote_addr}" "lsof -i:${p2p_port}") == "" ]]; then
                break
            else
                p2p_port=$(expr ${p2p_port} + 1)
            fi
        done

        ## check rpc port
        while [[ 0 -lt 1 ]]; do
            if [ -f "${DEPLOYMENT_CONF_PATH}/logs/prepare_log.txt" ] && [[ $(grep "ip_addr:${IP_ADDR}" "${DEPLOYMENT_CONF_PATH}/logs/prepare_log.txt" | grep "rpc_port:${rpc_port}") != "" ]]; then
                rpc_port="$(expr ${rpc_port} + 1)"
                continue
            fi
            if [[ $(xcmd "${remote_addr}" "lsof -i:${rpc_port}") == "" ]]; then
                break
            else
                rpc_port="$(expr ${rpc_port} + 1)"
            fi
        done

        ## check ws port
        while [[ 0 -lt 1 ]]; do
            if [ -f "${DEPLOYMENT_CONF_PATH}/logs/prepare_log.txt" ] && [[ $(grep "ip_addr:${IP_ADDR}" "${DEPLOYMENT_CONF_PATH}/logs/prepare_log.txt" | grep "ws_port:${ws_port}") != "" ]]; then
                ws_port="$(expr ${ws_port} + 1)"
                continue
            fi
            if [[ $(xcmd "${remote_addr}" "lsof -i:${ws_port}") == "" ]]; then
                break
            else
                ws_port="$(expr ${ws_port} + 1)"
            fi
        done

        ## generate deploy conf file
        # generate project path
        home="home"
        os=$(xcmd "${USER_NAME}@${IP_ADDR}" "uname")
        if [[ "${os}" == "Darwin" ]]; then
            home="Users"
        fi
        deploy_path="/${home}/${USER_NAME}/Venachain/${PROJECT_NAME}"

        # write value
        cp ${CONF_PATH}/deploy.conf.template ${PROJECT_CONF_PATH}/deploy_node-${node_id}.conf
        saveConf "${node_id}" "deploy_path" "${deploy_path}"
        saveConf "${node_id}" "user_name" "${USER_NAME}"
        saveConf "${node_id}" "ip_addr" "${IP_ADDR}"
        saveConf "${node_id}" "p2p_port" "${p2p_port}"
        saveConf "${node_id}" "rpc_port" "${rpc_port}"
        saveConf "${node_id}" "ws_port" "${ws_port}"
        saveConf "${node_id}" "log_dir" "${deploy_path}/data/node-${node_id}/logs"

        if [[ ! -f "${PROJECT_CONF_PATH}/deploy_node-${node_id}.conf" ]]; then
            printLog "error" "GENERATE ${PROJECT_CONF_PATH}/deploy_node-${node_id}.conf for ${remote_addr} FAILED"
            exit 1
        fi
        echo "[${PROJECT_NAME}] [node-${node_id}] ip_addr:${IP_ADDR} p2p_port:${p2p_port} rpc_port:${rpc_port} ws_port:${ws_port}" >>"${DEPLOYMENT_CONF_PATH}/logs/prepare_log.txt"
        printLog "success" "Generate ${PROJECT_CONF_PATH}/deploy_node-${node_id}.conf for ${remote_addr} succeeded"
    done
}

################################################# Main #################################################
function main() {
    checkEnv

    checkProjectExistence
    setupDirectoryStructure
    generateConfFile
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
    --address | -addr)
        shiftOption2 $#
        ADDRESS="${2}"
        shift 2
        ;;
    --cover)
        COVER="true"
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
