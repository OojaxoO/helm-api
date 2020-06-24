# helm-api

**支持多集群、多命名空间的helm部署http接口**

## 安装
git clone https://github.com/OojaxoO/helm-api.git  
cd helm-api  
make  
make install  

## 配置
vim /etc/profile  
export GO111MODULE=on  
GOPROXY=https://goproxy.io  
export GOPROXY  

source /etc/profile  

vim /opt/helm-api/conf/app.ini  
[database]  
Type = mysql  
User = root  
Password = test123  
Host = 127.0.0.1:3306  
Name = nvwa  
TablePrefix = kube_  
  
[http]  
Port = 10000  

## 运行
cd /opt/helm-api/
./helm-api  

## 使用
1. helm list -n namespace  
GET /charts/:cluster/:namespace/

2. helm get all -n namespace name  
GET /charts/:cluster/:namespace/:name

3. helm upgrade --install name chart -n namespace  
POST /charts/:cluster/:namespace/:name?chart=:chart

4. helm delete name -n namespace  
DELETE /charts/:cluster/:namespace/:name


