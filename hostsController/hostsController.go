package hostsController

import (
	"github.com/goodhosts/hostsfile"
	"github.com/sirupsen/logrus"
	"idv-login-go/config"
	"idv-login-go/constants"
	"idv-login-go/logger"
)

type HostsController struct {
	hosts *hostsfile.Hosts
}

var log *logrus.Logger
var conf *config.Config
var localhost = constants.Localhost
var host string

func New() *HostsController {
	log = logger.GetLogger()
	conf = config.GetConfig()
	host = conf.String("host")

	hosts, _ := hostsfile.NewHosts()
	return &HostsController{hosts: hosts}
}

func (h *HostsController) Exist() bool {
	return h.hosts.Has(localhost, host)
}

func (h *HostsController) Add() bool {
	if !h.IsWritable() {
		return false
	}
	// 添加 hosts
	err := h.hosts.Add(localhost, host)
	if err != nil {
		log.Errorf("添加 hosts 失败：%v", err)
		return false
	}

	err = h.hosts.Flush()
	if err != nil {
		log.Errorf("保存 hosts 失败：%v", err)
		return false
	}
	return true
}

func (h *HostsController) Remove() bool {
	if !h.IsWritable() {
		return false
	}
	// 移除
	err := h.hosts.Remove(localhost, host)
	if err != nil {
		log.Errorf("移除 hosts 失败：%v", err)
		return false
	}

	err = h.hosts.Flush()
	if err != nil {
		log.Errorf("保存 hosts 失败：%v", err)
		return false
	}
	return true
}

func (h *HostsController) IsWritable() bool {
	// 检查文件是否可写
	if !h.hosts.IsWritable() {
		log.Errorf("hosts 文件不可写")
		return false
	}
	return true
}
