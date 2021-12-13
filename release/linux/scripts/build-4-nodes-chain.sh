#!/bin/bash

###########################################################################################################
################################################# VRIABLES #################################################
###########################################################################################################

## path
SCRIPT_NAME="$(basename ${0})"
PROJECT_PATH=$(
    cd $(dirname $0)
    cd ../
    pwd
)

#############################################################################################################
################################################# FUNCTIONS #################################################
#############################################################################################################

################################################# Main #################################################
function main() {
    ./setup-genesis.sh --ip 127.0.0.1 --p2p_port 16791 --auto true
    ./init-node.sh --ip 127.0.0.1 --rpc_port 6791 --p2p_port 16791 --ws_port 26791 --auto true
    ./start-node.sh -n 0
    ./deploy-system-contract.sh --auto true

    ./init-node.sh -n 1 --ip 127.0.0.1 --rpc_port 6792 --p2p_port 16792 --ws_port 26792 --auto true
    ./start-node.sh -n 1
    ./add-node.sh -n 1
    ./update_to_consensus_node.sh -n 1

    ./init-node.sh -n 2 --ip 127.0.0.1 --rpc_port 6793 --p2p_port 16793 --ws_port 26793 --auto true
    ./start-node.sh -n 2
    ./add-node.sh -n 2
    ./update_to_consensus_node.sh -n 2

    ./init-node.sh -n 3 --ip 127.0.0.1 --rpc_port 6794 --p2p_port 16794 --ws_port 26794 --auto true
    ./start-node.sh -n 3
    ./add-node.sh -n 3
    ./update_to_consensus_node.sh -n 3
}

###########################################################################################################
#################################################  EXECUTE #################################################
###########################################################################################################
main
