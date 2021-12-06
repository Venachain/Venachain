#!/bin/bash

SCRIPT_NAME="$(basename ${0})"
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

VERSION=$(${BIN_PATH}/platone --version)

function usage() {
    cat <<EOF
#h0    DESCRIPTION
#h0        The deployment script for platone
#h0
#h0    Usage:
#h0        ${SCRIPT_NAME} <command> [command options] [arguments...]
#h0
#h0    COMMANDS
#c0        one                              start a node completely
#c0                                         default account password: 0
#c0        four                             start four node completely
#c0                                         default account password: 0
#c0        init                             initialize node. please setup genesis first
#c1        init OPTIONS
#c1            --nodeid, -n                 set node id (default=0)
#c1            --ip                         set node ip (default=127.0.0.1)
#c1            --rpc_port                   set node rpc port (default=6791)
#c1            --p2p_port                   set node p2p port (default=16791)
#c1            --ws_port                    set node ws port (default=26791)
#c1            --auto                       auto=true: will no prompt to create
#c1                                         the node key and init (default: false)
#c1            --help, -h                   show help
#c0        start                            try to start the specified node
#c4        start OPTIONS
#c4            --nodeid, -n                 start the specified node
#c4            --bootnodes, -b              Connect to the specified bootnodes node
#c4                                         The default is the first in the suggestObserverNodes
#c4                                         in genesis.json
#c4            --logsize, -s                Log block size (default: 67108864)
#c4            --logdir, -d                 log dir (default: ../data/node_dir/logs/)
#c4                                         The path connector '/' needs to be escaped
#c4                                         when set: eg ".\/logs"
#c4            --extraoptions, -e           extra platone command options when platone starts
#c4                                         (default: --debug)
#c4            --txcount, -c                max tx count in a block (default:1000)
#c4            --lightmode                  light node mode
#c4                                         option: lightnode, lightserver or ''
#c4            --all, -a                    start all node
#c4            --help, -h                   show help
#c0        stop                             try to stop the specified node
#c5        stop OPTIONS
#c5            --nodeid, -n                 stop the specified node
#c5            --all, -a                    stop all node
#c5            --help, -h                   show help
#c0        restart                          try to restart the specified node
#c6        restart OPTIONS
#c6            --nodeid, -n                 restart the specified node
#c6            --all, -a                    restart all node
#c6            --help, -h                   show help
#c0        console                          start an interactive JavaScript environment
#c7        console OPTIONS
#c7            --opennodeid , -n            open the specified node console
#c7                                         set the node id here
#c7            --closenodeid, -c            stop the specified node console
#c7                                         set the node id here
#c7            --closeall                   stop all node console
#c7            --help, -h                   show help
#c0        deploysys                        deploy the system contract
#c8        deploysys OPTIONS
#c8            --nodeid, -n                 the specified node id (default: 0)
#c8            --auto                       auto=true: will use the default node password: 0
#c8                                         to create the account and also
#c8                                         to unlock the account (default: false)
#c8            --help, -h                   show help
#c0        updatesys                        normal node update to consensus node
#c9        updatesys OPTIONS
#c9            --nodeid, -n                 the specified node id
#c9            --content, -c                update content (default: 'consensus')
#c9            --help, -h                   show help
#c0        addnode                          add normal node to system contract
#c10       addnode OPTIONS
#c10           --nodeid, -n                 the specified node id. must be specified
#c10           --desc                       the specified node desc
#c10           --p2p_port                   the specified node p2p_port
#c10                                        If the node specified by nodeid is local,
#c10                                        then you do not need to specify this option.
#c10           --rpc_port                   the specified node rpc_port
#c10                                        If the node specified by nodeid is local,
#c10                                        then you do not need to specify this option.
#c10           --ip                         the specified node ip
#c10                                        If the node specified by nodeid is local,
#c10                                        then you do not need to specify this option.
#c10           --pubkey                     the specified node pubkey
#c10                                        If the node specified by nodeid is local,
#c10                                        then you do not need to specify this option.
#c10           --account                    the specified node account
#c10                                        If the node specified by nodeid is local,
#c10                                        then you do not need to specify this option.
#c10           --help, -h                   show help
#c0        clear                            clear all nodes data
#c11       clear OPTIONS
#c11           --nodeid, -n                 clear specified node data
#c11           --all, -a                    clear all nodes data
#c11           --help, -h                   show help
#c0        unlock                           unlock node account
#c12       unlock OPTIONS
#c12           --nodeid, -n                 unlock account on specified node
#c12           --account, -a                account to unlock
#c12           --phrase, -p                 phrase of the account
#c12           --help, -h                   show help
#c0        get                              display all nodes in the system contract
#c0        setupgen                         create the genesis.json and compile sys contract
#c13       setupgen OPTIONS
#c13           --nodeid, -n                 the first node id (default: 0)
#c13           --ip                         the first node ip (default: 127.0.0.1)
#c13           --p2p_port                   the first node p2p_port (default: 16791)
#c13           --observerNodes, -o          set the genesis suggestObserverNodes
#c13                                        (default is the first node enode code)
#c13           --validatorNodes, -v         set the genesis validatorNodes
#c13                                        (default is the first node enode code)
#c13           --interpreter, -i            Select virtual machine interpreter in wasm, evm, all (default: wasm)
#c13           --auto                       auto=true: Will auto create new node keys and will
#c13                                        not compile system contracts again (default=false)
#c13           --help, -h                   show help
#c0        status                           show all node status
#c14       status OPTIONS                   show all node status
#c14           --nodeid, -n                 show the specified node status info
#c14           --all, -a                    show all  node status info
#c14           --help, -h                   show help
#c0        createacc                        create account
#c15       createacc OPTIONS
#c15           --nodeid, -n                 create account for specified node
#c15           --help, -h                   show help
#c0        remote                           remote deploy (recommended)
#c16       remote OPTIONS
#c16           deploy                       deploy nodes
#c16           prepare                      generate directory structure and deployment conf file
#c16           transfer                     transfer necessary file to target node
#c16           init                         initialize the target node
#c16           start                        start the target node
#c16           clear                        clear the target node
#c16           --help, -h                   show help
#c0        version                          show platone release version
#c0===============================================================
#c0    INFORMATION
#c0        version         PlatONE Version: ${VERSION}
#c0        author
#c0        copyright       Copyright
#c0        license
#c0
#c0===============================================================
#c0    HISTORY
#c0        2019/06/26  ty : create the deployment script
#c0        2021/08/11  wjw : update old scripts 
#c0                          create the remote deployment script
#c0
#c0===============================================================
EOF
}

