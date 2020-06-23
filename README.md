# helm-api

**支持多集群、多命名空间的helm部署http接口**

## 安装
git clone https://github.com/OojaxoO/helm-api.git
cd helm-api
make
make install

## 配置
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
/opt/helm-api/helm-api