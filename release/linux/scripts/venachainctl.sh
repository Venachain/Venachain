#!/bin/bash
CURRENT_PATH="$(cd "$(dirname "$0")";pwd)"
WORKSPACE_PATH="$(cd "${CURRENT_PATH}"/..;pwd)"
BIN_PATH="${WORKSPACE_PATH}/bin"
CONF_PATH="${WORKSPACE_PATH}/conf"
SCRIPT_PATH="${WORKSPACE_PATH}/scripts"
DATA_PATH="${WORKSPACE_PATH}/data"

VERSION="PLEASE BUILD VENACHAIN FIRST"
if [ -f ${BIN_PATH}/venachain ]; then
    VERSION="$(${BIN_PATH}/venachain --version)"
fi
SCRIPT_NAME="$(basename ${0})"
SCRIPT_ALIAS="$(echo ${SCRIPT_PATH}/${SCRIPT_NAME} | sed -e 's/\(.*\)\/scripts\/\(.*\).sh/\2/g')"
NODE_ID="0"

ENABLE=""
DISENABLE=""

cd "${CURRENT_PATH}"

function usage() {
    cat <<EOF
#h0    DESCRIPTION
#h0        The deployment script for venachain
#h0
#h0    USAGE:
#h0        ${SCRIPT_NAME} <command> [command options] [arguments...]
#h0
#h0    COMMANDS
#c0        one                              start a node completely (default account password: 0)
#c1        one OPTIONS
#c1            --help, -h                   show help
#c0        four                             start four node completely (default account password: 0)
#c2        four OPTIONS
#c2            --help, -h                   show help
#c0        setupgen                         create the genesis.json and compile sys contract
#c3        setupgen OPTIONS
#c3            --nodeid, -n                 the first node id (default: 0)
#c3            --interpreter, -i            select virtual machine interpreter
#c3                                         "wasm", "evm" and "all" are supported (default: all)
#c3            --validatorNodes, -v         set the genesis validatorNodes
#c3                                         (default: the first node enode code)
#c3            --ip                         the first node ip (default: 127.0.0.1)
#c3            --p2p_port, -p2p             the first node p2p_port (default: 16791)
#c3            --auto                       will read exit the deploy conf, node key and skip ip check
#c3            --help, -h                   show help
#c0        init                             initialize node. please setup genesis first
#c4        init OPTIONS
#c4            --nodeid, -n                 set node id (default: 0)
#c4            --ip                         set node ip (default: 127.0.0.1)
#c4            --rpc_port, -rpc             set node rpc port (default: 6791)
#c4            --p2p_port, -p2p             set node p2p port (default: 16791)
#c4            --ws_port, -ws               set node ws port (default: 26791)
#c4            --auto                       will read exit the deploy conf and node key 
#c4            --help, -h                   show help
#c0        start                            try to start the specified node
#c5        start OPTIONS
#c5            --nodeid, -n                 start the specified node, must be specified
#c5            --bootnodes, -b              connect to the specified bootnodes node
#c5                                         (default: the first in the suggestObserverNodes in genesis.json)
#c5            --logsize, -s                Log block size (default: 67108864)
#c5            --logdir, -d                 log dir (default: ../data/node_dir/logs/)
#c5            --extraoptions, -e           extra venachain command options when venachain starts
#c5                                         (default: --debug)
#c5            --txcount, -c                max tx count in a block (default: 1000)
#c5            --tx_global_slots, -tgs      max tx count in txpool (default: 4096)
#c5            --lightmode, -l              select lightnode mode
#c5                                         "lightnode" and "lightserver" are supported
#c5            --dbtype                     select database type
#c5                                         "leveldb" and "pebbledb" are supported (default: leveldb)
#c5            --all, -a                    start all nodes
#c5            --help, -h                   show help
#c0        deploysys                        deploy the system contract
#c6        deploysys OPTIONS
#c6            --nodeid, -n                 the specified node id (default: 0)
#c6            --auto                       will use the default node password: 0
#c6                                         to create the account and also to unlock the account
#c6            --help, -h                   show help
#c0        addnode                          add normal node to system contract
#c7        addnode OPTIONS
#c7            --nodeid, -n                 the specified node id, must be specified
#c7            --desc                       the specified node desc
#c7            --ip                         the specified node ip
#c7                                         If the node specified by nodeid is local,
#c7                                         then you do not need to specify this option
#c7            --rpc_port, -rpc             the specified node rpc_port
#c7                                         If the node specified by nodeid is local,
#c7                                         then you do not need to specify this option
#c7            --p2p_port, -p2p             the specified node p2p_port
#c7                                         If the node specified by nodeid is local,
#c7                                         then you do not need to specify this option
#c7            --pubkey                     the specified node pubkey
#c7                                         If the node specified by nodeid is local,
#c7                                         then you do not need to specify this option
#c7            --type                       select specified node type in "2" & "3"
#c7                                         "2" is observer, "3" is lightnode (default: 2)
#c7            --help, -h                   show help
#c0        updatesys                        update node type
#c8        updatesys OPTIONS
#c8            --nodeid, -n                 the specified node id, must be specified
#c8            --content, -c                update content 
#c8                                         "consensus" and "observer" are supported (default: consensus)
#c8            --help, -h                   show help
#c0        dcgen                            generate deploy conf file
#c9        dcgen OPTIONS
#c9            --nodeid, -n                 the node's name, must be specified
#c9            --auto                       will read exist file
#c9            --help, -h                   show help
#c0        keygen                           generate key pair
#c10       keygen OPTIONS
#c10           --nodeid, -n                 the specified node name, must be specified
#c10           --auto                       will read exit node key 
#c10           --help, -h                   show help
#c0        gengen                           generate genesis.json file
#c11       gengen OPTIONS
#c11           --nodeid, -n                 the first node id, must be specified
#c11           --interpreter, -i            select virtual machine interpreter, must be specified
#c11                                        "wasm", "evm" and "all" are supported
#c11           --validatorNodes, -v         set the genesis validatorNodes
#c11                                        (default: the first node enode code)
#c11           --ip                         the first node ip
#c11           --p2p_port, -p2p             the first node p2p_port
#c11           --help, -h                   show help
#c0        addadmin                         add super admin role and chain admin role
#c12       addadmin OPTIONS
#c12           --nodeid, -n                 the specified node name, must be specified
#c12           --help, -h                   show help
#c0        delete                           try to delete the specified node
#c13       delete OPTIONS
#c13           --nodeid, -n                 the specified node id, must be specified
#c13           --help, -h                   show help
#c0        stop                             try to stop the specified node
#c14       stop OPTIONS
#c14           --nodeid, -n                 stop the specified node
#c14           --all, -a                    stop all node
#c14           --help, -h                   show help
#c0        restart                          try to restart the specified node
#c15       restart OPTIONS
#c15           --nodeid, -n                 restart the specified node
#c15           --all, -a                    restart all node
#c15           --help, -h                   show help
#c0        clear                            try to stop and clear the node data
#c16       clear OPTIONS
#c16           --nodeid, -n                 clear specified node data
#c16           --all, -a                    clear all nodes data
#c16           --help, -h                   show help
#c0        createacc                        create account
#c17       createacc OPTIONS
#c17           --nodeid, -n                 the specified node name, must be specified
#c17           --create_keyfile, -ck        will create keyfile
#c17           --auto                       will use the default node password: 0
#c17                                        to create the account and also to unlock the account
#c17           --help, -h                   show help
#c0        unlock                           unlock node account
#c18       unlock OPTIONS
#c18           --nodeid, -n                 unlock account on specified node
#c18           --help, -h                   show help
#c0        console                          start an interactive JavaScript environment
#c19       console OPTIONS
#c19           --opennodeid , -n            open the specified node console
#c19                                        set the node id here
#c19           --closenodeid, -c            stop the specified node console
#c19                                        set the node id here
#c19           --closeall                   stop all node console
#c19           --help, -h                   show help
#c0        remote                           remote deploy
#c20       remote OPTIONS
#c20           deploy                       deploy nodes
#c20           prepare                      generate directory structure and deployment conf file
#c20           transfer                     transfer necessary file to target node
#c20           init                         initialize the target node
#c20           start                        start the target node
#c20           clear                        clear the target node
#c20           --help, -h                   show help
#c0        status                           show all node status
#c21       status OPTIONS                   show all node status
#c21           --nodeid, -n                 show the specified node status info
#c21           --all, -a                    show all nodes status info
#c21           --help, -h                   show help
#c0        get                              display all nodes in the system contract
#c0        version                          show venachain release version
#c0    =====================================================================================================
#c0    INFORMATION
#c0        version         Venachain Version: ${VERSION}
#c0        author
#c0        copyright       Copyright
#c0        license
#c0
#c0    =====================================================================================================
#c0    HISTORY
#c0        2019/06/26  ty : create the deployment scripts
#c0        2021/12/10  wjw : update old deployment scripts 
#c0                          create the remote deployment scripts
#c0    =====================================================================================================
EOF
}

