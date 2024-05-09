package constants

const (
	CaPath    = "./idv_ca.pem"
	CertPath  = "./idv_cert.pem"
	KeyPath   = "./idv_key.pem"
	IpHost    = "https://www.ip.cn/api/index"
	Icv       = "i3.15.0"
	Pcv       = "p3.15.0"
	Ccv       = "c3.15.0"
	Localhost = "127.0.0.1"
)

var (
	PcInfo = map[string]interface{}{
		"extra_unisdk_data": "",
		"from_game_id":      "h55",
		"src_app_channel":   "netease",
		"src_client_ip":     "",
		"src_client_type":   1,
		"src_jf_game_id":    "h55",
		"src_pay_channel":   "netease",
		"src_sdk_version":   "3.15.0",
		"src_udid":          "",
	}
	DebugMode = false
)
