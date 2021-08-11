###   Quick Start



#### 1. 节点操作

首先，进入`scripts`目录

1.1 启动单节点链：

   ```shell
   ./platonectl.sh one
   ```

1.2 在一台机器上快速启动四节点区块链:

   ```shell
   ./platonectl.sh four
   ```

1.3 查看链的运行状态：

```shell
./platonectl.sh status
```

1.4 停止某个节点，如停止节点0：

```shell
./platonectl.sh stop -n 0
```



#### 2.  RPC 默认端口

快速启动单节点的链：RPC端口为：6791

快速启动四节点的链，RPC端口分别为：6791-6794



#### 3. 默认账户

* keystore位置：

```
./data/node-[x]/keystore
```

* 账号地址：keystore文件中第一个address字段

* 默认账号解锁密码: **0**



#### 4. 日志

节点日志默认位置如下，主要打印节点运行信息：

```
./data/node-[x]/logs
```

wasm日志默认位置如下，主要打印合约调用中的输出：

```
./data/node-[x]/logs/wasm.log
```


# 一、快速部署流程介绍

## 1 准备工作

### 1.1 PlatONE源码下载及编译

```shell
# 获取PlatONE源码
git clone --recursive git@git-c.i.wxblockchain.com:PlatONE/src/node/PlatONE-Go.git

# 编译PlatONE
cd PlatONE-Go
make all
```

release/linux/bin下会生成可执行文件

### 1.2 部署目录搭建

在部署机执行部署任务的目录下新建文件夹，假定执行部署任务的目录为${platone_deploy}，新建完后的目录结构如下

```
${platone_deploy}
│   ├── deployment_conf
│   └── release
│       └── linux
│           ├── bin
│           ├── conf
│           └── scripts
```

```shell
1. 将release/linux/bin下的ethkey、repstr、platone、platonecli这4个可执行文件放入${platone_deploy}/release/linux/bin中

2. 将release/linux/conf下的 genesis.json.istanbul.template放入${platone_deploy}/release/linux/conf中

3. 将release/linux/scripts下的 所有文件放入${platone_deploy}/release/linux/scripts中
```

```
${platone_deploy}
├── deployment_conf
└── release
    └── linux
        ├── bin
        │   ├── ethkey
        │   ├── platone
        │   ├── platonecli
        │   └── repstr
        ├── conf
        │   └── genesis.json.istanbul.template
        └── scripts
            ├── clear.sh
            ├── deploy.sh
            ├── init.sh
            ├── local-add-node.sh
            ├── local-create-account.sh
            ├── local-deploy-system-contract.sh
            ├── local-keygen.sh
            ├── local-setup-genesis.sh
            ├── local-start-node.sh
            ├── local-update-to-consensus-node.sh
            ├── prepare.sh
            ├── start.sh
            └── transfer.sh
```

## 2 一键快速部署

### 2.1 本地单节点部署

```shell
cd ${platone_deploy}/release/linux/scripts
./deploy.sh -p ${project_name} -m one
```

### 2.2 本地四节点部署

```shell
cd ${platone_deploy}/release/linux/scripts
./deploy.sh -p ${project_name} -m four
```

### 2.3 指定地址部署

```shell
cd ${platone_deploy}/release/linux/scripts
# -m conf可省略
./deploy.sh -p ${project_name} -m conf -a wxuser@10.230.48.11,wxuser@10.230.48.11,wxuser@10.230.48.12,10.230.48.12 
```

### 2.4 自定义配置部署

首先根据要部署的目标节点配置信息，生成配置信息文件，配置信息文件模板如下：

```shell
## PlateONE Node Remote Deploy Configuration File ##

## NODE
deploy_path=
user_name=
ip_addr=
p2p_port=6791

## RPC
rpc_addr=0.0.0.0
rpc_port=
rpc_api=db,eth,net,web3,admin,personal,txpool,istanbul

## WEBSOCKET
ws_addr=0.0.0.0
ws_port=26791

## WASM_LOG
log_dir=
log_size=67108864

## NODE START
gcmode=archive
```

```shell
如果节点名为${node-name}，
1. 将其命名为deploy_node-${node_name}，
2. 放至${platone_deploy}/deployment_conf/${project_name}目录下
```

```shell
cd ${platone_deploy}/release/linux/scripts
# -m conf可省略
./deploy.sh -p ${project_name} -m conf	
```

### 2.5 节点信息

#### 2.5.1 存储位置

```shell
节点默认部署地址${deploy_path}：${HOME}/PlatONE/${project_name}
节点数据${node_dir}：${deploy_path}/data/node-0/
节点运行日志：${node_dir}/logs/platone_log/
节点公钥：${node_dir}/node.pubkey

部署地址、本机用户名、IP、p2p端口、RPC信息、websocket信息、log地址、gcmode信息在${node_dir}/deploy_conf/deploy_node-${node_name}.conf中
```

#### 2.5.2 目录结构

```
${deploy_path}
├── bin
│   ├── ethkey
│   ├── platone
│   ├── platonecli
│   └── repstr
├── conf
│   ├── firstnode.info
│   ├── genesis.json
│   ├── genesis.json.istanbul.template
│   ├── keyfile.account
│   ├── keyfile.json
│   └── keyfile.phrase
├── data
│   └── node-0
│       ├── deploy_conf
│       │   └── deploy_node-0.conf
│       ├── keystore
│       ├── logs
│       │   ├── platone_error.log
│       │   ├── platone_log
│       │   └── wasm_log
│       ├── node-0.ipc
│       ├── node.address
│       ├── node.prikey
│       ├── node.pubkey
│       └── platone
│           ├── chaindata
│           ├── extdb
│           ├── lightchaindata
│           ├── LOCK
│           └── transactions.rlp
└── scripts
    ├── clear.sh
    ├── deploy.sh
    ├── init.sh
    ├── local-add-node.sh
    ├── local-create-account.sh
    ├── local-deploy-system-contract.sh
    ├── local-keygen.sh
    ├── local-setup-genesis.sh
    ├── local-start-node.sh
    ├── local-update-to-consensus-node.sh
    ├── prepare.sh
    ├── start.sh
    └── transfer.sh
```

## 3 一键快速清理

所有操作必须需要保证${platone_deploy}/document_conf/${project_name}下有配置文件

### 3.1 停止节点进程

```shell
cd ${platone_deploy}/release/linux/scripts
# 停止所有节点，-n all 可省略
./clear.sh -p ${project_name} -n all -m stop
# 停止指定节点
./clear.sh -p ${project_name} -n 1,2 -m stop
```

### 3.2 清除节点文件

```shell
cd ${platone_deploy}/release/linux/scripts
# 清除所有节点文件，-n all 可省略
./clear.sh -p ${project_name} -n all -m clean
# 清除指定节点文件
./clear.sh -p ${project_name} -n 1,2 -m clean
```

```shell
1. 如果脚本识别到目标主机在${deploy_path}/data下已经没有任何node文件夹了，会自动删除项目目录${deploy_path}
会在${deploy_path}/../bak/${project_name}.bak.${timestamp}下会备份deploy_node-${node-name}.conf文件与conf目录
2. 如果脚本通过日志识别到当前项目所有节点已经被清除，就会提示是否删除部署机中该项目的全局文件与日志，如果选择否，则之后重新永远有的配置文件搭建区块链，则需要重新执行一边清除脚本或手动删除部署机中${platone_deploy}/deployment/${project_name}下的logs和global目录
```

### 3.3 将节点从区块链删除

需要保证firstnode处于可用状态，否则无法删除节点

```shell
cd ${platone_deploy}/release/linux/scripts
# 清除所有节点文件，-n all 可省略
./clear.sh -p ${project_name} -n all -m delete
# 清除指定节点文件
./clear.sh -p ${project_name} -n 1,2 -m delete
```

### 3.4 彻底移除节点

将节点从区块链删除，停止进程并备份、删除文件

```shell
cd ${platone_deploy}/release/linux/scripts
# -m deep 可省略
# 清除所有节点文件，-n all 可省略
./clear.sh -p ${project_name} -n all -m deep
# 清除指定节点文件
./clear.sh -p ${project_name} -n 1,2 -m deep
```

## 4 加入新节点

### 4.1 生成节点配置文件

有两种新增配置文件的方式，支持一次新增多个节点

#### 4.1.1 快速生成配置文件

```shell
cd ${platone_deploy}/release/linux/scripts
./prepare.sh -p ${project_name} -a wxuser@10.230.48.12,10.230.48.13
```

#### 4.2.2 自定义生成配置文件

```
1. 根据模板设置配置文件，并放入${platone_deploy}/document_conf/${project_name}下
```

### 4.2 部署新节点

```shell
cd ${platone_deploy}/release/linux/scripts
# 假设新节点名为${node_name}，-m conf 可省略，在跳出提示询问是否cover原来的目录时，输入n或no
./deploy.sh -p ${project_name} -n ${node_name} -m conf
```

## 5 部署中断后解决方式参考

### 5.1 锁定问题发生的大概位置

所有脚本运行时的输出信息格式如下

```shell
## type分为INFO、WARN、ERROR
 # INFO代表执行成功
 # WARN代表有需要警惕的情况可能，执行情况与原本操作初衷不同，但是不影响整体脚本运行
 # ERROR致命错误导致脚本运行停止
 
## shell_name是脚本名称

[type] [shell_name] : result
```

因此，找到第一个ERROR类型信息，通过shell_name知晓出错发生的是哪个脚本，再通过具体执行结果语句result，来定位错误具体位置

### 5.2 从头重新开始部署

如果原本执行的是自动生成配置文件的部署方式，那么直接重新执行原先命令即可，在跳出提示询问是否cover原来的目录时，输入y或yes

### 5.3 根据日志继续部署

在项目目录下已经有配置文件的情况下，排除完问题后，执行

```shell
# -n all，-m conf 可省略
./deploy.sh -p ${project_name} -n all -m conf
```

脚本会从上一次执行时最后记录的checkpoint后继续执行

### 5.4 日志概述

#### 5.4.1 配置文件日志

```shell
${platone_deploy}/deployment_conf/logs/prepare_log.txt
```

日志记录所有项目配置文件中的节点ip地址、端口号等信息。新建配置文件后会在日志中进行记录，清除节点文件后，会将相应节点的信息从日志中删除

#### 5.4.2 部署日志

```shell
${platone_deploy}/deployment_conf/${project_name}/logs/deploy_log.txt
```

日志记录了部署任务中已经完成的checkpoint。如果停止节点、清除节点文件，都会删除日志中相应的内容。

# 二、脚本使用介绍

## deploy.sh 一键部署脚本

### 功能

- 根据目标主机用户名与地址，自动生成配置文件并完成远程部署与区块链搭建
- 根据配置文件对节点进行部署，完成区块链搭建
- 根据模式选择1节点或4节点部署，自动生成配置文件并根据生成的配置文件进行部署，快速完成本地区块链搭建
- 即使主机上已有别的区块链了依旧支持部署新的区块链，不用停止已运行的其他platone进程
- 借助传输脚本、初始化脚本以及启动脚本的日志机制，根据checkpoint执行操作
- 借助传输脚本、初始化脚本以及启动脚本对本地部署的优化，可以迅速在本地完成部署

### 命令

```shell
"
USAGE: deploy.sh  [options] [value]

        OPTIONS:

           --project, -p              the specified project name. must be specified

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
```

- --project, -p:
  - 项目名称，属于必填项
- --node, -n:
  - 节点名称，默认node=all
  - 如果设置值为all，那么脚本会根据项目路径下所有配置文件依次进行部署
  - 如果设置多个指定节点，那么需要用","作为不同的节点名称的分隔符
- --mode, -m
  - 模式名称，默认mode=conf
  - conf：根据项目目录下配置文件进行部署
  - one：本地部署单节点，自动创建配置文件并部署
  - four：本地部署四节点，自动创建配置文件并部署
- --address, -a
  - 地址，只有在conf模式下有效
  - 格式为${USER_NAME}@${IP_ADDR}
  - 多个地址用','符号隔开，即使是同样的地址，如果要新建多个配置文件，那么需要写相应的个数
  - 脚本会根据地址自动生成配置文件后，进行部署

### 使用演示

#### 1 通过主机用户名与ip地址一键部署

模拟情况：1个节点为部署机本机（wxuser@10.230.48.11），另3个节点为远程目标机，且已有当前项目名称的目录

##### 命令输入

```shell
./deploy.sh -p test -a wxuser@10.230.48.11,wxuser@10.230.48.12,wxuser@10.230.48.13,wxuser@10.230.48.14
```

##### 打印输出

