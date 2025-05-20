package main

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gopkg.in/gomail.v2"
)

// 邮件发送
type Mailer struct {
	SMTPHost string
	SMTPPort int
	User     string
	Password string
}
func NewMailer(host string, port int, user, pwd string) *Mailer {
	return &Mailer{host, port, user, pwd}
}
func (m *Mailer) SendEmailWithFrom(from, to, subject, body string) error {
	message := gomail.NewMessage()
	message.SetHeader("From", from)
	message.SetHeader("To", to)
	message.SetHeader("Subject", subject)
	message.SetBody("text/html", body)
	dialer := gomail.NewDialer(m.SMTPHost, m.SMTPPort, m.User, m.Password)
	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	return dialer.DialAndSend(message)
}

// 去除首尾空白行
func trimBlankLines(s string) string {
	lines := strings.Split(s, "\n")
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	return strings.Join(lines, "\n")
}

// 消息体处理
type Message interface {
	ToHTML() (title, content string, err error)
}
func NewMessage(msgType string, data []byte) (Message, error) {
	switch msgType {
	case "text":
		var m TextMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, err
		}
		return &m, nil
	case "markdown":
		var m MarkdownMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, err
		}
		return &m, nil
	case "image":
		var m ImageMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, err
		}
		return &m, nil
	case "news":
		var m NewsMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, err
		}
		return &m, nil
	case "html":
		var m HTMLMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, err
		}
		return &m, nil
	default:
		return nil, errors.New("unknown message type")
	}
}

type TextMessage struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
}
func (msg *TextMessage) ToHTML() (string, string, error) {
	lines := strings.Split(strings.TrimSpace(msg.Text.Content), "\n")
	if len(lines) == 0 || strings.TrimSpace(msg.Text.Content) == "" {
		return "", "", errors.New("empty text content")
	}
	title := strings.TrimSpace(lines[0])
	content := ""
	if len(lines) > 1 {
		content = strings.Join(lines[1:], "\n")
		content = trimBlankLines(content)
	}
	if content == "" {
		content = title
	}
	contentHTML := fmt.Sprintf("<p>%s</p>", strings.ReplaceAll(content, "\n", "<br>"))
	return title, contentHTML, nil
}

type MarkdownMessage struct {
	MsgType  string `json:"msgtype"`
	Markdown struct {
		Content string `json:"content"`
	} `json:"markdown"`
}
func (msg *MarkdownMessage) ToHTML() (string, string, error) {
	content := msg.Markdown.Content
	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) == 0 || strings.TrimSpace(content) == "" {
		return "", "", errors.New("empty markdown content")
	}
	title := strings.TrimSpace(lines[0])
	body := ""
	if len(lines) > 1 {
		body = strings.Join(lines[1:], "\n")
		body = trimBlankLines(body)
	}
	if body == "" {
		body = title
	}
	contentHtml := "<pre>" + body + "</pre>"
	return title, contentHtml, nil
}

type ImageMessage struct {
	MsgType string `json:"msgtype"`
	Image   struct {
		Base64 string `json:"base64"`
		MD5    string `json:"md5"`
	} `json:"image"`
}
func (msg *ImageMessage) ToHTML() (string, string, error) {
	if msg.Image.Base64 == "" || msg.Image.MD5 == "" {
		return "", "", errors.New("missing image data or MD5")
	}
	title := fmt.Sprintf("Image Message - %s", time.Now().Format("2006-01-02 15:04:05"))
	content := fmt.Sprintf("<img src='data:image/png;base64,%s' alt='Image Message'/>", msg.Image.Base64)
	return title, content, nil
}

type NewsMessage struct {
	MsgType string `json:"msgtype"`
	News    struct {
		Articles []struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			URL         string `json:"url"`
			PicURL      string `json:"picurl"`
		} `json:"articles"`
	} `json:"news"`
}
func (msg *NewsMessage) ToHTML() (string, string, error) {
	if len(msg.News.Articles) == 0 {
		return "", "", errors.New("no articles in news message")
	}
	title := msg.News.Articles[0].Title
	htmlContent := "<ul>"
	for _, a := range msg.News.Articles {
		htmlContent += fmt.Sprintf("<li><a href='%s'><img src='%s' alt='%s'></a><p>%s</p></li>", a.URL, a.PicURL, a.Title, a.Description)
	}
	htmlContent += "</ul>"
	return title, htmlContent, nil
}