################################################# General Functions  #################################################
function showVersion() {
    echo "${VERSION}"
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
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* MISS OPTION VALUE! PLEASE SET THE VALUE **********"
        exit
    fi
}

function helpOption() {
    for op in "$@"; do
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
    node_conf_tmp=${DATA_PATH}/node-${1}/node.conf1
    if [[ $3 == "" ]]; then
        return
    fi
    if ! [[ -f "${node_conf}" ]]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* FILE ${node_conf} NOT FOUND **********"
        return
    fi
    cat $node_conf | sed "s#${2}=.*#${2}=${3}#g" | cat >$node_conf_tmp
    mv $node_conf_tmp $node_conf
}

function checkNodeStatusFullName() {
    if [[ -d ${DATA_PATH}/${1} ]] && [[ $1 == node-* ]]; then
        nodeid=$(echo ${1#*-})
        if [[ $(ps -ef | grep "platone --identity platone --datadir ${DATA_PATH}/node-${nodeid} " | grep -v grep | awk '{print $2}') != "" ]]; then
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
    if [[ $(ps -ef | grep "platone --identity platone --datadir ${DATA_PATH}/node-${1} " | grep -v grep | awk '{print $2}') != "" ]]; then
        return 1
    fi
    return 0
}

################################################# Local Deploy Functions #################################################
function setupGenesis() {
    helpOption "$@"
    if [[ $? -ne 0 ]]; then
        showUsage 13
        return
    fi
    nodeid=""
    ip=""
    p2p_port=""
    validator_nodes=""
    interpreter=""
    auto=""
    while [ ! $# -eq 0 ]; do
        case $1 in
        --nodeid | -n)
            shiftOption2 $#
            nodeid=$2
            shift 2
            ;;
        --ip)
            shiftOption2 $#
            ip=$2
            checkIp "${ip}"
            if [ $? -ne 0 ]; then
                echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* IP ${ip} NOT VALID **********"
                return
            fi
            shift 2
            ;;
        --p2p_port | -p)
            shiftOption2 $#
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
        --auto)
            auto="--auto"
            shift 1
            ;;
        *)
            showUsage 13
            return
            ;;
        esac
    done
    ## init deploy conf
    if [ ! -d ${DATA_PATH}/node-${nodeid} ]; then
        echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')]  : The node directory have not been created, Now to create it"
        mkdir -p ${DATA_PATH}/node-${nodeid}
    fi
    if [ ! -f ${DATA_PATH}/node-${nodeid}/deploy_node-${nodeid}.conf ]; then
        cp ${CONF_PATH}/deploy.conf.template ${DATA_PATH}/node-${nodeid}/deploy_node-${nodeid}.conf
    fi
    saveConf "${nodeid}" "deploy_path" "${WORKSPACE_PATH}"
    saveConf "${nodeid}" "user_name" "$(whoami)"
    saveConf "${nodeid}" "log_dir" "${DATA_PATH}/node-${nodeid}/logs"
    ## replace configuration
    if [[ "${ip}" != "" ]]; then
        saveConf "${nodeid}" "ip_addr" "${ip}"
    fi
    if [[ "${p2p_port}" != "" ]]; then
        saveConf "${nodeid}" "p2p_port" "${p2p_port}"
    fi
    ## generate param
    param=""
    if [[ "${validator_nodes}" != "" ]]; then
        param="${param} -v ${validator_nodes}"
    fi
    if [[ "${interpreter}" != "" ]]; then
        param="${param} -i ${interpreter}"
    fi
    if [[ "${auto}" != "" ]]; then
        param="${param} ${auto}"
    fi
    ## setupgen
    ./local-keygen.sh -n ${nodeid} ${auto}
    ./local-setup-genesis.sh -n ${nodeid} ${param}
}

