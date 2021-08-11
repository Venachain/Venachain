#!/bin/bash

###########################################################################################################
################################################# VRIABLES #################################################
###########################################################################################################
SCRIPT_NAME="$(basename ${0})"
OS=$(uname)
LOCAL_IP=$(ifconfig | grep inet | grep -v 127.0.0.1 | grep -v inet6 | awk '{print $2}')
if [[ "$(echo ${LOCAL_IP} | grep addr:)" != "" ]]; then
    LOCAL_IP=$(echo ${LOCAL_IP} | tr -s ':' ' ' | awk '{print $2}')
fi
DEPLOYMENT_PATH=$(
    cd $(dirname $0)
    cd ../../
    pwd
)
DEPLOYMENT_CONF_PATH="${DEPLOYMENT_PATH}/deployment_conf"
PROJECT_CONF_PATH=""
PROJECT_NAME=""

NODE="all"
MODE="deep"
SKIP=""

NODE_ID=""
USER_NAME=""
IP_ADDR=""
P2P_PORT=""
RPC_PORT=""
DEPLOY_PATH=""
BACKUP_PATH=""
FIRSTNODE_USER_NAME=""
FIRSTNODE_ID=""
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

           --project, -p              the specified project name. must be specified

           --node, -n                 the specified node name. 
                                      default='all': deploy all nodes by conf in deployment_conf
                                      use ',' to seperate the name of node

           --mode, -m                 the specified execute mode. 
                                      'delete': will delete the node from chain
                                      'clean': will clean the files, configuration files will be backed up
                                      'stop' : will stop the node 
                                      'deep': will do delete clean and stop
                                      default='deep'

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

################################################# Yes Or No #################################################
function yesOrNo() {
    read -p "" anw
    case $anw in
    [Yy][Ee][Ss] | [yY])
        return 0
        ;;
    [Nn][Oo] | [Nn])
        return 1
        ;;
    esac
    return 1
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
    USER_NAME=$(cat $1 | grep "user_name=" | sed -e 's/user_name=\(.*\)/\1/g')
    IP_ADDR=$(cat $1 | grep "ip_addr=" | sed -e 's/ip_addr=\(.*\)/\1/g')
    P2P_PORT=$(cat $1 | grep "p2p_port=" | sed -e 's/p2p_port=\(.*\)/\1/g')
    RPC_PORT=$(cat $1 | grep "rpc_port=" | sed -e 's/rpc_port=\(.*\)/\1/g')
    DEPLOY_PATH=$(cat $1 | grep "deploy_path=" | sed -e 's/deploy_path=\(.*\)/\1/g')

    if [[ "${USER_NAME}" == "" ]] || [[ "${IP_ADDR}" == "" ]] || [[ "${P2P_PORT}" == "" ]] || [[ "${DEPLOY_PATH}" == "" ]] || [[ "${RPC_PORT}" == "" ]]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* DEPLOY CONF MISS VALUE **********"
        return 1
    fi
    xcmd "${USER_NAME}@${IP_ADDR}" "[ -d ${DEPLOY_PATH} ]"
    if [ $? -eq 0 ]; then
        BACKUP_PATH=${DEPLOY_PATH}/../bak/${PROJECT_NAME}
        BIN_PATH=${DEPLOY_PATH}/bin
        CONF_PATH=${DEPLOY_PATH}/conf
    fi

    firstnode_info="${PROJECT_CONF_PATH}/global/firstnode.info"
    if [ -f ${firstnode_info} ]; then
        FIRSTNODE_USER_NAME=$(cat ${firstnode_info} | grep "user_name=" | sed -e 's/user_name=\(.*\)/\1/g')
        FIRSTNODE_ID=$(cat ${firstnode_info} | grep "node_id=" | sed -e 's/node_id=\(.*\)/\1/g')
        FIRSTNODE_IP_ADDR=$(cat ${firstnode_info} | grep "ip_addr=" | sed -e 's/ip_addr=\(.*\)/\1/g')
        FIRSTNODE_RPC_PORT=$(cat ${firstnode_info} | grep "rpc_port=" | sed -e 's/rpc_port=\(.*\)/\1/g')
    fi
}