type HTMLMessage struct {
	MsgType string `json:"msgtype"`
	HTML    struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	} `json:"html"`
}
func (msg *HTMLMessage) ToHTML() (string, string, error) {
	if msg.HTML.Content == "" {
		return "", "", errors.New("empty html content")
	}
	return msg.HTML.Title, msg.HTML.Content, nil
}

// 工具函数
func Base64Decode(encoded string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}
func errorResponse(c *fiber.Ctx, code, errcode int, errmsg string) error {
	return c.Status(code).JSON(fiber.Map{"errcode": errcode, "errmsg": errmsg})
}
func WrapHTML(body string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: sans-serif; background: #fff; color: #222; }
        .footer-time {
            color: #aaa;
            font-size: 0.92em;
            text-align: left;
            margin-top: 36px;
            padding-top: 10px;
        }
        hr {
            border: none;
            border-top: 1px solid #eee;
            margin-top: 32px;
            margin-bottom: 0;
        }
    </style>
</head>
<body>
    <div class="content">
        %s
        <hr>
        <div class="footer-time">%s</div>
    </div>
</body>
</html>`, body, time.Now().In(time.FixedZone("CST", 8*3600)).Format("2006-01-02 15:04:05"))
}
func StripHTML(input string) string {
	return input
}

// 主接口
func MailHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		data := c.Query("key")
		if data == "" {
			return errorResponse(c, 400, 40012, "missing key parameter")
		}
		decoded, err := Base64Decode(data)
		if err != nil {
			return errorResponse(c, 400, 40008, "invalid Base64 parameter")
		}
		parts := strings.Split(decoded, "|")
		if len(parts) < 3 || len(parts) > 5 {
			return errorResponse(c, 400, 40009, "invalid parameter format")
		}
		// 新参数处理
		hostport := parts[0] // 例: smtp.qq.com:465
		user := parts[1]     // 完整用户名: user@qq.com
		password := parts[2]
		fromAddress := user
		toAddress := user
		if len(parts) >= 4 && parts[3] != "" {
			fromAddress = parts[3]
		}
		if len(parts) == 5 && parts[4] != "" {
			toAddress = parts[4]
		}
		// host:port 拆分
		var smtpHost string
		var smtpPort int
		if idx := strings.LastIndex(hostport, ":"); idx > 0 {
			smtpHost = hostport[:idx]
			fmt.Sscanf(hostport[idx+1:], "%d", &smtpPort)
		}
		if smtpHost == "" || smtpPort == 0 {
			return errorResponse(c, 400, 40013, "invalid smtp host/port")
		}
		var body struct {
			MsgType string `json:"msgtype"`
		}
		if err := c.BodyParser(&body); err != nil {
			return errorResponse(c, 400, 40001, "Invalid JSON body")
		}
		msg, err := NewMessage(body.MsgType, c.Body())
		if err != nil {
			return errorResponse(c, 400, 40002, "Invalid message type")
		}
		title, content, err := msg.ToHTML()
		if err != nil {
			return errorResponse(c, 500, 40003, "Failed to generate HTML content")
		}
		if body.MsgType != "html" {
			content = WrapHTML(content)
		}
		title = StripHTML(title)
		log.Printf("Sending email to %s, type: %s, title: %s (from: %s)", toAddress, body.MsgType, title, fromAddress)
		m := NewMailer(smtpHost, smtpPort, user, password)
		if err := m.SendEmailWithFrom(fromAddress, toAddress, title, content); err != nil {
			return errorResponse(c, 500, 40004, fmt.Sprintf("Failed to send email: %v", err))
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"errcode": 0, "errmsg": "ok"})
	}
}

func main() {
	port := flag.Int("port", 8080, "listen port")
	flag.Parse()

	app := fiber.New()
	app.Post("/cgi-bin/webhook/send", MailHandler())
	log.Printf("Server running on port %d", *port)
	log.Fatal(app.Listen(fmt.Sprintf(":%d", *port)))
}