function init() {
    helpOption "$@"
    if [[ $? -ne 0 ]]; then
        showUsage 1
        return
    fi
    nodeid=""
    ip=""
    rpc_port=""
    ws_port=""
    p2p_port=""
    auto=""
    while [ ! $# -eq 0 ]; do
        case "$1" in
        --nodeid | -n)
            shiftOption2 $#
            nodeid=$2
            shift 2
            ;;
        --ip)
            shiftOption2 $#
            ip=$2
            checkIp "${ip}"
            if [ $? -ne 0 ]; then
                echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* IP ${ip} NOT VALID **********"
                return
            fi
            shift 2
            ;;
        --rpc_port)
            shiftOption2 $#
            rpc_port=$2
            shift 2
            ;;
        --ws_port)
            shiftOption2 $#
            ws_port=$2
            shift 2
            ;;
        --p2p_port)
            shiftOption2 $#
            p2p_port=$2
            shift 2
            ;;
        --auto)
            auto="--auto"
            shift 1
            ;;
        *)
            showUsage 1
            return
            ;;
        esac
    done
    ## init deploy conf
    if [ ! -f ${CONF_PATH}/genesis.json ]; then
        echo '[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* THERE NO GENESIS FILE, PLEASE GENERATE GENESIS FILE FIRST *********'
        return
    fi
    if [ ! -d ${DATA_PATH}/node-${nodeid} ]; then
        echo '[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : The node directory have not been created, Now to create it'
        mkdir -p ${DATA_PATH}/node-${nodeid}
    fi
    if [ ! -f ${DATA_PATH}/node-${nodeid}/deploy_node-${nodeid}.conf ]; then
        cp ${CONF_PATH}/deploy.conf.template ${DATA_PATH}/node-${nodeid}/deploy_node-${nodeid}.conf
    fi
    saveConf "${nodeid}" "deploy_path" "${WORKSPACE_PATH}"
    saveConf "${nodeid}" "user_name" "$(whoami)"
    saveConf "${nodeid}" "log_dir" "${DATA_PATH}/node-${nodeid}/logs"
    ## replace configuration
    if [[ "${ip}" != "" ]]; then
        saveConf "${nodeid}" "ip_addr" "${ip}"
    fi
    if [[ "${rpc_port}" != "" ]]; then
        saveConf "${nodeid}" "rpc_port" "${rpc_port}"
    fi
    if [[ "${ws_port}" != "" ]]; then
        saveConf "${nodeid}" "ws_port" "${ws_port}"
    fi
    if [[ "${p2p_port}" != "" ]]; then
        saveConf "${nodeid}" "p2p_port" "${p2p_port}"
    fi
    ## init
    ./local-keygen.sh -n ${nodeid} ${auto}
    ${BIN_PATH}/platone --datadir ${DATA_PATH}/node-${nodeid} init ${CONF_PATH}/genesis.json
}

