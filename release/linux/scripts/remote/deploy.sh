#!/bin/bash

###########################################################################################################
################################################# VRIABLES #################################################
###########################################################################################################
SCRIPT_NAME="$(basename ${0})"
SCRIPT_ALIAS="$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')"
LOCAL_IP="127.0.0.1"
DEPLOYMENT_PATH=$(
    cd $(dirname $0)
    cd ../../../
    pwd
)
USER_NAME=$USER

DEPLOYMENT_CONF_PATH="${DEPLOYMENT_PATH}/deployment_conf"
if [ ! -d "${DEPLOYMENT_CONF_PATH}" ]; then
    mkdir -p ${DEPLOYMENT_CONF_PATH}
fi
PROJECT="test"
PROJECT_CONF_PATH="${DEPLOYMENT_CONF_PATH}/${PROJECT}"

NODE="all"
MODE="conf"
ADDRESS=""

#############################################################################################################
################################################# FUNCTIONS #################################################
#############################################################################################################

################################################# Help #################################################
function help() {
    echo
    echo "
USAGE: ${SCRIPT_NAME}  [options] [value]

        OPTIONS:

           --project, -p              the specified project name. must be specified.

           --node, -n                 the specified node name. only used in conf mode. 
                                      default='all': deploy all nodes by conf in deployment_conf
                                      use ',' to seperate the name of node

           --mode, -m                 the specified deploy mode. 
                                      default='conf': deploy node by exist node deployment conf
                                      'one': automatically generate one node's deployment conf file and build the blockchain on local
                                      'four': automatically generate four nodes' deployment conf file and build the blockchain on local

           --address, -a              the specified node address. only used in conf mode.               
                                      nodes' deployment file will be generated automatically if set

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

################################################# Yes Or No #################################################
function yesOrNo() {
    read anw
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

################################################# CONF #################################################
function conf() {
    if [[ "${ADDRESS}" != "" ]]; then
        if [ -d "${PROJECT_CONF_PATH}" ]; then
            printLog "question" "${PROJECT_CONF_PATH} has already been existed, do you want to continue?"
            yesOrNo
            if [[ $? -ne 1 ]]; then
                exit
            fi
        fi
        ${DEPLOYMENT_PATH}/linux/scripts/remote/prepare.sh -p "${PROJECT}" -a "${ADDRESS}" --cover
    elif [[ "${NODE}" == "" ]]; then
        printLog "error" "NODE MUST BE SET IN CONF MODE"
        help
        exit
    elif [ ! -d "${PROJECT_CONF_PATH}" ]; then
        printLog "error" "${PROJECT_CONF_PATH} NOT CREATED"
        help
        exit
    fi

    cd ${PROJECT_CONF_PATH}
    for f in $(ls ./); 
    do
        if [ ! -f "${f}" ]; then
            continue
        fi
        node_id=$(echo ${f} | sed -e 's/\(.*\)deploy_node-\(.*\).conf/\2/g')
        ${DEPLOYMENT_PATH}/linux/scripts/remote/transfer.sh -p "${PROJECT}" -n ${node_id}
        ${DEPLOYMENT_PATH}/linux/scripts/remote/init.sh -p "${PROJECT}" -n ${node_id}
        ${DEPLOYMENT_PATH}/linux/scripts/remote/start.sh -p "${PROJECT}" -n ${node_id}
        cd ${PROJECT_CONF_PATH}
    done
}

################################################# ONE #################################################
function one() {
    if [ -d "${PROJECT_CONF_PATH}" ]; then
        printLog "question" "${PROJECT_CONF_PATH} has already been existed, do you want to overwrite it?"
        yesOrNo
        if [[ $? -ne 1 ]]; then
            exit
        fi
    fi
    ${DEPLOYMENT_PATH}/linux/scripts/remote/prepare.sh -p "${PROJECT}" -a "${USER_NAME}@${LOCAL_IP}"
    ${DEPLOYMENT_PATH}/linux/scripts/remote/transfer.sh -p "${PROJECT}"
    ${DEPLOYMENT_PATH}/linux/scripts/remote/init.sh -p "${PROJECT}"
    ${DEPLOYMENT_PATH}/linux/scripts/remote/start.sh -p "${PROJECT}"
}

################################################# FOUR #################################################
function four() {
    if [ -d "${PROJECT_CONF_PATH}" ]; then
        printLog "question" "${PROJECT_CONF_PATH} has already been existed, do you want to overwrite it?"
        yesOrNo
        if [[ $? -ne 1 ]]; then
            exit
        fi
    fi
    ${DEPLOYMENT_PATH}/linux/scripts/remote/prepare.sh -p "${PROJECT}" -a "${USER_NAME}@${LOCAL_IP},${USER_NAME}@${LOCAL_IP},${USER_NAME}@${LOCAL_IP},${USER_NAME}@${LOCAL_IP}"
    cd ${PROJECT_CONF_PATH}
    for f in $(ls ./); 
    do
        if [ ! -f "${f}" ]; then
            continue
        fi
        node_id=$(echo ${f} | sed -e 's/\(.*\)deploy_node-\(.*\).conf/\2/g')
        ${DEPLOYMENT_PATH}/linux/scripts/remote/transfer.sh -p "${PROJECT}" -n ${node_id}
        ${DEPLOYMENT_PATH}/linux/scripts/remote/init.sh -p "${PROJECT}" -n ${node_id}
        ${DEPLOYMENT_PATH}/linux/scripts/remote/start.sh -p "${PROJECT}" -n ${node_id}
        cd ${PROJECT_CONF_PATH}
    done
}

################################################# Main #################################################
function deploy() {
    case "${MODE}" in
    "conf")
        conf
        ;;
    "one")
        one
        ;;
    "four")
        four
        ;;
    *)
        printLog "error" "MODE ${MODE} NOT FOUND"
        help
        exit
        ;;
    esac
}

################################################# Main #################################################
function main() {
    deploy
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
        if [[ "$2" != "" ]]; then
            PROJECT=$2
        fi
        PROJECT_CONF_PATH="${DEPLOYMENT_CONF_PATH}/${PROJECT}"
        printLog "info" "Project's conf path: ${PROJECT_CONF_PATH}"
        ;;
    --node | -n)
        NODE=$2
        printLog "info" "Node ${NODE} will be deployed"
        ;;
    --mode | -m)
        MODE=$2
        ;;
    --address | -a)
        ADDRESS=$2
        ;;
    *)
        printLog "error" "COMMAND \"$1\" NOT FOUND"
        help
        exit
        ;;
    esac
    shiftOption2 $#
    shift 2
done
main