################################################# Check Remote Access #################################################
function checkRemoteAccess() {

    if [[ "${IP_ADDR}" == "${LOCAL_IP}" ]] || [[ "${IP_ADDR}" == "127.0.0.1" ]]; then
        return 0
    fi

    ## check ip connection
    ping -c 3 -w 3 "${IP_ADDR}" >/dev/null 2>&1
    if [[ $? -ne 0 ]]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* ${IP_ADDR} IS DOWN **********"
        return 1
    fi
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Check ip ${IP_ADDR} connection completed"

    ## check ssh connection
    timeout 3 ssh "${USER_NAME}@${IP_ADDR}" echo "permission" >/dev/null 2>&1
    if [[ $? -ne 0 ]]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* ${USER_NAME}@${IP_ADDR} DO NOT SUPPORT PASSWORDLESS ACCCESS **********"
        return 1
    fi
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Check ssh ${USER_NAME}@${IP_ADDR} access completed"
}

################################################# Delete Node #################################################
function deleteNode() {

    ## skip if clean all node
    if [[ "${MODE}" == "deep" ]] && [[ "${NODE}" == "all" ]]; then
        return 0
    fi
    checkRemoteAccess "$1"

    ## firstnode's info not valid
    if [[ "${FIRSTNODE_IP_ADDR}" == "" ]] || [[ "${FIRSTNODE_RPC_PORT}" == "" ]] || [[ "${FIRSTNODE_USER_NAME}" == "" ]] || [[ "${FIRSTNODE_ID}" == "" ]]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* FIRSTNODE INFO NOT VALID, PLEASE CHECK ${PROJECT_CONF_PATH}/global/firstnode.info **********"
        return 1
    fi

    ## check firstnode
    xcmd "${FIRSTNODE_USER_NAME}@${FIRSTNODE_IP_ADDR}" "lsof -i:${FIRSTNODE_RPC_PORT}" 1>/dev/null
    if [[ $? -ne 0 ]]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* DELETE NODE NODE-${NODE_ID} FAILED, FIRSTNODE IS DOWN *********"
        return 1
    fi

    delete_node_flag=$(xcmd "${USER_NAME}@${IP_ADDR}" "${BIN_PATH}/platonecli node delete \"${NODE_ID}\" --keyfile ${CONF_PATH}/keyfile.json --url ${FIRSTNODE_IP_ADDR}:${FIRSTNODE_RPC_PORT} <${CONF_PATH}/keyfile.phrase")
    if [[ $(echo "${delete_node_flag}" | grep "success") == "" ]]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* DELETE NODE NODE-${NODE_ID} FAILED, MAY BE IS DOWN *********"
        return 1
    fi
}

################################################# Stop Node #################################################
function stopNode() {
    checkRemoteAccess "$1"
    if [[ $? -ne 0 ]]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* CHECK REMOTE ACCESS TO NODE-${NODE_ID} FAILED **********"
        return 1
    fi
    pid_info=$(xcmd "${USER_NAME}@${IP_ADDR}" "lsof -i:${RPC_PORT}")
    pid=$(echo ${pid_info} | awk '{ print $11 }')
    if [[ $? -ne 0 ]] || [[ "${pid}" == "" ]]; then
        echo "[WARN] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : !!! GET PID OF ${USER_NAME}@${IP_ADDR}:${RPC_PORT} FAILED, MAYBE HAS ALREADY BEEN STOPPED !!!"
        return 0
    fi
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Get PID of ${USER_NAME}@${IP_ADDR}:${RPC_PORT} completed"

    xcmd "${USER_NAME}@${IP_ADDR}" "kill -9 ${pid}"
    xcmd "${USER_NAME}@${IP_ADDR}" "lsof -i:${RPC_PORT}"
    if [[ $? -eq 0 ]]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* KILL PID OF ${USER_NAME}@${IP_ADDR}:${RPC_PORT} FAILED **********"
        return 1
    fi
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Kill PID ${pid} of ${USER_NAME}@${IP_ADDR}:${RPC_PORT} completed"
    if [ -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ]; then
        if [[ "${OS}" == "Darwin" ]]; then
            sed -i '' "/\[node-${NODE_ID}\] : Start node*/d" "${PROJECT_CONF_PATH}"/logs/deploy_log.txt
        else
            sed -i "/\[node-${NODE_ID}\] : Start node*/d" "${PROJECT_CONF_PATH}"/logs/deploy_log.txt
        fi
    fi
}