################################################# General Functions  #################################################
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

function shiftOption2() {
    if [[ $1 -lt 2 ]]; then
        printLog "error" "MISS OPTION VALUE! PLEASE SET THE VALUE"
        exit 1
    fi
}

function helpOption() {
    for op in "$@"
    do
        if [[ "${op}" == "--help" ]] || [[ "${op}" == "-h" ]]; then
            return 1
        fi
    done
}

function showUsage() {
    if [[ "${1}" == "" ]]; then
        usage | grep -e "^#[ch]0 " | sed -e "s/^#[ch][0-9]*//g"
    fi
    usage | grep -e "^#h0 \|^#c${1} " | sed -e "s/^#[ch][0-9]*//g"
}

function saveConf() {
    node_conf="${DATA_PATH}/node-${1}/deploy_node-${1}.conf"
    node_conf_tmp="${DATA_PATH}/node-${1}/deploy_node-${1}.temp.conf"
    if [[ "${3}" == "" ]]; then
        return
    fi
    if ! [ -f "${node_conf}" ]; then
        printLog "error" "FILE ${node_conf} NOT FOUND"
        exit 1
    fi
    cat "${node_conf}" | sed "s#${2}=.*#${2}=${3}#g" | cat >"${node_conf_tmp}"
    mv "${node_conf_tmp}" "${node_conf}"
}

