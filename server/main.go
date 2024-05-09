package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/imroc/req/v3"
	"github.com/sirupsen/logrus"
	"idv-login-go/constants"
	"idv-login-go/logger"
	"net"
	"net/http"
	"strings"
	"time"
)

var cvType = struct {
	Pc   *string
	Ios  *string
	Code *string
}{
	Pc:   func() *string { s := constants.Pcv; return &s }(),
	Ios:  func() *string { s := constants.Icv; return &s }(),
	Code: func() *string { s := constants.Ccv; return &s }(),
}

type Server struct {
	targetHost   string
	redirectHost string
	urlRedirect  string
	client       *req.Client
	ginServer    *gin.Engine
}

func NewServer(targetHost string, targetIp string) *Server {
	log = logger.GetLogger()
	cli := req.C().EnableInsecureSkipVerify()
	if constants.DebugMode {
		cli.DevMode()
	}

	return &Server{
		targetHost:   targetHost,
		redirectHost: targetIp,
		urlRedirect:  fmt.Sprintf("https://%s", targetIp),
		client:       cli,
	}
}

var log *logrus.Logger

func (s *Server) Run(shutChan chan bool) {
	// 检查重定向情况
	ip, err := net.LookupHost(s.targetHost)
	if err != nil {
		log.Errorf("重定向失败：%v", err)
		return
	}

	if strings.Compare(constants.Localhost, ip[0]) != 0 {
		log.Errorf("重定向IP不一致，目标IP：%s，解析IP：%s", constants.Localhost, ip[0])
		return
	}

	// 检查端口占用
	if done, err := s.checkPort(); !done || err != nil {
		log.Errorf("端口检查失败：%v", err)
		return
	}

	// 启动代理服务器
	log.Info("启动代理服务器...")

	if !constants.DebugMode {
		gin.SetMode(gin.ReleaseMode)
	}
	s.ginServer = gin.Default()
	s.setupRoutes()

	srv := &http.Server{
		Addr:    ":443",
		Handler: s.ginServer,
	}

	// 使用TLS证书和私钥启动服务器
	go func() {
		if err := srv.ListenAndServeTLS(constants.CertPath, constants.KeyPath); err != nil && !errors.Is(err, http.ErrServerClosed) {
			{
				log.Fatalf("代理服务器运行失败：%v", err)
				return
			}
		}
	}()

	// 等待中断信号
	<-shutChan
	log.Info("代理服务器关闭...")

	// 创建一个 5 秒的超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 关闭 HTTP Server
	// 5秒内优雅关闭服务（将未处理完的请求处理完再关闭服务），超过5秒就超时退出
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("代理服务器关闭出错：%v", err)
	}
	log.Info("代理服务器已关闭")
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	g := s.ginServer
	// 修改登录方法
	g.GET("/mpay/games/:game_id/login_methods", s.handleLoginMethods)
	// 首次登录
	g.POST("/mpay/api/users/login/mobile/finish", s.handleFirstLogin)
	g.POST("/mpay/api/users/login/mobile/get_sms", s.handleFirstLogin)
	g.POST("/mpay/api/users/login/mobile/verify_sms", s.handleFirstLogin)
	g.POST("/mpay/games/:game_id/devices/:device_id/users", s.handleFirstLogin)
	// 登录
	g.GET("/mpay/games/:game_id/devices/:device_id/users/:user_id", s.handleLogin)
	// 更改审核状态
	g.GET("/mpay/games/pc_config", s.handlePcConfig)
	// 其他
	g.Any("/:path/*path", s.handleAllRest)
}

// handleAllRest 处理所有请求
func (s *Server) handleAllRest(c *gin.Context) {
	rsp, done := s.getProxyReturn(c, nil)
	if !done {
		return
	}

	s.returnByteToJson(c, rsp)
}

// handlePcConfig 更改审核状态
func (s *Server) handlePcConfig(c *gin.Context) {
	rsp, done := s.getProxyReturn(c, cvType.Ios)
	if !done {
		return
	}

	// 修改响应
	newBody := s.modifyResponse(rsp, func(newBody *map[string]interface{}) {
		if game, ok := (*newBody)["game"].(map[string]interface{}); ok {
			if config, ok := game["config"].(map[string]interface{}); ok {
				config["cv_review_status"] = 1
			}
		}
	})
	c.JSON(rsp.StatusCode, newBody)
}