################################################# Clean Node #################################################
function cleanNode() {

    ## check env
    checkRemoteAccess "$1"
    if [[ $? -ne 0 ]]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* CHECK REMOTE ACCESS TO NODE-${NODE_ID} FAILED **********"
        return 1
    fi

    ## clean node data
    xcmd "${USER_NAME}@${IP_ADDR}" "[ -d ${DEPLOY_PATH}/data/node-${NODE_ID} ]"
    if [ $? -ne 0 ]; then
        echo "[WARN] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : !!! ${USER_NAME}@${IP_ADDR}:${DEPLOY_PATH}/data/node-${NODE_ID} NOT FOUND, MAYBE HAS ALREADY BEEN CLEANED !!!"
    else
        # backup deployment conf
        xcmd "${USER_NAME}@${IP_ADDR}" "[ -f ${DEPLOY_PATH}/data/node-${NODE_ID}/deploy_node-${NODE_ID}.conf ]"
        if [ $? -eq 0 ]; then
            timestamp=$(date '+%Y%m%d%H%sM%S')
            xcmd "${USER_NAME}@${IP_ADDR}" "mkdir -p ${BACKUP_PATH}"
            xcmd "${USER_NAME}@${IP_ADDR}" "mv ${DEPLOY_PATH}/data/node-${NODE_ID}/deploy_node-${NODE_ID}.conf ${BACKUP_PATH}/deploy_node-${NODE_ID}.conf.bak.${timestamp}"
            xcmd "${USER_NAME}@${IP_ADDR}" "[ -f ${BACKUP_PATH}/deploy_node-${NODE_ID}.conf.bak.${timestamp} ]"
            if [[ $? -ne 0 ]]; then
                echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* BACKUP ${USER_NAME}@${IP_ADDR}:${DEPLOY_PATH}/data/node-${NODE_ID}/deploy_node-${NODE_ID}.conf FAILED **********"
                return 1
            fi
            echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Backup ${USER_NAME}@${IP_ADDR}:${DEPLOY_PATH}/data/node-${NODE_ID}/deploy_node-${NODE_ID}.conf to ${USER_NAME}@${IP_ADDR}:${BACKUP_PATH}/deploy_node-${NODE_ID}.conf.bak.${timestamp} completed"
        fi

        # remove node dir
        xcmd "${USER_NAME}@${IP_ADDR}" "rm -rf ${DEPLOY_PATH}/data/node-${NODE_ID}"
        xcmd "${USER_NAME}@${IP_ADDR}" "[ -d ${DEPLOY_PATH}/data/node-${NODE_ID} ]"
        if [ $? -eq 0 ]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* REMOVE ${USER_NAME}@${IP_ADDR}:${DEPLOY_PATH}/data/node-${NODE_ID} FAILED **********"
            return 1
        elif [ -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ]; then
            if [[ "${OS}" == "Darwin" ]]; then
                sed -i '' "/\[*\] \[node-${NODE_ID}\] : */d" ${PROJECT_CONF_PATH}/logs/deploy_log.txt
            else
                sed -i "/\[*\] \[node-${NODE_ID}\] : */d" ${PROJECT_CONF_PATH}/logs/deploy_log.txt
            fi
            echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Remove ${USER_NAME}@${IP_ADDR}:${DEPLOY_PATH}/data/node-${NODE_ID} completed"
        fi
    fi

    ## check project
    cnt=0
    cd "${PROJECT_CONF_PATH}"
    for f in $(ls ./); do
        if [ ! -f "${f}" ]; then
            continue
        fi
        node_id=$(echo ${f} | sed -e 's/\(.*\)deploy_node-\(.*\).conf/\2/g')
        xcmd "${USER_NAME}@${IP_ADDR}" "[ -d ${DEPLOY_PATH}/data/node-${node_id} ]"
        if [ $? -eq 0 ]; then
            cnt=$(expr ${cnt} + 1)
        fi
        cd "${PROJECT_CONF_PATH}"
    done
    if [[ ${cnt} -ne 0 ]]; then
        return 0
    fi

    ## clean project
    # clean scripts dir
    xcmd "${USER_NAME}@${IP_ADDR}" "[ -d ${DEPLOY_PATH}/scripts ]"
    if [ $? -eq 0 ]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "rm -rf ${DEPLOY_PATH}/scripts"
        xcmd "${USER_NAME}@${IP_ADDR}" "[ -d ${DEPLOY_PATH}/scripts ]"
        if [ $? -eq 0 ]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* REMOVE ${USER_NAME}@${IP_ADDR}:${DEPLOY_PATH}/scripts FAILED **********"
            return 1
        else
            echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Remove ${USER_NAME}@${IP_ADDR}:${DEPLOY_PATH}/scripts completed"
        fi
    fi

    # clean data dir
    xcmd "${USER_NAME}@${IP_ADDR}" "[ -d ${DEPLOY_PATH}/data ]"
    if [[ $? -eq 0 ]]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "rm -rf ${DEPLOY_PATH}/data"
        xcmd "${USER_NAME}@${IP_ADDR}" "[ -d ${DEPLOY_PATH}/data ]"
        if [[ $? -eq 0 ]]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* REMOVE ${USER_NAME}@${IP_ADDR}:${DEPLOY_PATH}/data FAILED **********"
            return 1
        else
            echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Remove ${USER_NAME}@${IP_ADDR}:${DEPLOY_PATH}/data completed"
        fi
    fi

    # clean bin dir
    xcmd "${USER_NAME}@${IP_ADDR}" "[ -d ${DEPLOY_PATH}/bin ]"
    if [[ $? -eq 0 ]]; then
        xcmd "${USER_NAME}@${IP_ADDR}" "rm -rf ${DEPLOY_PATH}/bin"
        xcmd "${USER_NAME}@${IP_ADDR}" "[ -d ${DEPLOY_PATH}/bin ]"
        if [[ $? -eq 0 ]]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* REMOVE ${USER_NAME}@${IP_ADDR}:${DEPLOY_PATH}/bin FAILED **********"
            return 1
        else
            echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Remove ${USER_NAME}@${IP_ADDR}:${DEPLOY_PATH}/bin completed"
        fi
    fi

    # backup conf dir
    xcmd "${USER_NAME}@${IP_ADDR}" "[ -d ${DEPLOY_PATH}/conf ]"
    if [ $? -eq 0 ]; then
        timestamp=$(date '+%Y%m%d%H%sM%S')
        xcmd "${USER_NAME}@${IP_ADDR}" "mkdir -p ${BACKUP_PATH}"
        xcmd "${USER_NAME}@${IP_ADDR}" "mv ${DEPLOY_PATH}/conf ${BACKUP_PATH}/conf"
        xcmd "${USER_NAME}@${IP_ADDR}" "mv ${BACKUP_PATH} ${BACKUP_PATH}.bak.${timestamp}"
        xcmd "${USER_NAME}@${IP_ADDR}" "[ -d ${BACKUP_PATH}.bak.${timestamp} ]"
        if [[ $? -ne 0 ]]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* BACKUP ${USER_NAME}@${IP_ADDR}:${DEPLOY_PATH}/conf FAILED **********"
            return 1
        else
            xcmd "${USER_NAME}@${IP_ADDR}" "rm -rf ${DEPLOY_PATH}"
            echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Backup ${USER_NAME}@${IP_ADDR}:${DEPLOY_PATH}/conf completed"
        fi
    fi

    if [ -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ]; then
        path_id=$(echo ${USER_NAME}@${IP_ADDR}:${DEPLOY_PATH} | sed 's/\//#/g')
        if [[ "${OS}" == "Darwin" ]]; then
            sed -i '' "/\[*\] \[${path_id}\] : */d" ${PROJECT_CONF_PATH}/logs/deploy_log.txt
        else
            sed -i "/\[*\] \[${path_id}\] : */d" ${PROJECT_CONF_PATH}/logs/deploy_log.txt
        fi
    fi

    if [ -f "${PROJECT_CONF_PATH}/logs/deploy_log.txt" ] && [[ $(cat ${PROJECT_CONF_PATH}/logs/deploy_log.txt) == "" ]]; then
        if [[ "${SKIP}" != "true" ]]; then
            echo "Do you want to remove ${PROJECT_CONF_PATH}/global and ${PROJECT_CONF_PATH}/logs? Yes or No(y/n):"
            yesOrNo
            if [ $? -ne 0 ]; then
                return 0
            fi
        fi
        rm -rf ${PROJECT_CONF_PATH}/global
        rm -rf ${PROJECT_CONF_PATH}/logs
        if [ -d ${PROJECT_CONF_PATH}/global ] || [ -d ${PROJECT_CONF_PATH}/logs ]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* BACKUP ${USER_NAME}@${IP_ADDR}:${DEPLOY_PATH}/conf FAILED **********"
            return 1
        fi
    fi
}