```
[INFO] [deploy] : Project's conf path: /home/wxuser/platone_deploy/deployment_conf/test
/home/wxuser/platone_deploy/deployment_conf/test has already been existed, do you want to cover it?
y

###########################################
####       prepare default files       ####
###########################################
[INFO] [prepare] : Backup /home/wxuser/platone_deploy/deployment_conf/test to /home/wxuser/platone_deploy/deployment_conf/bak/test.bak.20210803094342 completed

#### Start to clear Node-0 ####
[INFO] [clear] : Delete node-0 end
[INFO] [clear] : Get PID of wxuser@10.230.48.11:6791 completed
[INFO] [clear] : Kill PID 93008 of wxuser@10.230.48.11:6791 completed
[INFO] [clear] : Stop node-0 end
[INFO] [clear] : Backup wxuser@10.230.48.11:/home/wxuser/PlatONE/test/data/node-0/deploy_conf/deploy_node-0.conf to wxuser@10.230.48.11:/home/wxuser/PlatONE/test/../bak/test/deploy_node-0.conf.bak.20210803091627955022M42 completed
[INFO] [clear] : Remove wxuser@10.230.48.11:/home/wxuser/PlatONE/test/data/node-0 completed
[INFO] [clear] : Remove wxuser@10.230.48.11:/home/wxuser/PlatONE/test/scripts completed
[INFO] [clear] : Remove wxuser@10.230.48.11:/home/wxuser/PlatONE/test/data completed
[INFO] [clear] : Remove wxuser@10.230.48.11:/home/wxuser/PlatONE/test/bin completed
[INFO] [clear] : Backup wxuser@10.230.48.11:/home/wxuser/PlatONE/test/conf completed
[INFO] [clear] : Clean node-0 end

#### Start to clear Node-1 ####
[INFO] [clear] : Delete node-1 end
[INFO] [clear] : Check ip 10.230.48.12 connection completed
[INFO] [clear] : Check ssh wxuser@10.230.48.12 access completed
[INFO] [clear] : Get PID of wxuser@10.230.48.12:6793 completed
[INFO] [clear] : Kill PID 24440 of wxuser@10.230.48.12:6793 completed
[INFO] [clear] : Stop node-1 end
[INFO] [clear] : Check ip 10.230.48.12 connection completed
[INFO] [clear] : Check ssh wxuser@10.230.48.12 access completed
[INFO] [clear] : Backup wxuser@10.230.48.12:/home/wxuser/PlatONE/test/data/node-1/deploy_conf/deploy_node-1.conf to wxuser@10.230.48.12:/home/wxuser/PlatONE/test/../bak/test/deploy_node-1.conf.bak.20210803091627955028M48 completed
[INFO] [clear] : Remove wxuser@10.230.48.12:/home/wxuser/PlatONE/test/data/node-1 completed
[INFO] [clear] : Remove wxuser@10.230.48.12:/home/wxuser/PlatONE/test/scripts completed
[INFO] [clear] : Remove wxuser@10.230.48.12:/home/wxuser/PlatONE/test/data completed
[INFO] [clear] : Remove wxuser@10.230.48.12:/home/wxuser/PlatONE/test/bin completed
[INFO] [clear] : Backup wxuser@10.230.48.12:/home/wxuser/PlatONE/test/conf completed
[INFO] [clear] : Clean node-1 end

#### Start to clear Node-2 ####
[INFO] [clear] : Delete node-2 end
[INFO] [clear] : Check ip 10.230.48.13 connection completed
[INFO] [clear] : Check ssh wxuser@10.230.48.13 access completed
[INFO] [clear] : Get PID of wxuser@10.230.48.13:6793 completed
[INFO] [clear] : Kill PID 98110 of wxuser@10.230.48.13:6793 completed
[INFO] [clear] : Stop node-2 end
[INFO] [clear] : Check ip 10.230.48.13 connection completed
[INFO] [clear] : Check ssh wxuser@10.230.48.13 access completed
[INFO] [clear] : Backup wxuser@10.230.48.13:/home/wxuser/PlatONE/test/data/node-2/deploy_conf/deploy_node-2.conf to wxuser@10.230.48.13:/home/wxuser/PlatONE/test/../bak/test/deploy_node-2.conf.bak.20210803091627955041M01 completed
[INFO] [clear] : Remove wxuser@10.230.48.13:/home/wxuser/PlatONE/test/data/node-2 completed
[INFO] [clear] : Remove wxuser@10.230.48.13:/home/wxuser/PlatONE/test/scripts completed
[INFO] [clear] : Remove wxuser@10.230.48.13:/home/wxuser/PlatONE/test/data completed
[INFO] [clear] : Remove wxuser@10.230.48.13:/home/wxuser/PlatONE/test/bin completed
[INFO] [clear] : Backup wxuser@10.230.48.13:/home/wxuser/PlatONE/test/conf completed
[INFO] [clear] : Clean node-2 end

#### Start to clear Node-3 ####
[INFO] [clear] : Delete node-3 end
[INFO] [clear] : Check ip 10.230.48.14 connection completed
[INFO] [clear] : Check ssh wxuser@10.230.48.14 access completed
[INFO] [clear] : Get PID of wxuser@10.230.48.14:6793 completed
[INFO] [clear] : Kill PID 21817 of wxuser@10.230.48.14:6793 completed
[INFO] [clear] : Stop node-3 end
[INFO] [clear] : Check ip 10.230.48.14 connection completed
[INFO] [clear] : Check ssh wxuser@10.230.48.14 access completed
[INFO] [clear] : Backup wxuser@10.230.48.14:/home/wxuser/PlatONE/test/data/node-3/deploy_conf/deploy_node-3.conf to wxuser@10.230.48.14:/home/wxuser/PlatONE/test/../bak/test/deploy_node-3.conf.bak.20210803091627955053M13 completed
[INFO] [clear] : Remove wxuser@10.230.48.14:/home/wxuser/PlatONE/test/data/node-3 completed
[INFO] [clear] : Remove wxuser@10.230.48.14:/home/wxuser/PlatONE/test/scripts completed
[INFO] [clear] : Remove wxuser@10.230.48.14:/home/wxuser/PlatONE/test/data completed
[INFO] [clear] : Remove wxuser@10.230.48.14:/home/wxuser/PlatONE/test/bin completed
[INFO] [clear] : Backup wxuser@10.230.48.14:/home/wxuser/PlatONE/test/conf completed
[INFO] [clear] : Clean node-3 end

[INFO] [clear] : Clear action end
[INFO] [prepare] : Set up directory structure completed

################ Generate Configuration File For wxuser@10.230.48.11 Start ################
[INFO] [prepare] : Generate /home/wxuser/platone_deploy/deployment_conf/test/deploy_node-0.conf for wxuser@10.230.48.11 completed

################ Generate Configuration File For wxuser@10.230.48.12 Start ################
[INFO] [prepare] : Check ip 10.230.48.12 connection completed
[INFO] [prepare] : Check ssh wxuser@10.230.48.12 access completed
[INFO] [prepare] : Generate /home/wxuser/platone_deploy/deployment_conf/test/deploy_node-1.conf for wxuser@10.230.48.12 completed

################ Generate Configuration File For wxuser@10.230.48.13 Start ################
[INFO] [prepare] : Check ip 10.230.48.13 connection completed
[INFO] [prepare] : Check ssh wxuser@10.230.48.13 access completed
[INFO] [prepare] : Generate /home/wxuser/platone_deploy/deployment_conf/test/deploy_node-2.conf for wxuser@10.230.48.13 completed

################ Generate Configuration File For wxuser@10.230.48.14 Start ################
[INFO] [prepare] : Check ip 10.230.48.14 connection completed
[INFO] [prepare] : Check ssh wxuser@10.230.48.14 access completed
[INFO] [prepare] : Generate /home/wxuser/platone_deploy/deployment_conf/test/deploy_node-3.conf for wxuser@10.230.48.14 completed

###########################################
####       transfer file to nodes      ####
###########################################

################ Transfer file to Node-0 ################
[INFO] [transfer] : Create /home/wxuser/PlatONE/test/conf completed
[INFO] [transfer] : Create /home/wxuser/PlatONE/test/scripts completed
[INFO] [transfer] : Create /home/wxuser/PlatONE/test/bin completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/conf/genesis.json.istanbul.template completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/clear.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/deploy.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/init.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-add-node.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-create-account.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-deploy-system-contract.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-keygen.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-setup-genesis.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-start-node.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-update-to-consensus-node.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/prepare.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/start.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/transfer.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/bin/ethkey completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/bin/platone completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/bin/platonecli completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/bin/repstr completed
[INFO] [transfer] : Create /home/wxuser/PlatONE/test/data/node-0/deploy_conf completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/data/node-0/deploy_conf/deploy_node-0.conf completed
[INFO] [transfer] : Transfer files to Node-0 completed

......

################ Transfer file to Node-3 ################
[INFO] [transfer] : Create /home/wxuser/PlatONE/test/conf completed
[INFO] [transfer] : Create /home/wxuser/PlatONE/test/scripts completed
[INFO] [transfer] : Create /home/wxuser/PlatONE/test/bin completed
genesis.json.istanbul.template                                                                                                                                                                                                              100%  356     0.4KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/conf/genesis.json.istanbul.template completed
clear.sh                                                                                                                                                                                                                                    100%   23KB  22.7KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/clear.sh completed
deploy.sh                                                                                                                                                                                                                                   100% 6711     6.6KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/deploy.sh completed
init.sh                                                                                                                                                                                                                                     100%   15KB  15.0KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/init.sh completed
local-add-node.sh                                                                                                                                                                                                                           100% 5356     5.2KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-add-node.sh completed
local-create-account.sh                                                                                                                                                                                                                     100% 5905     5.8KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-create-account.sh completed
local-deploy-system-contract.sh                                                                                                                                                                                                             100% 6519     6.4KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-deploy-system-contract.sh completed
local-keygen.sh                                                                                                                                                                                                                             100% 5984     5.8KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-keygen.sh completed
local-setup-genesis.sh                                                                                                                                                                                                                      100% 7030     6.9KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-setup-genesis.sh completed
local-start-node.sh                                                                                                                                                                                                                         100% 7328     7.2KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-start-node.sh completed
local-update-to-consensus-node.sh                                                                                                                                                                                                           100% 5050     4.9KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-update-to-consensus-node.sh completed
prepare.sh                                                                                                                                                                                                                                  100%   15KB  14.8KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/prepare.sh completed
start.sh                                                                                                                                                                                                                                    100%   16KB  16.2KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/start.sh completed
transfer.sh                                                                                                                                                                                                                                 100%   16KB  15.8KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/transfer.sh completed
ethkey                                                                                                                                                                                                                                      100%   26MB  26.1MB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/bin/ethkey completed
platone                                                                                                                                                                                                                                     100%   36MB  35.8MB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/bin/platone completed
platonecli                                                                                                                                                                                                                                  100%   31MB  30.9MB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/bin/platonecli completed
repstr                                                                                                                                                                                                                                      100%   19KB  19.1KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/bin/repstr completed
[INFO] [transfer] : Create /home/wxuser/PlatONE/test/data/node-3/deploy_conf completed
deploy_node-3.conf                                                                                                                                                                                                                          100%  404     0.4KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/data/node-3/deploy_conf/deploy_node-3.conf completed
[INFO] [transfer] : Transfer files to Node-3 completed

[INFO] [transfer] : Transfer completed

###########################################
####             init nodes            ####
###########################################

################ Init Node-0 ################
[INFO] [keygen] : ## Node-0 Keygen Start ##
[INFO] [keygen] : Files: /home/wxuser/PlatONE/test/data/node-0/node.address, /home/wxuser/PlatONE/test/data/node-0/node.prikey, /home/wxuser/PlatONE/test/data/node-0/node.pubkey
        Node-0's address: 1b827a7e68179d099Db16782ceD04Aaa1F0914b9
        Node-0's private key: 425feff81167b55c1b4a93530f2f5ac2198a95abd47bf46cd40251ffa8ce187a
        Node-0's public key: 1a8565b3978090f73de58e2cb70d95affba150dd941da8271a57866c73a27f62d891451b0bc858755513f5d4e98923f85ebdcc2908ba31dacf776a4db471a623
[INFO] [keygen] : Node-0 keygen succeeded
[INFO] [init] : Generate key for node-0 completed
[INFO] [setup-genesis] : ## Setup Genesis Start ##
[INFO] [setup-genesis] : File: /home/wxuser/PlatONE/test/conf/genesis.json
[INFO] [setup-genesis] : Genesis:
{
    "config": {
    "chainId": 300,
    "interpreter": "all",
    "istanbul": {
        "timeout": 10000,
        "period": 1,
        "policy": 0,
        "firstValidatorNode": "enode://1a8565b3978090f73de58e2cb70d95affba150dd941da8271a57866c73a27f62d891451b0bc858755513f5d4e98923f85ebdcc2908ba31dacf776a4db471a623@10.230.48.11:16791"
    }
  },
  "timestamp": "1627955113",
  "extraData": "0x00",
  "alloc": {
    "0x1b827a7e68179d099Db16782ceD04Aaa1F0914b9": {
      "balance": "100000000000000000000"
    }
  }
}
[INFO] [setup-genesis] : Setup genesis succeeded
[INFO] [init] : Setup genesis file completed
[INFO] [init] : Get genesis file completed
[INFO] [init] : Setup firstnode info completed
[INFO] [init] : Sync firstnode info to node-0 completed
******************************************************************************************************************************************************************************
INFO [08-03|09:45:13.370] Maximum peer count                       ETH=50 LES=0 total=50 RoutineID=1
INFO [08-03|09:45:13.370] Allocated cache and file handles         database=/home/wxuser/PlatONE/test/data/node-0/platone/chaindata cache=16 handles=16 RoutineID=1
INFO [08-03|09:45:13.387] Persisted trie from memory database      nodes=13 size=2.35kB time=113.061µs gcnodes=0 gcsize=0.00B gctime=0s livenodes=1 livesize=0.00B RoutineID=1
INFO [08-03|09:45:13.387] Successfully wrote genesis state         database=chaindata                                               hash=dad2d7…b975ee RoutineID=1
INFO [08-03|09:45:13.388] Allocated cache and file handles         database=/home/wxuser/PlatONE/test/data/node-0/platone/lightchaindata cache=16 handles=16 RoutineID=1
INFO [08-03|09:45:13.408] Persisted trie from memory database      nodes=13 size=2.35kB time=95.991µs  gcnodes=0 gcsize=0.00B gctime=0s livenodes=1 livesize=0.00B RoutineID=1
INFO [08-03|09:45:13.408] Successfully wrote genesis state         database=lightchaindata                                               hash=dad2d7…b975ee RoutineID=1
******************************************************************************************************************************************************************************
[INFO] [init] :Init genesis on node-0 completed
[INFO] [init]: Init node Node-0 completed

......

################ Init Node-3 ################
[INFO] [keygen] : ## Node-3 Keygen Start ##
[INFO] [keygen] : Files: /home/wxuser/PlatONE/test/data/node-3/node.address, /home/wxuser/PlatONE/test/data/node-3/node.prikey, /home/wxuser/PlatONE/test/data/node-3/node.pubkey
        Node-3's address: C840EBD765675Fbf0738bDEBbcF3154A795F3246
        Node-3's private key: 4b68f183faee391678f4d62736f6feaf6d72d3270cfed4a3bc7bd8d31220a14e
        Node-3's public key: 39b0dc39ca810e0e1dd7d548a26c97a63e15ed4d85e1b7bbd20a46d01014f2c58996d48e7e0c4f3cabf11ec4554e6633c743e22182089ff4fe27622782aa8046
[INFO] [keygen] : Node-3 keygen succeeded
[INFO] [init] : Generate key for node-3 completed
genesis.json                                                                                                                                                                                                                                100%  514     0.5KB/s   00:00
[INFO] [init] : Send genesis file to node-3 completed
firstnode.info                                                                                                                                                                                                                              100%   62     0.1KB/s   00:00
[INFO] [init] : Sync firstnode info to node-3 completed
******************************************************************************************************************************************************************************
INFO [08-03|09:45:21.129] Maximum peer count                       ETH=50 LES=0 total=50 RoutineID=1
INFO [08-03|09:45:21.130] Allocated cache and file handles         database=/home/wxuser/PlatONE/test/data/node-3/platone/chaindata cache=16 handles=16 RoutineID=1
INFO [08-03|09:45:21.145] Persisted trie from memory database      nodes=13 size=2.35kB time=117.503µs gcnodes=0 gcsize=0.00B gctime=0s livenodes=1 livesize=0.00B RoutineID=1
INFO [08-03|09:45:21.146] Successfully wrote genesis state         database=chaindata                                               hash=dad2d7…b975ee RoutineID=1
INFO [08-03|09:45:21.146] Allocated cache and file handles         database=/home/wxuser/PlatONE/test/data/node-3/platone/lightchaindata cache=16 handles=16 RoutineID=1
INFO [08-03|09:45:21.168] Persisted trie from memory database      nodes=13 size=2.35kB time=154.024µs gcnodes=0 gcsize=0.00B gctime=0s livenodes=1 livesize=0.00B RoutineID=1
INFO [08-03|09:45:21.169] Successfully wrote genesis state         database=lightchaindata                                               hash=dad2d7…b975ee RoutineID=1
******************************************************************************************************************************************************************************
[INFO] [init] :Init genesis on node-3 completed
[INFO] [init]: Init node Node-3 completed

[INFO] [init] : Init completed

###########################################
####            start  nodes           ####
###########################################

################ Start first node Node-0 ################
[INFO] [start-node] : ## Run node-0 ##
[INFO] [start-node] : Node's url: 10.230.48.11:6791
[INFO] [start-node] : Run node-0 succeeded
[INFO] [start] : Run node Node-0 completed
[INFO] [create-account] : ## Create account Start ##
[WARN] [create-account] : !!! An account will be created. The default password is 0 !!!
[INFO] [create-account] : Account:
        New account: 0x378a967fa993401f0afe69c318401ab03eb6f79a
        Passphrase: 0
[INFO] [create-account] : Create account succeeded
[INFO] [start] : Create account completed
[INFO] [start] : Get keyfile completed
[INFO] [deploy-system-contract] : ## Deploy System Contract Start ##
[INFO] [deploy-system-contract] : Set Node-0 as super admin completed
[INFO] [deploy-system-contract] : Set Node-0 as chain admin completed
[INFO] [deploy-system-contract] : Deploy system contract completed
[INFO] [start] : Deploy system contract completed
[INFO] [add-node] : ## Add Node-0 Start ##
[INFO] [add-node] : Add Node-0 succeeded
[INFO] [start] : Add node node-0 completed
[INFO] [update-to-consensus-node] : ## Update Node-0 To Consensus Node Start ##
[INFO] [update-to-consensus-node] : Update Node-0 to consensus node succeeded
[INFO] [start] : Update node node-0 to consensus node completed
[INFO] [start] : Start firstnode Node-0 completed

......

################ Start Node-3 ################
keyfile.json                                                                                                                                                                                                                                100%  491     0.5KB/s   00:00
keyfile.phrase                                                                                                                                                                                                                              100%    2     0.0KB/s   00:00
[INFO] [start] : Send keyfile to Node-3 completed
[INFO] [start-node] : ## Run node-3 ##
[INFO] [start-node] : Node's url: 10.230.48.14:6793
[INFO] [start-node] : Run node-3 succeeded
[INFO] [start] : Run node Node-3 completed
[INFO] [add-node] : ## Add Node-3 Start ##
[INFO] [add-node] : Add Node-3 succeeded
[INFO] [start] : Add node node-3 completed
[INFO] [update-to-consensus-node] : ## Update Node-3 To Consensus Node Start ##
[INFO] [update-to-consensus-node] : Update Node-3 to consensus node succeeded
[INFO] [start] : Update node node-3 to consensus node completed
[INFO] [start] : Start node Node-3 completed

[INFO] [start] : Start completed
```

