package main

import (
	"github.com/qinguanri/fileclan/handler"
	"github.com/qinguanri/fileclan/middlewares"
	"flag"
	"fmt"
	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

const (
	// FileClan配置文件
	CONFIG_FILE = `.` + string(os.PathSeparator) + `fileclan.conf`
)

func main() {
	flag.Parse()
	glog.Infoln("FileClan ApiServer Starting...")

	selfPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		glog.Fatalln("cannot touch os.Args[0] filepath, maybe disk problem: %v", err)
	}

	// 加载本程序所需要的配置文件fileclan.conf
	if err = middlewares.LoadConfig(filepath.Join(selfPath, CONFIG_FILE)); err != nil {
		panic(err)
	}

	// 将配置文件内容打印到屏幕
	fmt.Println(ConsoleOut(*middlewares.Conf))

	if err = middlewares.InitBackend(); err != nil {
		glog.Fatalln("init backend failed, err: %v", err)
		panic(err)
	}

	router := fasthttprouter.New()
	router.GET("/hello", handler.Hello)
	router.PUT("/api/v1/object", handler.PutObject)
	router.HEAD("/api/v1/object", handler.HeadObject)
	router.GET("/api/v1/object", handler.GetObject)
	router.DELETE("/api/v1/object", handler.DeleteObject)

	fmt.Println("ListenAndServe ...")

	//启动fasthttp
	go func(allowedIP []string) {
		IPAllowHandler := func(ctx *fasthttp.RequestCtx) {
			ip := ctx.RemoteIP().String()
			isAllowed := false

			for _, v := range allowedIP {
				if v == ip {
					isAllowed = true
					break
				}
			}

			if !isAllowed {
				glog.Warningln(ip + " Access Unauthorized.")
				ctx.Error("IP Access Unauthorized.", 403)
				return
			}

			router.Handler(ctx)
		}
		if err := fasthttp.ListenAndServe(middlewares.Conf.Server.Addr, IPAllowHandler); err != nil {
			glog.Fatalln("fileClan error:", err)
		}
	}(middlewares.Conf.Server.AllowedIPs)

	// 阻塞在这里，捕获ctrl+c的终止服务信号，平滑退出
	chExit := make(chan os.Signal, 1)
	signal.Notify(chExit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	select {
	case <-chExit:
	}

	defer func() {
		if err := recover(); err != nil {
			glog.Fatalln("panic: %v", err)
		}
		glog.Infoln("FileClan EXIT!")
		glog.Flush()
	}()
}
