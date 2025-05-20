# wecom-webhook-push-mail

本项目兼容企业微信机器人 webhook 接口和消息格式，接收到的消息内容自动转发到指定邮箱。

## 功能特性

- 支持企业微信机器人 Webhook 消息格式（text/markdown/news/image）
- 支持自定义邮件服务商，支持任意 SMTP（QQ/163/自建邮局等）
- 所有认证信息通过 Base64 编码存储在 URL 中
- 服务器不存储任何用户信息
- Go 单文件实现，依赖极少，易于维护

---

## 注意事项

🚨 **部署服务前需先确认可连接目标邮件服务商 SMTP 端口**

---

## 快速部署

### 一、二进制部署

#### 1. 下载 Release 包

前往 [Releases 页面](https://github.com/MxxApp/wecom-webhook-push-mail/releases) 下载对应架构的二进制包，解压得到 `wwpm`。

#### 2. 启动服务

```bash
# 默认 8080 端口
./wwpm

# 或自定义端口
./wwpm -port=8080
```

---

### 二、Docker 部署

```bash
docker run -d -p 8080:8080 ghcr.io/mxxapp/wwpm:latest
```

---

## 消息推送用法

向 `/cgi-bin/webhook/send?key=xxxx` 发 POST 请求，Body 为企业微信机器人兼容 JSON，  
**key** 参数为如下字段用 `|` 分割后 base64 编码：

| 顺序 | 含义                | 示例                       |
|------|---------------------|----------------------------|
| 1    | SMTP主机:端口       | smtp.qq.com:465            |
| 2    | 完整用户名          | user@qq.com                |
| 3    | 密码                | xxxx123456                 |
| 4    | 完整发件人邮箱(可选) | sender@qq.com              |
| 5    | 完整收件人邮箱(可选) | receiver@qq.com            |

### key参数编码例子

```bash
echo -n "smtp.qq.com:465|user@qq.com|xxxx123456|sender@qq.com|receiver@qq.com" | base64 | tr -d '\n'; echo
```

### curl 请求示例

假设你 base64 后结果为 `c210cC5xcS5jb206NDY1fHVzZXJA...`，则：

```bash
curl -X POST "http://127.0.0.1:8080/cgi-bin/webhook/send?key=c210cC5xcS5jb206NDY1fHVzZXJA..." \
     -H "Content-Type: application/json" \
     -d '{
            "msgtype":"text",
            "text": {
                "content": "邮件标题\n这里是正文内容\n支持多行"
            }
         }'
```

---

## 支持的消息类型

- `text`
- `markdown`
- `image`
- `news`

消息体请参考 [WeCom 官方文档](https://developer.work.weixin.qq.com/document/path/90236) 或本项目示例。

---

## 常见问题

- 部分企业邮箱需开启 SMTP/IMAP，并使用授权码而非明文密码
- 若收不到邮件请检查垃圾箱或 SMTP 配置
- 发送失败请检查 SMTP 服务器地址/端口、用户名、密码是否正确

---