#### 2 已有配置文件的情况下，一键部署

模拟情况：1个节点为部署机本机（wxuser@10.230.48.11），另3个节点为远程目标机，且尚未有当前项目名称的目录

##### 命令输入

```shell
./deploy.sh -p test
```

##### 打印输出

```
[INFO] [deploy] : Project's conf path: /home/wxuser/platone_deploy/deployment_conf/test

###########################################
####       transfer file to nodes      ####
###########################################

################ Transfer file to Node-0 ################
[INFO] [transfer] : Create /home/wxuser/PlatONE/test/conf completed
[INFO] [transfer] : Create /home/wxuser/PlatONE/test/scripts completed
[INFO] [transfer] : Create /home/wxuser/PlatONE/test/bin completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/conf/genesis.json.istanbul.template completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/clear.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/deploy.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/init.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-add-node.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-create-account.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-deploy-system-contract.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-keygen.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-setup-genesis.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-start-node.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-update-to-consensus-node.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/prepare.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/start.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/transfer.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/bin/ethkey completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/bin/platone completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/bin/platonecli completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/bin/repstr completed
[INFO] [transfer] : Create /home/wxuser/PlatONE/test/data/node-0/deploy_conf completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/data/node-0/deploy_conf/deploy_node-0.conf completed
[INFO] [transfer] : Transfer files to Node-0 completed

......

################ Transfer file to Node-3 ################
[INFO] [transfer] : Create /home/wxuser/PlatONE/test/conf completed
[INFO] [transfer] : Create /home/wxuser/PlatONE/test/scripts completed
[INFO] [transfer] : Create /home/wxuser/PlatONE/test/bin completed
genesis.json.istanbul.template                                                                                                                                                                                                      100%  356     0.4KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/conf/genesis.json.istanbul.template completed
clear.sh                                                                                                                                                                                                                            100%   23KB  22.7KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/clear.sh completed
deploy.sh                                                                                                                                                                                                                           100% 6711     6.6KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/deploy.sh completed
init.sh                                                                                                                                                                                                                             100%   15KB  15.0KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/init.sh completed
local-add-node.sh                                                                                                                                                                                                                   100% 5356     5.2KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-add-node.sh completed
local-create-account.sh                                                                                                                                                                                                             100% 5905     5.8KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-create-account.sh completed
local-deploy-system-contract.sh                                                                                                                                                                                                     100% 6519     6.4KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-deploy-system-contract.sh completed
local-keygen.sh                                                                                                                                                                                                                     100% 5984     5.8KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-keygen.sh completed
local-setup-genesis.sh                                                                                                                                                                                                              100% 7030     6.9KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-setup-genesis.sh completed
local-start-node.sh                                                                                                                                                                                                                 100% 7328     7.2KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-start-node.sh completed
local-update-to-consensus-node.sh                                                                                                                                                                                                   100% 5050     4.9KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-update-to-consensus-node.sh completed
prepare.sh                                                                                                                                                                                                                          100%   15KB  14.8KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/prepare.sh completed
start.sh                                                                                                                                                                                                                            100%   16KB  16.2KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/start.sh completed
transfer.sh                                                                                                                                                                                                                         100%   16KB  15.8KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/transfer.sh completed
ethkey                                                                                                                                                                                                                              100%   26MB  26.1MB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/bin/ethkey completed
platone                                                                                                                                                                                                                             100%   36MB  35.8MB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/bin/platone completed
platonecli                                                                                                                                                                                                                          100%   31MB  30.9MB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/bin/platonecli completed
repstr                                                                                                                                                                                                                              100%   19KB  19.1KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/bin/repstr completed
[INFO] [transfer] : Create /home/wxuser/PlatONE/test/data/node-3/deploy_conf completed
deploy_node-3.conf                                                                                                                                                                                                                  100%  404     0.4KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/data/node-3/deploy_conf/deploy_node-3.conf completed
[INFO] [transfer] : Transfer files to Node-3 completed

[INFO] [transfer] : Transfer completed

###########################################
####             init nodes            ####
###########################################

################ Init Node-0 ################
[INFO] [keygen] : ## Node-0 Keygen Start ##
[INFO] [keygen] : Files: /home/wxuser/PlatONE/test/data/node-0/node.address, /home/wxuser/PlatONE/test/data/node-0/node.prikey, /home/wxuser/PlatONE/test/data/node-0/node.pubkey
        Node-0's address: A62fFB5792076a35a1F3d2c8D2bc2AF87299B9cF
        Node-0's private key: 1482c3ae9e032354c6c4307b583b5692ae636e16eae83a07a08763d781a1832c
        Node-0's public key: 181bb59f6a0f3f62394a9e78ff358be178ec2f45be46f34039960f2850046742108ce1cf580e05533dbd5da3fddb3728aff67365ab825e24e91d86f325d121eb
[INFO] [keygen] : Node-0 keygen succeeded
[INFO] [init] : Generate key for node-0 completed
[INFO] [setup-genesis] : ## Setup Genesis Start ##
[INFO] [setup-genesis] : File: /home/wxuser/PlatONE/test/conf/genesis.json
[INFO] [setup-genesis] : Genesis:
{
    "config": {
    "chainId": 300,
    "interpreter": "all",
    "istanbul": {
        "timeout": 10000,
        "period": 1,
        "policy": 0,
        "firstValidatorNode": "enode://181bb59f6a0f3f62394a9e78ff358be178ec2f45be46f34039960f2850046742108ce1cf580e05533dbd5da3fddb3728aff67365ab825e24e91d86f325d121eb@10.230.48.11:16791"
    }
  },
  "timestamp": "1627955438",
  "extraData": "0x00",
  "alloc": {
    "0xA62fFB5792076a35a1F3d2c8D2bc2AF87299B9cF": {
      "balance": "100000000000000000000"
    }
  }
}
[INFO] [setup-genesis] : Setup genesis succeeded
[INFO] [init] : Setup genesis file completed
[INFO] [init] : Get genesis file completed
[INFO] [init] : Setup firstnode info completed
[INFO] [init] : Sync firstnode info to node-0 completed
******************************************************************************************************************************************************************************
INFO [08-03|09:50:38.680] Maximum peer count                       ETH=50 LES=0 total=50 RoutineID=1
INFO [08-03|09:50:38.680] Allocated cache and file handles         database=/home/wxuser/PlatONE/test/data/node-0/platone/chaindata cache=16 handles=16 RoutineID=1
INFO [08-03|09:50:38.695] Persisted trie from memory database      nodes=12 size=2.27kB time=136.397µs gcnodes=0 gcsize=0.00B gctime=0s livenodes=1 livesize=0.00B RoutineID=1
INFO [08-03|09:50:38.696] Successfully wrote genesis state         database=chaindata                                               hash=37c055…7b55f3 RoutineID=1
INFO [08-03|09:50:38.696] Allocated cache and file handles         database=/home/wxuser/PlatONE/test/data/node-0/platone/lightchaindata cache=16 handles=16 RoutineID=1
INFO [08-03|09:50:38.717] Persisted trie from memory database      nodes=12 size=2.27kB time=165.256µs gcnodes=0 gcsize=0.00B gctime=0s livenodes=1 livesize=0.00B RoutineID=1
INFO [08-03|09:50:38.718] Successfully wrote genesis state         database=lightchaindata                                               hash=37c055…7b55f3 RoutineID=1
******************************************************************************************************************************************************************************
[INFO] [init] :Init genesis on node-0 completed
[INFO] [init]: Init node Node-0 completed

......

################ Init Node-3 ################
[INFO] [keygen] : ## Node-3 Keygen Start ##
[INFO] [keygen] : Files: /home/wxuser/PlatONE/test/data/node-3/node.address, /home/wxuser/PlatONE/test/data/node-3/node.prikey, /home/wxuser/PlatONE/test/data/node-3/node.pubkey
        Node-3's address: 5145307a11Dd23e640BBDfc2ed8E7b92efac48F0
        Node-3's private key: 5888a7fb10074d09b77047b3b588f84cd4f89ff6b74c23b518a47ac5924f8b79
        Node-3's public key: fd1b4afa703b3ad6ddb61d9ebbf928939c32ca63b4687a33bdf7b06f28ddfe1b14cf3fca72c4f8f661c955760c7bbf557721e6f1c5b9f79270f82a209025a79a
[INFO] [keygen] : Node-3 keygen succeeded
[INFO] [init] : Generate key for node-3 completed
genesis.json                                                                                                                                                                                                                        100%  514     0.5KB/s   00:00
[INFO] [init] : Send genesis file to node-3 completed
firstnode.info                                                                                                                                                                                                                      100%   62     0.1KB/s   00:00
[INFO] [init] : Sync firstnode info to node-3 completed
******************************************************************************************************************************************************************************
INFO [08-03|09:50:46.595] Maximum peer count                       ETH=50 LES=0 total=50 RoutineID=1
INFO [08-03|09:50:46.595] Allocated cache and file handles         database=/home/wxuser/PlatONE/test/data/node-3/platone/chaindata cache=16 handles=16 RoutineID=1
INFO [08-03|09:50:46.610] Persisted trie from memory database      nodes=12 size=2.27kB time=178.726µs gcnodes=0 gcsize=0.00B gctime=0s livenodes=1 livesize=0.00B RoutineID=1
INFO [08-03|09:50:46.611] Successfully wrote genesis state         database=chaindata                                               hash=37c055…7b55f3 RoutineID=1
INFO [08-03|09:50:46.612] Allocated cache and file handles         database=/home/wxuser/PlatONE/test/data/node-3/platone/lightchaindata cache=16 handles=16 RoutineID=1
INFO [08-03|09:50:46.632] Persisted trie from memory database      nodes=12 size=2.27kB time=140.55µs  gcnodes=0 gcsize=0.00B gctime=0s livenodes=1 livesize=0.00B RoutineID=1
INFO [08-03|09:50:46.632] Successfully wrote genesis state         database=lightchaindata                                               hash=37c055…7b55f3 RoutineID=1
******************************************************************************************************************************************************************************
[INFO] [init] :Init genesis on node-3 completed
[INFO] [init]: Init node Node-3 completed

[INFO] [init] : Init completed

###########################################
####            start  nodes           ####
###########################################

################ Start first node Node-0 ################
[INFO] [start-node] : ## Run node-0 ##
[INFO] [start-node] : Node's url: 10.230.48.11:6791
[INFO] [start-node] : Run node-0 succeeded
[INFO] [start] : Run node Node-0 completed
[INFO] [create-account] : ## Create account Start ##
[WARN] [create-account] : !!! An account will be created. The default password is 0 !!!
[INFO] [create-account] : Account:
        New account: 0xa471d262659949c4aec745f89cdcb65437abb9a0
        Passphrase: 0
[INFO] [create-account] : Create account succeeded
[INFO] [start] : Create account completed
[INFO] [start] : Get keyfile completed
[INFO] [deploy-system-contract] : ## Deploy System Contract Start ##
[INFO] [deploy-system-contract] : Set Node-0 as super admin completed
[INFO] [deploy-system-contract] : Set Node-0 as chain admin completed
[INFO] [deploy-system-contract] : Deploy system contract completed
[INFO] [start] : Deploy system contract completed
[INFO] [add-node] : ## Add Node-0 Start ##
[INFO] [add-node] : Add Node-0 succeeded
[INFO] [start] : Add node node-0 completed
[INFO] [update-to-consensus-node] : ## Update Node-0 To Consensus Node Start ##
[INFO] [update-to-consensus-node] : Update Node-0 to consensus node succeeded
[INFO] [start] : Update node node-0 to consensus node completed
[INFO] [start] : Start firstnode Node-0 completed

......

################ Start Node-3 ################
keyfile.json                                                                                                                                                                                                                        100%  491     0.5KB/s   00:00
keyfile.phrase                                                                                                                                                                                                                      100%    2     0.0KB/s   00:00
[INFO] [start] : Send keyfile to Node-3 completed
[INFO] [start-node] : ## Run node-3 ##
[INFO] [start-node] : Node's url: 10.230.48.14:6793
[INFO] [start-node] : Run node-3 succeeded
[INFO] [start] : Run node Node-3 completed
[INFO] [add-node] : ## Add Node-3 Start ##
[INFO] [add-node] : Add Node-3 succeeded
[INFO] [start] : Add node node-3 completed
[INFO] [update-to-consensus-node] : ## Update Node-3 To Consensus Node Start ##
[INFO] [update-to-consensus-node] : Update Node-3 to consensus node succeeded
[INFO] [start] : Update node node-3 to consensus node completed
[INFO] [start] : Start node Node-3 completed

[INFO] [start] : Start completed
```

