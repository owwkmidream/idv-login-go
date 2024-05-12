package main

import (
	"flag"
	"fmt"
	"github.com/getlantern/elevate"
	"github.com/sirupsen/logrus"
	"idv-login-go/config"
	"idv-login-go/logger"
	"idv-login-go/windowController"
	"os"
	"path/filepath"
	"runtime"
)

var (
	log  *logrus.Logger
	conf *config.Config
)

type BootArgs struct {
	DontAdmin bool
}

func main() {
	// 解析参数
	args := ParseBootArgs()
	if !args.DontAdmin && runtime.GOOS != "linux" {
		cmd := elevate.Command(os.Args[0], "--noadmin")
		cmd.Start()
		os.Exit(0)
	}
	windowController.GetWindowController().HideWindow()
	// 切换工作目录
	ex, err := os.Executable()
	if err != nil {
		fmt.Printf("获取可执行文件路径失败：%v\n", err)
	} else {
		exPath := filepath.Dir(ex)
		err = os.Chdir(exPath)
		if err != nil {
			fmt.Printf("切换工作目录失败：%v\n", err)
		} else {
			fmt.Printf("切换工作目录：%s\n", exPath)
		}
	}

	// 启动
	log = logger.GetLogger()
	conf = config.GetConfig()

	app := newTray()
	app.run()
}

func ParseBootArgs() *BootArgs {
	var args BootArgs
	// 使用flag包解析命令行参数
	flag.BoolVar(&args.DontAdmin, "noadmin", false, "不要升级权限")

	// 解析flag
	flag.Parse()

	return &args
}
