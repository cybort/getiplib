package ipconfig

var F_ip_section_file string = "all/continue_detect_ipsection.txt"

var F_same_ip = "all/merge_result.txt"

var F_breakpoint_file string = "all/breakpoint.info"

var F_Middle string = "all/middle_result_store.txt.again"

var Taobao_url string = "http://ip.taobao.com/service/getIpInfo.php?ip="

var F_already_IpInfo string = "all/uniq_ipinfo.txt.valid"

const BATCH_NUM = 100

var Taobaoip = [2]string{"140.205.140.33", "42.120.226.92"}

const UrlSuffix = "/service/getIpInfo.php?ip="
const TaobaoHost = "ip.taobao.com"

const (
	Goon        = "same network"
	Leftmove    = "left equal, left move to right"
	Rightmove   = "right equal, right move to left"
	Morenetwork = "!!!more network!!!"
)
