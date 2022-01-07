#!/bin/bash


Version=${1}
[[ ${Version} == "" ]] && echo "ERROR: Please set release version; example command:
     \"${0} v0.0.0.0.0\"" && exit

cd ..
Venachain_Project_name="Venachain"
Java_Project_name="java-sdk"
BCWasm_project_name="BCWasm"

Venachain_Linux_Dir="${Venachain_Project_name}/release/linux"
Venachain_CMD_SystemContract="${Venachain_Project_name}/cmd/SysContracts"

Venachain_linux_name="Venachain_linux_${Version}"
BCWasm_linux_name="BCWasm_linux_release.${Version}"
Java_sdk_linux_name="java_sdk_linux_${Version}"
End_with=".tar.gz"

release_md_name="release.md"

function create_release_note() {
cat <<EOF
链部署指南

[Venachain快速搭链教程](https://venachain-docs.readthedocs.io/zh/latest/documents/quick/env.html)

Asset

->上传${Venachain_linux_name}

WASM合约开发库

[Venachain合约指导文档](https://venachain-docs.readthedocs.io/zh/latest/documents/develop/smart_contract.html)

Asset

->上传${BCWasm_linux_name}


SDK工具

[SDK使用说明](https://git-c.i.wxblockchain.com/PlatONE/doc/Venachain_WIKI/blob/v0.9.0/zh-cn/SDK/%5BChinese-Simplified%5D-SDK%E4%BD%BF%E7%94%A8%E8%AF%B4%E6%98%8E.md)

Asset

->上传${Java_sdk_linux_name}

Release Change Log

[change_log文档](https://git-c.i.wxblockchain.com/PlatONE/src/node/venachain/blob/develop/CHANGELOG.md)
EOF
}

function env() {
    if [[ -d ${Venachain_Project_name} ]]; then
        echo "${Venachain_Project_name} already exists."
    else
        git clone --recursive https://github.com/Venachain/Venachain.git
    fi

    if [[ -d ${Java_Project_name} ]]; then
        echo "${Java_Project_name} already exists"
    else
        git clone --recursive https://git-c.i.wxblockchain.com/PlatONE/src/node/java-sdk.git
    fi
    rm -rf ${Java_Project_name}/.git
}

function compile() {
    cd ${Venachain_Project_name} && make all && cd ..
}

function create_Venachain_linux() {
    [[ -d ${Venachain_linux_name} ]] && rm -rf ${Venachain_linux_name}
    mkdir ${Venachain_linux_name}
    cp -rf ${Venachain_Linux_Dir}/* ${Venachain_linux_name}/
    tar -zcvf ${Venachain_linux_name}${End_with} ${Venachain_linux_name}
}

function create_bcwasm_linux() {
    [[ -d ${BCWasm_project_name} ]] && rm -rf ${BCWasm_project_name}
    mkdir ${BCWasm_project_name}
    cp -rf ${Venachain_CMD_SystemContract}/* ${BCWasm_project_name}/
    cp ${Venachain_Project_name}/release/linux/bin/ctool ${BCWasm_project_name}/external/bin/
    rm -rf ${BCWasm_project_name}/systemContract
    rm -rf ${BCWasm_project_name}/build
    tar -zcvf ${BCWasm_linux_name}${End_with} ${BCWasm_project_name}
}

function create_sdk_linux() {
    tar -zcvf ${Java_sdk_linux_name}${End_with} ${Java_Project_name}
}

function tag() {
    cd ${Venachain_Project_name}
    git tag -a ${Version} -m "Release"
    git push --tags
    cd ..
}

function clean() {
    rm -rf ${Venachain_linux_name}
    #rm -rf ${Java_Project_name}
    rm -rf ${BCWasm_project_name}
}

function main() {
    echo "#################################################################################"
    echo "note: Please change the version number in Venachain before executing this script"
    echo "#################################################################################"
    sleep 3
    echo "#################################################################################"
    echo "note: If it is github, please set the change log differently"
    echo "#################################################################################"
    sleep 3

    #env
    compile

    create_Venachain_linux
    create_bcwasm_linux
    #create_sdk_linux

    tag

    clean
    echo "#################################################################################"
    echo "note: The release pkg massage format:"
    echo "#################################################################################"
    create_release_note
}

main