function start() {
    nid=""
    bns=""
    logsize=""
    logdir=""
    extraoptions=""
    txcount=""
    tx_global_slots=""
    lightmode=""
    all="false"
    if [[ $# -eq 0 ]]; then
        showUsage 4
        exit
    fi
    while [ ! $# -eq 0 ]; do
        case "$1" in
        --nodeid | -n)
            shiftOption2 $#
            nodeIsRunning $2
            if [[ $? -ne 0 ]]; then
                echo "[WARN] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')]: !!! The Node Is Running ...; node_id: $2 !!!"
                return
            fi
            echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Start node: ${2}"
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
            logsize=$2
            shift 2
            ;;
        --logdir | -d)
            shiftOption2 $#
            logdir=$2
            shift 2
            ;;
        --extraoptions | -e)
            shiftOption2 $#
            extraoptions=$2
            shift 2
            ;;
        --txcount | -c)
            shiftOption2 $#
            txcount=$2
            shift 2
            ;;
        --tx_global_slots | -tgs)
            shiftOption2 $#
            tx_global_slots=$2
            shift 2
            ;;
        --lightmode)
            shiftOption2 $#
            lightmode=$2
            shift 2
            ;;
        --all | -a)
            echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Start all nodes"
            all=true
            shift 1
            ;;
        *)
            showUsage 4
            exit
            ;;
        esac
    done

    if [[ $all == true ]]; then
        checkAllNodeStatus
        for d in ${DISENABLE}; do
            echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Start all disable nodes"
            saveConf $d bootnodes "${bns}"
            saveConf $d logsize "${logsize}"
            saveConf $d logdir "${logdir}"
            saveConf $d extraoptions "${extraoptions}"
            saveConf $d txcount "${txcount}"
            saveConf $d tx_global_slots = "${tx_global_slots}"
            ./local-run-node.sh -n $d --lightmode "${lightmode}"
        done
        exit
    fi
    saveConf $nid bootnodes "${bns}"
    saveConf $nid logsize "${logsize}"
    saveConf $nid logdir "${logdir}"
    saveConf $nid extraoptions "${extraoptions}"
    saveConf $nid txcount "${txcount}"
    saveConf $nid tx_global_slots "${tx_global_slots}"
    ./local-run-node.sh -n $nid --lightmode "${lightmode}"
}

function deploySys() {
    helpOption "$@"
    if [[ $? -ne 0 ]]; then
        showUsage 8
        return
    fi
    ./local-create-account.sh "$@" "--admin"
    ./local-add-admin-role.sh "$@"
    ./local-add-node.sh "$@"
    ./local-update-node.sh "$@"
}

