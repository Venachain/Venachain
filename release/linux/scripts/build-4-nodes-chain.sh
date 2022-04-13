#!/bin/bash

###########################################################################################################
################################################# VRIABLES #################################################
###########################################################################################################

## path
PROJECT_PATH="$(
    cd $(dirname ${0})
    cd ../
    pwd
)"
SCRIPT_PATH="${PROJECT_PATH}/scripts"

## global
SCRIPT_NAME="$(basename ${0})"

#############################################################################################################
################################################# FUNCTIONS #################################################
#############################################################################################################

################################################# Help #################################################
function help() {
    echo
    echo "
USAGE: ${SCRIPT_NAME}  [options] [value]

        OPTIONS:

            --help, -h                  show help
"
}

################################################# Check Shift Option #################################################
function shiftOption2() {
    if [[ $1 -lt 2 ]]; then
        printLog "error" "MISS OPTION VALUE! PLEASE SET THE VALUE"
        exit
    fi
}

################################################# Four #################################################
function four() {
    "${SCRIPT_PATH}"/venachainctl.sh setupgen -n "0" --ip "127.0.0.1" --p2p_port "16791" --auto "true"
    if [[ $? -eq 1 ]]; then
        exit 1
    fi
    "${SCRIPT_PATH}"/venachainctl.sh init -n "0" --ip "127.0.0.1" --rpc_port "6791" --p2p_port "16791" --ws_port "26791" --auto "true"
    if [[ $? -eq 1 ]]; then
        exit 1
    fi
    "${SCRIPT_PATH}"/venachainctl.sh start -n "0"
    if [[ $? -eq 1 ]]; then
        exit 1
    fi
    "${SCRIPT_PATH}"/venachainctl.sh deploysys --auto "true"
    if [[ $? -eq 1 ]]; then
        exit 1
    fi

    "${SCRIPT_PATH}"/venachainctl.sh init -n "1" --ip "127.0.0.1" --rpc_port "6792" --p2p_port "16792" --ws_port "26792" --auto "true"
    if [[ $? -eq 1 ]]; then
        exit 1
    fi
    "${SCRIPT_PATH}"/venachainctl.sh start -n "1"
    if [[ $? -eq 1 ]]; then
        exit 1
    fi
    "${SCRIPT_PATH}"/venachainctl.sh addnode -n "1"
    if [[ $? -eq 1 ]]; then
        exit 1
    fi
    "${SCRIPT_PATH}"/venachainctl.sh updatesys -n "1"
    if [[ $? -eq 1 ]]; then
        exit 1
    fi

    "${SCRIPT_PATH}"/venachainctl.sh init -n "2" --ip "127.0.0.1" --rpc_port "6793" --p2p_port "16793" --ws_port "26793" --auto "true"
    if [[ $? -eq 1 ]]; then
        exit 1
    fi
    "${SCRIPT_PATH}"/venachainctl.sh start -n "2"
    if [[ $? -eq 1 ]]; then
        exit 1
    fi
    "${SCRIPT_PATH}"/venachainctl.sh addnode -n "2"
    if [[ $? -eq 1 ]]; then
        exit 1
    fi
    "${SCRIPT_PATH}"/venachainctl.sh updatesys -n "2"
    if [[ $? -eq 1 ]]; then
        exit 1
    fi

    "${SCRIPT_PATH}"/venachainctl.sh init -n "3" --ip "127.0.0.1" --rpc_port "6794" --p2p_port "16794" --ws_port "26794" --auto "true"
    if [[ $? -eq 1 ]]; then
        exit 1
    fi
    "${SCRIPT_PATH}"/venachainctl.sh start -n "3"
    if [[ $? -eq 1 ]]; then
        exit 1
    fi
    "${SCRIPT_PATH}"/venachainctl.sh addnode -n "3"
    if [[ $? -eq 1 ]]; then
        exit 1
    fi
    "${SCRIPT_PATH}"/venachainctl.sh updatesys -n "3"
    if [[ $? -eq 1 ]]; then
        exit 1
    fi
}

################################################# Main #################################################
function main() {
    assignDefault
    readParam

    four
}

###########################################################################################################
#################################################  EXECUTE #################################################
###########################################################################################################
while [ ! $# -eq 0 ]
do
    case "${1}" in
        *)
        printLog "error" "COMMAND \"${1}\" NOT FOUND"
        help
        exit 1
        ;;
    esac
done
main
