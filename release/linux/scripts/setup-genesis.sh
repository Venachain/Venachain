#!/bin/bash

###########################################################################################################
################################################# VRIABLES #################################################
###########################################################################################################

## path
SCRIPT_NAME="$(basename ${0})"
SCRIPT_ALIAS="$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')"
OS=$(uname)
PROJECT_PATH=$(
    cd $(dirname $0)
    cd ../
    pwd
)

## param
nodeid=""
ip_addr=""
p2p_port=""
validator_nodes=""
interpreter=""
auto=""

#############################################################################################################
################################################# FUNCTIONS #################################################
#############################################################################################################

################################################# Help #################################################
function help() {
    echo
    echo "
USAGE: ${SCRIPT_NAME}  [options] [value]

        OPTIONS:

           --nodeid, -n                 the first node id (default: 0)

           --ip                         the first node ip (default: 127.0.0.1)

           --p2p_port, -p               the first node p2p_port (default: 16791)

           --validatorNodes, -v         set the genesis validatorNodes
                                        (default: the first node enode code)

           --interpreter, -i            Select virtual machine interpreter in wasm, evm, all
                                        (default: all)

           --auto, -a                   will no prompt to create the node key and skip ip check
                                        
           --help, -h                   show help
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

################################################# Main #################################################
function main() {
    if [[ "${nodeid}" == "" ]]; then
        nodeid="0"
    fi

    flag_ip_addr=""
    flag_p2p_port=""
    flag_validator_nodes=""
    flag_interpreter=""
    flag_auto=""
    if [[ "${ip_addr}" != "" ]]; then
        flag_ip_addr=" --ip ${ip_addr} "
    fi
    if [[ "${p2p_port}" != "" ]]; then
        flag_p2p_port=" -p ${p2p_port} "
    fi
    if [[ "${validator_nodes}" != "" ]]; then
        flag_validator_nodes=" -v ${validator_nodes} "
    fi
    if [[ "${interpreter}" != "" ]] ; then
        flag_interpreter=" -i ${interpreter} "
    fi
    if [[ "${auto}" == "true" ]]; then
        flag_auto=" --auto "
    fi

    ./local/generate-deployconf.sh -n ${nodeid} ${flag_auto}
    ./local/generate-key.sh -n ${nodeid} ${flag_auto}
    ./local/generate-genesis.sh -n ${nodeid} ${flag_ip_addr} ${flag_p2p_port} ${flag_validator_nodes} ${flag_interpreter} ${flag_auto}
}

###########################################################################################################
#################################################  EXECUTE #################################################
###########################################################################################################
while [ ! $# -eq 0 ]; do
    case $1 in
    --nodeid | -n)
        shiftOption2 $#
        nodeid=$2
        shift 2
        ;;
    --ip)
        shiftOption2 $#
        ip_addr=$2
        shift 2
        ;;
    --p2p_port | -p)
        p2p_port=$2
        shift 2
        ;;
    --validatorNodes | -v)
        shiftOption2 $#
        validator_nodes=$2
        shift 2
        ;;
    --interpreter | -i)
        shiftOption2 $#
        interpreter=$2
        shift 2
        ;;
    --auto | -a)
        shiftOption2 $#
        auto=$2
        shift 2
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
