package controllers

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/astaxie/beego"

	"../../models"
)

// OriMessageController 处理查看原始邮件内容的请求
type OriMessageController struct {
	beego.Controller
}

// Get 处理 GET 请求
func (controller *OriMessageController) Get() {
	controller.Post()
}

// Post 处理 POST 请求
func (controller *OriMessageController) Post() {
	var uidl = controller.Ctx.Input.Param(":uidl")
	if uidl == "" {
		controller.Abort(strconv.Itoa(http.StatusBadRequest))
	}

	var email models.Email
	err := gSrvConfig.Ormer.QueryTable("email").
		Filter("Uidl", uidl).
		One(&email, "Message", "Uidl")

	if err != nil {
		log.Println(err)
		controller.Abort(strconv.Itoa(http.StatusInternalServerError))
	}

	// 从sqlite3导入数据到mysql之后，发生了转义的问题
	// 如果全部修复，貌似也没啥意义，就这里用到的时候再处理一下吧
	if strings.Index(email.Message, "`Content-Type`") != -1 {
		email.Message = strings.Replace(email.Message, "`", "'", -1)
	}

	// 替换src="cid:"的内容
	var re = regexp.MustCompile(`src="cid:([^"]+)"`)
	email.Message = re.ReplaceAllString(email.Message,
		fmt.Sprintf(`src="/downloads/%s/cid/$1"`, url.QueryEscape(email.Uidl)))

	// 有的数据记录里面已经是src="downloads"了，此时需要替换一下（老的数据）
	// 新的数据里面保存的都是src="cid:xxx"
	email.Message = strings.Replace(email.Message, `src="downloads`, `src="/downloads`, -1)

	controller.Ctx.ResponseWriter.Header().Set("Content-Type", "text/html; charset=utf-8")
	controller.Ctx.WriteString(email.Message)
}