// handleLogin 登录
func (s *Server) handleLogin(c *gin.Context) {
	rsp, done := s.getProxyReturn(c, cvType.Ios)
	if !done {
		return
	}

	// 修改响应
	newBody := s.modifyResponse(rsp, func(newBody *map[string]interface{}) {
		if user, ok := (*newBody)["user"].(map[string]interface{}); ok {
			user["pc_ext_info"] = constants.PcInfo
		}
	})
	c.JSON(rsp.StatusCode, newBody)
}

// handleFirstLogin 首次登录
func (s *Server) handleFirstLogin(c *gin.Context) {
	rsp, done := s.getProxyReturn(c, cvType.Ios)
	if !done {
		return
	}

	s.returnByteToJson(c, rsp)
}

// handleLoginMethods 修改登录方法
func (s *Server) handleLoginMethods(c *gin.Context) {
	rsp, done := s.getProxyReturn(c, cvType.Pc)
	if !done {
		return
	}

	// 修改响应
	newBody := s.modifyResponse(rsp, func(newBody *map[string]interface{}) {
		(*newBody)["select_platform"] = true
		(*newBody)["qrcode_select_platform"] = true
		if config, ok := (*newBody)["config"].(map[string]interface{}); ok {
			for _, v := range config {
				if configMap, ok := v.(map[string]interface{}); ok {
					configMap["select_platforms"] = []interface{}{0, 1, 2, 3, 4}
				}
			}
		}
	})
	c.JSON(rsp.StatusCode, newBody)
}

// modifyResponse 修改body
func (s *Server) modifyResponse(rsp *req.Response, callback func(body *map[string]interface{})) map[string]interface{} {
	var newBody map[string]interface{}
	json.Unmarshal(rsp.Bytes(), &newBody)

	// 处理
	callback(&newBody)
	return newBody
}

// returnByteToJson 返回byte转json
func (s *Server) returnByteToJson(c *gin.Context, rsp *req.Response) {
	var temp gin.H
	json.Unmarshal(rsp.Bytes(), &temp)
	c.JSON(rsp.StatusCode, temp)
}

// getProxyReturn 获取代理返回
func (s *Server) getProxyReturn(c *gin.Context, cv *string) (*req.Response, bool) {
	reqs := c.Request

	rsp := s.proxy(reqs, cv)
	if rsp.Err != nil {
		log.Errorf("请求失败：%v", rsp.Err)
		c.JSON(http.StatusInternalServerError, gin.H{"reason": rsp.Err.Error()})
		return rsp, false
	}
	return rsp, true
}

// proxy 代理请求
func (s *Server) proxy(r *http.Request, cv *string) *req.Response {
	client := s.client
	urlPath := r.URL.Path

	// 处理url
	reqs := client.R()
	// 设置header
	s.setReqValues(r.Header, reqs.SetHeader)

	// 判断是否覆盖cv
	if cv != nil {
		// 判断方法
		if r.Method == http.MethodGet {
			newQuery := r.URL.Query()
			newQuery.Set("cv", *cv)
			s.setReqValues(newQuery, reqs.AddQueryParam)
		} else if r.Method == http.MethodPost {
			r.ParseForm()
			r.PostForm.Set("cv", *cv)
			newBody := r.PostForm.Encode()
			reqs = reqs.SetBodyString(newBody)
		}
	}
	rsp, _ := reqs.Send(r.Method, s.urlRedirect+urlPath)

	return rsp
}

func (s *Server) checkPort() (bool, error) {
	ln, err := net.Listen("tcp", ":443")
	if err != nil {
		return false, err
	}
	defer ln.Close()
	return true, nil
}

func (s *Server) setReqValues(data map[string][]string, callback func(string, string) *req.Request) {
	for k, v := range data {
		for _, vv := range v {
			callback(k, vv)
		}
	}
}
