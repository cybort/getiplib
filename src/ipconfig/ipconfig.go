package ipconfig

var F_ip_section_file string = "all/ipquery_range.txt"
var F_exception = "etc/exception_line.txt"
var F_not_same = "all/ip_not_the_same_network.txt"
var F_ip_result_detected string = "all/network_detected.txt"
var F_breakpoint_file string = "etc/breakpoint.info"
var Taobao_url string = "http://ip.taobao.com/service/getIpInfo.php?ip="

var Taobaoip = [2]string{"140.205.140.33", "42.120.226.92"}

const UrlSuffix = "/service/getIpInfo.php?ip="
const TaobaoHost = "ip.taobao.com"

const (
	Goon        = "same network"
	Leftmove    = "left equal, left move to right"
	Rightmove   = "right equal, right move to left"
	Morenetwork = "!!!more network!!!"
)