function checkNodeStatusFullName() {
    if [ -d "${DATA_PATH}/${1}" ] && [[ "${1}" == node-* ]]; then
        nodeid="$(echo ${1#*-})"
        if [[ "$(ps -ef | grep "venachain --identity venachain --datadir ${DATA_PATH}/node-${nodeid} " | grep -v grep | awk '{print $2}')" != "" ]]; then
            ENABLE="$(echo "${ENABLE} ${nodeid}")"
        else
            DISENABLE="$(echo ${DISENABLE} ${nodeid})"
        fi
    fi
}

function checkAllNodeStatus() {
    nodes="$(ls ${DATA_PATH})"
    for n in ${nodes}; do
        checkNodeStatusFullName "${n}"
    done
}

function nodeIsRunning() {
    if [[ "$(ps -ef | grep "venachain --identity venachain --datadir ${DATA_PATH}/node-${1} " | grep -v grep | awk '{print $2}')" != "" ]]; then
        return 1
    fi
    return 0
}

################################################# Local Deploy Functions #################################################
function setupGenesis() {
    helpOption "$@"
    if [[ $? -eq 1 ]]; then
        showUsage 3
        exit 1
    fi
    "${SCRIPT_PATH}"/setup-genesis.sh "$@"
}

function init() {
    helpOption "$@"
    if [[ $? -eq 1 ]]; then
        showUsage 4
        exit 1
    fi
    "${SCRIPT_PATH}"/init-node.sh "$@"
}

