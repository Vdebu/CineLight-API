package mailer

import (
	"bytes"
	"embed"
	"github.com/go-mail/mail"
	"html/template"
	"time"
)

// 将静态的模板嵌入程序
//
//go:embed "templates"
var templateFS embed.FS

// 用于发送邮件
type Mailer struct {
	dialer *mail.Dialer // 用于连接SMTP服务器
	sender string       // 发件人的邮箱
}

// 创建新的Mailer
func New(host string, port int, username, password, sender string) Mailer {
	// 根据给定的SMTP信息初始化dialer
	dialer := mail.NewDialer(host, port, username, password)
	dialer.Timeout = 5 * time.Second
	return Mailer{
		dialer: dialer,
		sender: sender,
	}
}

// 发送邮件 接收收件人，模板文件名称，输入模板的数据
func (m *Mailer) Send(recipient, templateFile string, data interface{}) error {
	// 解析特定的模板(所有->*.tmpl.html)
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}
	// 先向尝试向缓冲区写入数据
	subject := new(bytes.Buffer)
	// 写入subject的模板
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}
	// 写入plainBody模板
	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}
	// 写入htmlBody模板
	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}
	// 初始化mail.Message实例 -> 存储用于发送的邮件主体信息
	msg := mail.NewMessage()
	// 设置收件人的姓名与发件人的邮箱
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.sender)
	// 设置主题 -> 将先前写入缓冲区的内容转换成string输出
	msg.SetHeader("Subject", subject.String())
	// 设置body文本类型 -> 将先前写入缓冲区的内容转换成string输出
	msg.SetBody("text/plain", plainBody.String())
	// 添加可选的文本选项 AddAlternative只能在SetBody之后进行调用
	msg.AddAlternative("text/html", htmlBody.String())
	// 尝试发送邮件 三次重试机会
	for i := 1; i <= 3; i++ {
		err = m.dialer.DialAndSend(msg)
		if nil == err {
			// 发送成功返回nil
			return nil
		}
		// 等待0.5s继续重试
		time.Sleep(500 * time.Millisecond)
	}
	// 经过三次重试还是发生了错误 将错误进行返回
	return err
}
