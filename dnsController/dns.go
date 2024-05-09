package dnsController

import (
	"github.com/imroc/req/v3"
	"idv-login-go/config"
	"idv-login-go/constants"
)

type DnsController struct {
	dnsHost string
	params  map[string]string
	client  *req.Client
}

var conf *config.Config

func NewDnsController() *DnsController {
	conf = config.GetConfig()
	dC := &DnsController{
		dnsHost: conf.String("hostDNS"),
		params: map[string]string{
			"name":               conf.String("host"),
			"short":              "true",
			"edns_client_subnet": "",
		},
		client: req.C(),
	}
	// 获取本地IP
	var ips struct {
		Ip string `json:"ip"`
	}
	resp, _ := dC.client.R().SetQueryParam("type", "0").
		SetSuccessResult(&ips).
		Get(constants.IpHost)

	if resp.IsSuccessState() {
		dC.params["edns_client_subnet"] = ips.Ip
	}

	return dC
}

func (d *DnsController) Resolve() (string, error) {
	var ips []string
	err := d.client.Get(d.dnsHost).
		SetQueryParams(d.params).
		Do().
		Into(&ips)

	if err != nil {
		return "", err
	}

	return ips[0], nil
}
