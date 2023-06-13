一些常用软件的安装命令整理

============GO===============
#下载二进制包
wget https://golang.google.cn/dl/go1.18.linux-amd64.tar.gz
#将下载的二进制包解压至 /usr/local目录
sudo tar -C /usr/local -xzf go1.18.linux-amd64.tar.gz
mkdir $HOME/go
#将以下内容添加至环境变量 ~/.bashrc
export GOPATH=$HOME/go
export GOROOT=/usr/local/go
export PATH=$GOROOT/bin:$PATH
#更新环境变量
source  ~/.bashrc 
#设置代理
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct

============NVM==============
#下载nvm安装脚本
wget https://gitee.com/real__cool/fabric_install/raw/main/nvminstall.sh
#安装nvm；安装屏幕输出内容添加环境变量
./nvminstall.sh
#换源
npm config set registry https://registry.npm.taobao.org

===========Fabric=============
#下载Fabric网络组件
wget https://gitee.com/real__cool/fabric_install/raw/main/bootstrap.sh
#给脚本添加可执行权限
chmod +x bootstrap.sh
#下载Fabric相关材料 后边两个版本号分别是Fabric、CA的版本号
./bootstrap.sh 1.4.4 1.4.4
#启动BYFN网络，最后输出END则代表环境搭建测试成功
cd fabric-samples/first-network/
./byfn.sh up

==========DOCKER============
#下载docker
curl -fsSL https://get.docker.com | bash -s docker --mirror Aliyun
#下载docker-compose
sudo curl -L "https://ghproxy.com/https://github.com/docker/compose/releases/download/v2.2.2/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
#给docker compose加上权限
sudo chmod +x /usr/local/bin/docker-compose
#添加当前用户到docker用户组
sudo usermod -aG docker $USER
newgrp docker
#向/etc/docker/daemon.json写入docker 镜像源
{
    "registry-mirrors": ["https://punulfd2.mirror.aliyuncs.com"]
}
#重启docker
sudo systemctl restart docker