function start() {
    nid=""
    bns=""
    log_size=""
    log_dir=""
    extra_options=""
    tx_count=""
    tx_global_slots=""
    lightmode=""
    dbtype=""
    all=""
    if [[ $# -eq 0 ]]; then
        showUsage 3
        exit 1
    fi
    while [ ! $# -eq 0 ]; do
        case "${1}" in
        --nodeid | -n)
            shiftOption2 $#
            nodeIsRunning "${2}"
            if [[ $? -eq 1 ]]; then
                printLog "warn" "Node-$2 is Running"
                return
            fi
            printLog "info" "Start node: ${2}"
            nid="${2}"
            shift 2
            ;;
        --bootnodes | -b)
            shiftOption2 $#
            bns="${2}"
            shift 2
            ;;
        --logsize | -s)
            shiftOption2 $#
            log_size="${2}"
            shift 2
            ;;
        --logdir | -d)
            shiftOption2 $#
            log_dir="${2}"
            shift 2
            ;;
        --extraoptions | -e)
            shiftOption2 $#
            extra_options="${2}"
            shift 2
            ;;
        --txcount | -c)
            shiftOption2 $#
            tx_count="${2}"
            shift 2
            ;;
        --tx_global_slots | -tgs)
            shiftOption2 $#
            tx_global_slots="${2}"
            shift 2
            ;;
        --lightmode | -l)
            shiftOption2 $#
            lightmode="${2}"
            shift 2
            ;;
        --dbtype)
            shiftOption2 $#
            dbtype="${2}"
            shift 2
            ;;
        --all | -a)
            printLog "info" "Start all nodes"
            all="true"
            shift 1
            ;;
        *)
            showUsage 5
            exit 1
            ;;
        esac
    done

    if [[ "${all}" == true ]]; then
        checkAllNodeStatus
        for d in ${DISENABLE}
        do
            printLog "info" "Start all disable nodes"
            saveConf "${d}" "bootnodes" "${bns}"
            saveConf "${d}" "log_size" "${log_size}"
            saveConf "${d}" "log_dir" "${log_dir}"
            saveConf "${d}" "extra_options" "${extra_options}"
            saveConf "${d}" "tx_count" "${tx_count}"
            saveConf "${d}" "tx_global_slots" "${tx_global_slots}"
            saveConf "${d}" "lightmode" "${lightmode}"
            saveConf "${d}" "dbtype" "${dbtype}"
            "${SCRIPT_PATH}"/start-node.sh -n "${d}"
        done
    else
        saveConf "${nid}" "bootnodes" "${bns}"
        saveConf "${nid}" "log_size" "${log_size}"
        saveConf "${nid}" "log_dir" "${log_dir}"
        saveConf "${nid}" "extra_options" "${extra_options}"
        saveConf "${nid}" "tx_count" "${tx_count}"
        saveConf "${nid}" "tx_global_slots" "${tx_global_slots}"
        saveConf "${nid}" "lightmode" "${lightmode}"
        saveConf "${nid}" "dbtype" "${dbtype}"
        "${SCRIPT_PATH}"/start-node.sh -n "${nid}"
    fi
}

function deploySys() {
    helpOption "$@"
    if [[ $? -eq 1 ]]; then
        showUsage 6
        exit 1
    fi
    "${SCRIPT_PATH}"/deploy-system-contract.sh "$@"
}

function addNode() {
    helpOption "$@"
    if [[ $? -eq 1 ]]; then
        showUsage 7
        exit 1
    fi
    "${SCRIPT_PATH}"/add-node.sh "$@"
}

function updateSys() {
    helpOption "$@"
    if [[ $? -eq 1 ]]; then
        showUsage 8
        exit 1
    fi
    "${SCRIPT_PATH}"/update_to_consensus_node.sh "$@"
}

function dcgen() {
    helpOption "$@"
    if [[ $? -eq 1 ]]; then
        showUsage 9
        exit 1
    fi
    "${SCRIPT_PATH}"/local/generate-deployconf.sh "$@"
}

function keygen() {
    helpOption "$@"
    if [[ $? -eq 1 ]]; then
        showUsage 10
        exit 1
    fi
    "${SCRIPT_PATH}"/local/generate-key.sh "$@"
}

function gengen() {
    helpOption "$@"
    if [[ $? -eq 1 ]]; then
        showUsage 11
        exit 1
    fi
    "${SCRIPT_PATH}"/local/generate-genesis.sh "$@"
}

function addadmin() {
    helpOption "$@"
    if [[ $? -eq 1 ]]; then
        showUsage 12
        exit 1
    fi
    "${SCRIPT_PATH}"/local/add-admin-role.sh "$@"
}

