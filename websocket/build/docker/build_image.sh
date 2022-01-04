# 镜像版本
TAG=1.0
# 镜像私有仓库
REGISTRY=docker.hgpark.cn/yftest
# 镜像名称
NAME=duyong/wssvc
USERNAME=duyong
# 私有仓库镜像名称 必须是这种格式
IMAGE_NAME=$REGISTRY/$NAME

mkdir -p ./tmp
cp ../../_output/wssvc ./tmp/
cp -r ../../configs ./tmp/


# 删除latest
docker rmi -f $IMAGE_NAME

# 登录私有仓库
docker login --username=$USERNAME $REGISTRY

# 构建镜像
docker build   -t $IMAGE_NAME  .

# 推送历史版本
docker tag  $IMAGE_NAME $IMAGE_NAME:$TAG
docker push $IMAGE_NAME:$TAG

# 推送latest版本 后续覆盖
docker push $IMAGE_NAME

# 删除版本号镜像 留下latest
docker rmi -f $IMAGE_NAME:$TAG

rm -rf  ./tmp
