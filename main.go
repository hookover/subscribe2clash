package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/whoisix/subscribe2clash/api"
	"github.com/whoisix/subscribe2clash/pkg/clash/acl"
)

var (
	gc         bool
	h          bool
	baseFile   string
	outputFile string
	origin     string
	listenAddr string
	listenPort string
)

func init() {
	flag.BoolVar(&h, "h", false, "help")
	flag.BoolVar(&gc, "gc", false, "生成clash配置文件")
	flag.StringVar(&origin, "origin", "github", "acl规则获取地址。cn：国内镜像，github：github获取")
	flag.StringVar(&baseFile, "b", "", "clash基础配置文件")
	flag.StringVar(&outputFile, "o", "", "输出clash文件名")
	flag.StringVar(&listenAddr, "l", "0.0.0.0", "listen address")
	flag.StringVar(&listenPort, "p", "8162", "listen port")
	flag.Parse()
}

func main() {
	if h {
		flag.Usage()
		return
	}

	if gc {
		var options []acl.GenOption
		if origin != "" {
			options = append(options, acl.WithOrigin(origin))
		}
		if baseFile != "" {
			options = append(options, acl.WithBaseFile(baseFile))
		}
		if outputFile != "" {
			options = append(options, acl.WithOutputFile(outputFile))
		}
		acl.GenerateConfig(options...)
		return
	}

	acl.GenerateConfig()

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.GET("/", api.Clash)

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", listenAddr, listenPort),
		Handler: router,
	}

	go func() {
		// 服务连接
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 等待中断信号以优雅地关闭服务器（设置 5 秒的超时时间）
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")

}
