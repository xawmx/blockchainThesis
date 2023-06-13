# 文件

```
chaincode/demo:       测试 chaincode
chaincode/callback:   hyperleger/caliper测试用例

pbft:                 可插拔 PBFT 共识算法简单实现
rbft:                 可插拔 PBFT-Byz 共识算法简单实现
vpbft:                 可插拔 VPBFT 共识算法简单实现

solo-network:          solo共识配置
pbft-network:          pbft共识配置 
rbft-network:          rbft共识配置
vpbft-network:         vpbft共识配置
multi-channel-network: solo多链配置
```

# 链码

| 函数 |       功能       |    参数    |
| :-------: | :--------------: | :--------------------: |
|  open  | 开户 | 账户名, 金额 |
|  query  | 查询 | 账户名 |
|  delete  | 销户 | 账户名 |

# 编译

编译 pbft 说明文件：`./pbft/doc.md` 

编译 rbft 说明文件：`./rbft/doc.md`

编译 vpbft 说明文件：`./vpbft/doc.md`

# 测试

```
$ npx caliper launch master --caliper-workspace <pbft或rbft或vpbft-network> --caliper-benchconfig benchmarks/config.yaml --caliper-networkconfig benchmarks/network.yaml
```