function addNode() {
    helpOption "$@"
    if [[ $? -ne 0 ]]; then
        showUsage 10
        return
    fi

    nodeid=""
    ip=""
    p2p_port=""
    rpc_port=""
    ip=""
    pubkey=""
    desc=""
    account=""
    while [ ! $# -eq 0 ]; do
        case "$1" in
        --nodeid | -n)
            nodeid=$2
            ;;
        --p2p_port)
            p2p_port=$2
            ;;
        --rpc_port)
            rpc_port=$2
            ;;
        --ip)
            ip=$2
            checkIp "${ip}"
            if [ $? -ne 0 ]; then
                echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* IP ${ip} NOT VALID **********"
                return
            fi
            ;;
        --pubkey)
            pubkey=$2
            ;;
        --desc)
            desc=$2
            ;;
        --account)
            account=$2
            ;;
        *)
            showUsage 10
            return
            ;;
        esac
        shiftOption2 $#
        shift 2
    done
    ## replace configuration
    if [[ "${p2p_port}" != "" ]]; then
        saveConf "${nodeid}" "p2p_port" "${p2p_port}"
    fi
    if [[ "${rpc_port}" != "" ]]; then
        saveConf "${nodeid}" "rpc_port" "${rpc_port}"
    fi
    if [[ "${ip}" != "" ]]; then
        saveConf "${nodeid}" "ip_addr" "${ip}"
    fi
    ## generate param
    param=""
    if [[ "${pubkey}" != "" ]]; then
        param="${param} --pubkey ${pubkey}"
    fi
    if [[ "${desc}" != "" ]]; then
        param = "${param} --desc ${desc}"
    fi
    if [[ "${account}" != "" ]]; then
        param = "${param} --account ${account}"
    fi
    ## addnode
    ./local-add-node.sh -n ${nodeid} ${param}
}

function updateSys() {
    helpOption "$@"
    if [[ $? -ne 0 ]]; then
        ./local-update-node.sh "-h"
        return
    fi
    ./local-update-node.sh "$@"
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
        killall "platone"
    }

    case "$1" in
    --nodeid | -n)
        shiftOption2 $#
        pid=$(ps -ef | grep "platone --identity platone --datadir ${DATA_PATH}/node-${2} " | grep -v grep | awk '{print $2}')
        if [[ $pid != "" ]]; then
            echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Stop node: ${2}"
            kill $pid
            sleep 1
        fi
        ;;
    --all | -a)
        echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Stop all nodes"
        stopAll
        ;;
    *) showUsage 5 ;;
    esac
}

