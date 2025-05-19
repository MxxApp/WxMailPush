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

	"github.com/BurntSushi/toml"
	"github.com/gofiber/fiber/v2"
	"gopkg.in/gomail.v2"
)

// 配置结构体
type Provider struct {
	EmailSuffix string `toml:"email_suffix"`
	SMTPHost    string `toml:"smtp_host"`
	SMTPPort    int    `toml:"smtp_port"`
}
type Providers map[string]Provider
type Config struct {
	Port      int       `toml:"port"`
	Providers Providers `toml:"providers"`
}
func LoadConfig(filename string) (*Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(filename, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
func (c *Config) GetProvider(name string) (*Provider, error) {
	p, ok := c.Providers[name]
	if !ok {
		return nil, fmt.Errorf("provider '%s' not found", name)
	}
	return &p, nil
}

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
	if len(lines) == 0 {
		return "", "", errors.New("empty text content")
	}
	title := strings.TrimSpace(lines[0])
	content := fmt.Sprintf("<p>%s</p>", strings.ReplaceAll(msg.Text.Content, "\n", "<br>"))
	return title, content, nil
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
	if len(lines) == 0 {
		return "", "", errors.New("empty markdown content")
	}
	title := strings.TrimSpace(lines[0])
	contentHtml := "<pre>" + content + "</pre>"
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
func MailHandler(cfg *Config) fiber.Handler {
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
		provider, username, password := parts[0], parts[1], parts[2]
		customuser, customto := "", ""
		if len(parts) >= 4 {
			customuser = parts[3]
		}
		if len(parts) == 5 {
			customto = parts[4]
		}
		providerCfg, err := cfg.GetProvider(provider)
		if err != nil {
			return errorResponse(c, 404, 40010, "provider not found")
		}
		emailAddress := username + providerCfg.EmailSuffix
		fromAddress := emailAddress
		toAddress := emailAddress
		if customuser != "" {
			fromAddress = customuser + providerCfg.EmailSuffix
		}
		if customto != "" {
			toAddress = customto
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
		content = WrapHTML(content)
		title = StripHTML(title)
		log.Printf("Sending email to %s, type: %s, title: %s (from: %s)", toAddress, body.MsgType, title, fromAddress)
		m := NewMailer(providerCfg.SMTPHost, providerCfg.SMTPPort, emailAddress, password)
		if err := m.SendEmailWithFrom(fromAddress, toAddress, title, content); err != nil {
			return errorResponse(c, 500, 40004, fmt.Sprintf("Failed to send email: %v", err))
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"errcode": 0, "errmsg": "ok"})
	}
}

func main() {
	configPath := flag.String("config", "config.toml", "path to config file")
	flag.Parse()

	cfg, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	app := fiber.New()
	app.Post("/cgi-bin/webhook/send", MailHandler(cfg))
	port := 8080
	if cfg.Port > 0 {
		port = cfg.Port
	}
	log.Printf("Server running on port %d", port)
	log.Fatal(app.Listen(fmt.Sprintf(":%d", port)))
}