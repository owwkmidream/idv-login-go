package main

import (
	"github.com/getlantern/systray"
	"idv-login-go/certController"
	"idv-login-go/constants"
	"idv-login-go/dnsController"
	"idv-login-go/hostsController"
	"idv-login-go/icon"
	"idv-login-go/server"
	"idv-login-go/windowController"
	"os"
)

type tray struct {
	mQuit         *systray.MenuItem
	mStart        *systray.MenuItem
	mStop         *systray.MenuItem
	mRestart      *systray.MenuItem
	mToggleWindow *systray.MenuItem
	serv          *server.Server
	shutChan      chan bool
}

func newTray() *tray {
	return &tray{}
}
func (t *tray) start() {
	t.mStart.Disable()
	t.mStop.Enable()

	if !t.init() { // 启动失败
		t.mStart.Enable()
	}
}

func (t *tray) stop() {
	t.mStart.Enable()
	t.mStop.Disable()

	// 关闭代理服务器，移除DNS
	select {
	case t.shutChan <- true:
	default: // 防止阻塞
	}
	// 进行hosts操作
	hostC := hostsController.New()
	if !hostC.IsWritable() {
		log.Info("文件不可写，请关闭杀毒软件或使用管理员权限运行本程序")
	}
	if hostC.Exist() {
		hostC.Remove()
		log.Info("hosts移除完成")
	}
}

func (t *tray) run() {
	systray.Run(t.onReady, t.onExit)
}
func (t *tray) onReady() {
	log.Info("程序启动")
	t.createMenuListening()
	t.start() // 默认进行启动
}

func (t *tray) init() bool {
	// 进行hosts操作
	hostC := hostsController.New()
	if !hostC.IsWritable() {
		log.Info("文件不可写，请关闭杀毒软件或使用管理员权限运行本程序")
		return false
	}
	if !hostC.Exist() {
		log.Info("hosts中不存在，添加")
		hostC.Add()
	}
	log.Info("hosts准备完成")

	// 检查证书是否存在

	if err := func() error {
		if _, errCaCert := os.Stat(constants.CaPath); errCaCert != nil {
			return errCaCert
		}
		if _, errWebCert := os.Stat(constants.CertPath); errWebCert != nil {
			return errWebCert
		}
		if _, errKey := os.Stat(constants.KeyPath); errKey != nil {
			return errKey
		}
		return nil
	}(); os.IsNotExist(err) {
		// 生成证书
		certM := certController.New()
		certM.GenerateCA()
		certM.GenerateCert([]string{conf.String("host")})

		// 导出证书和key
		certM.ExportCert(constants.CaPath, certM.CaCert)
		certM.ExportCert(constants.CertPath, certM.WebCert)
		certM.ExportKey(constants.KeyPath)

		// 导入CA证书
		if done, err := certM.ImportToRoot(constants.CaPath); !done {
			// 删除证书文件
			os.Remove(constants.CaPath)
			os.Remove(constants.CertPath)
			os.Remove(constants.KeyPath)

			log.Fatalf("导入CA证书失败：%v", err)
			return false
		}
	}
	log.Infof("证书准备完成")

	// 解析DNS
	dnsC := dnsController.NewDnsController()
	ip, err := dnsC.Resolve()
	if err != nil {
		log.Errorf("DNS解析失败：%v\n将使用默认IP", err)
		ip = conf.String("defaultIP")
	}
	log.Infof("DNS解析结果：%s", ip)

	// 创建一个 channel 用于发送终止信号
	t.shutChan = make(chan bool)

	go func() { // 启动代理服务器
		t.serv = server.NewServer(conf.String("host"), ip)
		t.serv.Run(t.shutChan)
	}()
	return true
}

func (t *tray) createMenuListening() {
	t.mStart = systray.AddMenuItem("启动", "启动")
	t.mStop = systray.AddMenuItem("停止", "停止")
	t.mRestart = systray.AddMenuItem("重启", "重启")
	t.mToggleWindow = systray.AddMenuItem("显示窗口", "显示窗口")
	t.mQuit = systray.AddMenuItem("退出", "退出")

	systray.SetIcon(icon.Icon)
	systray.SetTitle("idv-login-go")
	systray.SetTooltip("第五人格登录助手")

	go func() {
		for {
			select {
			case <-t.mStart.ClickedCh:
				t.start()
			case <-t.mStop.ClickedCh:
				t.stop()
			case <-t.mRestart.ClickedCh:
				t.stop()
				t.start()
			case <-t.mToggleWindow.ClickedCh:
				wC := windowController.GetWindowController()
				wC.ToggleWindow()
				if wC.Status == 1 {
					t.mToggleWindow.SetTitle("隐藏窗口")
				} else {
					t.mToggleWindow.SetTitle("显示窗口")
				}
			case <-t.mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func (t *tray) onExit() {
	t.stop()
}