################################################# Clear Node #################################################
function clearNode() {

    ## delete mode
    if [[ "${MODE}" == "delete" ]]; then
        deleteNode "$1"
        if [[ $? -ne 0 ]]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* DELETE NODE-${NODE_ID} FAILED **********"
            return 1
        fi
        echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Delete node-${NODE_ID} end"

    ## stop mode
    elif [[ "${MODE}" == "stop" ]]; then
        stopNode "$1"
        if [[ $? -ne 0 ]]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* STOP NODE-${NODE_ID} FAILED **********"
            return 1
        fi
        echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Stop node-${NODE_ID} end"

    ## clean mode
    elif [[ "${MODE}" == "clean" ]]; then
        cleanNode "$1"
        if [[ $? -ne 0 ]]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* CLEAN NODE-${NODE_ID} FAILED **********"
            return 1
        fi
        echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Clean node-${NODE_ID} end"

    ## deep mode
    elif [[ "${MODE}" == "deep" ]]; then
        deleteNode "$1"
        if [[ $? -ne 0 ]]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* DELETE NODE-${NODE_ID} FAILED **********"
            return 1
        fi
        echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Delete node-${NODE_ID} end"

        stopNode "$1"
        if [[ $? -ne 0 ]]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* STOP NODE-${NODE_ID} FAILED **********"
            return 1
        fi
        echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Stop node-${NODE_ID} end"

        cleanNode "$1"
        if [[ $? -ne 0 ]]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* CLEAN NODE-${NODE_ID} FAILED **********"
            return 1
        fi
        echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Clean node-${NODE_ID} end"
    else
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* MODE ${MODE} NOT FOUND **********"
    fi
}

