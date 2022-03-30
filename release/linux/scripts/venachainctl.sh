#!/bin/bash

SCRIPT_NAME="$(basename ${0})"
SCRIPT_ALIAS="$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')"
CURRENT_PATH=$(pwd)
WORKSPACE_PATH=$(cd ${CURRENT_PATH}/.. && echo ${PWD})
BIN_PATH=${WORKSPACE_PATH}/bin
CONF_PATH=${WORKSPACE_PATH}/conf
SCRIPT_PATH=${WORKSPACE_PATH}/scripts
DATA_PATH=${WORKSPACE_PATH}/data
ENABLE=""
DISENABLE=""
NODE_ID=0
cd ${CURRENT_PATH}

VERSION="PLEASE BUILD VENACHAIN FIRST"
if [ -f ${BIN_PATH}/venachain ]; then
    VERSION="$(${BIN_PATH}/venachain --version)"
fi


function usage() {
    cat <<EOF
#h0    DESCRIPTION
#h0        The deployment script for venachain
#h0
#h0    USAGE:
#h0        ${SCRIPT_NAME} <command> [command options] [arguments...]
#h0
#h0    COMMANDS
#c0        one                              start a node completely
#c0                                         (default account password: 0)
#c0        four                             start four node completely
#c0                                         (default account password: 0)
#c0        setupgen                         create the genesis.json and compile sys contract
#c1        setupgen OPTIONS
#c1            --nodeid, -n                 the first node id (default: 0)
#c1            --ip                         the first node ip (default: 127.0.0.1)
#c1            --p2p_port                   the first node p2p_port (default: 16791)
#c1            --validatorNodes, -v         set the genesis validatorNodes
#c1                                         (default: the first node enode code)
#c1            --interpreter, -i            select virtual machine interpreter in wasm, evm, all 
#c1                                         (default: all)
#c1            --auto, -a                   will no prompt to create the node key and skip ip check
#c1            --help, -h                   show help
#c0        init                             initialize node. please setup genesis first
#c2        init OPTIONS
#c2            --nodeid, -n                 set node id (default: 0)
#c2            --ip                         set node ip (default: 127.0.0.1)
#c2            --rpc_port, -r               set node rpc port (default: 6791)
#c2            --p2p_port, -p               set node p2p port (default: 16791)
#c2            --ws_port, -w                set node ws port (default: 26791)
#c2            --auto, -a                   will no prompt to create the node key and init
#c2            --help, -h                   show help
#c0        start                            try to start the specified node
#c3        start OPTIONS
#c3            --nodeid, -n                 start the specified node (default: 0)
#c3            --bootnodes, -b              connect to the specified bootnodes node
#c3                                         (default: the first in the suggestObserverNodes in genesis.json)
#c3            --logsize, -s                Log block size (default: 67108864)
#c3            --logdir, -d                 log dir (default: ../data/node_dir/logs/)
#c3            --extraoptions, -e           extra venachain command options when venachain starts
#c3                                         (default: --debug)
#c3            --txcount, -c                max tx count in a block (default: 1000)
#c3            --tx_global_slots, -tgs      max tx count in txpool (default: 4096)
#c3            --lightmode, -l              select lightnode mode in lightnode, lightserver
#c3            --dbtype                     select database type (default: leveldb)
#c3            --all, -a                    start all nodes
#c3            --help, -h                   show help
#c0        deploysys                        deploy the system contract
#c4        deploysys OPTIONS
#c4            --nodeid, -n                 the specified node id (default: 0)
#c4            --auto                       will use the default node password: 0
#c4                                         to create the account and also to unlock the account
#c4            --help, -h                   show help
#c0        addnode                          add normal node to system contract
#c5        addnode OPTIONS
#c5            --nodeid, -n                 the specified node id. must be specified
#c5            --desc, -d                   the specified node desc
#c5            --p2p_port, -p               the specified node p2p_port
#c5                                         If the node specified by nodeid is local,
#c5                                         then you do not need to specify this option.
#c5            --rpc_port, -r               the specified node rpc_port
#c5                                         If the node specified by nodeid is local,
#c5                                         then you do not need to specify this option.
#c5            --ip                         the specified node ip
#c5                                         If the node specified by nodeid is local,
#c5                                         then you do not need to specify this option.
#c5            --pubkey                     the specified node pubkey
#c5                                         If the node specified by nodeid is local,
#c5                                         then you do not need to specify this option.
#c5            --type                       select specified node type in 2 & 3
#c5                                         2 is observer, 3 is lightnode (default: 2)
#c5            --help, -h                   show help
#c0        updatesys                        update node type
#c6        updatesys OPTIONS
#c6            --nodeid, -n                 the specified node id. must be specified
#c6            --content, -c                update content (default: consensus)
#c6            --help, -h                   show help
#c0        stop                             try to stop the specified node
#c7        stop OPTIONS
#c7            --nodeid, -n                 stop the specified node
#c7            --all, -a                    stop all node
#c7            --help, -h                   show help
#c0        restart                          try to restart the specified node
#c8        restart OPTIONS
#c8            --nodeid, -n                 restart the specified node
#c8            --all, -a                    restart all node
#c8            --help, -h                   show help
#c0        clear                            try to stop and clear the node data
#c8        clear OPTIONS
#c9            --nodeid, -n                 clear specified node data
#c9            --all, -a                    clear all nodes data
#c9            --help, -h                   show help
#c0        createacc                        create account
#c10       createacc OPTIONS
#c10           --nodeid, -n                 create account for specified node
#c10           --auto, -a                   create account with default phrase 0
#c10           --create_keyfile, -ck        will create keyfile
#c10           --help, -h                   show help
#c0        unlock                           unlock node account
#c11       unlock OPTIONS
#c11           --nodeid, -n                 unlock account on specified node
#c11           --help, -h                   show help
#c0        console                          start an interactive JavaScript environment
#c12       console OPTIONS
#c12           --opennodeid , -n            open the specified node console
#c12                                        set the node id here
#c12           --closenodeid, -c            stop the specified node console
#c12                                        set the node id here
#c12           --closeall                   stop all node console
#c12           --help, -h                   show help
#c0        remote                           remote deploy
#c13       remote OPTIONS
#c13           deploy                       deploy nodes
#c13           prepare                      generate directory structure and deployment conf file
#c13           transfer                     transfer necessary file to target node
#c13           init                         initialize the target node
#c13           start                        start the target node
#c13           clear                        clear the target node
#c13           --help, -h                   show help
#c0        status                           show all node status
#c14       status OPTIONS                   show all node status
#c14           --nodeid, -n                 show the specified node status info
#c14           --all, -a                    show all nodes status info
#c14           --help, -h                   show help
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
#c0        2019/06/26  ty : create the deployment script
#c0        2021/12/10  wjw : update old scripts 
#c0                          create the remote deployment script
#c0
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

function showUsage() {
    if [[ $1 == "" ]]; then
        usage | grep -e "^#[ch]0 " | sed -e "s/^#[ch][0-9]*//g"
        return
    fi
    usage | grep -e "^#h0 \|^#c${1} " | sed -e "s/^#[ch][0-9]*//g"
}

function shiftOption2() {
    if [[ $1 -lt 2 ]]; then
        printLog "error" "MISS OPTION VALUE! PLEASE SET THE VALUE"
        exit
    fi
}

function helpOption() {
    for op in "$@"
    do
        if [[ $op == "--help" ]] || [[ $op == "-h" ]]; then
            return 1
        fi
    done
}

function checkIp() {
    ip=$1
    check=$(echo $ip | awk -F. '$1<=255&&$2<=255&&$3<=255&&$4<=255{print "yes"}')
    if echo $ip | grep -E "^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$" >/dev/null; then
        if [ ${check:-no} == "yes" ]; then
            return 0
        fi
    fi
    return 1
}

function saveConf() {
    node_conf=${DATA_PATH}/node-${1}/deploy_node-${1}.conf
    node_conf_tmp=${DATA_PATH}/node-${1}/deploy_node-${1}.temp.conf
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

function checkNodeStatusFullName() {
    if [[ -d ${DATA_PATH}/${1} ]] && [[ $1 == node-* ]]; then
        nodeid=$(echo ${1#*-})
        if [[ $(ps -ef | grep "venachain --identity venachain --datadir ${DATA_PATH}/node-${nodeid} " | grep -v grep | awk '{print $2}') != "" ]]; then
            ENABLE=$(echo "${ENABLE} ${nodeid}")
        else
            DISENABLE=$(echo ${DISENABLE} ${nodeid})
        fi
    fi
}

function checkAllNodeStatus() {
    nodes=$(ls ${DATA_PATH})
    for n in ${nodes}; do
        checkNodeStatusFullName $n
    done
}

function nodeIsRunning() {
    if [[ $(ps -ef | grep "venachain --identity venachain --datadir ${DATA_PATH}/node-${1} " | grep -v grep | awk '{print $2}') != "" ]]; then
        return 1
    fi
    return 0
}

################################################# Local Deploy Functions #################################################
function setupGenesis() {
    helpOption "$@"
    if [[ $? -ne 0 ]]; then
        showUsage 1
        return
    fi
    ./setup-genesis.sh "$@"
}

function init() {
    helpOption "$@"
    if [[ $? -ne 0 ]]; then
        showUsage 2
        return
    fi
    ./init-node.sh "$@"
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
        return
    fi
    while [ ! $# -eq 0 ]; do
        case "$1" in
        --nodeid | -n)
            shiftOption2 $#
            nodeIsRunning $2
            if [[ $? -ne 0 ]]; then
                printLog "warn" "The Node-$2 is Running"
                return
            fi
            printLog "info" "Start node: ${2}"
            nid=$2
            shift 2
            ;;
        --bootnodes | -b)
            shiftOption2 $#
            bns=$2
            shift 2
            ;;
        --logsize | -s)
            shiftOption2 $#
            log_size=$2
            shift 2
            ;;
        --logdir | -d)
            shiftOption2 $#
            log_dir=$2
            shift 2
            ;;
        --extraoptions | -e)
            shiftOption2 $#
            extra_options=$2
            shift 2
            ;;
        --txcount | -c)
            shiftOption2 $#
            tx_count=$2
            shift 2
            ;;
        --tx_global_slots | -tgs)
            shiftOption2 $#
            tx_global_slots=$2
            shift 2
            ;;
        --lightmode | -l)
            shiftOption2 $#
            lightmode=$2
            shift 2
            ;;
        --dbtype)
            shiftOption2 $#
            dbtype=$2
            shift 2
            ;;
        --all | -a)
            printLog "info" "Start all nodes"
            all=true
            shift 1
            ;;
        *)
            showUsage 3
            exit
            ;;
        esac
    done

    if [[ $all == true ]]; then
        checkAllNodeStatus
        for d in ${DISENABLE}
        do
            printLog "info" "Start all disable nodes"
            saveConf $d bootnodes "${bns}"
            saveConf $d log_size "${log_size}"
            saveConf $d log_dir "${log_dir}"
            saveConf $d extra_options "${extra_options}"
            saveConf $d tx_count "${tx_count}"
            saveConf $d tx_global_slots "${tx_global_slots}"
            saveConf $d lightmode "${lightmode}"
            saveConf $d dbtype "${dbtype}"
            ./start-node.sh -n $d
        done
        exit
    fi
    saveConf $nid bootnodes "${bns}"
    saveConf $nid log_size "${log_size}"
    saveConf $nid log_dir "${log_dir}"
    saveConf $nid extra_options "${extra_options}"
    saveConf $nid tx_count "${tx_count}"
    saveConf $nid tx_global_slots "${tx_global_slots}"
    saveConf $nid lightmode "${lightmode}"
    saveConf $nid dbtype "${dbtype}"
    ./start-node.sh -n $nid
}

