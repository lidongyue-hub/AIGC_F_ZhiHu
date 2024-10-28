# 
一个类似知乎的问答系统（服务端）

- [x] 用户注册
- [x] 用户登录
- [x] 查看个人信息
- [x] 查看自己发布的问题、回答以及点赞记录

---

- [x] 发布问题
- [x] 查看问题
- [x] 修改问题
- [x] 删除问题
- [x] 首页推荐
- [x] 问题热榜

---

- [x] 回答问题
- [x] 查看回答
- [x] 修改回答
- [x] 删除回答
- [x] 回答列表
- [x] 回答点赞点踩

## 环境依赖

- [Gin](https://github.com/gin-gonic/gin): 轻量级Web框架
- [GORM](http://gorm.io/docs/index.html): ORM工具，本项目需要配合MySQL使用
- [Go-Redis](https://github.com/go-redis/redis): Golang Redis客户端，用于缓存相关功能
- [godotenv](https://github.com/joho/godotenv): 开发环境下的环境变量工具，方便配置环境变量
- [Jwt-Go](https://github.com/dgrijalva/jwt-go): Golang JWT组件，本项目使用基于 jwt 实现的 token 来做身份验证
- [crypto](https://pkg.go.dev/golang.org/x/crypto): Golang 加密算法库，本项目使用其中的 bcrypto 算法来加密用户密码
- [cron](https://github.com/robfig/cron): Golang 定时任务库，用于 Redis 缓存同步

## 目录结构

```
├── api              API控制层，负责处理请求
│   ├── v1           具体API版本
├── cache            redis 缓存相关
├── conf             项目的静态配置
├── cron             定时任务
├── middleware       中间件
├── model            数据库模型以及相关操作
├── routes           路由配置
├── serializer       将实体映射成不同的viewmodel，以及常用的响应信息
├── service          将比较复杂的业务从API层分离出来
| main.go            项目入口
```


### 1. 直接运行

go版本1.14、go module、执行 `go run main.go` 

### 2. docker部署

用 docker 部署本项目， 用Dockerfile 和 docker-compose.yml



![image](https://github.com/user-attachments/assets/560016de-a186-4399-97a7-2b8122a692a1)