#### 3 本地一键部署单节点

##### 命令输入

```shell
./deploy.sh -p test -m one
```

##### 打印输出

```
[INFO] [deploy] : Project's conf path: /home/wxuser/platone_deploy/deployment_conf/test

###########################################
####       prepare default files       ####
###########################################

[INFO] [prepare] : Set up directory structure completed

################ Generate Configuration File For wxuser@10.230.48.11 Start ################
[INFO] [prepare] : Generate /home/wxuser/platone_deploy/deployment_conf/test/deploy_node-0.conf for wxuser@10.230.48.11 completed

###########################################
####       transfer file to nodes      ####
###########################################

################ Transfer file to Node-0 ################
[INFO] [transfer] : Create /home/wxuser/PlatONE/test/conf completed
[INFO] [transfer] : Create /home/wxuser/PlatONE/test/scripts completed
[INFO] [transfer] : Create /home/wxuser/PlatONE/test/bin completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/conf/genesis.json.istanbul.template completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/clear.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/deploy.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/init.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-add-node.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-create-account.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-deploy-system-contract.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-keygen.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-setup-genesis.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-start-node.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-update-to-consensus-node.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/prepare.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/start.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/transfer.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/bin/ethkey completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/bin/platone completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/bin/platonecli completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/bin/repstr completed
[INFO] [transfer] : Create /home/wxuser/PlatONE/test/data/node-0/deploy_conf completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/data/node-0/deploy_conf/deploy_node-0.conf completed
[INFO] [transfer] : Transfer files to Node-0 completed

[INFO] [transfer] : Transfer completed

###########################################
####             init nodes            ####
###########################################

################ Init Node-0 ################
[INFO] [keygen] : ## Node-0 Keygen Start ##
[INFO] [keygen] : Files: /home/wxuser/PlatONE/test/data/node-0/node.address, /home/wxuser/PlatONE/test/data/node-0/node.prikey, /home/wxuser/PlatONE/test/data/node-0/node.pubkey
        Node-0's address: 5Bb12346D3D6ae7fA6CbDA7D607da3715E47207A
        Node-0's private key: 5a8cea195e37df1e2ba395feea8ba5ca25d874fc11894e6d7cbdefdbcedc18e4
        Node-0's public key: 6a7bdacc1a58ae2fffa4dec517f6f52279d92688548eea78114874f1538f2f48d68cc62a5f23b29711cded67fe3e6bdf0b5de0c3b4c680b39d005ce5a528f098
[INFO] [keygen] : Node-0 keygen succeeded
[INFO] [init] : Generate key for node-0 completed
[INFO] [setup-genesis] : ## Setup Genesis Start ##
[INFO] [setup-genesis] : File: /home/wxuser/PlatONE/test/conf/genesis.json
[INFO] [setup-genesis] : Genesis:
{
    "config": {
    "chainId": 300,
    "interpreter": "all",
    "istanbul": {
        "timeout": 10000,
        "period": 1,
        "policy": 0,
        "firstValidatorNode": "enode://6a7bdacc1a58ae2fffa4dec517f6f52279d92688548eea78114874f1538f2f48d68cc62a5f23b29711cded67fe3e6bdf0b5de0c3b4c680b39d005ce5a528f098@10.230.48.11:16791"
    }
  },
  "timestamp": "1627956419",
  "extraData": "0x00",
  "alloc": {
    "0x5Bb12346D3D6ae7fA6CbDA7D607da3715E47207A": {
      "balance": "100000000000000000000"
    }
  }
}
[INFO] [setup-genesis] : Setup genesis succeeded
[INFO] [init] : Setup genesis file completed
[INFO] [init] : Get genesis file completed
[INFO] [init] : Setup firstnode info completed
[INFO] [init] : Sync firstnode info to node-0 completed
******************************************************************************************************************************************************************************
INFO [08-03|10:07:00.149] Maximum peer count                       ETH=50 LES=0 total=50 RoutineID=1
INFO [08-03|10:07:00.150] Allocated cache and file handles         database=/home/wxuser/PlatONE/test/data/node-0/platone/chaindata cache=16 handles=16 RoutineID=1
INFO [08-03|10:07:00.165] Persisted trie from memory database      nodes=12 size=2.27kB time=95.879µs gcnodes=0 gcsize=0.00B gctime=0s livenodes=1 livesize=0.00B RoutineID=1
INFO [08-03|10:07:00.166] Successfully wrote genesis state         database=chaindata                                               hash=b729be…0c0ed5 RoutineID=1
INFO [08-03|10:07:00.166] Allocated cache and file handles         database=/home/wxuser/PlatONE/test/data/node-0/platone/lightchaindata cache=16 handles=16 RoutineID=1
INFO [08-03|10:07:00.187] Persisted trie from memory database      nodes=12 size=2.27kB time=217.835µs gcnodes=0 gcsize=0.00B gctime=0s livenodes=1 livesize=0.00B RoutineID=1
INFO [08-03|10:07:00.188] Successfully wrote genesis state         database=lightchaindata                                               hash=b729be…0c0ed5 RoutineID=1
******************************************************************************************************************************************************************************
[INFO] [init] :Init genesis on node-0 completed
[INFO] [init]: Init node Node-0 completed

[INFO] [init] : Init completed

###########################################
####            start  nodes           ####
###########################################

################ Start first node Node-0 ################
[INFO] [start-node] : ## Run node-0 ##
[INFO] [start-node] : Node's url: 10.230.48.11:6791
[INFO] [start-node] : Run node-0 succeeded
[INFO] [start] : Run node Node-0 completed
[INFO] [create-account] : ## Create account Start ##
[WARN] [create-account] : !!! An account will be created. The default password is 0 !!!
[INFO] [create-account] : Account:
        New account: 0x13565cd218415ccd4e2114fd2024d4a010912d93
        Passphrase: 0
[INFO] [create-account] : Create account succeeded
[INFO] [start] : Create account completed
[INFO] [start] : Get keyfile completed
[INFO] [deploy-system-contract] : ## Deploy System Contract Start ##
[INFO] [deploy-system-contract] : Set Node-0 as super admin completed
[INFO] [deploy-system-contract] : Set Node-0 as chain admin completed
[INFO] [deploy-system-contract] : Deploy system contract completed
[INFO] [start] : Deploy system contract completed
[INFO] [add-node] : ## Add Node-0 Start ##
[INFO] [add-node] : Add Node-0 succeeded
[INFO] [start] : Add node node-0 completed
[INFO] [update-to-consensus-node] : ## Update Node-0 To Consensus Node Start ##
[INFO] [update-to-consensus-node] : Update Node-0 to consensus node succeeded
[INFO] [start] : Update node node-0 to consensus node completed
[INFO] [start] : Start firstnode Node-0 completed

[INFO] [start] : Start completed
```

#### 4 本地一键部署四节点

##### 命令输入

```shell
./deploy.sh -p test -m four
```

##### 打印输出

```
[INFO] [deploy] : Project's conf path: /home/wxuser/platone_deploy/deployment_conf/test

###########################################
####       prepare default files       ####
###########################################

[INFO] [prepare] : Set up directory structure completed

################ Generate Configuration File For wxuser@10.230.48.11 Start ################
[INFO] [prepare] : Generate /home/wxuser/platone_deploy/deployment_conf/test/deploy_node-0.conf for wxuser@10.230.48.11 completed

################ Generate Configuration File For wxuser@10.230.48.11 Start ################
[INFO] [prepare] : Generate /home/wxuser/platone_deploy/deployment_conf/test/deploy_node-1.conf for wxuser@10.230.48.11 completed

################ Generate Configuration File For wxuser@10.230.48.11 Start ################
[INFO] [prepare] : Generate /home/wxuser/platone_deploy/deployment_conf/test/deploy_node-2.conf for wxuser@10.230.48.11 completed

################ Generate Configuration File For wxuser@10.230.48.11 Start ################
[INFO] [prepare] : Generate /home/wxuser/platone_deploy/deployment_conf/test/deploy_node-3.conf for wxuser@10.230.48.11 completed

###########################################
####       transfer file to nodes      ####
###########################################

################ Transfer file to Node-0 ################
[INFO] [transfer] : Create /home/wxuser/PlatONE/test/conf completed
[INFO] [transfer] : Create /home/wxuser/PlatONE/test/scripts completed
[INFO] [transfer] : Create /home/wxuser/PlatONE/test/bin completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/conf/genesis.json.istanbul.template completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/clear.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/deploy.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/init.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-add-node.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-create-account.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-deploy-system-contract.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-keygen.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-setup-genesis.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-start-node.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-update-to-consensus-node.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/prepare.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/start.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/transfer.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/bin/ethkey completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/bin/platone completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/bin/platonecli completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/bin/repstr completed
[INFO] [transfer] : Create /home/wxuser/PlatONE/test/data/node-0/deploy_conf completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/data/node-0/deploy_conf/deploy_node-0.conf completed
[INFO] [transfer] : Transfer files to Node-0 completed

################ Transfer file to Node-1 ################
[INFO] [transfer] : Create /home/wxuser/PlatONE/test/data/node-1/deploy_conf completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/data/node-1/deploy_conf/deploy_node-1.conf completed
[INFO] [transfer] : Transfer files to Node-1 completed

################ Transfer file to Node-2 ################
[INFO] [transfer] : Create /home/wxuser/PlatONE/test/data/node-2/deploy_conf completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/data/node-2/deploy_conf/deploy_node-2.conf completed
[INFO] [transfer] : Transfer files to Node-2 completed

################ Transfer file to Node-3 ################
[INFO] [transfer] : Create /home/wxuser/PlatONE/test/data/node-3/deploy_conf completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/data/node-3/deploy_conf/deploy_node-3.conf completed
[INFO] [transfer] : Transfer files to Node-3 completed

[INFO] [transfer] : Transfer completed

###########################################
####             init nodes            ####
###########################################

################ Init Node-0 ################
[INFO] [keygen] : ## Node-0 Keygen Start ##
[INFO] [keygen] : Files: /home/wxuser/PlatONE/test/data/node-0/node.address, /home/wxuser/PlatONE/test/data/node-0/node.prikey, /home/wxuser/PlatONE/test/data/node-0/node.pubkey
        Node-0's address: E9F41A5C29BE6449b63183bC3fDC0C9d40d39bF8
        Node-0's private key: d1279c5e9f4ca9c0aaf93e2534ab850eefd3f9f0e006ac0431409e9dc497d74a
        Node-0's public key: 6e75b28e04e6080890032cfb154d8f7b62d6d2231fd6168a60e4fa854e8cf06a43e5f7acc4f728ec06cc8b21b6a8ebd6eef48e46fee19fa9047788585789aed3
[INFO] [keygen] : Node-0 keygen succeeded
[INFO] [init] : Generate key for node-0 completed
[INFO] [setup-genesis] : ## Setup Genesis Start ##
[INFO] [setup-genesis] : File: /home/wxuser/PlatONE/test/conf/genesis.json
[INFO] [setup-genesis] : Genesis:
{
    "config": {
    "chainId": 300,
    "interpreter": "all",
    "istanbul": {
        "timeout": 10000,
        "period": 1,
        "policy": 0,
        "firstValidatorNode": "enode://6e75b28e04e6080890032cfb154d8f7b62d6d2231fd6168a60e4fa854e8cf06a43e5f7acc4f728ec06cc8b21b6a8ebd6eef48e46fee19fa9047788585789aed3@10.230.48.11:16791"
    }
  },
  "timestamp": "1627956725",
  "extraData": "0x00",
  "alloc": {
    "0xE9F41A5C29BE6449b63183bC3fDC0C9d40d39bF8": {
      "balance": "100000000000000000000"
    }
  }
}
[INFO] [setup-genesis] : Setup genesis succeeded
[INFO] [init] : Setup genesis file completed
[INFO] [init] : Get genesis file completed
[INFO] [init] : Setup firstnode info completed
[INFO] [init] : Sync firstnode info to node-0 completed
******************************************************************************************************************************************************************************
INFO [08-03|10:12:05.752] Maximum peer count                       ETH=50 LES=0 total=50 RoutineID=1
INFO [08-03|10:12:05.752] Allocated cache and file handles         database=/home/wxuser/PlatONE/test/data/node-0/platone/chaindata cache=16 handles=16 RoutineID=1
INFO [08-03|10:12:05.770] Persisted trie from memory database      nodes=12 size=2.27kB time=276.455µs gcnodes=0 gcsize=0.00B gctime=0s livenodes=1 livesize=0.00B RoutineID=1
INFO [08-03|10:12:05.771] Successfully wrote genesis state         database=chaindata                                               hash=f81d4a…02a054 RoutineID=1
INFO [08-03|10:12:05.771] Allocated cache and file handles         database=/home/wxuser/PlatONE/test/data/node-0/platone/lightchaindata cache=16 handles=16 RoutineID=1
INFO [08-03|10:12:05.790] Persisted trie from memory database      nodes=12 size=2.27kB time=133.979µs gcnodes=0 gcsize=0.00B gctime=0s livenodes=1 livesize=0.00B RoutineID=1
INFO [08-03|10:12:05.791] Successfully wrote genesis state         database=lightchaindata                                               hash=f81d4a…02a054 RoutineID=1
******************************************************************************************************************************************************************************
[INFO] [init] :Init genesis on node-0 completed
[INFO] [init]: Init node Node-0 completed

......

################ Init Node-3 ################
[INFO] [keygen] : ## Node-3 Keygen Start ##
[INFO] [keygen] : Files: /home/wxuser/PlatONE/test/data/node-3/node.address, /home/wxuser/PlatONE/test/data/node-3/node.prikey, /home/wxuser/PlatONE/test/data/node-3/node.pubkey
        Node-3's address: 5B66581b061793c9d70be95137FBe395d1f6177e
        Node-3's private key: 36e8b1d2224dde9eb7d28d64161207cf802ccbc62f097982cfb11e6970d52767
        Node-3's public key: 28005cdf7c9489e636cdbb30207f9fa0b825a211740e65470e0ddb2f488479740ebfa5b9df81b3bf767d3e0ab92f5417777891d3436d6d0486394e0e2c34f476
[INFO] [keygen] : Node-3 keygen succeeded
[INFO] [init] : Generate key for node-3 completed
[INFO] [init] : Send genesis file to node-3 completed
[INFO] [init] : Sync firstnode info to node-3 completed
******************************************************************************************************************************************************************************
INFO [08-03|10:12:07.244] Maximum peer count                       ETH=50 LES=0 total=50 RoutineID=1
INFO [08-03|10:12:07.244] Allocated cache and file handles         database=/home/wxuser/PlatONE/test/data/node-3/platone/chaindata cache=16 handles=16 RoutineID=1
INFO [08-03|10:12:07.259] Persisted trie from memory database      nodes=12 size=2.27kB time=138.423µs gcnodes=0 gcsize=0.00B gctime=0s livenodes=1 livesize=0.00B RoutineID=1
INFO [08-03|10:12:07.260] Successfully wrote genesis state         database=chaindata                                               hash=f81d4a…02a054 RoutineID=1
INFO [08-03|10:12:07.260] Allocated cache and file handles         database=/home/wxuser/PlatONE/test/data/node-3/platone/lightchaindata cache=16 handles=16 RoutineID=1
INFO [08-03|10:12:07.280] Persisted trie from memory database      nodes=12 size=2.27kB time=144.714µs gcnodes=0 gcsize=0.00B gctime=0s livenodes=1 livesize=0.00B RoutineID=1
INFO [08-03|10:12:07.281] Successfully wrote genesis state         database=lightchaindata                                               hash=f81d4a…02a054 RoutineID=1
******************************************************************************************************************************************************************************
[INFO] [init] :Init genesis on node-3 completed
[INFO] [init]: Init node Node-3 completed

[INFO] [init] : Init completed

###########################################
####            start  nodes           ####
###########################################

################ Start first node Node-0 ################
[INFO] [start-node] : ## Run node-0 ##
[INFO] [start-node] : Node's url: 10.230.48.11:6791
[INFO] [start-node] : Run node-0 succeeded
[INFO] [start] : Run node Node-0 completed
[INFO] [create-account] : ## Create account Start ##
[WARN] [create-account] : !!! An account will be created. The default password is 0 !!!
[INFO] [create-account] : Account:
        New account: 0x300a877d0d81c929a9294a6e697216068e9e2d54
        Passphrase: 0
[INFO] [create-account] : Create account succeeded
[INFO] [start] : Create account completed
[INFO] [start] : Get keyfile completed
[INFO] [deploy-system-contract] : ## Deploy System Contract Start ##
[INFO] [deploy-system-contract] : Set Node-0 as super admin completed
[INFO] [deploy-system-contract] : Set Node-0 as chain admin completed
[INFO] [deploy-system-contract] : Deploy system contract completed
[INFO] [start] : Deploy system contract completed
[INFO] [add-node] : ## Add Node-0 Start ##
[INFO] [add-node] : Add Node-0 succeeded
[INFO] [start] : Add node node-0 completed
[INFO] [update-to-consensus-node] : ## Update Node-0 To Consensus Node Start ##
[INFO] [update-to-consensus-node] : Update Node-0 to consensus node succeeded
[INFO] [start] : Update node node-0 to consensus node completed
[INFO] [start] : Start firstnode Node-0 completed

......

################ Start Node-3 ################
[INFO] [start] : Send keyfile to Node-3 completed
[INFO] [start-node] : ## Run node-3 ##
[INFO] [start-node] : Node's url: 10.230.48.11:6795
[INFO] [start-node] : Run node-3 succeeded
[INFO] [start] : Run node Node-3 completed
[INFO] [add-node] : ## Add Node-3 Start ##
[INFO] [add-node] : Add Node-3 succeeded
[INFO] [start] : Add node node-3 completed
[INFO] [update-to-consensus-node] : ## Update Node-3 To Consensus Node Start ##
[INFO] [update-to-consensus-node] : Update Node-3 to consensus node succeeded
[INFO] [start] : Update node node-3 to consensus node completed
[INFO] [start] : Start node Node-3 completed

[INFO] [start] : Start completed
```

