# 联盟链网络测试

# 1.Caliper 测试共识算法在fabric的性能（linux）

### 测试数据

| 节点数（n=2i）                        | 4    | 8    | 16   | 32   | 64   | 128  | 256  | 512  |
| ------------------------------------- | ---- | ---- | ---- | ---- | ---- | ---- | ---- | ---- |
| 对应满足n=3f+1条件节点数n             | 4    | 7    | 16   | 31   | 64   | 127  | 256  | 511  |
| 对应拜占庭节点数f                     | 1    | 2    | 5    | 10   | 21   | 42   | 85   | 170  |
| 网络中节点总数，peer为5（order+peer） | 9    | 12   | 21   | 36   | 69   | 132  | 261  | 516  |

### 安装caliper依赖

#安装caliper-cli

```sql
npm install -g --only=prod @hyperledger/caliper-cli@0.3.2
```

#绑定sut，这里出错了就重试几次

```sql
caliper bind --caliper-bind-sut fabric:1.4.4 --caliper-bind-args=-g
```

#下载caliper-core

```sql
cd /fabric-sample/chaincode/demo/callback
npm install @hyperledger/caliper-core@0.3.2
```

#修改脚本，参数为节点数量

```sql
./bench.sh 7
./docker.sh 7
```

#caliper测试

```sql
sudo ./scripts/utils.sh down && npx caliper launch master --caliper-workspace ./ --caliper-benchconfig benchmarks/config.yaml --caliper-networkconfig benchmarks/network.yaml
```

#输出的report.html 就是本次测试的报告，可以在浏览器中查看

#将报告下载至Windows

```sql
apt install lrzsz
sz report.html
```

# 2.使用Apache JMeter做平台接口性能测试详解（Windows）

由于JMeter是Java写的，所以需要提前安装好Java开发环境
解压下载的安装包后，直接双击目录[apache](https://so.csdn.net/so/search?q=apache&spm=1001.2101.3001.7020)-jmeter-5.4.1\bin下**jmeter.bat**启动jmeter

### 1. 添加线程组

右击"Test Plan"—>Add—>Threads(Users)—>Thread Group

新建的线程组如下:

Number of Threads(users)：虚拟用户数，默认为1，表示模拟多少个虚拟用户访问测试的接口/系统

Ramp-up period(seconds)：虚拟用户增长时长，默认为1，表示模拟多长时间内测试完接口/系统

Loop Count：循环次数，默认为1，表示一个虚拟用户做多少次测试

### 2. 添加测试请求

右击新建的"Thread Group"—>Add—>Sampler—>Http Request

新建测试请求

### 3. 添加结果树

右击"Http Request—>Add—>Listener—>View Results Tree

添加结果树后，点击启动按钮就可以测试了

结果返回中有中文时，会出现乱码，解决办法如下：

右击请求(Http Request)—>Add—>Post Processors—>BeanShell PostProcessor

添加BeanShell后，在Script脚本区输入脚本prev.setDataEncoding(“UTF-8”)

### 4. 添加报告统计

右击请求(Http Request)—>Add—>Listener—>Summary Report

发送请求后，可以查看请求报告，例如：请求次数、错误率等

### 5. 添加请求头管理器

在实际测试中，经常会在请求上加一些请求头参数

右击线程组(Thread Group)—>Add—>Config Element—>Http Header Manager

新建请求头管理器后，可以添加相应请求头参数

### 6.post请求头JSON格式配置

因为请求参数的格式为JSON格式，所以需要在请求中添加配置元件（Config Element）--> HTTP信息头管理器（HTTP Header Manager）

在HTTP信息头中说明发送的请求参数格式为Json格式。

Content-Type:application/json;charset=UTF-8