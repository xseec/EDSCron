package cronx

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/zeromicro/go-zero/core/logx"
	"gopkg.in/gomail.v2"
)

// 预定义的邮件主题
const (
	SubjectDlgdWarning    MailSubject = "[警告]代理购电异常"
	SubjectHolidayNotice  MailSubject = "[通知]节假日更新"
	SubjectEmptyData      MailSubject = "[警告]空记录"
	SubjectTwdlNotice     MailSubject = "[通知]台湾电价表更新"
	SubjectTwCarbonNotice MailSubject = "[通知]台湾碳排因子更新"
)

// MailConfig 邮件服务器配置
type MailConfig struct {
	Host     string `json:"host"`     // SMTP服务器地址
	Port     int    `json:"port"`     // SMTP服务器端口
	Username string `json:"username"` // 发件人账号
	Password string `json:"password"` // 发件人密码/授权码
	Addr     string `json:"addr"`     // 收件人邮箱地址
}

// MailSubject 邮件主题类型
type MailSubject string

// MailTemplate 邮件模板接口
type MailTemplate interface {
	Subject() MailSubject
	Body() (string, error)
}

// DlgdWarningTemplate 电价获取失败模板
type DlgdWarningTemplate struct {
	Remark string
	Rule   string
}

func (t DlgdWarningTemplate) Subject() MailSubject {
	return SubjectDlgdWarning
}

func (t DlgdWarningTemplate) Body() (string, error) {
	const tpl = `<p>备注无<b>{{.Remark}}</b>规定</p><p><blockquote>{{.Rule}}</blockquote></p>`
	return renderTemplate(tpl, t)
}

// HolidayNoticeTemplate 节假日获取成功模板
type HolidayNoticeTemplate struct {
	Area    string
	Details string
}

func (t HolidayNoticeTemplate) Subject() MailSubject {
	return SubjectHolidayNotice
}

func (t HolidayNoticeTemplate) Body() (string, error) {
	const tpl = `<p><b>{{.Area}}</b>假期安排</p>{{.Details}}`
	return renderTemplate(tpl, t)
}

// EmptyDataTemplate 空数据模板
type EmptyDataTemplate struct {
	TaskDetail interface{}
}

func (t EmptyDataTemplate) Subject() MailSubject {
	return SubjectEmptyData
}

func (t EmptyDataTemplate) Body() (string, error) {
	const tpl = `数据源格式可能发生变化, 任务详情: {{.TaskDetail}}`
	return renderTemplate(tpl, t)
}

// TwdlTemplate 台湾电价模板
type TwdlTemplate struct {
	Calendar    string
	StartDate   string
	RecordCount int
	TargetCount int
	Details     template.HTML
}

func (t TwdlTemplate) Subject() MailSubject {
	return SubjectTwdlNotice
}

func (t TwdlTemplate) Body() (string, error) {
	const tpl = `<p>生效时间 : <b>{{.StartDate}}</b></p>
<p>电价条数 : <b>{{.RecordCount}}</b> (目标值 : {{.TargetCount}})</p>
<p>离峰日 : <b>{{.Calendar}}</b></p>
<p>列表详情 : </p><div>{{.Details | safeHTML}}</div>`
	return renderTemplate(tpl, t)
}

type TwCarbonTemplate struct {
	Year  int
	Value float64
}

func (t TwCarbonTemplate) Subject() MailSubject {
	return SubjectTwCarbonNotice
}

func (t TwCarbonTemplate) Body() (string, error) {
	const tpl = `<p>台湾{{.Year}}年度電力排碳係數 = <b>{{.Value}}</b>公斤 CO<sub>2</sub>e/度</p>`
	return renderTemplate(tpl, t)
}

// renderTemplate 渲染模板
func renderTemplate(tpl string, data interface{}) (string, error) {
	tmpl, err := template.New("mail").
		Funcs(template.FuncMap{
			"safeHTML": func(s template.HTML) template.HTML {
				return s
			},
		}).
		Parse(tpl)
	if err != nil {
		return "", fmt.Errorf("解析模板失败: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("执行模板失败: %w", err)
	}

	return buf.String(), nil
}

// emptyValueErr 检查空值并发送通知邮件
func emptyValueErr[T any](m *MailConfig, task any, v *[]T) (*[]T, error) {
	if v == nil || len(*v) == 0 {
		if m != nil {
			tpl := EmptyDataTemplate{TaskDetail: task}
			m.Send(tpl) // 发送空数据通知
		}
		return nil, fmt.Errorf("空数据错误: %v", task)
	}
	return v, nil
}

// Send 发送邮件通知
func (m *MailConfig) Send(template MailTemplate, files ...string) error {
	body, err := template.Body()
	if err != nil {
		return fmt.Errorf("生成邮件内容失败: %w", err)
	}

	msg := gomail.NewMessage()
	msg.SetHeader("From", msg.FormatAddress(m.Username, "EDS服务"))
	msg.SetHeader("To", m.Addr)
	msg.SetHeader("Subject", string(template.Subject()))
	msg.SetBody("text/html", body)
	for _, file := range files {
		msg.Attach(file)
	}

	dialer := gomail.NewDialer(m.Host, m.Port, m.Username, m.Password)
	if err := dialer.DialAndSend(msg); err != nil {
		logx.Errorf("发送邮件失败: %v", err)
		return fmt.Errorf("发送邮件失败: %w", err)
	}

	return nil
}