function deploySys() {
    helpOption "$@"
    if [[ $? -ne 0 ]]; then
        showUsage 4
        return
    fi
    ./deploy-system-contract.sh "$@"
}

function addNode() {
    helpOption "$@"
    if [[ $? -ne 0 ]]; then
        showUsage 5
        return
    fi
    ./add-node.sh "$@"
}

function updateSys() {
    helpOption "$@"
    if [[ $? -ne 0 ]]; then
        showUsage 6
        return
    fi
    ./update_to_consensus_node.sh "$@"
}

function stop() {
    stopAll() {
        nodes=$(ls ${DATA_PATH})
        for n in ${nodes}; do
            if [[ -d ${DATA_PATH}/$n ]] && [[ $n == node-* ]]; then
                nodeid=$(echo ${n#*-})
                stop --nodeid $nodeid
            fi
        done
    }

    case "$1" in
    --nodeid | -n)
        shiftOption2 $#
        pid=$(ps -ef | grep "venachain --identity venachain --datadir ${DATA_PATH}/node-${2} " | grep -v grep | awk '{print $2}')
        if [[ $pid != "" ]]; then
            printLog "info" "Stop node-${2}"
            kill -9 $pid
            sleep 1
        fi
        ;;
    --all | -a)
        echo printLog "info" "Stop all nodes"
        stopAll
        ;;
    *) 
        showUsage 7
        ;;
    esac
}

