#!/bin/bash

###########################################################################################################
################################################# VRIABLES #################################################
###########################################################################################################

## path
CURRENT_PATH="$(cd "$(dirname "$0")";pwd)"
PROJECT_PATH="$(
    cd $(dirname ${0})
    cd ../
    pwd
)"
SCRIPT_PATH="${PROJECT_PATH}/scripts"
DATA_PATH="${PROJECT_PATH}/data"

## global
SCRIPT_NAME="$(basename ${0})"
SCRIPT_ALIAS="$(echo ${CURRENT_PATH}/${SCRIPT_NAME} | sed -e 's/\(.*\)\/scripts\/\(.*\).sh/\2/g')"
NODE_ID=""
NODE_DIR=""
IP_ADDR=""
P2P_PORT=""
VALIDATOR_NODES=""
INTERPRETER=""
AUTO=""

## param
interpreter=""
ip_addr=""
p2p_port=""

#############################################################################################################
################################################# FUNCTIONS #################################################
#############################################################################################################

################################################# Help #################################################
function help() {
    echo
    echo "
USAGE: ${SCRIPT_NAME}  [options] [value]

        OPTIONS:

            --nodeid, -n                the first node id (default: 0)

            --interpreter, -i           Select virtual machine interpreter
                                        \"wasm\", \"evm\" and \"all\" are supported (default: all)

            --validatorNodes, -v        set the genesis validatorNodes
                                        (default: the first node enode code)

            --ip                        the first node ip (default: 127.0.0.1)

            --p2p_port, -p2p            the first node p2p_port (default: 16791)

            --auto                      will read exit the deploy conf, node key and skip ip check

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
        exit 1
    fi
}

################################################# Yes Or No #################################################
function yesOrNo() {
    read -p "" anw
    case "${anw}" in
    [Yy][Ee][Ss] | [yY])
        return 0
        ;;
    [Nn][Oo] | [Nn])
        return 1
        ;;
    esac
    return 1
}

################################################# Save Config #################################################
function saveConf() {
    conf="${NODE_DIR}/deploy_node-${NODE_ID}.conf"
    conf_tmp="${NODE_DIR}/deploy_node-${NODE_ID}.tmp.conf"
    if [[ "${2}" == "" ]]; then
        return 0
    fi
    cat "${conf}" | sed "s#${1}=.*#${1}=${2}#g" | cat >"${conf_tmp}"
    mv "${conf_tmp}" "${conf}"
}

################################################# Check Conf #################################################
function checkConf() {
    ref="$(cat "${DEPLOY_CONF}" | grep "${1}=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    if [[ "${ref}" != "" ]]; then
        return 1
    fi
    return 0
}

################################################# Check Env #################################################
function checkEnv() {
    if [[ "${NODE_ID}" == "" ]]; then
        NODE_ID="0"
    fi
    NODE_DIR="${DATA_PATH}/node-${NODE_ID}"
    DEPLOY_CONF="${NODE_DIR}/deploy_node-${NODE_ID}.conf"
}

################################################# Assign Default #################################################
function assignDefault() { 
    INTERPRETER="all"
    IP_ADDR="127.0.0.1"
    P2P_PORT="16791"
}

################################################# Read File #################################################
function readFile() {
    checkConf "ip_addr"
    if [[ $? -eq 1 ]]; then
        IP_ADDR="$(cat ${DEPLOY_CONF} | grep "ip_addr=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    fi

    checkConf "p2p_port"
    if [[ $? -eq 1 ]]; then
        P2P_PORT="$(cat ${DEPLOY_CONF} | grep "p2p_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    fi
}

################################################# Read Param #################################################
function readParam() {
    if [[ "${interpreter}" != "" ]]; then
        INTERPRETER="${interpreter}"
    fi
    if [[ "${ip_addr}" != "" ]]; then
        IP_ADDR="${ip_addr}"
    fi
    if [[ "${p2p_port}" != "" ]]; then
        P2P_PORT="${p2p_port}"
    fi
}

################################################# Check Ip #################################################
function checkIp() {
    ip=$1
    check=$(echo $ip|awk -F. '$1<=255&&$2<=255&&$3<=255&&$4<=255{print "yes"}')
    if echo $ip|grep -E "^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$">/dev/null; then
        if [ ${check:-no} == "yes" ]; then
            return 1
        fi
    fi
    return 0
}

################################################# Setup Genesis #################################################
function setupGenesis() {
    ## setup flag
    flag_auto=""
    if [[ "${AUTO}" == "true" ]]; then
        flag_auto=" --auto "
    fi

    ## generate deploy conf
    "${SCRIPT_PATH}"/venachainctl.sh dcgen -n "${NODE_ID}" ${flag_auto}
    res=$?
    if [[ ${res} -eq 1 ]]; then
        exit 1
    fi
    if [[ ${res} -eq 2 ]]; then
        if [[ -f "${NODE_DIR}/deploy_node-${NODE_ID}.conf" ]]; then
            printLog "warn" "Deploy conf Exists, Will Read Them Automatically"
        else 
            printLog "warn" "Please Put Your Deploy Conf File to the directory ${NODE_DIR}"
            exit 1
        fi
    fi

    assignDefault
    readFile
    readParam
    saveConf "p2p_port" "${P2P_PORT}"

    ## generate key
    "${SCRIPT_PATH}"/venachainctl.sh keygen -n "${NODE_ID}" ${flag_auto} 
    res=$?
    if [[ ${res} -eq 1 ]]; then
        exit 1
    fi
    if [[ ${res} -eq 2 ]]; then
        if [[ -f "${DATA_PATH}/node-${NODE_ID}/node.prikey" ]] && [[ -f "${DATA_PATH}/node-${NODE_ID}/node.pubkey" ]] && [[ -f "${DATA_PATH}/node-${NODE_ID}/node.address" ]]; then
            printLog "warn" "Key Already Exists, Will Read Them Automatically"
        else 
            printLog "warn" "Please Put Your Nodekey file \"node.prikey\",\"node.pubkey\",\"node.address\" to the directory ${NODE_DIR}"
            exit 1
        fi
    fi

    ## generate genesis
    ## check ip
    if [[ "${AUTO}" != "true" ]]; then
        checkIp "${IP_ADDR}"
        if [ $? -ne 1 ]; then
            while true
            do
                printLog "warn" "Invalid Ip!"
                printLog "question" "Please in put a valid ip address: "
                read input 
                checkIp "${input}"
                if [ $? -eq 1 ]; then
                    IP_ADDR="${input}"
                    break
                fi
            done
        fi
    fi
    saveConf "ip_addr" "${IP_ADDR}"

    "${SCRIPT_PATH}"/venachainctl.sh gengen -n "${NODE_ID}" -i "${INTERPRETER}" -v "${VALIDATOR_NODES}" --ip "${IP_ADDR}" -p2p "${P2P_PORT}" 
    if [[ $? -eq 1 ]]; then
        exit 1
    fi
}

################################################# Main #################################################
function main() {
    checkEnv
    assignDefault
    readParam

    setupGenesis
}

###########################################################################################################
#################################################  EXECUTE #################################################
###########################################################################################################
while [ ! $# -eq 0 ]; do
    case "${1}" in
    --nodeid | -n)
        shiftOption2 $#
        NODE_ID="${2}"
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
    --ip)
        shiftOption2 $#
        ip_addr="${2}"
        shift 2
        ;;
    --p2p_port | -p2p)
        p2p_port="${2}"
        shift 2
        ;;
    --auto)
        shiftOption2 $#
        AUTO="${2}"
        shift 2
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
