package main

import (
	"flag"
	"fmt"
	"log"
	"mime"
	"os"
	"path"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"

	"./backend"
	"./bindata"
	cf "./config"
	"./frontend"
	"./util"
)

// 最终所发布文件的编译日期
var BuildDate = "1989-06-04 03:04:05"

// 最终所发布文件的编译版本(git commit id)
var BuildVersion = "default"

func printHelp() {
	fmt.Printf("BUILD_DATE: %s\n", BuildDate)
	fmt.Printf("BUILD_VERSION: %s\n", BuildVersion)
}

func main() {
	var configfile = flag.String("config", "config.json", "The config file path")
	var version = flag.Bool("version", false, "Print version")
	flag.Parse()

	if *version {
		printHelp()
		return
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var initMode = false
	config, err := cf.NewConfig(*configfile)
	if err != nil {
		initMode = true
		log.Println("Read config file failed, enter init mode.")
		config = cf.DefaultConfig()
	}

	if len(os.Args) >= 2 && strings.Index(os.Args[1], "-config=") == 0 {
		// 为了兼容 	orm.RunCommand() 的参数处理逻辑，当发现传递 -config= 参数的时候，
		// 人肉把它删掉，例如：
		// os.Args = ["main", "-config=x.json", "orm", "sqlall"]
		// =>
		// os.Args = ["main", "orm", "sqlall"]
		// https://github.com/astaxie/beego/blob/a144769515b8ddfb69046b3ef9e29c38bab94f3b/orm/cmd.go#L51
		os.Args = append(os.Args[0:1], os.Args[2:]...)
	}

	orm.Debug = config.Debug
	orm.RegisterDriver("mysql", orm.DR_MySQL)
	orm.RegisterDriver("sqlite3", orm.DR_Sqlite)

	if !initMode {
		// [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
		srv := config.Service.Db
		if srv.Type == "" || srv.Type == "mysql" {
			dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8",
				srv.Username, srv.Password, srv.Host, srv.Port, srv.Name)
			orm.RegisterDataBase("default", "mysql", dsn)
		} else if srv.Type == "sqlite" {
			orm.RegisterDataBase("default", "sqlite3",
				path.Join(config.BaseDir, "db", srv.Name+".db"))
		} else {
			log.Fatalf("Invalid Database Type: %s\n", srv.Type)
		}

		orm.RunCommand()
		config.Ormer = orm.NewOrm()
	}

	beego.HttpPort = config.HTTP.Port
	beego.HttpAddr = "localhost"
	beego.DirectoryIndex = true
	beego.EnableGzip = true
	if !config.Debug {
		beego.RunMode = "prod"
	}

	// 静态文件的根目录
	// var root = "static"

	// beego.StaticDir["/"] = root

	// 特殊的静态文件目录
	// beego.SetStaticPath("/src", path.Join(root, "src"))
	// beego.SetStaticPath("/asset", path.Join(root, "asset"))
	// beego.SetStaticPath("/dep", path.Join(root, "dep"))

	// 一个特殊的资源文件，只在开发的模式下启用
	beego.InsertFilter("/src/common/css/main.less",
		beego.BeforeStatic, util.StyleFilter("static"))
	beego.InsertFilter("/src/common/css/preview.less",
		beego.BeforeStatic, util.StyleFilter("static"))
	beego.InsertFilter("/src/common/css/ckeditor.less",
		beego.BeforeStatic, util.StyleFilter("static"))
	beego.InsertFilter("/", beego.BeforeStatic, bindataProvider)
	beego.InsertFilter("/index.html", beego.BeforeStatic, bindataProvider)
	beego.InsertFilter("/preview.html", beego.BeforeStatic, bindataProvider)
	beego.InsertFilter("/src/*", beego.BeforeStatic, bindataProvider)
	beego.InsertFilter("/asset/*", beego.BeforeStatic, bindataProvider)
	beego.InsertFilter("/dep/*", beego.BeforeStatic, bindataProvider)

	beego.SetStaticPath("/downloads", path.Join(config.BaseDir, "downloads"))
	beego.SetStaticPath("/raw", path.Join(config.BaseDir, "raw"))

	go backend.Run(config)
	go backend.FlushContacts(config)

	frontend.Run(config)
}

// 调用bindata.Asset的接口，返回数据
func bindataProvider(ctx *context.Context) {
	var url = ctx.Request.URL.Path
	if url == "/" || url == "" {
		url = "index.html"
	} else {
		// 删掉最前面的 /
		url = url[1:]
	}
	if orm.Debug {
		log.Println(url)
	}

	data, err := bindata.Asset(url)
	if err != nil {
		log.Println(err)
		return
	}

	ct := mime.TypeByExtension(path.Ext(url))
	if ct == "" {
		ct = "text/plain; charset=utf-8"
	}

	ctx.ResponseWriter.Header().Set("Content-Type", ct)
	ctx.ResponseWriter.Write(data)
}
