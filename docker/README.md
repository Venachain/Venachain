
docker编译部署环境
------


## docker 安装

docker安装 [http://get.daocloud.io/#install-docker-for-mac-windows](http://get.daocloud.io/#install-docker-for-mac-windows)


docker-hub国内源设置 [https://www.daocloud.io/mirror#accelerator-doc](https://www.daocloud.io/mirror#accelerator-doc)

## 构建docker镜像

```shell
cd Venachain/docker

docker build -t venachain:dev .
```

## 启动容器

```shell
export PathToVenachain=/home/gexin/Venachain
docker run -itd -p 6791:6791 16791:16791 26791:26791 -v ${PathToVenachain}:/Venachain --name venachain venachain:dev /bin/bash
```

## 进入容器


```shell
docker exec -it venachain /bin/bash
```

## 编译或者搭链

```
cd /Venachain/
make clean & make all

cd release/linux/script/
./venachainctl.sh one
```