## clear.sh 清理节点脚本

### 功能

- 停止节点运行进程、从区块链将节点删除、清除节点的相关文件
- 支持对全节点或指定节点进行操作

### 命令

```shell
USAGE: clear.sh  [options] [value]

        OPTIONS:

           --project, -p              the specified project name. must be specified

           --node, -n                 the specified node name.
                                      default='all': deploy all nodes by conf in deployment_conf
                                      use ',' to seperate the name of node

           --mode, -m                 the specified execute mode.
                                      'delete': will delete the node from chain
                                      'clean': will clean the files, configuration files will be backed up
                                      'stop' : will stop the node
                                      'deep': will do delete clean and stop

           --help, -h                 show help
```

- --project, -p:
  - 项目名称，属于必填项
- --node, -n:
  - 节点名称，不填默认为node=all
  - 如果设置值为all，那么脚本会根据项目路径下所有配置文件向对应的节点依次进行清理节点
  - 如果设置多个指定节点，那么需要用","作为不同的节点名称的分隔符
- --mode, m
  - 模式名称，不填默认为mode=deep
  - delete：通过platonecli将节点从区块链删除，即status变为2
  - clean：将目标机中的node_dir清除，备份deploy_node配置文件。如果项目下所有node_dir都已经被清除，则会自动备份conf下的文件，并将bin、scripts、data，conf目录都删除
  - stop：根据节点的端口号信息，获取PID并杀掉进程
  - deep：按照delete->clean->stop的顺序依次执行。"deep all"模式下跳过delete操作。

### 使用演示（按序号依次执行）

#### 1 清理指定节点

模拟情况：1个节点为部署机本机（wxuser@10.230.48.11），另3个节点为远程目标机

##### 命令输入

```shell
./clear.sh -p test -n 2
```

##### 打印输出

```
################ Start to clear Node-2 ################
[INFO] [clear] : Delete node-2 end
[INFO] [clear] : Check ip 10.230.48.12 connection completed
[INFO] [clear] : Check ssh wxuser@10.230.48.12 access completed
[INFO] [clear] : Get PID of wxuser@10.230.48.12:6793 completed
[INFO] [clear] : Kill PID 759 of wxuser@10.230.48.12:6793 completed
[INFO] [clear] : Stop node-2 end
[INFO] [clear] : Check ip 10.230.48.12 connection completed
[INFO] [clear] : Check ssh wxuser@10.230.48.12 access completed
[INFO] [clear] : Backup wxuser@10.230.48.12:/home/wxuser/PlatONE/test/data/node-2/deploy_conf/deploy_node-2.conf to wxuser@10.230.48.12:/home/wxuser/PlatONE/test/../bak/test/deploy_node-2.conf.bak.20210802171627898159M59 completed
[INFO] [clear] : Remove wxuser@10.230.48.12:/home/wxuser/PlatONE/test/data/node-2 completed
[INFO] [clear] : Clean node-2 end

[INFO] [clear] : Clear action end
```

#### 2 清理全节点

模拟情况：2个节点为部署机本机（wxuser@10.230.48.11），另2个节点为远程目标机，其中node-2已经清理完毕

##### 命令输入

```shell
./clear.sh -p test
```

##### 打印输出

```
#### Start to clear Node-0 ####
[INFO] [clear] : Delete node-0 end
[INFO] [clear] : Get PID of wxuser@10.230.48.11:6801 completed
[INFO] [clear] : Kill PID 62276 of wxuser@10.230.48.11:6801 completed
[INFO] [clear] : Stop node-0 end
[INFO] [clear] : Backup wxuser@10.230.48.11:/home/wxuser/PlatONE/test/data/node-0/deploy_conf/deploy_node-0.conf to wxuser@10.230.48.11:/home/wxuser/PlatONE/test/../bak/test/deploy_node-0.conf.bak.20210802171627898341M01 completed
[INFO] [clear] : Remove wxuser@10.230.48.11:/home/wxuser/PlatONE/test/data/node-0 completed
[INFO] [clear] : Clean node-0 end

#### Start to clear Node-1 ####
[INFO] [clear] : Delete node-1 end
[INFO] [clear] : Get PID of wxuser@10.230.48.11:6802 completed
[INFO] [clear] : Kill PID 62674 of wxuser@10.230.48.11:6802 completed
[INFO] [clear] : Stop node-1 end
[INFO] [clear] : Backup wxuser@10.230.48.11:/home/wxuser/PlatONE/test/data/node-1/deploy_conf/deploy_node-1.conf to wxuser@10.230.48.11:/home/wxuser/PlatONE/test/../bak/test/deploy_node-1.conf.bak.20210802171627898341M01 completed
[INFO] [clear] : Remove wxuser@10.230.48.11:/home/wxuser/PlatONE/test/data/node-1 completed
[INFO] [clear] : Remove wxuser@10.230.48.11:/home/wxuser/PlatONE/test/scripts completed
[INFO] [clear] : Remove wxuser@10.230.48.11:/home/wxuser/PlatONE/test/data completed
[INFO] [clear] : Remove wxuser@10.230.48.11:/home/wxuser/PlatONE/test/bin completed
[INFO] [clear] : Backup wxuser@10.230.48.11:/home/wxuser/PlatONE/test/conf completed
[INFO] [clear] : Clean node-1 end

#### Start to clear Node-2 ####
[INFO] [clear] : Delete node-2 end
[INFO] [clear] : Check ip 10.230.48.12 connection completed
[INFO] [clear] : Check ssh wxuser@10.230.48.12 access completed
[WARN] [clear] : !!! GET PID OF wxuser@10.230.48.12:6793 FAILED, MAYBE HAS ALREADY BEEN STOPPED !!!
[INFO] [clear] : Stop node-2 end
[INFO] [clear] : Check ip 10.230.48.12 connection completed
[INFO] [clear] : Check ssh wxuser@10.230.48.12 access completed
[WARN] [clear] : !!! wxuser@10.230.48.12:/home/wxuser/PlatONE/test/data/node-2 NOT FOUND, MAYBE HAS ALREADY BEEN CLEANED !!!
[INFO] [clear] : Clean node-2 end

#### Start to clear Node-3 ####
[INFO] [clear] : Delete node-3 end
[INFO] [clear] : Check ip 10.230.48.12 connection completed
[INFO] [clear] : Check ssh wxuser@10.230.48.12 access completed
[INFO] [clear] : Get PID of wxuser@10.230.48.12:6794 completed
[INFO] [clear] : Kill PID 1287 of wxuser@10.230.48.12:6794 completed
[INFO] [clear] : Stop node-3 end
[INFO] [clear] : Check ip 10.230.48.12 connection completed
[INFO] [clear] : Check ssh wxuser@10.230.48.12 access completed
[INFO] [clear] : Backup wxuser@10.230.48.12:/home/wxuser/PlatONE/test/data/node-3/deploy_conf/deploy_node-3.conf to wxuser@10.230.48.12:/home/wxuser/PlatONE/test/../bak/test/deploy_node-3.conf.bak.20210802171627898354M14 completed
[INFO] [clear] : Remove wxuser@10.230.48.12:/home/wxuser/PlatONE/test/data/node-3 completed
[INFO] [clear] : Remove wxuser@10.230.48.12:/home/wxuser/PlatONE/test/scripts completed
[INFO] [clear] : Remove wxuser@10.230.48.12:/home/wxuser/PlatONE/test/data completed
[INFO] [clear] : Remove wxuser@10.230.48.12:/home/wxuser/PlatONE/test/bin completed
[INFO] [clear] : Backup wxuser@10.230.48.12:/home/wxuser/PlatONE/test/conf completed
Do you want to remove /home/wxuser/platone_deploy/deployment_conf/test/global and /home/wxuser/platone_deploy/deployment_conf/test/logs? Yes or No(y/n):
y
[INFO] [clear] : Clean node-3 end

[INFO] [clear] : Clear action end
```

##### 目录结构（目标主机PlatONE目录）

```
├── bak
│   ├── test.bak.20210802161627894290M30
│   │   ├── conf
│   │   │   ├── firstnode.info
│   │   │   ├── genesis.json
│   │   │   ├── genesis.json.istanbul.template
│   │   │   ├── keyfile.account
│   │   │   ├── keyfile.json
│   │   │   └── keyfile.phrase
│   │   ├── deploy_node-0.conf.bak.20210802161627894290M30
│   │   └── deploy_node-1.conf.bak.20210802161627894290M30
```

## prepare.sh 生成配置文件脚本

### 功能

- 根据用户名和ip地址自动生成配置文件
- 能够保证不会与目标主机已有的其他节点的配置文件中的端口号冲突
- 能够保证不会使用目标主机已被占用的端口号

### 命令

```shell
"
USAGE: prepare.sh  [options] [value]

        OPTIONS:

           --project, -p             the specified project name. must be specified

           --address, -a             nodes' addresses. must be specified

           --cover                   will backup the project directory if exists

           --help, -h                show help
"
```

- --project, -p:
  - 项目名称，属于必填项
  - 会新建项目路径：XXX/platone_deploy/deployment_conf/${project_name}
  - 所有生成的配置文件都会放于项目路径下
- --address, -a:
  - 节点地址，属于必填项
  - 格式为：${USER_NAME}@${IP_ADDR}
  - 多个地址之间用","隔开
