package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"ipconfig"
	"iputil"
	"net/http"
	"testing"
)

func TestTaobaoIp(t *testing.T) {
	ip := "203.119.80.11"
	url := fmt.Sprintf("%s%s", ipconfig.Taobao_url, ip)
	req, _ := http.NewRequest("GET", url, nil)
	req.Host = ipconfig.TaobaoHost
	resp, _ := http.DefaultClient.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	dat := iputil.AliIp{}
	if err := json.Unmarshal(body, &dat); err == nil {
		fmt.Printf("%+v\n", dat)
		fmt.Printf("%s\n", dat.Ip)
		fmt.Printf("%s\n", dat.Country)

	} else {
		fmt.Println(err)
	}
}

func invalidtestCountry(t *testing.T) {
	var bline_us string = "150.242.5.255|150.242.5.255|1|150.242.5.255|中国:CN|中国科技网:1000114|华北:100000|北京市:110100|北京市:110000"
	var bline_nan1 string = "150.242.6.255|150.242.6.255|1|150.242.6.255|中国:CN|中国电信:1000115|华北:100000|北京市:110100|天津市:110002"
	var bline_au string = "150.242.7.255|150.242.7.255|1|150.242.7.255|中国:CN|中国联通:1000115|华北:100000|北京市:110100|北京市:110000"
	var bline_nan2 string = "150.242.8.255|150.242.8.255|1|150.242.8.255|中国:CN|中国电信:1000115|华北:100000|北京市:110100|天津市:110002"
	var bline_nan3 string = "150.242.9.255|150.242.9.255|1|150.242.8.255|中国:CN|中国电信:1000115|华北:100000|北京市:110100|天津市:110002"

	//ip_us := "150.242.5.255"
	//ip_nan1 := "150.242.6.255"
	//ip_au := "150.242.7.255"
	//ip_nan2 := "150.242.8.255"
	ipmap_us := iputil.ConstrucIpMapFromStr(bline_us)
	ipmap_nan1 := iputil.ConstrucIpMapFromStr(bline_nan1)
	ipmap_au := iputil.ConstrucIpMapFromStr(bline_au)
	ipmap_nan2 := iputil.ConstrucIpMapFromStr(bline_nan2)
	ipmap_nan3 := iputil.ConstrucIpMapFromStr(bline_nan3)

	ret := iputil.QualifiedIpAtLevel("country", ipmap_us, ipmap_nan1, ipmap_au)
	if ret != ipconfig.Goon {
		t.Errorf("want continue detect , but get %s", ret)
	}
	ret = iputil.QualifiedIpAtLevel("isp", ipmap_nan2, ipmap_nan1, ipmap_au)
	if ret != ipconfig.Leftmove {
		t.Errorf("want %s , but get %s", ipconfig.Leftmove, ret)
	}
	ret = iputil.QualifiedIpAtLevel("isp", ipmap_nan2, ipmap_nan1, ipmap_au)
	if ret != ipconfig.Leftmove {
		t.Errorf("want %s , but get %s", ipconfig.Leftmove, ret)
	}
	ret = iputil.QualifiedIpAtLevel("isp", ipmap_nan2, ipmap_nan1, ipmap_nan3)
	if ret != ipconfig.Goon {
		t.Errorf("want %s , but get %s", ipconfig.Goon, ret)
	}

	iputil.QualifiedIpAtRegion(ipmap_nan2, ipmap_nan1, ipmap_nan3)
	if ret != ipconfig.Goon {
		t.Errorf("want %s , but get %s", ipconfig.Goon, ret)
	}

	ret = iputil.QualifiedIpAtRegion(ipmap_nan2, ipmap_nan1, ipmap_au)
	if ret != ipconfig.Leftmove {
		t.Errorf("want %s , but get %s", ipconfig.Leftmove, ret)
	}
	ret = iputil.QualifiedIpAtRegion(ipmap_nan2, ipmap_au, ipmap_nan1)
	if ret != ipconfig.Rightmove {
		t.Errorf("want %s , but get %s", ipconfig.Rightmove, ret)
	}
	ret = iputil.QualifiedIpAtRegion(ipmap_nan2, ipmap_au, ipmap_us)
	if ret != ipconfig.Morenetwork {
		t.Errorf("want %s , but get %s", ipconfig.Rightmove, ret)
	}

}
