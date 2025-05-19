# wecom-webhook-push-mail

本项目兼容企业微信机器人 webhook 接口和消息格式， 接收到的消息内容自动转发到指定邮箱。

## 功能特性

- 支持企业微信机器人 Webhook 消息格式（text/markdown/news/image）
- 自定义邮件服务商，可按需配置（QQ/163/自建邮局）
- 所有认证信息通过 Base64 编码存储在 URL 中
- 服务器不存储任何用户信息
- Go 单文件实现，依赖极少，易于维护

---

## 注意事项

🚨 **部署服务前需先确认可连接目标邮件服务商 SMTP 端口**

## 快速部署

### 一、二进制部署

#### 1. 下载 Release 包

前往 [Releases 页面](https://github.com/MxxApp/wecom-webhook-push-mail/releases) 下载对应架构的二进制包，解压得到 `wwpm` 和 `config.toml`。

#### 2. 编辑配置文件

参考 `config.toml` 示例：

```toml
# 服务端口
port = 8080

# 邮件服务商
[providers.163]
email_suffix = "@163.com"
smtp_host = "smtp.163.com"
smtp_port = 465

[providers.qq]
email_suffix = "@qq.com"
smtp_host = "smtp.qq.com"
smtp_port = 465
```

#### 3. 启动服务

```sh
./wwpm -config=config.toml
# 或指定自定义配置文件路径
./wwpm -config=/path/to/your_config.toml
```

---

### 二、Docker 部署

#### 1. 拉取镜像

支持多架构（x86_64/arm64），可从 GitHub Packages 拉取：

```sh
docker pull ghcr.io/mxxapp/wwpm:latest
# 或指定版本
docker pull ghcr.io/mxxapp/wwpm:v1.0.0
```

#### 2. 运行容器

- **挂载配置**（推荐，方便更新配置）：

```sh
docker run -d \
  -v /your/local/config.toml:/app/config.toml \
  -p 8080:8080 \
  ghcr.io/mxxapp/wwpm:latest
```

- **自定义配置路径**：

```sh
docker run -d \
  -v /your/local/myconfig.toml:/app/myconfig.toml \
  -p 8080:8080 \
  ghcr.io/mxxapp/wwpm:latest \
  -config=/app/myconfig.toml
```

---

## 消息推送用法

向 `/cgi-bin/webhook/send?key=xxxx` 发 POST 请求，Body 为企业微信机器人兼容 JSON，  
`key` 为 `服务商|用户名|密码` Base64 编码，示例：

- 例子（163邮箱，发给自己）：  
  原文：`163|your163user|your163passwd`  
  Base64：`MTYzfHlvdXIxNjN1c2VyfHlvdXIxNjNwYXNzd2Q=`

- curl 示例：

```sh
curl -X POST 'http://localhost:8080/cgi-bin/webhook/send?key=MTYzfHlvdXIxNjN1c2VyfHlvdXIxNjNwYXNzd2Q=' \
  -H 'Content-Type: application/json' \
  -d '{"msgtype":"text","text":{"content":"hello\n这是一封测试邮件"}}'
```

---

## 参数说明

- `/cgi-bin/webhook/send?key=Base64参数`
    - Base64参数格式:  
      `服务商|用户名|密码`  
      `服务商|用户名|密码|发件人(不包含@后缀)`  
      `服务商|用户名|密码|发件人(不包含@后缀)|收件人完整地址`  
      `服务商|用户名|密码||收件人完整地址`
- POST Body：企业微信机器人消息格式
- 支持消息类型: `text`, `markdown`, `news`, `image`

---

## 配置文件说明

- `port`：服务监听端口
- `[providers.xxx]`：邮件服务商配置
    - `email_suffix`：邮箱后缀
    - `smtp_host`：SMTP服务器
    - `smtp_port`：SMTP端口

---

## 常见问题

- 部分企业邮箱需开启 SMTP/IMAP，并使用授权码而非明文密码
- 若收不到邮件请检查垃圾箱或 SMTP 配置
