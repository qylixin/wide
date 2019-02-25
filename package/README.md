# 1 配置说明：

## 1.1 将docker-compose中commond启动命令中的“-ip=”和“-channel=”的值都改为本机器的外网ip

## 1.2 ./conf/wide.json中的“BcapAddress”中的地址和端口修改为wischain后套bcap-apiserver服务实际的地址

# 2 启动说明(在wide目录下执行)：

## 2.1 启动命令：docker-compose up -d

## 2.2 停止命令：docker-compose down