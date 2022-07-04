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
USER_NAME=$USER
LOCAL_IP="127.0.0.1"
SCRIPT_NAME="$(basename ${0})"
SCRIPT_ALIAS="$(echo ${CURRENT_PATH}/${SCRIPT_NAME} | sed -e 's/\(.*\)\/scripts\/\(.*\).sh/\2/g')"
PROJECT_NAME=""
NODE=""
INTERPRETER=""
VALIDATOR_NODES=""
MODE=""
ADDRESS=""
ALL=""

PROJECT_CONF_PATH=""

## param
interpreter=""
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

            --project, -p               the project name, must be specified

            --interpreter, -i           Select virtual machine interpreter, must be specified for new project
                                        \"wasm\", \"evm\" and \"all\" are supported
                                        (default: all)

            --validatorNodes, -v        set the genesis validatorNodes, must be specified for new project
                                        (default: the first node enode code)

            --mode, -m                  the specified deploy mode
                                        \"conf\", \"one\", \"four\" are supported (default: conf)
                                        \"conf\": deploy node by exist node deploy conf file
                                        \"one\": deploy a one-node chain locally
                                        \"four\": deploy a four-nodes chain locally

            --node, -n                  the node name, only used in conf mode
                                        use \",\" to seperate the name of node

            --address, -addr            the specified node address, only used in conf mode
                                        deploy conf file will be generated automatically

            --all, -a                   deploy all nodes

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
    read anw
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

################################################# Check Env #################################################
function checkEnv() {
    PROJECT_CONF_PATH="${DEPLOYMENT_CONF_PATH}/projects/${PROJECT_NAME}"

    if [[ "${PROJECT_NAME}" == "" ]]; then
        printLog "error" "PROJECT NAME NOT SET"
        exit 1
    fi
}

################################################# Assaign Default #################################################
function assignDefault() {
    INTERPRETER="all"
    MODE="conf"
}

################################################# Read Param #################################################
function readParam() {
    if [[ "${interpreter}" != "" ]]; then
        INTERPRETER="${interpreter}"
    fi

    if [[ "${mode}" != "" ]]; then
        MODE="${mode}"
    fi
}

################################################# Deploy One Node #################################################
function deployOneNode() {
    node_id="${1}"
    "${DEPLOYMENT_PATH}/linux/scripts"/venachainctl.sh remote transfer -p "${PROJECT_NAME}" -n "${node_id}"
    if [[ $? -eq 1 ]]; then
        return 1
    fi
    "${DEPLOYMENT_PATH}/linux/scripts"/venachainctl.sh remote init --project "${PROJECT_NAME}" --interpreter "${INTERPRETER}" --validatorNodes "${VALIDATOR_NODES}" --node "${node_id}"
    if [[ $? -eq 1 ]]; then
        return 1
    fi
    "${DEPLOYMENT_PATH}/linux/scripts"/venachainctl.sh remote start -p "${PROJECT_NAME}" -n "${node_id}"
    if [[ $? -eq 1 ]]; then
        return 1
    fi
}