- --cover:
  - 可选参数，在已存在项目路径的情况下进行覆盖操作
  - 覆盖操作：按项目路径下的配置文件清除所有节点，再对项目路径下的文件进行备份并移除
    - 如果项目路径已存在，选择了cover，会自动进行覆盖操作
    - 如果项目路径已存在，单未选择cover，那么会弹出提示手动进行yesOrNo选择是否进行覆盖操作
      - 如果选择了覆盖，则进行覆盖操作
      - 如果没有选择覆盖，则会在原项目路径下生成配置文件，会提前检测已存在配置文件设置的端口

### 使用演示（按序号依次进行）

#### 1 项目路径不存在时，按地址生成配置文件

模拟情况：1个节点为部署机本机（wxuser@10.230.48.11），另3个节点为远程目标机

##### 命令输入

```shell
./prepare.sh -p test -a wxuser@10.230.48.11,wxuser@10.230.48.12,wxuser@10.230.48.13,wxuser@10.230.48.14
```

##### 打印输出

```
###########################################
####       prepare default files       ####
###########################################
[INFO] [prepare] : Set up directory structure completed

################ Generate Configuration File For wxuser@10.230.48.11 Start ################
[INFO] [prepare] : Generate /home/wxuser/platone_deploy/deployment_conf/test/deploy_node-0.conf for wxuser@10.230.48.11 completed

################ Generate Configuration File For wxuser@10.230.48.12 Start ################
[INFO] [prepare] : Check ip 10.230.48.12 connection completed
[INFO] [prepare] : Check ssh wxuser@10.230.48.12 access completed
[INFO] [prepare] : Generate /home/wxuser/platone_deploy/deployment_conf/test/deploy_node-1.conf for wxuser@10.230.48.12 completed

################ Generate Configuration File For wxuser@10.230.48.13 Start ################
[INFO] [prepare] : Check ip 10.230.48.13 connection completed
[INFO] [prepare] : Check ssh wxuser@10.230.48.13 access completed
[INFO] [prepare] : Generate /home/wxuser/platone_deploy/deployment_conf/test/deploy_node-2.conf for wxuser@10.230.48.13 completed

################ Generate Configuration File For wxuser@10.230.48.14 Start ################
[INFO] [prepare] : Check ip 10.230.48.14 connection completed
[INFO] [prepare] : Check ssh wxuser@10.230.48.14 access completed
[INFO] [prepare] : Generate /home/wxuser/platone_deploy/deployment_conf/test/deploy_node-3.conf for wxuser@10.230.48.14 completed
```

##### 目录结构（部署机部署项目目录）

```
├── deployment_conf
│   ├── logs
│   │   └── prepare_log.txt
│   └── test
│       ├── deploy_node-0.conf
│       ├── deploy_node-1.conf
│       ├── deploy_node-2.conf
│       ├── deploy_node-3.conf
│       ├── global
│       └── logs
```

#### 2 项目路径存在时，不选择覆盖操作

##### 命令输入

```shell
./prepare.sh -p test -a wxuser@10.230.48.11,wxuser@10.230.48.12,wxuser@10.230.48.13,wxuser@10.230.48.14
```

##### 打印输出

```
###########################################
####       prepare default files       ####
###########################################
/home/wxuser/platone_deploy/deployment_conf/test already exists, do you want to cover it? Yes or No(y/n): n
[WARN] [prepare] : !!! New Conf Files Will Be Generated In Exist Path !!!

################ Generate Configuration File For wxuser@10.230.48.11 Start ################
[INFO] [prepare] : Generate /home/wxuser/platone_deploy/deployment_conf/test/deploy_node-4.conf for wxuser@10.230.48.11 completed

################ Generate Configuration File For wxuser@10.230.48.12 Start ################
[INFO] [prepare] : Check ip 10.230.48.12 connection completed
[INFO] [prepare] : Check ssh wxuser@10.230.48.12 access completed
[INFO] [prepare] : Generate /home/wxuser/platone_deploy/deployment_conf/test/deploy_node-5.conf for wxuser@10.230.48.12 completed

################ Generate Configuration File For wxuser@10.230.48.13 Start ################
[INFO] [prepare] : Check ip 10.230.48.13 connection completed
[INFO] [prepare] : Check ssh wxuser@10.230.48.13 access completed
[INFO] [prepare] : Generate /home/wxuser/platone_deploy/deployment_conf/test/deploy_node-6.conf for wxuser@10.230.48.13 completed

################ Generate Configuration File For wxuser@10.230.48.14 Start ################
[INFO] [prepare] : Check ip 10.230.48.14 connection completed
[INFO] [prepare] : Check ssh wxuser@10.230.48.14 access completed
[INFO] [prepare] : Generate /home/wxuser/platone_deploy/deployment_conf/test/deploy_node-7.conf for wxuser@10.230.48.14 completed
```

##### 目录结构（部署机）

```
├── deployment_conf
│   ├── logs
│   │   └── prepare_log.txt
│   └── test
│       ├── deploy_node-0.conf
│       ├── deploy_node-1.conf
│       ├── deploy_node-2.conf
│       ├── deploy_node-3.conf
│       ├── deploy_node-4.conf
│       ├── deploy_node-5.conf
│       ├── deploy_node-6.conf
│       ├── deploy_node-7.conf
│       ├── global
│   
└── logs
```

#### 3 项目路径存在时，进行覆盖操作

模拟情况：2个节点为部署文件所在的本机（wxuser@10.230.48.11），另2个节点为远程机

##### 命令输入

```shell
./prepare.sh -p test -a wxuser@10.230.48.11,wxuser@10.230.48.11,wxuser@10.230.48.12,wxuser@10.230.48.12 --cover
```

##### 打印输出

```
###########################################
####       prepare default files       ####
###########################################
[INFO] [prepare] : Backup /home/wxuser/platone_deploy/deployment_conf/test to /home/wxuser/platone_deploy/deployment_conf/bak/test.bak.20210802144651 completed

#### Start to clear Node-0 ####
[INFO] [clear] : Delete node-0 end
[WARN] [clear] : !!! wxuser@10.230.48.11:/home/wxuser/PlatONE/test/data/node-0 NOT FOUND, MAYBE HAS ALREADY BEEN CLEANED !!!
[INFO] [clear] : Clean node-0 end
[WARN] [clear] : !!! GET PID OF wxuser@10.230.48.11:6801 FAILED, MAYBE HAS ALREADY BEEN STOPPED !!!
[INFO] [clear] : Stop node-0 end

......


#### Start to clear Node-7 ####
[INFO] [clear] : Delete node-7 end
[INFO] [clear] : Check ip 10.230.48.14 connection completed
[INFO] [clear] : Check ssh wxuser@10.230.48.14 access completed
[WARN] [clear] : !!! wxuser@10.230.48.14:/home/wxuser/PlatONE/test/data/node-7 NOT FOUND, MAYBE HAS ALREADY BEEN CLEANED !!!
[INFO] [clear] : Clean node-7 end
[INFO] [clear] : Check ip 10.230.48.14 connection completed
[INFO] [clear] : Check ssh wxuser@10.230.48.14 access completed
[WARN] [clear] : !!! GET PID OF wxuser@10.230.48.14:6794 FAILED, MAYBE HAS ALREADY BEEN STOPPED !!!
[INFO] [clear] : Stop node-7 end

[INFO] [clear] : Clear action end
[INFO] [prepare] : Set up directory structure completed

################ Generate Configuration File For wxuser@10.230.48.11 Start ################
[INFO] [prepare] : Generate /home/wxuser/platone_deploy/deployment_conf/test/deploy_node-0.conf for wxuser@10.230.48.11 completed

################ Generate Configuration File For wxuser@10.230.48.11 Start ################
[INFO] [prepare] : Generate /home/wxuser/platone_deploy/deployment_conf/test/deploy_node-1.conf for wxuser@10.230.48.11 completed

################ Generate Configuration File For wxuser@10.230.48.12 Start ################
[INFO] [prepare] : Check ip 10.230.48.12 connection completed
[INFO] [prepare] : Check ssh wxuser@10.230.48.12 access completed
[INFO] [prepare] : Generate /home/wxuser/platone_deploy/deployment_conf/test/deploy_node-2.conf for wxuser@10.230.48.12 completed

################ Generate Configuration File For wxuser@10.230.48.12 Start ################
[INFO] [prepare] : Check ip 10.230.48.12 connection completed
[INFO] [prepare] : Check ssh wxuser@10.230.48.12 access completed
[INFO] [prepare] : Generate /home/wxuser/platone_deploy/deployment_conf/test/deploy_node-3.conf for wxuser@10.230.48.12 completed
```

##### 目录结构（部署机部署项目目录）

```
├── deployment_conf
│   ├── bak
│   │   └── test.bak.20210802144651
│   │       ├── deploy_node-0.conf
│   │       ├── deploy_node-1.conf
│   │       ├── deploy_node-2.conf
│   │       ├── deploy_node-3.conf
│   │       ├── deploy_node-4.conf
│   │       ├── deploy_node-5.conf
│   │       ├── deploy_node-6.conf
│   │       ├── deploy_node-7.conf
│   │       ├── global
│   │       └── logs
│   ├── logs
│   │   └── prepare_log.txt
│   └── test
│       ├── deploy_node-0.conf
│       ├── deploy_node-1.conf
│       ├── deploy_node-2.conf
│       ├── deploy_node-3.conf
│       ├── global
│       └── logs
```

## transfer.sh 传输脚本

### 功能

- 根据项目目录下的配置文件，向要目标主机传输部署活动中需要的文件。
- 可指定节点传输

- 具有日志机制，执行成功后会记录日志。假如因中断或异常停止，下一次执行会根据日志跳过已完成的步骤。此外，根据日志，同一个地址目录(${user_name}@{ip_addr}:${deploy_path})下只会传输一次bin、conf和scripts目录及其下尚未传输的文件。
- 根据地址判断是否为远程活动，执行对应的远程登录/传输命令

### 命令

```shell
"
USAGE: transfer.sh  [options] [value]

        OPTIONS:

           --project, -p              the specified project name. must be specified

           --node, -n                 the specified node name. only used in conf mode.
                                      default='all': deploy all nodes by conf in deployment_conf
                                      use ',' to seperate the name of node

           --help, -h                 show help
"
```

- --project, -p:
  - 项目名称，属于必填项
  - 项目路径: XXX/platone_deploy/deployment_conf/${project_name}

- --node, -n:
  - 节点名称，不填默认为node=all
  - 脚本会根据项目路径下的deploy_node-${node_name}.conf配置文件进行传输操作
  - 如果设置值为all，那么脚本会根据项目路径下所有配置文件向对应的节点依次进行文件传输
  - 如果设置多个指定节点，那么需要用","作为不同的节点名称的分隔符

### 使用演示（按序号依次执行）

#### 1 指定节点传输

模拟情况：node-3为远程机（wxuser@10.230.48.13）

##### 命令输入

```shell
./transfer.sh -p test -n 3
```

##### 打印输出

```
###########################################
####       transfer file to nodes      ####
###########################################

################ Transfer file to Node-3 ################
[INFO] [transfer] : Create /home/wxuser/PlatONE/test/conf completed
[INFO] [transfer] : Create /home/wxuser/PlatONE/test/scripts completed
[INFO] [transfer] : Create /home/wxuser/PlatONE/test/bin completed
genesis.json.istanbul.template                                                                                                                                                       100%  356     0.4KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/conf/genesis.json.istanbul.template completed
clear.sh                                                                                                                                                                             100%   23KB  22.6KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/clear.sh completed
deploy.sh                                                                                                                                                                            100% 6711     6.6KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/deploy.sh completed
init.sh                                                                                                                                                                              100%   15KB  15.0KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/init.sh completed
local-add-node.sh                                                                                                                                                                    100% 5356     5.2KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-add-node.sh completed
local-create-account.sh                                                                                                                                                              100% 5905     5.8KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-create-account.sh completed
local-deploy-system-contract.sh                                                                                                                                                      100% 6519     6.4KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-deploy-system-contract.sh completed
local-keygen.sh                                                                                                                                                                      100% 5984     5.8KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-keygen.sh completed
local-setup-genesis.sh                                                                                                                                                               100% 7030     6.9KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-setup-genesis.sh completed
local-start-node.sh                                                                                                                                                                  100% 7328     7.2KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-start-node.sh completed
local-update-to-consensus-node.sh                                                                                                                                                    100% 5050     4.9KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-update-to-consensus-node.sh completed
prepare.sh                                                                                                                                                                           100%   15KB  14.8KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/prepare.sh completed
start.sh                                                                                                                                                                             100%   16KB  16.2KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/start.sh completed
transfer.sh                                                                                                                                                                          100%   16KB  15.8KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/transfer.sh completed
ethkey                                                                                                                                                                               100%   26MB  26.1MB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/bin/ethkey completed
platone                                                                                                                                                                              100%   36MB  35.8MB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/bin/platone completed
platonecli                                                                                                                                                                           100%   31MB  30.9MB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/bin/platonecli completed
repstr                                                                                                                                                                               100%   19KB  19.1KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/bin/repstr completed
[INFO] [transfer] : Create /home/wxuser/PlatONE/test/data/node-3/deploy_conf completed
deploy_node-3.conf                                                                                                                                                                   100%  404     0.4KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/data/node-3/deploy_conf/deploy_node-3.conf completed
[INFO] [transfer] : Transfer files to Node-3 completed
```

##### 目录结构（目标主机PlatONE目录）

```
└── test
    ├── bin
    │   ├── ethkey
    │   ├── platone
    │   ├── platonecli
    │   └── repstr
    ├── conf
    │   └── genesis.json.istanbul.template
    ├── data
    │   └── node-3
    │       └── deploy_conf
    │           └── deploy_node-3.conf
    └── scripts
        ├── clear.sh
        ├── deploy.sh
        ├── init.sh
        ├── local-add-node.sh
        ├── local-create-account.sh
        ├── local-deploy-system-contract.sh
        ├── local-keygen.sh
        ├── local-setup-genesis.sh
        ├── local-start-node.sh
        ├── local-update-to-consensus-node.sh
        ├── prepare.sh
        ├── start.sh
        └── transfer.sh
```

#### 2 全节点传输

模拟情况：node-3已完成传输（wxuser@10.230.48.13），node-0、node-1在一台主机，node-2、node-3在一台主机

##### 命令输入

```shell
./transfer.sh -p test
```

##### 打印输出