function restart() {
    case "$1" in
    --nodeid | -n)
        shiftOption2 $#
        checkNodeStatusFullName "node-${2}"
        if [[ ${ENABLE} == "" ]]; then
            printLog "warn" "The Node Is Not Running"
            echo printLog "info" "To start the node-$2"
            start -n $2
            exit
        fi
        stop -n $2
        start -n $2
        ;;
    --all | -a)
        printLog "info" "Restart all running nodes"
        checkAllNodeStatus
        for e in ${ENABLE}; do
            stop -n $e
            start -n $e
        done
        ;;
    *) 
        showUsage 8
        ;;
    esac
}

function clearConf() {
    if [[ -f ${CONF_PATH}/${1} ]]; then
        mkdir -p ${CONF_PATH}/bak
        mv ${CONF_PATH}/${1} ${CONF_PATH}/bak/${1}.bak.$(date '+%Y%m%d%H%M%S')
    fi
}

function clear() {
    case "$1" in
    --nodeid | -n)
        shiftOption2 $#
        stop -n $2
        printLog "info" "Clear node-${2}"
        NODE_DIR=${WORKSPACE_PATH}/data/node-${2}
        printLog "info" "Clean NODE_DIR: ${NODE_DIR}"
        rm -rf ${NODE_DIR}
        ;;
    --all | -a)
        stop -a
        printLog "info" "Clear all nodes data"
        data=${WORKSPACE_PATH}/data
        rm -rf ${data}/*
        clearConf genesis.json
        clearConf firstnode.info
        clearConf keyfile.json
        clearConf keyfile.phrase
        clearConf keyfile.account
        ;;
    *) 
        showUsage 9
        ;;
    esac
}

################################################# Account Operation #################################################
function createAcc() {
    helpOption "$@"
    if [[ $? -ne 0 ]]; then
        showUsage 10
        return
    fi
    ./local/create-account.sh "$@"
}

function unlockAccount() {
    printLog "info" "Unlock node account, nodeid: ${NODE_ID}"
    printLog "question" "Please input your account password"
    read pw
        
    # get node owner address
    keystore=${DATA_PATH}/node-${NODE_ID}/keystore/
    echo $keystore
    keys=$(ls $keystore)
    echo "$keys"
    for k in $keys
    do
        keyinfo=$(cat $keystore/$k | sed s/[[:space:]]//g)
        keyinfo=${keyinfo,,}sss
        account=${keyinfo:12:40}
        echo "account: 0x${account}"
        break
    done

    printLog "info" "Unlock command: "
    echo "curl -X POST  -H 'Content-Type: application/json' --data '{\"jsonrpc\":\"2.0\",\"method\": \"personal_unlockAccount\", \"params\": [\"0x${account}\",\"${pw}\",0],\"id\":1}' http://${1}:${2}"
    curl -X POST -H "Content-Type: application/json" --data "{\"jsonrpc\":\"2.0\",\"method\": \"personal_unlockAccount\", \"params\": [\"0x${account}\",\"${pw}\",0],\"id\":1}" http://${1}:${2}
}

function unlock() {
    case "$1" in
    --nodeid | -n)
        shiftOption2 $#
        NODE_ID=$2
        nodeHome=${DATA_PATH}/node-${NODE_ID}
        echo $nodeHome
        ip=$(cat ${nodeHome}/deploy_node-${NODE_ID}.conf | grep "ip_addr=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
        rpc_port=$(cat ${nodeHome}/deploy_node-${NODE_ID}.conf | grep "rpc_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
        shift 2
        ;;
    *)
        showUsage 11
        exit
        ;;
    esac
    unlockAccount ${ip} ${rpc_port}
}

################################################# Console #################################################
function console() {
    case "$1" in
    --opennodeid | -n)
        shiftOption2 $#
        cd ${DATA_PATH}/node-${2}/
        rpc_port=$(cat deploy_node-$2.conf | grep "rpc_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
        ip=$(cat deploy_node-$2.conf | grep "ip_addr=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
        cd ${BIN_PATH}
        ./venachain attach http://${ip}:${rpc_port}
        cd ${CURRENT_PATH}
        ;;
    --closenodeid | -c)
        shiftOption2 $#
        cd ${DATA_PATH}/node-${2}/
        rpc_port=$(cat deploy_node-$2.conf | grep "rpc_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
        ip=$(cat deploy_node-$2.conf | grep "ip_addr=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
        pid=$(ps -ef | grep "venachain attach http://${ip}:${rpc_port}" | grep -v grep | awk '{print $2}')
        if [[ "${pid}" != "" ]]; then
            kill -9 ${pid}
        fi
        cd ${CURRENT_PATH}
        ;;
    --closeall)
        while [[ 0 -lt 1 ]]; 
        do
            pid=$(ps -ef | grep "venachain attach" | head -n 1 | grep -v grep | awk '{print $2}')
            if [[ "${pid}" == "" ]]; then
                break
            fi
            kill -9 ${pid}
        done
        cd ${CURRENT_PATH}
        ;;
    *) 
        showUsage 12 
        ;;
    esac
}

################################################# Node Info #################################################
function echoInformation() {
    echo "                  ${1}: ${2}"
}

function showNodeInformation() {
    nodeHome=${DATA_PATH}/node-${1}
    echo "          node info:"

    keystore=${nodeHome}/keystore
    if [[ -d $keystore ]]; then
        keys=$(ls $keystore)
        for k in $keys; do
            keyinfo=$(cat ${keystore}/${k} | sed s/[[:space:]]//g)
            account=${keyinfo:12:40}
            echo "                  account: ${account}"
            break
        done
    fi
    echoInformation "node.address" "$(cat ${nodeHome}/node.address)"
    echoInformation "node.pubkey" "$(cat ${nodeHome}/node.pubkey)"
    echoInformation "node.ip_addr" "$(cat ${nodeHome}/deploy_node-${1}.conf | grep "ip_addr=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    echoInformation "node.rpc_port" "$(cat ${nodeHome}/deploy_node-${1}.conf | grep "rpc_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    echoInformation "node.p2p_port" "$(cat ${nodeHome}/deploy_node-${1}.conf | grep "p2p_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
    echoInformation "node.ws_port" "$(cat ${nodeHome}/deploy_node-${1}.conf | grep "ws_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')"
}

function showNodeInformation1() {
    case $1 in
    enable)
        echo "running -> node_id:  ${2}"
        ;;
    disable)
        echo "disable -> node_id:  ${2}"
        ;;
    esac
    showNodeInformation $2
}

function show() {
    case "$1" in
    --nodeid | -n)
        shiftOption2 $#
        checkNodeStatusFullName "node-$2"
        ;;
    --all | -a)
        checkAllNodeStatus
        ;;
    *) 
        showUsage 14 
        ;;
    esac
    for e in ${ENABLE}; do
        showNodeInformation1 enable $e
    done

    for d in ${DISENABLE}; do
        showNodeInformation1 disable $d
    done
}

function getAllNodes() {
    if [ ! -f ${CONF_PATH}/firstnode.info ] || [ ! -f ${CONF_PATH}/keyfile.json ]; then
        printLog "error" "MISS CONF FILE"
        return
    fi
    firstnode_ip_addr=$(cat ${CONF_PATH}/firstnode.info | grep "ip_addr=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    firstnode_rpc_port=$(cat ${CONF_PATH}/firstnode.info | grep "rpc_port=" | sed -e 's/\(.*\)=\(.*\)/\2/g')
    ${BIN_PATH}/vcl node query --all --keyfile ${CONF_PATH}/keyfile.json --url ${firstnode_ip_addr}:${firstnode_rpc_port}
}

function showVersion() {
    echo "${VERSION}"
}

################################################# Quick Local Deploy Functions #################################################
function one() {
    ./setup-genesis.sh -n 0 --auto "true"
    ./init-node.sh -n 0 --auto "true"
    ./start-node.sh
    ./deploy-system-contract.sh --auto "true"
}

function four() {
    ./build-4-nodes-chain.sh "$@"
}

################################################# Remote Deploy Functions #################################################
function remote() {
    case "$1" in
    deploy)
        shift
        ./remote/deploy.sh "$@"
        ;;
    prepare)
        shift
        ./remote/prepare.sh "$@"
        ;;
    transfer)
        shift
        ./remote/transfer.sh "$@"
        ;;
    init)
        shift
        ./remote/init.sh "$@"
        ;;
    start)
        shift
        ./remote/start.sh "$@"
        ;;
    clear)
        shift
        ./remote/clear.sh "$@"
        ;;
    *)
        showUsage 13
        ;;
    esac
}

################################################# Commands #################################################
case $1 in
one)
    shift
    one
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
    ;;
esac