################################################# CONF #################################################
function conf() {
    if [[ "${ADDRESS}" == "" ]] && [[ "${ALL}" != "true" ]] && [[ "${NODE}" == "" ]]; then
        printLog "error" "NODE'S ADDRESS OR NODE'S NAME MUST BE SET"
        exit 1
    fi

    if [[ "${ADDRESS}" != "" ]]; then
        "${DEPLOYMENT_PATH}/linux/scripts"/venachainctl.sh remote prepare -p "${PROJECT_NAME}" -addr "${ADDRESS}"
        res=$?
        if [[ ${res} -ne 0 ]]; then
            exit ${res}
        fi
        
        addr_num=$(expr $(echo "${ADDRESS}" | grep -o "," | wc -l) + 1)
        NODE=""
        while read line;
        do
            node_id="$(echo ${line} | sed -e 's/\(.*\)\[node-\(.*\)\]\(.*\)/\2/g')"
            if [[ "${NODE}" != "" ]]; then
                NODE="${NODE},"
            fi
            NODE="${NODE}${node_id}"
        done <<< "$(cat ${DEPLOYMENT_CONF_PATH}/logs/prepare_log.txt | grep "\[${PROJECT_NAME}\]" | tail -${addr_num})"
    fi

    cd "${PROJECT_CONF_PATH}"
    if [[ "${ALL}" == "true" ]]; then
        for file in $(ls ./); 
        do
            if [ ! -f "${file}" ]; then
                continue
            fi
            node_id=$(echo ${file} | sed -e 's/\(.*\)deploy_node-\(.*\).conf/\2/g')
            deployOneNode "${node_id}"
            if [[ $? -eq 1 ]]; then
                printLog "error" "DEPLOY NODE-${node_id} FAILED"
                exit 1
            fi
            cd "${PROJECT_CONF_PATH}"
        done
    else 
        for node_id in $(echo "${NODE}" | sed 's/,/\n/g'); do
            if [ ! -f "${PROJECT_CONF_PATH}/deploy_node-${node_id}.conf" ]; then
                printLog "error" "FILE deploy_node-${node_id}.conf NOT FOUND"
                exit 1
            fi
            deployOneNode "${node_id}"
            if [[ $? -eq 1 ]]; then
                printLog "error" "DEPLOY NODE-${node_id} FAILED"
                exit 1
            fi
            cd "${PROJECT_CONF_PATH}"
        done
    fi
}

################################################# ONE #################################################
function one() {
    if [ -d "${PROJECT_CONF_PATH}" ]; then
        printLog "question" "${PROJECT_CONF_PATH} has already been existed, do you want to overwrite it?"
        yesOrNo
        if [[ $? -ne 1 ]]; then
            exit 2
        fi
    fi

    "${DEPLOYMENT_PATH}/linux/scripts"/venachainctl.sh remote prepare -p "${PROJECT_NAME}" -addr "${USER_NAME}@${LOCAL_IP}" --cover
    if [[ $? -eq 1 ]]; then
        exit 1
    fi
    deployOneNode "0"
    if [[ $? -eq 1 ]]; then
        printLog "error" "DEPLOY ONE FAILED"
        exit 1
    fi
}

################################################# FOUR #################################################
function four() {
    if [ -d "${PROJECT_CONF_PATH}" ]; then
        printLog "question" "${PROJECT_CONF_PATH} has already been existed, do you want to overwrite it?"
        yesOrNo
        if [[ $? -ne 1 ]]; then
            exit 2
        fi
    fi

    "${DEPLOYMENT_PATH}/linux/scripts"/venachainctl.sh remote prepare -p "${PROJECT_NAME}" -addr "${USER_NAME}@${LOCAL_IP},${USER_NAME}@${LOCAL_IP},${USER_NAME}@${LOCAL_IP},${USER_NAME}@${LOCAL_IP}" --cover
    if [[ $? -eq 1 ]]; then
        exit 1
    fi
    cd "${PROJECT_CONF_PATH}"
    for file in $(ls ./); 
    do
        if [ ! -f "${file}" ]; then
            continue
        fi
        node_id=$(echo ${file} | sed -e 's/\(.*\)deploy_node-\(.*\).conf/\2/g')
        deployOneNode "${node_id}"
        if [[ $? -eq 1 ]]; then
            printLog "error" "DEPLOY FOUR FAILED"
        exit 1
    fi
        cd "${PROJECT_CONF_PATH}"
    done
}

################################################# Deploy #################################################
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
        exit 1
        ;;
    esac
}

################################################# Main #################################################
function main() {
    checkEnv
    assignDefault
    readParam

    deploy
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
    --interpreter | -i)
        shiftOption2 $#
        interpreter="${2}"
        shift 2
        ;;
    --validatorNodes | -v)
        shiftOption2 $#
        VALIDATOR_NODES="${2}"
        shift 2
        ;;
    --mode | -m)
        shiftOption2 $#
        mode="${2}"
        shift 2
        ;;
    --node | -n)
        shiftOption2 $#
        NODE="${2}"
        shift 2
        ;;
    --address | -addr)
        shiftOption2 $#
        ADDRESS="${2}"
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