```
###########################################
####       transfer file to nodes      ####
###########################################

################ Transfer file to Node-0 ################
[INFO] [transfer] : Create /home/wxuser/PlatONE/test/conf completed
[INFO] [transfer] : Create /home/wxuser/PlatONE/test/scripts completed
[INFO] [transfer] : Create /home/wxuser/PlatONE/test/bin completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/conf/genesis.json.istanbul.template completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/clear.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/deploy.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/init.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-add-node.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-create-account.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-deploy-system-contract.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-keygen.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-setup-genesis.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-start-node.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/local-update-to-consensus-node.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/prepare.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/start.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/scripts/transfer.sh completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/bin/ethkey completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/bin/platone completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/bin/platonecli completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/bin/repstr completed
[INFO] [transfer] : Create /home/wxuser/PlatONE/test/data/node-0/deploy_conf completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/data/node-0/deploy_conf/deploy_node-0.conf completed
[INFO] [transfer] : Transfer files to Node-0 completed

################ Transfer file to Node-1 ################
[INFO] [transfer] : Create /home/wxuser/PlatONE/test/data/node-1/deploy_conf completed
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/data/node-1/deploy_conf/deploy_node-1.conf completed
[INFO] [transfer] : Transfer files to Node-1 completed

################ Transfer file to Node-2 ################
[INFO] [transfer] : Create /home/wxuser/PlatONE/test/data/node-2/deploy_conf completed
deploy_node-2.conf                                                                                                                                                                   100%  404     0.4KB/s   00:00
[INFO] [transfer] : Transfer /home/wxuser/PlatONE/test/data/node-2/deploy_conf/deploy_node-2.conf completed
[INFO] [transfer] : Transfer files to Node-2 completed

[INFO] [transfer] : Transfer completed
```

#### 3 指定多节点传输

模拟情况：所有节点已经完成传输

##### 命令输入

```shell
./transfer.sh -p test -n 0,1
```

##### 打印输出

```
###########################################
####       transfer file to nodes      ####
###########################################

[INFO] [transfer] : Transfer completed
```

## init.sh 初始化脚本

### 功能

- 进行启动前的初始化工作，包括生成密钥与地址、获取genesis.json文件以及初始化节点
- 支持指定节点初始化
- 具有日志机制，checkpoint设在密钥生成、genesis.json生成、全局信息同步、初始化节点
- 根据地址判断是否为远程活动，执行对应的远程登录/传输命令

### 命令

```shell
"
USAGE: init.sh  [options] [value]

        OPTIONS:

           --project, -p              the specified project name. must be specified

           --node, -n                 the specified node name. only used in conf mode.
                                      default='all': deploy all nodes by conf in deployment_conf
                                      use ',' to seperate the name of node

           --help, -h                 show help
"
```

- --project, -p:
  - 项目名称，属于必填项
  - 项目路径: XXX/platone_deploy/deployment_conf/${project_name}
- --node, -n:
  - 节点名称，不填默认为node=all
  - 脚本会根据项目路径下的deploy_node-${node_name}.conf配置文件进行初始化操作
  - 如果设置值为all，那么脚本会根据项目路径下所有配置文件向对应的节点依次进行初始化操作。只有“all”时才会选定firstnode并生成genesis.json，如果第一次部署项目，需要使用“all”
  - 如果设置多个指定节点，那么需要用","作为不同的节点名称的分隔符。指定节点初始化只有在项目已经存在firstnode并且已经有genesis.json文件时，才可以正常执行。

### 使用演示（按序号依次执行）

#### 1 指定节点初始化

模拟情况：还没有完成firstnode初始化

##### 命令输入

```shell
./init.sh -p test -n 2
```

##### 打印输出

```
###########################################
####             init nodes            ####
###########################################
[ERROR] [init] : ********* FIRSTNODE MUST INIT **********
```

#### 2 全节点初始化

模拟情况：2个节点为部署文件所在的本机（wxuser@10.230.48.11），另2个节点为远程机

##### 命令输入

```shell
./init.sh -p test
```

##### 打印输出

```
###########################################
####             init nodes            ####
###########################################
[ERROR] [init] : ********* FIRSTNODE MUST INIT **********
wxuser@WXTVOTR01-plt:~/platone_deploy/release/linux/scripts$ ./init.sh -p test

###########################################
####             init nodes            ####
###########################################

################ Init Node-0 ################
[INFO] [keygen] : ## Node-0 Keygen Start ##
[INFO] [keygen] : Files: /home/wxuser/PlatONE/test/data/node-0/node.address, /home/wxuser/PlatONE/test/data/node-0/node.prikey, /home/wxuser/PlatONE/test/data/node-0/node.pubkey
        Node-0's address: b65f89ad16E011dcbAb3905fa675083a2c551339
        Node-0's private key: 1d48419ebf947679fe38e8412e7875bfc71d29eab1aff9212e45743b83c7579e
        Node-0's public key: a65dfc6ee888b517a3ad0983c1e213b4eb1156b37d58e3d767d6ccf968cab011d56beb15184baa9d7e39e449e4e55024acc7e10fc2f1b609f403b309430d63be
[INFO] [keygen] : Node-0 keygen succeeded
[INFO] [init] : Generate key for node-0 completed
[INFO] [setup-genesis] : ## Setup Genesis Start ##
[INFO] [setup-genesis] : File: /home/wxuser/PlatONE/test/conf/genesis.json
[INFO] [setup-genesis] : Genesis:
{
    "config": {
    "chainId": 300,
    "interpreter": "all",
    "istanbul": {
        "timeout": 10000,
        "period": 1,
        "policy": 0,
        "firstValidatorNode": "enode://a65dfc6ee888b517a3ad0983c1e213b4eb1156b37d58e3d767d6ccf968cab011d56beb15184baa9d7e39e449e4e55024acc7e10fc2f1b609f403b309430d63be@10.230.48.11:16801"
    }
  },
  "timestamp": "1627888607",
  "extraData": "0x00",
  "alloc": {
    "0xb65f89ad16E011dcbAb3905fa675083a2c551339": {
      "balance": "100000000000000000000"
    }
  }
}
[INFO] [setup-genesis] : Setup genesis succeeded
[INFO] [init] : Setup genesis file completed
[INFO] [init] : Get genesis file completed
[INFO] [init] : Setup firstnode info completed
[INFO] [init] : Sync firstnode info to node-0 completed
******************************************************************************************************************************************************************************
INFO [08-02|15:16:47.440] Maximum peer count                       ETH=50 LES=0 total=50 RoutineID=1
INFO [08-02|15:16:47.441] Allocated cache and file handles         database=/home/wxuser/PlatONE/test/data/node-0/platone/chaindata cache=16 handles=16 RoutineID=1
INFO [08-02|15:16:47.465] Persisted trie from memory database      nodes=13 size=2.35kB time=102.498µs gcnodes=0 gcsize=0.00B gctime=0s livenodes=1 livesize=0.00B RoutineID=1
INFO [08-02|15:16:47.465] Successfully wrote genesis state         database=chaindata                                               hash=17767e…e7f056 RoutineID=1
INFO [08-02|15:16:47.466] Allocated cache and file handles         database=/home/wxuser/PlatONE/test/data/node-0/platone/lightchaindata cache=16 handles=16 RoutineID=1
INFO [08-02|15:16:47.487] Persisted trie from memory database      nodes=13 size=2.35kB time=224.184µs gcnodes=0 gcsize=0.00B gctime=0s livenodes=1 livesize=0.00B RoutineID=1
INFO [08-02|15:16:47.489] Successfully wrote genesis state         database=lightchaindata                                               hash=17767e…e7f056 RoutineID=1
******************************************************************************************************************************************************************************
[INFO] [init] :Init genesis on node-0 completed
[INFO] [init] : Init node Node-0 completed

################ Init Node-1 ################
[INFO] [keygen] : ## Node-1 Keygen Start ##
[INFO] [keygen] : Files: /home/wxuser/PlatONE/test/data/node-1/node.address, /home/wxuser/PlatONE/test/data/node-1/node.prikey, /home/wxuser/PlatONE/test/data/node-1/node.pubkey
        Node-1's address: fC6d2Cf03244c689fE5487723Ef86819b07Eed50
        Node-1's private key: 6a598fe1cea9c3bf9fa9f4bbfe074ecd83e732439477867a7e6f302de59a89d7
        Node-1's public key: 1e6b4771d938cb14d2fa45b00560ffb4b16424001f91419f17963b5f044a6e05681fbd792d0cde3c9495723823925777757a564991ee1abdbbb97b398345a722
[INFO] [keygen] : Node-1 keygen succeeded
[INFO] [init] : Generate key for node-1 completed
[INFO] [init] : Send genesis file to node-1 completed
[INFO] [init] : Sync firstnode info to node-1 completed
******************************************************************************************************************************************************************************
INFO [08-02|15:16:47.945] Maximum peer count                       ETH=50 LES=0 total=50 RoutineID=1
INFO [08-02|15:16:47.946] Allocated cache and file handles         database=/home/wxuser/PlatONE/test/data/node-1/platone/chaindata cache=16 handles=16 RoutineID=1
INFO [08-02|15:16:47.970] Persisted trie from memory database      nodes=13 size=2.35kB time=109.355µs gcnodes=0 gcsize=0.00B gctime=0s livenodes=1 livesize=0.00B RoutineID=1
INFO [08-02|15:16:47.971] Successfully wrote genesis state         database=chaindata                                               hash=17767e…e7f056 RoutineID=1
INFO [08-02|15:16:47.971] Allocated cache and file handles         database=/home/wxuser/PlatONE/test/data/node-1/platone/lightchaindata cache=16 handles=16 RoutineID=1
INFO [08-02|15:16:47.993] Persisted trie from memory database      nodes=13 size=2.35kB time=299.071µs gcnodes=0 gcsize=0.00B gctime=0s livenodes=1 livesize=0.00B RoutineID=1
INFO [08-02|15:16:47.994] Successfully wrote genesis state         database=lightchaindata                                               hash=17767e…e7f056 RoutineID=1
******************************************************************************************************************************************************************************
[INFO] [init] :Init genesis on node-1 completed
[INFO] [init] : Init node Node-1 completed

################ Init Node-2 ################
[INFO] [keygen] : ## Node-2 Keygen Start ##
[INFO] [keygen] : Files: /home/wxuser/PlatONE/test/data/node-2/node.address, /home/wxuser/PlatONE/test/data/node-2/node.prikey, /home/wxuser/PlatONE/test/data/node-2/node.pubkey
        Node-2's address: e99b21B5ED4Cde288106b792015e8e361d1C0f57
        Node-2's private key: 0e567dbdc2997fda922b2fc5bd0dda01c3a623087e220a0f7a89dd9983d47b32
        Node-2's public key: ca1b7e7b5bcf37939327265dd114599162aafee524402e7db067849b806bdb547916bc718d4d5e6306fd45d8e9ded7c0ee07c8482303acaaf9b84da36a364432
[INFO] [keygen] : Node-2 keygen succeeded
[INFO] [init] : Generate key for node-2 completed
genesis.json                                                                                                                                                                         100%  514     0.5KB/s   00:00
[INFO] [init] : Send genesis file to node-2 completed
firstnode.info                                                                                                                                                                       100%   62     0.1KB/s   00:00
[INFO] [init] : Sync firstnode info to node-2 completed
******************************************************************************************************************************************************************************
INFO [08-02|15:16:50.367] Maximum peer count                       ETH=50 LES=0 total=50 RoutineID=1
INFO [08-02|15:16:50.368] Allocated cache and file handles         database=/home/wxuser/PlatONE/test/data/node-2/platone/chaindata cache=16 handles=16 RoutineID=1
INFO [08-02|15:16:50.384] Persisted trie from memory database      nodes=13 size=2.35kB time=104.763µs gcnodes=0 gcsize=0.00B gctime=0s livenodes=1 livesize=0.00B RoutineID=1
INFO [08-02|15:16:50.384] Successfully wrote genesis state         database=chaindata                                               hash=17767e…e7f056 RoutineID=1
INFO [08-02|15:16:50.384] Allocated cache and file handles         database=/home/wxuser/PlatONE/test/data/node-2/platone/lightchaindata cache=16 handles=16 RoutineID=1
INFO [08-02|15:16:50.405] Persisted trie from memory database      nodes=13 size=2.35kB time=137.845µs gcnodes=0 gcsize=0.00B gctime=0s livenodes=1 livesize=0.00B RoutineID=1
INFO [08-02|15:16:50.406] Successfully wrote genesis state         database=lightchaindata                                               hash=17767e…e7f056 RoutineID=1
******************************************************************************************************************************************************************************
[INFO] [init] :Init genesis on node-2 completed
[INFO] [init] : Init node Node-2 completed

################ Init Node-3 ################
[INFO] [keygen] : ## Node-3 Keygen Start ##
[INFO] [keygen] : Files: /home/wxuser/PlatONE/test/data/node-3/node.address, /home/wxuser/PlatONE/test/data/node-3/node.prikey, /home/wxuser/PlatONE/test/data/node-3/node.pubkey
        Node-3's address: Da43C3b691AAaA246760179932a6597b535b6B96
        Node-3's private key: 51459611a965f2e9234edfa7840004263cf2e43a72a0a04da747671f66ead8aa
        Node-3's public key: ec9493efdd059439db0a828552dabdc3ddc15db296a9360e9fca572041e05247deecbb6891dc49856e90495a54894c32e2173c56d6e0f5ff589e754674a94e30
[INFO] [keygen] : Node-3 keygen succeeded
[INFO] [init] : Generate key for node-3 completed
genesis.json                                                                                                                                                                         100%  514     0.5KB/s   00:00
[INFO] [init] : Send genesis file to node-3 completed
firstnode.info                                                                                                                                                                       100%   62     0.1KB/s   00:00
[INFO] [init] : Sync firstnode info to node-3 completed
******************************************************************************************************************************************************************************
INFO [08-02|15:16:52.971] Maximum peer count                       ETH=50 LES=0 total=50 RoutineID=1
INFO [08-02|15:16:52.971] Allocated cache and file handles         database=/home/wxuser/PlatONE/test/data/node-3/platone/chaindata cache=16 handles=16 RoutineID=1
INFO [08-02|15:16:52.986] Persisted trie from memory database      nodes=13 size=2.35kB time=137.118µs gcnodes=0 gcsize=0.00B gctime=0s livenodes=1 livesize=0.00B RoutineID=1
INFO [08-02|15:16:52.987] Successfully wrote genesis state         database=chaindata                                               hash=17767e…e7f056 RoutineID=1
INFO [08-02|15:16:52.987] Allocated cache and file handles         database=/home/wxuser/PlatONE/test/data/node-3/platone/lightchaindata cache=16 handles=16 RoutineID=1
INFO [08-02|15:16:53.007] Persisted trie from memory database      nodes=13 size=2.35kB time=134.019µs gcnodes=0 gcsize=0.00B gctime=0s livenodes=1 livesize=0.00B RoutineID=1
INFO [08-02|15:16:53.008] Successfully wrote genesis state         database=lightchaindata                                               hash=17767e…e7f056 RoutineID=1
******************************************************************************************************************************************************************************
[INFO] [init] :Init genesis on node-3 completed
[INFO] [init] : Init node Node-3 completed

[INFO] [init] : Init completed
```

