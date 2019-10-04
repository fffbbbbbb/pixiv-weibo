package common

import (
	"strconv"

	"github.com/astaxie/beego/logs"
	"gopkg.in/gomail.v2"
)

func SendMail(body string) {
	//定义收件人
	mailTo := []string{
		"812129358@qq.com",
	}
	//邮件主题为"Hello"
	subject := "你服务器有东西挂了"
	//定义邮箱服务器连接信息，如果是阿里邮箱 pass填密码，qq邮箱填授权码
	mailConn := map[string]string{
		"user": "1706100077@e.gzhu.edu.cn",
		"pass": "bYYxnrYhQzTgYaeQ",
		"host": "smtp.exmail.qq.com",
		"port": "465",
	}

	port, _ := strconv.Atoi(mailConn["port"]) //转换端口类型为int

	m := gomail.NewMessage()
	m.SetHeader("From", "spider"+"<"+mailConn["user"]+">") //这种方式可以添加别名，即“XD Game”， 也可以直接用<code>m.SetHeader("From",mailConn["user"])</code> 读者可以自行实验下效果
	m.SetHeader("To", mailTo...)                           //发送给多个用户
	m.SetHeader("Subject", subject)                        //设置邮件主题
	m.SetBody("text/html", body)                           //设置邮件正文

	d := gomail.NewDialer(mailConn["host"], port, mailConn["user"], mailConn["pass"])

	err := d.DialAndSend(m)
	logs.Error(err)

}
