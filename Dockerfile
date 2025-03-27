FROM golang:1.23-alpine
LABEL authors="Vdebu"

#设置国内的镜像源用于下载依赖
ENV GOPROXY=https://goproxy.cn,direct \
    GOSUMDB=sum.golang.google.cn
#设置当前的工作目录
WORKDIR /

#复制项目的依赖列表
COPY go.mod go.sum ./

#下载所有的依赖项 由于这里有vendor所以可以不用额外下载额外的依赖项
#RUN go mod download

#将项目的所有文件拷贝到工作目录下
COPY . .

#在工作目录下构建项目(编译api二进制文件)
RUN go build -o api ./cmd/api

#声明服务器运行时暴露的端口(只是声明不会直接进行端口映射起文档化与提示作用)
EXPOSE 3939

#运行程序
CMD ["sh", "-c", "echo 'Starting backend...' && ./api"]