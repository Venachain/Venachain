#!/bin/bash

###########################################################################################################
################################################# VRIABLES #################################################
###########################################################################################################
SCRIPT_NAME="$(basename ${0})"
LOCAL_IP="127.0.0.1"
DEPLOYMENT_PATH=$(
    cd $(dirname $0)
    cd ../../
    pwd
)
USER_NAME=$USER

DEPLOYMENT_CONF_PATH="${DEPLOYMENT_PATH}/deployment_conf"
if [ ! -d "${DEPLOYMENT_CONF_PATH}" ]; then
    mkdir -p ${DEPLOYMENT_CONF_PATH}
fi
PROJECT="test"
PROJECT_CONF_PATH=""

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

################################################# CONF #################################################
function conf() {
    if [[ "${ADDRESS}" != "" ]]; then
        if [ -d "${PROJECT_CONF_PATH}" ]; then
            echo "${PROJECT_CONF_PATH} has already been existed, do you want to cover it?"
            yesOrNo
            if [[ $? -ne 0 ]]; then
                exit
            fi
        fi
        ./prepare.sh -p "${PROJECT}" -a "${ADDRESS}" --cover
    elif [[ "${NODE}" == "" ]]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* NODE MUST BE SET IN CONF MODE **********"
        help
        exit
    elif [ ! -d "${PROJECT_CONF_PATH}" ]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* ${PROJECT_CONF_PATH} NOT CREATED **********"
        help
        exit
    fi

    ./transfer.sh -p "${PROJECT}" -n "${NODE}"
    ./init.sh -p "${PROJECT}" -n "${NODE}"
    ./start.sh -p "${PROJECT}" -n "${NODE}"
}

################################################# ONE #################################################
function one() {
    if [ -d "${PROJECT_CONF_PATH}" ]; then
        echo "${PROJECT_CONF_PATH} has already been existed, do you want to cover it?"
        yesOrNo
        if [[ $? -ne 0 ]]; then
            exit
        fi
    fi
    ./prepare.sh -p "${PROJECT}" -a "${USER_NAME}@${LOCAL_IP}"
    ./transfer.sh -p "${PROJECT}"
    ./init.sh -p "${PROJECT}"
    ./start.sh -p "${PROJECT}"
}

################################################# FOUR #################################################
function four() {
    if [ -d "${PROJECT_CONF_PATH}" ]; then
        echo "${PROJECT_CONF_PATH} has already been existed, do you want to cover it?"
        yesOrNo
        if [[ $? -ne 0 ]]; then
            exit
        fi
    fi
    ./prepare.sh -p "${PROJECT}" -a "${USER_NAME}@${LOCAL_IP},${USER_NAME}@${LOCAL_IP},${USER_NAME}@${LOCAL_IP},${USER_NAME}@${LOCAL_IP}"
    ./transfer.sh -p "${PROJECT}"
    ./init.sh -p "${PROJECT}"
    ./start.sh -p "${PROJECT}"
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
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* MODE ${MODE} NOT FOUND **********"
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
        echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Project's conf path: ${PROJECT_CONF_PATH}"
        ;;
    --node | -n)
        NODE=$2
        echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Node ${NODE} will be deployed"
        ;;
    --mode | -m)
        MODE=$2
        ;;
    --address | -a)
        ADDRESS=$2
        ;;
    *)
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* COMMAND \"$1\" NOT FOUND **********"
        help
        exit
        ;;
    esac
    shiftOption2 $#
    shift 2
done
main
