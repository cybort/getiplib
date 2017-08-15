package iputil

import (
	"iputil"
	"testing"
)

func TestIpMap(t *testing.T) {
	var line string = "150.242.167.250|150.242.167.255|6|中国|北京市|北京市|长城互联网|华北"
	ipmap := iputil.ConstrucIpMapFromStr(line)
	if ipmap["ip"] != "150.242.167.250" {
		t.Errorf("want 150.242.167.250 but get %s", ipmap["ip"])
	}
	if ipmap["end"] != "150.242.167.255" {
		t.Errorf("want 150.242.167.255 but get %s", ipmap["end"])
	}
	if ipmap["len"] != "6" {
		t.Errorf("want 6 but get %s", ipmap["len"])
	}
	if ipmap["country"] != "中国" {
		t.Errorf("want chinese but get %s", ipmap["country"])
	}
	if ipmap["isp"] != "长城互联网" {
		t.Errorf("want 长城互联网 but get %s", ipmap["isp"])
	}
	if ipmap["city"] != "北京市" {
		t.Errorf("want 北京市 but get %s", ipmap["city"])
	}
	if ipmap["area"] != "华北" {
		t.Errorf("want 华北 but get %s", ipmap["region"])
	}
	if ipmap["region"] != "北京市" {
		t.Errorf("want 北京市 but get %s", ipmap["region"])
	}
}

func TestIpinfoGetFromTaobao(t *testing.T) {
	taobaoUrl := "http://ip.taobao.com/service/getIpInfo.php?ip="
	posInfo, ok := iputil.ParseUrlToMap(nil, taobaoUrl, "1.1.8.0")
	if !ok {
		t.Errorf("http request failed, %d", ok)
	}
	if posInfo.Code != 0 {
		t.Errorf("response code error, expect 0 but get %d", posInfo.Code)
	}
	if posInfo.Country != "中国" {
		t.Errorf("country error, expect 中国 but get %d", posInfo.Country)
	}
	if posInfo.Region != "广东省" {
		t.Errorf("region error, expect 广东省 but get %d", posInfo.Region)
	}
	if posInfo.Isp != "电信" {
		t.Errorf("isp error, expect 电信 but get %d", posInfo.Isp)
	}
	if posInfo.Ip != "1.1.8.0" {
		t.Errorf("ip error, expect 1.1.8.0 but get %d", posInfo.Ip)
	}
}