## start.sh 启动脚本

### 功能

- 对节点进行启动，包括启动节点、部署系统合约、将节点加入区块链以及将节点更新为共识节点
- 支持指定节点启动
- 具有日志机制，checkpoint设在创建keyfile、部署系统合约、keyfile相关文件同步、添加节点至区块链以及更新节点为共识节点
- 根据地址判断是否为远程活动，执行对应的远程登录/传输命令

### 命令

```shell
"
USAGE: start.sh  [options] [value]

        OPTIONS:

           --project, -p              the specified project name. must be specified

           --node, -n                 the specified node name. only used in conf mode.
                                      default='all': deploy all nodes by conf in deployment_conf
                                      use ',' to seperate the name of node

           --help, -h                 show help
"
```

- --project, -p:
  - 项目名称，属于必填项
  - 项目路径: XXX/platone_deploy/deployment_conf/${project_name}
- --node, -n:
  - 节点名称，不填默认为node=all
  - 脚本会根据项目路径下的deploy_node-${node_name}.conf配置文件进行启动操作
  - 如果设置值为all，那么脚本会根据项目路径下所有配置文件向对应的节点依次进行启动。只有设置为“all”时，才会启动firstnode。如果系统中还没有节点启动，需要使用“all”来一键启动
  - 如果设置多个指定节点，那么需要用","作为不同的节点名称的分隔符。只有firstnode已启动并且在区块链中成为共识节点后，才可以使用指定节点的方式启动其他节点。

### 使用演示（按序号依次执行）

#### 1 指定节点启动

模拟情况：还没有完成firstnode启动

##### 命令输入

```shell
./start.sh -p test -n 1
```

##### 打印输出

```
###########################################
####            start  nodes           ####
###########################################
[ERROR] [start] : ********* FIRSTNODE NOT STARTED **********
```

#### 2 全节点启动

模拟情况：2个节点为部署文件所在的本机（wxuser@10.230.48.11），另2个节点为远程机

##### 命令输入

```shell
./start.sh -p test
```

##### 打印输出

```
###########################################
####            start  nodes           ####
###########################################

################ Start first node Node-0 ################
[INFO] [start-node] : ## Run node-0 ##
[INFO] [start-node] : Node's url: 10.230.48.11:6801
[INFO] [start-node] : Run node-0 succeeded
[INFO] [start] : Run node Node-0 completed
[INFO] [create-account] : ## Create account Start ##
[WARN] [create-account] : !!! An account will be created. The default password is 0 !!!
[INFO] [create-account] : Account:
        New account: 0x1287e1f77a1397a759a606f395038098f0609195
        Passphrase: 0
[INFO] [create-account] : Create account succeeded
[INFO] [start] : Create account completed
[INFO] [start] : Get keyfile completed
[INFO] [deploy-system-contract] : ## Deploy System Contract Start ##
[INFO] [deploy-system-contract] : Set Node-0 as super admin completed
[INFO] [deploy-system-contract] : Set Node-0 as chain admin completed
[INFO] [deploy-system-contract] : Deploy system contract completed
[INFO] [start] : Deploy system contract completed
[INFO] [add-node] : ## Add Node-0 Start ##
[INFO] [add-node] : Add Node-0 succeeded
[INFO] [start] : Add node node-0 completed
[INFO] [update-to-consensus-node] : ## Update Node-0 To Consensus Node Start ##
[INFO] [update-to-consensus-node] : Update Node-0 to consensus node succeeded
[INFO] [start] : Update node node-0 to consensus node completed
[INFO] [start] : Start firstnode Node-0 completed

################ Start Node-1 ################
[INFO] [start] : Send keyfile to Node-1 completed
[INFO] [start-node] : ## Run node-1 ##
[INFO] [start-node] : Node's url: 10.230.48.11:6802
[INFO] [start-node] : Run node-1 succeeded
[INFO] [start] : Run node Node-1 completed
[INFO] [add-node] : ## Add Node-1 Start ##
[INFO] [add-node] : Add Node-1 succeeded
[INFO] [start] : Add node node-1 completed
[INFO] [update-to-consensus-node] : ## Update Node-1 To Consensus Node Start ##
[INFO] [update-to-consensus-node] : Update Node-1 to consensus node succeeded
[INFO] [start] : Update node node-1 to consensus node completed
[INFO] [start] : Start node Node-1 completed

################ Start Node-2 ################
keyfile.json                                                                                                                                                                         100%  491     0.5KB/s   00:00
keyfile.phrase                                                                                                                                                                       100%    2     0.0KB/s   00:00
[INFO] [start] : Send keyfile to Node-2 completed
[INFO] [start-node] : ## Run node-2 ##
[INFO] [start-node] : Node's url: 10.230.48.12:6793
[INFO] [start-node] : Run node-2 succeeded
[INFO] [start] : Run node Node-2 completed
[INFO] [add-node] : ## Add Node-2 Start ##
[INFO] [add-node] : Add Node-2 succeeded
[INFO] [start] : Add node node-2 completed
[INFO] [update-to-consensus-node] : ## Update Node-2 To Consensus Node Start ##
[INFO] [update-to-consensus-node] : Update Node-2 to consensus node succeeded
[INFO] [start] : Update node node-2 to consensus node completed
[INFO] [start] : Start node Node-2 completed

################ Start Node-3 ################
keyfile.json                                                                                                                                                                         100%  491     0.5KB/s   00:00
keyfile.phrase                                                                                                                                                                       100%    2     0.0KB/s   00:00
[INFO] [start] : Send keyfile to Node-3 completed
[INFO] [start-node] : ## Run node-3 ##
[INFO] [start-node] : Node's url: 10.230.48.12:6794
[INFO] [start-node] : Run node-3 succeeded
[INFO] [start] : Run node Node-3 completed
[INFO] [add-node] : ## Add Node-3 Start ##
[INFO] [add-node] : Add Node-3 succeeded
[INFO] [start] : Add node node-3 completed
[INFO] [update-to-consensus-node] : ## Update Node-3 To Consensus Node Start ##
[INFO] [update-to-consensus-node] : Update Node-3 to consensus node succeeded
[INFO] [start] : Update node node-3 to consensus node completed
[INFO] [start] : Start node Node-3 completed

[INFO] [start] : Start completed
```

## local-keygen.sh 本地生成密钥脚本

### 功能

- 用于在节点本地生成密钥

### 使用

```shell
"
USAGE: local-keygen.sh  [options] [value]

        OPTIONS:

           --node, -n                   the specified node name. must be specified

           --help, -h                   show help
"
```

- --node, -n:
  - 节点名称，属于必填项
  - 脚本会根据node_name获取node_dir路径，来存放地址、公钥以及私钥
  - 如果已经存在密钥信息，会自动备份移除
  - 只支持单节点操作

### 使用演示（本地）

##### 命令输入

```shell
./local-keygen.sh -n 1
```

##### 打印输出

```
[INFO] [keygen] : ## Node-1 Keygen Start ##
[INFO] [keygen] : Backup /home/wxuser/PlatONE/test_show/data/node-1/node.address completed
[INFO] [keygen] : Backup /home/wxuser/PlatONE/test_show/data/node-1/node.prikey succ
[INFO] [keygen] : Backup /home/wxuser/PlatONE/test_show/data/node-1/node.pubkey succ
[INFO] [keygen] : Files: /home/wxuser/PlatONE/test_show/data/node-1/node.address, /home/wxuser/PlatONE/test_show/data/node-1/node.prikey, /home/wxuser/PlatONE/test_show/data/node-1/node.pubkey
        Node-1's address: 8233e2CFcEfCD0C3E9c7677E997DA1f9C047a355
        Node-1's private key: 1c19c60bdde6abf3916a8c967f508c42b619205462d086f6482203a38d33db5d
        Node-1's public key: a0e3f48c8371e1099b7387778b1314ef17676627c934b885d1596e5f3f8bc98e0471baaf59cccac6e5c9c0ef51e892dd4cb55db158406b6ef925e1ca812a3412
[INFO] [keygen] : Node-1 keygen succeeded
```

## local-setup-genesis.sh 本地设置创世信息脚本

### 功能

- 生成用于节点初始化的genesis.json文件

### 使用

```shell
 "
USAGE: local-setup-genesis.sh  [options] [value]

        OPTIONS:

           --node, -n                   the specified node name. must be specified

           --help, -h                   show help
"
```

- --node, -n:
  - 节点名称，属于必填项
  - 脚本会根据deploy_node-${node_name}.conf配置文件获取节点的信息生成genesis.json
  - 如果已经存在genesis.json，会将自动备份移除
  - 只支持单节点操作

### 使用演示（本地）

##### 命令输入

```shell
./local-setup-genesis.sh -n 0
```

##### 打印输出

```
[INFO] [setup-genesis] : ## Setup Genesis Start ##
[INFO] [setup-genesis] : Backup /home/wxuser/PlatONE/test_show/conf/genesis.json completed
[INFO] [setup-genesis] : File: /home/wxuser/PlatONE/test_show/conf/genesis.json
[INFO] [setup-genesis] : Genesis:
{
    "config": {
    "chainId": 300,
    "interpreter": "all",
    "istanbul": {
        "timeout": 10000,
        "period": 1,
        "policy": 0,
        "firstValidatorNode": "enode://55bd98f0fd77b4da5844759e34345ed23d8bfe06332a90c6ad4b5451ce802a5a5b2fd76b472e40abf3dd5b0687fd9e8a584465570e090e960e3dfe23c3115ab0@10.230.48.11:16803"
    }
  },
  "timestamp": "1627890663",
  "extraData": "0x00",
  "alloc": {
    "0xb0BDBf7376665c8D4610B5342D6b785689682bdA": {
      "balance": "100000000000000000000"
    }
  }
}
```

## local-start-node.sh 本地运行节点脚本

### 功能

- 根据配置文件的信息，使用platone使节点运行起来
- 配置文件中如果配置了extra_options和pprof，这两项才会生效

### 命令

```shell
"
USAGE: local-start-node.sh  [options] [value]

        OPTIONS:

           --node, -n                   the specified node name. must be specified

           --help, -h                   show help
"
```

- --node, -n:
  - 节点名称，属于必填项
  - 脚本会根据deploy_node-${node_name}.conf配置文件运行节点
  - 只支持单节点操作

### 使用演示（本地）

##### 命令输入

```shell
./local-start-node.sh -n 1
```

##### 打印输出

```
[INFO] [start-node] : ## Run node-1 ##
[INFO] [start-node] : Node's url: 10.230.48.12:6795
[INFO] [start-node] : Run node-1 succeeded
```

## local-create-account 本地创建账户脚本

### 功能

- firstnode生成account，通过account生成keyfile.json
- 默认phrase为0
- 如果${node-dir}/keystore下已经有UTC文件了，则会直接清除并生成新的

### 命令

```shell
"
USAGE: local-create-account.sh  [options] [value]

        OPTIONS:

           --node, -n                   the specified node name. must be specified

           --help, -h                   show help
"
```

### 使用演示（本地）

##### 命令输入

```
./local-create-account.sh -n 0
```

##### 打印输出

```
[INFO] [create-account] : ## Create account Start ##
[WARN] [create-account] : !!! An account will be created. The default password is 0 !!!
[INFO] [create-account] : Account:
        New account: 0xc511fcbf891f54ee5b418655b8713786b48594b1
        Passphrase: 0
[INFO] [create-account] : Create account succeeded
```

## local-deploy-system-contract.sh 本地部署系统合约脚本

### 功能

- 将firstnode设置为超级管理员与链管理员

### 命令

```shell
"
USAGE: local-deploy-system-contract.sh  [options] [value]

        OPTIONS:

           --node, -n                   the specified node name. must be specified

           --help, -h                   show help
"
```

- --node, -n:
  - 节点名称，属于必填项
  - 脚本会根据node_name，赋予节点超级管理员和链管理员的权限
  - 只支持单节点操作

### 使用演示（本机）

##### 命令输入

```shell
./local-deploy-system-contract.sh -n 0
```

##### 打印输出

```
[INFO] [deploy-system-contract] : ## Deploy System Contract Start ##
[INFO] [deploy-system-contract] : Set Node-0 as super admin completed
[INFO] [deploy-system-contract] : Set Node-0 as chain admin completed
[INFO] [deploy-system-contract] : Deploy system contract succeeded
```

## local-add-node.sh 本地添加节点至区块链脚本

### 功能

- 通过platonecli将节点添加至区块链

### 命令

```shell
"
USAGE: local-add-node.sh  [options] [value]

        OPTIONS:

           --node, -n                   the specified node name. must be specified

           --help, -h                   show help
"
```

- --node, -n:
  - 节点名称，属于必填项
  - 脚本会根据node_name，将节点添加至区块链
  - 只支持单节点操作

### 使用演示（本机）

##### 命令输入

```shell
./local-add-node.sh -n 0
```

##### 打印输出

```
[INFO] [add-node] : ## Add Node-0 Start ##
[INFO] [add-node] : Add Node-0 succeeded
```

## local-update-node-to-consensus.sh 本地更新共识节点脚本

### 功能

- 通过platonecli将节点更新为共识节点

### 命令

```shell
"
USAGE: local-update-to-consensus-node.sh  [options] [value]

        OPTIONS:

           --node, -n                   the specified node name. must be specified

           --help, -h                   show help
"
```

- --node, -n:
  - 节点名称，属于必填项
  - 脚本会根据node_name，将节点更新为共识节点
  - 只支持单节点操作

### 使用演示（本机）

##### 命令输入

```shell
./local-update-node-to-consensus.sh -n 0
```

##### 打印输出

```
[INFO] [update-to-consensus-node] : ## Update Node-0 To Consensus Node Start ##
[INFO] [update-to-consensus-node] : Update Node-0 to consensus node succeeded
```

