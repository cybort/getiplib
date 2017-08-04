package ipconfig

var F_need_check_ipsection string = "all/need_checking.ipsection"

var F_verified_same_ipsection = "all/verified_result.ipsection"
var F_Middle string = "all/middle_result_store.txt"

var F_not_same_ipsection = "all/not_same.ipsection"

var F_breakpoint_file string = "all/breakpoint.info"

var Taobao_url string = "http://ip.taobao.com/service/getIpInfo.php?ip="

const BATCH_NUM = 10

var Taobaoip = [2]string{"140.205.140.33", "42.120.226.92"}
var Alins_arr = [3]string{"203.107.0.208", "121.43.18.42", "101.200.28.73"}

const UrlSuffix = "/service/getIpInfo.php?ip="
const TaobaoHost = "ip.taobao.com"

const CheckType = ".dns"

const (
	Goon        = "same network"
	Leftmove    = "left equal, left move to right"
	Rightmove   = "right equal, right move to left"
	Morenetwork = "!!!more network!!!"
)
