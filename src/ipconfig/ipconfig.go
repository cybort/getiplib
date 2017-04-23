package ipconfig

var F_ip_section_file string = "all/merge_break_ip.txt"
var F_same_ip = "all/ip_in_the_same_network.txt"
var F_ip_result_detected string = "all/network_detected.txt"
var F_breakpoint_file string = "all/breakpoint.info"
var F_already_IpInfo string = "all/uniq_ipinfo.txt.valid"
var F_Middle string = "all/middle_result_store.txt.again"

var Taobao_url string = "http://ip.taobao.com/service/getIpInfo.php?ip="

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
