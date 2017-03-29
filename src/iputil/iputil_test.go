package iputil

import (
	"iputil"
	"testing"
)

func TestIpMap(t *testing.T) {
	var line string = "150.242.167.250|150.242.167.255|6|150.242.167.250|中国:CN|长城互联网:1000313|华北:100000|北京市:110100|北京市:110000"
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

	line = "154.16.24.56|154.16.24.61|6|154.16.24.56|澳大利亚:AU|:|:|:|:"
	ipmap = iputil.ConstrucIpMapFromStr(line)
	if ipmap["ip"] != "154.16.24.56" {
		t.Errorf("want 154.16.24.56 but get %s", ipmap["ip"])
	}
	if ipmap["end"] != "154.16.24.61" {
		t.Errorf("want 1154.16.24.61 but get %s", ipmap["end"])
	}
	if ipmap["len"] != "6" {
		t.Errorf("want 6 but get %s", ipmap["len"])
	}
	if ipmap["country"] != "澳大利亚" {
		t.Errorf("want 澳大利亚 but get %s", ipmap["country"])
	}
	if ipmap["isp"] != "" {
		t.Errorf("want '' but get %s", ipmap["isp"])
	}
	if ipmap["city"] != "" {
		t.Errorf("want '' but get %s", ipmap["city"])
	}
	if ipmap["area"] != "" {
		t.Errorf("want '' but get %s", ipmap["region"])
	}
	if ipmap["region"] != "" {
		t.Errorf("want '' but get %s", ipmap["region"])
	}
}