function delete() {
    helpOption "$@"
    if [[ $? -eq 1 ]]; then
        showUsage 13
        exit 1
    fi
    "${SCRIPT_PATH}"/local/delete-node.sh "$@"
}

function stop() {
    stopAll() {
        nodes="$(ls ${DATA_PATH})"
        for n in ${nodes}; do
            if [ -d "${DATA_PATH}/${n}" ] && [[ "${n}" == node-* ]]; then
                nodeid="$(echo ${n#*-})"
                stop --nodeid "${nodeid}"
            fi
        done
    }

    case "$1" in
    --nodeid | -n)
        shiftOption2 $#
        pid="$(ps -ef | grep "venachain --identity venachain --datadir ${DATA_PATH}/node-${2} " | grep -v grep | awk '{print $2}')"
        if [[ "${pid}" != "" ]]; then
            printLog "info" "Stop node-${2}"
            kill -9 "${pid}"
            sleep 1
        fi
        ;;
    --all | -a)
        printLog "info" "Stop all nodes"
        stopAll
        ;;
    *)
        showUsage 14
        exit 1
        ;;
    esac
}

function restart() {
    case "$1" in
    --nodeid | -n)
        shiftOption2 $#
        checkNodeStatusFullName "node-${2}"
        if [[ "${ENABLE}" == "" ]]; then
            printLog "warn" "The Node Is Not Running"
            echo printLog "info" "To start the node-${2}"
            start -n "${2}"
        else
            stop -n "${2}"
            start -n "${2}"
        fi
        ;;
    --all | -a)
        printLog "info" "Restart all running nodes"
        checkAllNodeStatus
        for e in ${ENABLE}; do
            stop -n "${e}"
            start -n "${e}"
        done
        ;;
    *)
        showUsage 15
        exit 1
        ;;
    esac
}

function clearConf() {
    if [ -f "${CONF_PATH}/${1}" ]; then
        mkdir -p "${CONF_PATH}/bak"
        mv "${CONF_PATH}/${1}" "${CONF_PATH}/bak/${1}.bak.$(date '+%Y%m%d%H%M%S')"
    fi
}

