package saver

import (
	"fmt"
	"log"
	"path"
	"regexp"

	".."
	"../../models"
	"../parser"
	"../storage"
)

// EmailSaveFallback 用来当邮件内容解析失败之后，但是也需要保存邮件的信息，此时就
// 尽可能从邮件头里面得到信息，然后存储到数据库里面去
func EmailSaveFallback(data []byte, uidl, message string, config *models.ServerConfig) {
	var email, _ = parser.NewEmailFallback(data)
	if email != nil {
		email.Uidl = uidl
		email.Message = message
		EmailSave(email, config)
	}
}

// EmailSave 用来把邮件的保存到数据库，包括相关的Tags信息
// 如果参数 email.Id 存在并且大于0的话，会执行 Update 操作，不是 Insert 操作
func EmailSave(email *models.Email, config *models.ServerConfig) {
	var err error

	if email.Id > 0 {
		// 如果存在 email.Id 的话，那么不应该插入新的数据，而是应该更新数据
		// 比如 rebuild.go 修复数据的时候
		_, err = config.Ormer.Update(email)
	} else {
		email.Id, err = config.Ormer.Insert(email)
	}

	if err != nil {
		log.Println(err)
		return
	}

	// 看看是否有Tags的信息，如果有的话，需要更新Tags
	if email.Tags != nil && len(email.Tags) > 0 {
		// 初始化Tag的信息，有的话，得到Id，没有的话，插入之后也得到了Id
		for _, tag := range email.Tags {
			_, tag.Id, _ = config.Ormer.ReadOrCreate(tag, "Name")
		}

		var m2m = config.Ormer.QueryM2M(email, "Tags")
		_, err = m2m.Add(email.Tags)
		if err != nil {
			log.Println("[ ADD TAG]", email.Uidl, err)
		}
	}
}

// EmailResourceSave 把邮件相关的资源保存下来，以本地文件或者网盘的文件的形式
func EmailResourceSave(email *models.Email, config *models.ServerConfig) {
	re := regexp.MustCompile(`src="cid:([^"]+)"`)
	sm := re.FindAllSubmatch([]byte(email.Message), -1)

	// 先保存cid
	for _, match := range sm {
		var key = string(match[1])
		if value, ok := email.ResourceBundle[key]; ok {
			// 如果存在的话，那么这个文件需要写入cid目录
			var dst = path.Join(config.BaseDir, "downloads", email.Uidl, "cid", key)
			var data = value.Body
			storage.NewDiskStorage(dst, data, 0644).Save()

			// 写完之后删除，最后剩下的就放到att目录即可
			delete(email.ResourceBundle, key)
		}
	}

	// 再保存att
	if len(email.ResourceBundle) <= 0 {
		return
	}

	for _, value := range email.ResourceBundle {
		var filename string
		if value.Name != "" {
			filename = value.Name
		} else if value.ContentID != "" {
			filename = value.ContentID
		} else {
			continue
		}

		var dst = path.Join(config.BaseDir, "downloads", email.Uidl, "att", filename)
		var data = value.Body

		// TODO(user) 应该通过 chan 传递数据过去，而不是每次启动一个新的 goroutine
		go storage.NewDiskStorage(dst, data, 0644).Save()

		// dst需要重新计算
		var token = config.Service.Netdisk.AccessToken
		if token != "" {
			var uidl = util.StripInvalidCharacter(email.Uidl)
			var name = util.StripInvalidCharacter(filename)
			if uidl != "" && name != "" {
				// 一般不会超过1000个字节，所以不考虑超长的情况了
				var dst = fmt.Sprintf("/apps/dropbox/%s/%s/%s/%s",
					config.Pop3.Domain, config.Pop3.Username, uidl, name)
				if len([]byte(dst)) > 1000 {
					log.Println(dst, "was too long")
					continue
				}

				// TODO(user) 应该通过 chan 传递数据过去，而不是每次启动一个新的 goroutine
				go storage.NewNetdiskStorage(token, dst, data).Save()
			}
		}
	}
}
