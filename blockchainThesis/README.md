# 基于智能合约的输配电行业生产制造流程监管平台设计

```
blockchain-trace-bcnetwork：区块链网络，可直接将文件上传至服务器，然后启动里面的脚本

blockchain-trace-pc：PC端,使用的是RuoYi-Vue

blockchain-trace-basic-data：系统基础数据后台，使用的是RuoYi

前端：Vue.js , Element UI , mpvue

后端：SpringBoot , Mybatis , FastDFS , Node.js , Redis , MySQL

区块链：Fabric1.2

智能合约：Golang

环境：Ubuntu16.04 64位(建议8核 8G以上，2G也能运行)，Docker,  Docker-compose 
```

# 安装教程

## 一.fabric网络（blockchain-trace-bcnetwork）

### 1.确保环境配置好

> node.js 12

> docker

> docker-compose

> Redis

> FastDFS

> Mysql8

> go语言环境

### 2.上传代码到linux服务器

```
blockchain-trace-bcnetwork
```

### 3.启动平台相关服务

#### 启动fastdfs

##### 开启tracker 服务

```
docker run -dti --network=host --name tracker -v /var/fdfs/tracker:/var/fdfs delron/fastdfs tracker
```

##### 开启storage服务

```
docker run -dti --network=host --name storage -e TRACKER_SERVER=172.18.232.158:22122 -v /var/fdfs/storage:/var/fdfs delron/fastdfs storage
```

### 4.运行basic_network目录下的start.sh文件

#### 启动节点

```
chmod -R 777 start.sh 
./start.sh
```

### 5.运行webapp目录下的./start.sh

#### 启动合约

先给webapp目录下的所有sh文件授权，举个例子

```
chmod -R 777 startDriverCC.sh  
./start.sh
```

### 6.执行npm install安装依赖

```
npm install
```

### 7.注册用户

```
node enrollAdmin.js
node registerUser.js
```

执行node registerUser.js可能会安装失败，删除一下hfc-key-store后重新执行，如果还是失败，可能就是npm install出问题，注意node版本，使用node12

### 8.启动node服务(node服务就是一个中间件，连接前端和fabric网络)

```
./node-start.sh 
```

到这里，联盟链网络部署完成

## 二.后台服务部署（blockchain-trace-basic-data）

#### docker启动redis与mysql并挂载到服务器本地目录

##### redis 

参考https://blog.csdn.net/weixin_45821811/article/details/116211724

```
一定要注意安装的redis与官网下载的redis.conf是同一版本，否则无法启动
docker run -p 6379:6379 --privileged=true -d -v /edgex/uestc/redisData/data:/data -v /edgex/uestc/redisData/conf/redis.conf:/etc/redis/redis.conf --name redis redis:5.0.14 redis-server /etc/redis/redis.conf --appendonly yes
```

##### mysql

```
docker run -itd --name mysql  -p 3300:3306 -v /edgex/uestc/mysqlData/data/:/var/lib/mysql -v /edgex/uestc/mysqlData/conf/:/etc/mysql/conf.d -e MYSQL_ROOT_PASSWORD=123456 mysql:5.7
```

#### 1.修改application.yml文件中的Redis地址和fastdfs地址

#### 2.修改application-druid.yml文件中mysql地址

本地启动springboot项目，也可直接将\blockchain-trace-basic-data\ruoyi-admin\target下打包好的ruoyi-admin.jar包上传至linux服务器，运行以下命令进行后台服务的启停

```
sudo ./basic.sh start
sudo ./basic.sh stop
sudo ./basic.sh restart
sudo ./basic.sh status
```

## 三.前端服务部署（blockchain-trace-pc）

### 1.安装依赖

> npm install --registry=[https://registry.npm.taobao.org](https://registry.npm.taobao.org/)

### 2.修改连接区块链网络地址

main.js，ip地址修改为区块链网络所在服务器地址

```
Vue.prototype.$httpUrl = "http://localhost:8080/route";
```

### 3.启动项目

```
./console-start.sh 
```

以上，整个平台部署完成。

![image-20230412164125162](D:\study\blockchainThesis\README.assets\image-20230412164125162.png)