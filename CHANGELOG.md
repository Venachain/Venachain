# Changelog

## [1.1.2]
### Features
* [contract] 预编译合约添加paillier合约 --饶应典
### Improvements
* [chain] 优化 setState without getCommittedState --张玉坚
* [vcl] vcl生成funcArgs时，将去除param全部空格改为仅去除param的首尾空格 --吴经文
* [scripts] venachainctl unlock支持解锁指定account --吴经文
* [scripts] 完成1.1脚本遗留的一些优化点 --吴经文

## [1.1.1]

### Bug Fixes
* [consensus] 修复共识层执行区块的状态校验和交易重放检查问题 --陈明晶

### Improvements
* [deploy] 部署脚本逻辑与代码优化 --吴经文

## [1.1.0]
### Features
* [contract] 增加存证预编译合约 --李京京
* [chain] 增加DAG信息，可以并行执行交易 --张玉坚
* [other] 轻节点功能完善 --陈明晶
* [storage] 存储插件化，新增可选存储引擎PebbleDB --曾梦露
* [other] Venachain支持节点License，对节点可用性进行控制 --杜代栋
* [chain] 区块打包交易个数自动控制 --张伟
* [deploy] 重构部署脚本支持多机一键部署 --吴经文
* [contract] 预编译合约添加bulletproof合约 --陈炫慧

### Improvements
* [other] 节点间同步的交易可并行进行验证 --陈炫慧
* [download] 区块同步优化重复同步 --陈明晶
* [evm] 合约支持SM3 --李京京
* [rpc] namespace增加venachain --张玉坚
* [chain] 交易池优化-增加各个阶段的交易池预处理 --陈明晶
* [chain] 并行处理添加参数用于是否开启 --张玉坚
* [contract] 优化存证预编译合约 --陈炫慧
* [chain] 调整区块缓存个数 --陈明晶
* [scripts] 添加设置交易池最大交易个数 --吴经文
* [deploy] 脚本不再使用repstr二进制文件，使用sed指令替代 --吴经文

### Bug Fixes
* [evm] 合约运行中的错误信息可以正常返回 --张玉坚
* [chain] badBlock问题修复 --张玉坚
* [chain] 修复并行处理时，因资源竞争失败后的依赖关系处理 --张玉坚
* [vm] 修复通过CNS调用合约防火墙未生效问题 --陈明晶
* [vm] 修复在VM防火墙验证相关logs未生效或错误问题 --陈明晶
* [node] 修复新增节点的类型可以设置成共识节点的问题 --陈炫慧
* [vm] 修复CNS预编译合约中获取合约地址失败时返回值不正确的问题 --崔璨
* [console] 修复venachain namespace缺少方法问题 --陈明晶
* [build] 修复编译时，因为release/linux下没有bin目录而报错 --吴经文
* [build] 修复项目第一次make clean时，因为life/resolver/sig/openssl下没有Makefile文件而报错 --吴经文
* [deploy] 修复远程部署或容器化部署在通过ifconfig判断ip地址是否是本机时，出现误判的bug --吴经文

## [1.0.0.0.0]
### Breaking Changes
* [system contract] 系统合约重构成预编译合约形式
* [other] 删除eip，DAO等版本升级的Hard Fork和兼容性检查;
* [other] 删除Rinkeby，Testnet;删除ChainConfig的EmptyBlock设置;删除Clique；删除difficulty；删除dev模式；
* [other] 删除默认配置,并重写了genesis初始化逻辑。
* [other] 交易处理生命周期全流程优化  - 张玉坚
* [p2p] 修改protocol协议为venachainV1，交易广播添加hash广播  - 张玉坚
* [other] 交易结构中删除了txtype字段

### Features
* [chain] 添加一链多账本功能（群组预编译化系统合约等等）
* [chain] 支持VRF共识机制 - 于宗坤
* [chain] venachain命令行工具支持通过replay区块的方式完成非兼容性升级 - 于宗坤
* [chain] 增加隐私token的功能 --王琪
* [contract]　增加范围证明[range proof]的验证功能  --王琪
* [chain] 链bp参数修改，与之前版本兼容 --王琪
* [other] 可视化运维平台
* [other] 新的链交互工具 vcl

### Improvements
* [contract] 系统合约缓存机制重构。
* [contract] 禁止利用CNS调用系统预编译合约
* [scripts] 启动脚本添加设置区块最大交易数 --张玉坚
* [other] 版本管理采用mod模型 - 汤勇，于宗坤，杜满想
* [other] 删除whisper,swarm,mobile,cmd/wnode（Whisper node）- 杜满想
* [other] 删除pow相关逻辑(reorg,sidechain),删除cbft - 杜满想

