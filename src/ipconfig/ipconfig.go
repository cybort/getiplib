package ipconfig

var (
	FILESUFFIX                = ".temp"
	FileNeedIpsectionCheck    = "all/need_checking.ipsection" + FILESUFFIX
	FileVerifiedSameIpsection = "all/verified_result.ipsection" + FILESUFFIX
	FileMiddleDetectedResult  = "all/middle_result_store.txt" + FILESUFFIX
	FileNotSameIpsection      = "all/not_same.ipsection" + FILESUFFIX
	FileInvalidResIpsection   = "all/invalid_detected.ipsection" + FILESUFFIX
	FileBreakpoint            = "all/breakpoint.info" + FILESUFFIX
	TaobaoUrl                 = "http://ip.taobao.com/service/getIpInfo.php?ip="
)

const (
	Goon        = "same network"
	Leftmove    = "left equal, left move to right"
	Rightmove   = "right equal, right move to left"
	Morenetwork = "!!!more network!!!"
	BATCHNUM    = 2
)