function clear() {
    case "$1" in
    --nodeid | -n)
        shiftOption2 $#
        stop -n "${2}"
        printLog "info" "Clear node-${2}"
        NODE_DIR="${DATA_PATH}/node-${2}"
        printLog "info" "Clean NODE_DIR: ${NODE_DIR}"
        rm -rf "${NODE_DIR}"
        ;;
    --all | -a)
        stop -a
        printLog "info" "Clear all nodes data"
        rm -rf "${DATA_PATH}"/*
        clearConf "genesis.json"
        clearConf "firstnode.info"
        clearConf "keyfile.json"
        clearConf "keyfile.phrase"
        clearConf "keyfile.account"
        ;;
    *)
        showUsage 16
        exit 1
        ;;
    esac
}

################################################# Account Operation #################################################
function createAcc() {
    helpOption "$@"
    if [[ $? -eq 1 ]]; then
        showUsage 17
        exit 1
    fi
    "${SCRIPT_PATH}/local"/create-account.sh "$@"
}

function unlockAccount() {
    printLog "info" "Unlock node account, nodeid: ${NODE_ID}"
    printLog "question" "Please input your account password"
    read pw

    # get node owner address
    keystore="${DATA_PATH}/node-${NODE_ID}/keystore/"
    echo "${keystore}"
    keys="$(ls $keystore)"
    echo "${keys}"
    for k in ${keys}
    do
        keyinfo="$(cat ${keystore}/${k} | sed s/[[:space:]]//g)"
        keyinfo="${keyinfo,,}sss"
        account="${keyinfo:12:40}"
        echo "account: 0x${account}"
        break
    done

    printLog "info" "Unlock command: "
    echo "curl -X POST  -H 'Content-Type: application/json' --data '{\"jsonrpc\":\"2.0\",\"method\": \"personal_unlockAccount\", \"params\": [\"0x${account}\",\"${pw}\",0],\"id\":1}' http://${1}:${2}"
    curl -X POST -H "Content-Type: application/json" --data "{\"jsonrpc\":\"2.0\",\"method\": \"personal_unlockAccount\", \"params\": [\"0x${account}\",\"${pw}\",0],\"id\":1}" "http://${1}:${2}"
}

function unlock() {
    case "$1" in
    --nodeid | -n)
        shiftOption2 $#
        NODE_ID="${2}"
        NODE_DIR="${DATA_PATH}/node-${NODE_ID}"
        echo "${NODE_DIR}"
        ip="$(cat ${NODE_DIR}/deploy_node-${NODE_ID}.conf | grep "ip_addr=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
        rpc_port="$(cat ${NODE_DIR}/deploy_node-${NODE_ID}.conf | grep "rpc_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
        shift 2
        ;;
    *)
        showUsage 18
        exit 1
        ;;
    esac
    unlockAccount "${ip}" "${rpc_port}"
}

################################################# Console #################################################
function console() {
    case "${1}" in
    --opennodeid | -n)
        shiftOption2 $#
        rpc_port="$(cat ${DATA_PATH}/node-${2}/deploy_node-$2.conf | grep "rpc_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
        ip="$(cat ${DATA_PATH}/node-${2}/deploy_node-$2.conf | grep "ip_addr=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
        "${BIN_PATH}/venachain" attach "http://${ip}:${rpc_port}"
        cd "${CURRENT_PATH}"
        ;;
    --closenodeid | -c)
        shiftOption2 $#
        rpc_port="$(cat ${DATA_PATH}/node-${2}/deploy_node-$2.conf | grep "rpc_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
        ip="$(cat ${DATA_PATH}/node-${2}/deploy_node-$2.conf | grep "ip_addr=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
        pid="$(ps -ef | grep "venachain attach http://${ip}:${rpc_port}" | grep -v grep | awk '{print $2}')"
        if [[ "${pid}" != "" ]]; then
            kill -9 "${pid}"
        fi
        cd "${CURRENT_PATH}"
        ;;
    --closeall)
        while [[ 0 -lt 1 ]];
        do
            pid="$(ps -ef | grep "venachain attach" | head -n 1 | grep -v grep | awk '{print $2}')"
            if [[ "${pid}" == "" ]]; then
                break
            fi
            kill -9 "${pid}"
        done
        cd "${CURRENT_PATH}"
        ;;
    *)
        showUsage 19
        exit 1
        ;;
    esac
}

################################################# Node Info #################################################
function echoInformation() {
    echo "                  ${1}: ${2}"
}

function showNodeInformation() {
    NODE_DIR="${DATA_PATH}/node-${1}"
    echo "          node info:"

    keystore="${NODE_DIR}/keystore"
    if [ -d "${keystore}" ]; then
        keys="$(ls ${keystore})"
        for k in ${keys}; do
            keyinfo="$(cat ${keystore}/${k} | sed s/[[:space:]]//g)"
            account="${keyinfo:12:40}"
            echo "                  account: ${account}"
            break
        done
    fi
    echoInformation "node.address" "$(cat ${NODE_DIR}/node.address)"
    echoInformation "node.pubkey" "$(cat ${NODE_DIR}/node.pubkey)"
    echoInformation "node.ip_addr" "$(cat ${NODE_DIR}/deploy_node-${1}.conf | grep "ip_addr=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    echoInformation "node.rpc_port" "$(cat ${NODE_DIR}/deploy_node-${1}.conf | grep "rpc_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    echoInformation "node.p2p_port" "$(cat ${NODE_DIR}/deploy_node-${1}.conf | grep "p2p_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    echoInformation "node.ws_port" "$(cat ${NODE_DIR}/deploy_node-${1}.conf | grep "ws_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
}

function showNodeInformation1() {
    case "${1}" in
    enable)
        echo "running -> node_id:  ${2}"
        ;;
    disable)
        echo "disable -> node_id:  ${2}"
        ;;
    esac
    showNodeInformation "${2}"
}

function show() {
    case "$1" in
    --nodeid | -n)
        shiftOption2 $#
        checkNodeStatusFullName "node-${2}"
        ;;
    --all | -a)
        checkAllNodeStatus
        ;;
    *) 
        showUsage 21
        exit 1
        ;;
    esac
    for e in ${ENABLE}; do
        showNodeInformation1 enable "${e}"
    done

    for d in ${DISENABLE}; do
        showNodeInformation1 disable "${d}"
    done
}

function getAllNodes() {
    if [ ! -f "${CONF_PATH}/firstnode.info" ]; then
        printLog "error" "FILE ${CONF_PATH}/firstnode.info NOT FOUND"
        return
    fi
    firstnode_id="$(cat ${CONF_PATH}/firstnode.info | grep "node_id=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    firstnode_ip_addr="$(cat ${CONF_PATH}/firstnode.info | grep "ip_addr=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    firstnode_rpc_port="$(cat ${CONF_PATH}/firstnode.info | grep "rpc_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    firstnode_deploy_path="$(cat ${CONF_PATH}/firstnode.info | grep "deploy_path=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    "${BIN_PATH}"/vcl node query --all --url "${firstnode_ip_addr}:${firstnode_rpc_port}"
}

function showVersion() {
    echo "${VERSION}"
}

################################################# Quick Local Deploy Functions #################################################
function one() {
    while [ ! $# -eq 0 ]; do
        case "${1}" in
        *)
            showUsage 1
            exit 1
            ;;
        esac
    done

    setupGenesis --ip "127.0.0.1" --p2p_port "16791" --auto "true"
    if [[ $? -eq 1 ]]; then
        exit 1
    fi
    init --ip "127.0.0.1" --rpc_port "6791" --p2p_port "16791" --ws_port "26791" --auto "true"
    if [[ $? -eq 1 ]]; then
        exit 1
    fi
    start -n "0"
    if [[ $? -eq 1 ]]; then
        exit 1
    fi
    deploySys --auto "true"
    if [[ $? -eq 1 ]]; then
        exit 1
    fi
}

function four() {
    helpOption "$@"
    if [[ $? -eq 1 ]]; then
        showUsage 2
        exit 1
    fi
    "${SCRIPT_PATH}"/build-4-nodes-chain.sh "$@"
}

################################################# Remote Deploy Functions #################################################
function remote() {
    case "${1}" in
    deploy)
        shift
        "${SCRIPT_PATH}/remote"/deploy.sh "$@"
        ;;
    prepare)
        shift
        "${SCRIPT_PATH}/remote"/prepare.sh "$@"
        ;;
    transfer)
        shift
        "${SCRIPT_PATH}/remote"/transfer.sh "$@"
        ;;
    init)
        shift
        "${SCRIPT_PATH}/remote"/init.sh "$@"
        ;;
    start)
        shift
        "${SCRIPT_PATH}/remote"/start.sh "$@"
        ;;
    clear)
        shift
        "${SCRIPT_PATH}/remote"/clear.sh "$@"
        ;;
    *)
        showUsage 20
        exit 1
        ;;
    esac
}

################################################# Commands #################################################
case "${1}" in
one)
    shift
    one "$@"
    ;;
four)
    shift
    four "$@"
    ;;
setupgen)
    shift
    setupGenesis "$@"
    ;;
init)
    shift
    init "$@"
    ;;
start)
    shift
    start "$@"
    ;;
deploysys)
    shift
    deploySys "$@"
    ;;
addnode)
    shift
    addNode "$@"
    ;;
updatesys)
    shift
    updateSys "$@"
    ;;
dcgen)
    shift
    dcgen "$@"
    ;;
keygen)
    shift
    keygen "$@"
    ;;
gengen)
    shift
    gengen "$@"
    ;;
addadmin)
    shift
    addadmin "$@"
    ;;
delete)
    shift
    delete "$@"
    ;;
stop)
    shift
    stop "$@"
    ;;
restart)
    shift
    restart "$@"
    ;;
clear)
    shift
    clear "$@"
    ;;
createacc)
    shift
    createAcc "$@"
    ;;
unlock)
    shift
    unlock "$@"
    ;;
console)
    shift
    console "$@"
    ;;
remote)
    shift
    remote "$@"
    ;;
status)
    shift
    show "$@"
    ;;
get)
    shift
    getAllNodes
    ;;
version | -v)
    shift
    showVersion
    ;;
*)
    shift
    showUsage
    exit 1
    ;;
esac