### Bug Fixes
* [chain] 修复bad block问题 --张玉坚
* [p2p] peer的id使用 publicKey[:16]
* [evm] evm合约部署可以配置参数
* [other] 修复cmd工具对VRF参数的修改

## [0.9.12] 2020-08-25
### Breaking Changes
### Improvements
* [chain] genesis时间戳自动设置为当前系统时间 --葛鑫

### Features
### Bug Fixes

## [0.9.11] 2020-06-04
### Breaking Changes
### Improvements
### Features
* [chain] WASM虚拟机对大浮点数和大整数的支持 --于宗坤
* [other] Venachain-CDT

### Bug Fixes


## [0.9.10] 2020-05-07
### Breaking Changes
### Improvements
### Features
* [chain] 交易处理執行時gasPrice寫入receipt功能。 --潘晨

### Bug Fixes
* [chain] import功能bug修复。  --葛鑫
* [chain] 解决Venachain终端log输出时，日志等级设置失效问题。 --汤勇

## [0.9.9]
### Breaking Changes
### Improvements
* [chain] 在共识模块中直接同步写入区块，以提高区块链交易处理性能。 --葛鑫
### Features
* [chain] 交易处理流程引入根据交易消耗gas扣除用户特定token的功能。 --潘晨
### Bug Fixes

## [0.9.8]
### Breaking Changes
### Features
* [contract] Wasm合约支持float型计算 -- 王琪，王忠莉，朱冰心，潘晨，吴启迪，杜满想

### Bug Fixes
* [ctool] 返回值是uint32类型时无法解析。 -- 杜满想
* [contract] Wasm合约无法打印uint64类型的变量。 -- 王琪，杜满想
* [contract]　修复sm2验签某些公钥解析失败的bug。--潘晨

## [0.9.7] -- 2020-01-16
### Bug Fixes
* [contract]　修改secp256k1验签功能所用的hash函數。--潘晨

## [0.9.6] -- 2020-01-15
### Features
* [contract]　增加secp256k1和r1的验签功能。--潘晨

### Bug Fixes
* [chain] txpool先去重再验签。 -- 葛鑫

## [0.9.5] -- 2020-01-03
### Bug Fixes
* [chain] 修复第一个节点数据清空无法再加入网络的问题。--汤勇
* [chain] 共识模块中共识结束直接写入区块数据，可能会造成并发问题，修改为由p2p.fetcher异步写入。 --葛鑫 
* [chain] 共识消息处理中，投票类消息单独用一个Event Channel消息处理。 -- 葛鑫

## [0.9.4] -- 2020-12-20
### Bug Fixes
* [contract] 修复合约调用ecrecover时，若签名无效，则虚拟机执行失败的问题，现改为返回nil。--潘晨

## [0.9.3] -- 2019-12-11
### Bug Fixes
* [chain] 并发访问所有链接的节点map集合时，出现并发读写错误，导致节点宕机。--汤勇

## [0.9.2] -- 2019-12-06
### Bug Fixes
* [chain] 区块执行时间过长时，共识无法正常工作，不能继续出块。--葛鑫

## [0.9.1] -- 2019-11-22
### Features
* [contract] 调用一个没有在CNS中注册的合约时报错，receipt的status设为false。-- 简海波，葛鑫

### Improvements
* [contract] 简化了wasm与solidity兼容调用方式。-- 汤勇
* [chain] 添加对版本的支持，`./venachain  --version`可以打印当前版本。 -- 葛鑫
* [contract] 删除sm密码库的静态库文件，改为用源码编译，方便为以后的跨平台做准备。 -- 潘晨
* [chain] 暂时注释掉VC（verifiable computation，可验证计算）和nizkpail相关代码为跨平台做准备。-- 杜满想

### Bug Fixes
* [node] 以前在节点管理合约中, 删除节点后, 此节点就无法恢复正常状态.  目前 是支持可以更改为正常状态了。 -- 汤勇
* [chain] 各个节点时间不一致情况下搭链，搭链成功后向cnsManager合约注册一个新合约，会导致各节点状态不一致。  -- 葛鑫
* [contract] 合约中Event参数不支持int类型。 -- 葛鑫
* [contract] 合约中调用assert方法无法打印。-- 黄赛杰
