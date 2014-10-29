package backend

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"github.com/astaxie/beego/orm"
	pop3 "github.com/bytbox/go-pop3"

	"../filter"
	"../models"
	"../util/parser"
	"../util/saver"
)

func Receiver(config *models.ServerConfig) error {
	log.Println("Receiving Mails...")

	var hostname = fmt.Sprintf("%s:%d", config.Pop3.Host, config.Pop3.Port)
	var client *pop3.Client
	var err error
	if config.Pop3.Tls {
		client, err = pop3.DialTLS(hostname)
	} else {
		client, err = pop3.Dial(hostname)
	}

	if err != nil {
		return err
	}
	defer client.Quit()

	err = client.Auth(config.Pop3.Username, config.Pop3.Password)
	if err != nil {
		return err
	}

	// 开始拉数据
	msgs, _, err := client.ListAll()
	if err != nil {
		return err
	}

	filters, err := filter.NewFilters(config.Service.Filter.Config)
	if err != nil {
		log.Println(err)
	}

	uidls, err := client.UidlAll()
	if err != nil {
		return err
	}

	// 在循环的过程中，如果有错误，打印出来，不要中断循环
	for msgid, msg := range msgs {
		var uidl = uidls[msg]
		var email = &models.Email{Uidl: string(uidl)}
		var err = config.Ormer.Read(email, "Uidl")

		if err == nil && email.Id > 0 {
			// 存在记录
			var kmos = config.Pop3.KeepMailOnServer
			if kmos > 0 &&
				time.Since(email.Date).Hours() > float64(24*kmos) {
				// 判断 date 和 当前的日期差别，如果大于 config.Pop3.Delete_older_mails 的配置
				// 说明可以删除
				log.Printf("[ DELETE] %d -> %s (%f)\n", msgid, uidl,
					time.Since(email.Date).Hours())
				client.Dele(msgid)
			}
			continue
		} else if err != orm.ErrNoRows {
			// 肯定有其它错误了
			log.Println(err)
			continue
		}

		raw, err := client.Retr(msg)
		if err != nil {
			log.Println(err)
			continue
		}

		// 如果不存在的话，保存原始的文件
		var rawFile = path.Join(config.BaseDir, "raw", uidl+".txt")
		if _, err := os.Stat(rawFile); err != nil {
			err = ioutil.WriteFile(rawFile, []byte(raw), 0644)
			if err != nil {
				log.Println(err)
			}
		}

		// 解析邮件的正文，得到 Email 对象
		email, err = parser.NewEmail([]byte(raw))
		if err != nil {
			log.Println(uidl, err)
			continue
		}
		email.Uidl = string(uidl)

		// 执行过滤器
		if filters != nil {
			err = filter.RunFilter(email, filters[:])
			if err != nil {
				log.Println(uidl, err)
			}
		}

		// 保存到数据库
		saver.EmailSave(email, config)

		// 保存邮件的资源和附件
		saver.EmailResourceSave(email, config)

		log.Printf("[ SAVE] %d -> %s (%d)\n", msg, uidl, email.Id)
	}

	return nil
}