function restart() {
    case "$1" in
    --nodeid | -n)
        shiftOption2 $#
        checkNodeStatusFullName "node-${2}"
        if [[ ${ENABLE} == "" ]]; then
            echo "[WARN] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : !!! The Node Is Not Running !!!"
            echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : To start the node..."
            start -n $2
            exit
        fi
        stop -n $2
        start -n $2
        ;;
    --all | -a)
        echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Restart all running nodes"
        checkAllNodeStatus
        for e in ${ENABLE}; do
            stop -n $e
            start -n $e
        done
        ;;
    *) showUsage 6 ;;
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
        echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Clear node id: ${2}"
        NODE_DIR=${WORKSPACE_PATH}/data/node-${2}
        echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Clean NODE_DIR: ${NODE_DIR}"
        rm -rf ${NODE_DIR}
        ;;
    --all | -a)
        stop -a
        echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Clear all nodes data"
        data=${WORKSPACE_PATH}/data
        rm -rf ${data}/*
        clearConf genesis.json
        clearConf firstnode.info
        clearConf keyfile.json
        clearConf keyfile.phrase
        clearConf keyfile.account
        ;;
    *) showUsage 11 ;;
    esac
}

################################################# Quick Local Deploy Functions #################################################
function one() {
    stop --all
    clear --all
    setupGenesis -n 0 --auto
    init -n 0 --auto
    start -n 0
    deploySys -n 0 --auto
}

function four() {
    ## prepare env
    stop --all
    clear --all
    ## init
    setupGenesis -n 0 --ip 127.0.0.1 --p2p_port 16791 --auto
    init -n 0 --ip 127.0.0.1 --p2p_port 16791 --rpc_port 6791 --ws_port 26791 --auto
    init -n 1 --ip 127.0.0.1 --p2p_port 16792 --rpc_port 6792 --ws_port 26792 --auto
    init -n 2 --ip 127.0.0.1 --p2p_port 16793 --rpc_port 6793 --ws_port 26793 --auto
    init -n 3 --ip 127.0.0.1 --p2p_port 16794 --rpc_port 6794 --ws_port 26794 --auto
    ## start
    start -n 0
    deploySys -n 0 --auto
    start -n 1
    start -n 2
    start -n 3
    ## add node
    addNode -n 1
    addNode -n 2
    addNode -n 3
    ## update sys
    updateSys -n 1
    updateSys -n 2
    updateSys -n 3
}

################################################# Console #################################################
function console() {
    case "$1" in
    --opennodeid | -n)
        shiftOption2 $#
        cd ${DATA_PATH}/node-${2}/
        rpc_port=$(cat deploy_node-$2.conf | grep "rpc_port=" | sed -e 's/rpc_port=\(.*\)/\1/g')
        ip=$(cat deploy_node-$2.conf | grep "ip_addr=" | sed -e 's/ip_addr=\(.*\)/\1/g')
        cd ${BIN_PATH}
        ./platone attach http://${ip}:${rpc_port}
        cd ${CURRENT_PATH}
        ;;
    --closenodeid | -c)
        shiftOption2 $#
        cd ${DATA_PATH}/node-${2}/
        rpc_port=$(cat deploy_node-$2.conf | grep "rpc_port=" | sed -e 's/rpc_port=\(.*\)/\1/g')
        ip=$(cat deploy_node-$2.conf | grep "ip_addr=" | sed -e 's/ip_addr=\(.*\)/\1/g')
        pid=$(ps -ef | grep "platone attach http://${ip}:${rpc_port}" | grep -v grep | awk '{print $2}')
        cd ${CURRENT_PATH}
        ;;
    --closeall)
        killall "platone attach"
        ;;
    *) showUsage 7 ;;
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
    echoInformation "node.ip_addr" "$(cat ${nodeHome}/deploy_node-${1}.conf | grep "ip_addr=" | sed -e 's/ip_addr=\(.*\)/\1/g')"
    echoInformation "node.rpc_port" "$(cat ${nodeHome}/deploy_node-${1}.conf | grep "rpc_port=" | sed -e 's/rpc_port=\(.*\)/\1/g')"
    echoInformation "node.p2p_port" "$(cat ${nodeHome}/deploy_node-${1}.conf | grep "p2p_port=" | sed -e 's/p2p_port=\(.*\)/\1/g')"
    echoInformation "node.ws_port" "$(cat ${nodeHome}/deploy_node-${1}.conf | grep "ws_port=" | sed -e 's/ws_port=\(.*\)/\1/g')"
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
    *) showUsage 14 ;;
    esac
    for e in ${ENABLE}; do
        showNodeInformation1 enable $e
    done

    for d in ${DISENABLE}; do
        showNodeInformation1 disable $d
    done
}

function getAllNodes() {
    if [ ! -f ${CONF_PATH}/firstnode.info ] || [ ! -f ${CONF_PATH}/keyfile.json ] || [ ! -f ${CONF_PATH}/keyfile.phrase ]; then
        echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* MISS CONF FILE **********"
        return
    fi
    firstnode_ip_addr=$(cat ${CONF_PATH}/firstnode.info | grep "ip_addr=" | sed -e 's/ip_addr=\(.*\)/\1/g')
    firstnode_rpc_port=$(cat ${CONF_PATH}/firstnode.info | grep "rpc_port=" | sed -e 's/rpc_port=\(.*\)/\1/g')
    ${BIN_PATH}/platonecli node query --all --keyfile ${CONF_PATH}/keyfile.json --url ${firstnode_ip_addr}:${firstnode_rpc_port} <${CONF_PATH}/keyfile.phrase
}

################################################# Account Operation #################################################
function createAcc() {
    if [[ $? -ne 0 ]]; then
        showUsage 15
        return
    fi
    ./local-create-account.sh "$@"
}

function unlockAccount() {
    IP=${1}
    PORT=${2}
    pw=${3}
    account=${4}

    echo "[INFO] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : Unlock node account, nodeid: ${NODE_ID}"

    if [ -z ${account} ]; then
        # get node owner address
        keystore=${DATA_PATH}/node-${NODE_ID}/keystore/
        echo $keystore
        keys=$(ls $keystore)
        echo "$keys"
        for k in $keys; do
            keyinfo=$(cat $keystore/$k | sed s/[[:space:]]//g)
            account=${keyinfo:12:40}
            account="0x${account}"
            echo "account: ${account}"
            break
        done
    fi

    if [ -z ${pw} ]; then
        read -p "Your account password?: " pw
    fi

    echo "curl -X POST  -H 'Content-Type: application/json' --data '{\"jsonrpc\":\"2.0\",\"method\": \"personal_unlockAccount\", \"params\": [\"${account}\",\"${pw}\",0],\"id\":1}' http://${IP}:${PORT}"
    curl -X POST -H "Content-Type: application/json" --data "{\"jsonrpc\":\"2.0\",\"method\": \"personal_unlockAccount\", \"params\": [\"${account}\",\"${pw}\",0],\"id\":1}" http://${IP}:$PORT}
}

function unlockAcc() {

    ACC=""
    PHRASE=""

    while [ ! $# -eq 0 ]; do
        case "$1" in
        --nodeid | -n)
            nodeHome=${DATA_PATH}/node-${2}
            NODE_ID=$2
            echo $nodeHome
            # ip=$(cat ${nodeHome}/deploy_node-${2}.conf | grep "ip_addr=" | sed -e 's/ip_addr=\(.*\)/\1/g')
            # rpc_port=$(cat ${nodeHome}/deploy_node-${2}.conf | grep "rpc_port=" | sed -e 's/rpc_port=\(.*\)/\1/g')
            if [ ! -f ${CONF_PATH}/firstnode.info ]; then
                echo "[ERROR] [$(echo $0 | sed -e 's/\(.*\)\/\(.*\).sh/\2/g')] : ********* MISS FIRSTNODE INFO FILE **********"
                return
            fi
            ip=$(cat ${CONF_PATH}/firstnode.info | grep "ip_addr=" | sed -e 's/ip_addr=\(.*\)/\1/g')
            rpc_port=$(cat ${CONF_PATH}/firstnode.info | grep "rpc_port=" | sed -e 's/rpc_port=\(.*\)/\1/g')
            echo ${ip} ${rpc_port}
            ;;
        --account | -a)
            echo "account: $2"
            ACC=${2}
            ;;
        --phrase | -p)
            PHRASE=${2}
            ;;
        *)
            showUsage 12
            exit
            ;;
        esac
        shiftOption2 $#
        shift 2
    done
    unlockAccount ${ip} ${rpc_port} ${PHRASE} ${ACC}
}

################################################# Remote Deploy Functions #################################################
function remote() {
    case "$1" in
    deploy)
        shift
        ./deploy.sh "$@"
        ;;
    prepare)
        shift
        ./prepare.sh "$@"
        ;;
    transfer)
        shift
        ./transfer.sh "$@"
        ;;
    init)
        shift
        ./init.sh "$@"
        ;;
    start)
        shift
        ./start.sh "$@"
        ;;
    --clear)
        shift
        ./clear.sh "$@"
        ;;
    *)
        showUsage 16
        return
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
console)
    shift
    console "$@"
    ;;
status)
    shift
    show "$@"
    ;;
get)
    shift
    getAllNodes
    ;;
createacc)
    shift
    createAcc "$@"
    ;;
unlock)
    shift
    unlockAcc "$@"
    ;;
version | -v)
    showVersion
    ;;
remote)
    shift
    remote "$@"
    ;;
*)
    shift
    showUsage
    ;;
esac