################################################# Clear All Node #################################################
function clearAllNode() {
    cd "${PROJECT_CONF_PATH}"
    for file in $(ls ./); do
        if [ ! -f "${file}" ]; then
            continue
        fi
        NODE_ID=""
        USER_NAME=""
        IP_ADDR=""
        P2P_PORT=""
        DEPLOY_PATH=""

        NODE_ID=$(echo "${file}" | sed -e 's/\(.*\)deploy_node-\(.*\).conf/\2/g')

        echo
        echo "#### Start to clear Node-${NODE_ID} ####"
        readFile "${file}"
        if [[ $? -ne 0 ]]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* READ FILE ${file} FAILED **********"
            cd "${PROJECT_CONF_PATH}"
            continue
        fi
        clearNode "${file}"
        if [[ $? -ne 0 ]]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* CLEAR NODE-${NODE_ID} FAILED **********"
        fi
        cd "${PROJECT_CONF_PATH}"
    done
}

################################################# Clear Specified Node #################################################
function clearSpecifiedNode() {
    cd "${PROJECT_CONF_PATH}"
    for name in $(echo "${NODE}" | sed 's/,/\n/g'); do
        NODE_ID=""
        USER_NAME=""
        IP_ADDR=""
        P2P_PORT=""
        DEPLOY_PATH=""

        NODE_ID="${name}"
        file="deploy_node-${name}.conf"
        if [ ! -f "${PROJECT_CONF_PATH}/${file}" ]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* ${PROJECT_CONF_PATH}/${file} NOT EXISTS **********"
            return 1
        fi

        echo
        echo "################ Start to clear Node-${NODE_ID} ################"
        readFile "${file}"
        if [[ "${NODE_ID}" == "${FIRSTNODE_ID}" ]]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : !!! If Clear Firstnode, Many Services Will Not Be Usable !!!"
            echo "Are you sure to clear firstnode node-${NODE_ID}? Yes or No(y/n):"
            yesOrNo
            if [ $? -ne 0 ]; then
                continue
            fi
        fi

        clearNode "${file}"
        if [[ $? -ne 0 ]]; then
            echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* CLEAR NODE-${NODE_ID} FAILED **********"
        fi
        cd "${PROJECT_CONF_PATH}"
    done
}

################################################# Main #################################################
function main() {
    if [ ! -d "${PROJECT_CONF_PATH}" ]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* ${PROJECT_CONF_PATH} NOT EXISTS **********"
        exit
    fi

    if [[ "${NODE}" == "all" ]]; then
        clearAllNode
    else
        clearSpecifiedNode
    fi
    echo
    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Clear action end "
}

###########################################################################################################
#################################################  EXECUTE #################################################
###########################################################################################################
if [ $# -eq 0 ]; then
    help
    exit
fi
while [ ! $# -eq 0 ]; do
    case $1 in
    --project | -p)
        shiftOption2 $#
        PROJECT_NAME=$2
        PROJECT_CONF_PATH="${DEPLOYMENT_CONF_PATH}/$2"
        shift 2
        ;;
    --node | -n)
        shiftOption2 $#
        NODE=$2
        shift 2
        ;;
    --mode | -m)
        shiftOption2 $#
        MODE=$2
        shift 2
        ;;
    --skip)
        SKIP="true"
        shift 1
        ;;
    *)
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* COMMAND \"$1\" NOT FOUND **********"
        help
        exit
        ;;
    esac
done
main
