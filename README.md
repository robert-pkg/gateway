### gateway 

gateway is based on micro-go micro-service framework.

### 服务命名

分类 | 命名规则
---|---
内部服务    | community.service.服务名
对外的服务  | community.interface.服务名

网关只分发 community.interface 开头的服务

### 使用到的 http状态码

状态码 | 说明
---|---
200 | OK 请求成功
400 | BadRequest 客户端请求的语法错误 （http header的信息不正确）
401 | Unauthorized 令牌验证不通过
500 | Internal Server Error 服务器内部错误，无法完成请求

### 传输

网关接收http请求，分发到具体服务。 
- 传输通道：grpc
- 数据：json

### example

具体参加 github.com/robert-pkg/XXX4Go/tools/py
