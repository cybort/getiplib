package iputil

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/apsdehal/go-logger"
	"io/ioutil"
	"ipconfig"
	//"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func init() {
}

type AliIp struct {
	Code   int `json:"code"`
	IpInfo `json:"data"`
}

type IpInfo struct {
	Ip        string `json:"ip"`
	Country   string `json:"country"`
	CountryId string `json:"country_id"`
	Area      string `json:"area"`
	AreaId    string `json:"area_id"`
	Region    string `json:"region"`
	RegionId  string `json:"region_id"`
	Isp       string `json:"isp"`
	IspId     string `json:"isp_id"`
	City      string `json:"city"`
	CityId    string `json:"city_id"`
	County    string `json:"county"`
	CountyId  string `json:"county_id"`
}

// Convert uint to net.IP http://www.sharejs.com
func InetNtoa(ipnr int64) net.IP {
	var bytes [4]byte
	bytes[0] = byte(ipnr & 0xFF)
	bytes[1] = byte((ipnr >> 8) & 0xFF)
	bytes[2] = byte((ipnr >> 16) & 0xFF)
	bytes[3] = byte((ipnr >> 24) & 0xFF)

	return net.IPv4(bytes[3], bytes[2], bytes[1], bytes[0])
}

func InetAtonInt(ip string) int64 {
	addr := net.ParseIP(ip)
	return InetAton(addr)
}

// Convert net.IP to int64 ,  http://www.sharejs.com
func InetAton(ipnr net.IP) int64 {
	bits := strings.Split(ipnr.String(), ".")

	b0, _ := strconv.Atoi(bits[0])
	b1, _ := strconv.Atoi(bits[1])
	b2, _ := strconv.Atoi(bits[2])
	b3, _ := strconv.Atoi(bits[3])

	var sum int64

	sum += int64(b0) << 24
	sum += int64(b1) << 16
	sum += int64(b2) << 8
	sum += int64(b3)

	return sum
}

func IpRangeToCidr(startip, endip string) string {
	return "1"
}

func InetNtoaStr(intip int64) string {
	addr_net := InetNtoa(intip)
	return addr_net.String()
}

func GenEndIp(start_ip string, span int64) string {
	addr := net.ParseIP(start_ip)
	ip_int := InetAton(addr)
	ip_int += span - 1
	addr_net := InetNtoa(ip_int)
	return addr_net.String()
}
func DeepCopy(dst, src interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}

func ParseUrlToMap(log *logger.Logger, taobaoUrl, ip string) (map[string]string, bool) {
	url := fmt.Sprintf("%s%s", taobaoUrl, ip)
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, false
	}
	defer resp.Body.Close()
	defer func() {
		if r := recover(); r != nil {
			log.ErrorF("!!!!!get panic info, recoverit %s", r)
		}
	}()
	body, _ := ioutil.ReadAll(resp.Body)
	var dat map[string]interface{}
	if err := json.Unmarshal(body, &dat); err == nil {
		md, ok := dat["data"].(map[string]interface{})
		if ok {
			rtnValue := make(map[string]string)
			for k, v := range md {
				vv := v.(string)
				if vv == "" {
					rtnValue[k] = "*"
				} else {
					rtnValue[k] = vv
				}
			}
			return rtnValue, true
		}

		return nil, false
	}
	return nil, false
}
func UsefulInfoForPrint(md map[string]string) string {
	return fmt.Sprintf("%s|%s|%s|%s", md["country"], md["region"], md["city"], md["isp"])
}

func Format2Output(md map[string]string) string {
	//210.78.22.0|210.78.22.255|210.78.22.0/24|中国 上海市 上海市 联通 华东
	return fmt.Sprintf("%s|%s|%s|%s|%s", md["country"], md["region"], md["city"], md["isp"], md["area"])
}

func AllKeyInfoFormat_to_output(md map[string]string) string {
	suffix := Format2Output(md)
	return fmt.Sprintf("%s %s %s %s", md["ip"], md["end"], md["len"], suffix)
}

func ConstrucIpMapFromStr(ipinfoline string) map[string]string {
	ipinfo := strings.Split(ipinfoline, "|")
	if len(ipinfo) < 8 {
		return nil
	}
	var tempMap map[string]string = make(map[string]string, 0)

	tempMap["ip"] = ipinfo[0]
	tempMap["end"] = ipinfo[1]
	tempMap["cidr"] = ipinfo[2]

	if ipinfo[3] == "" {
		tempMap["country"] = "*"
	} else {
		tempMap["country"] = ipinfo[3]
	}
	if ipinfo[4] == "" {
		tempMap["region"] = "*"
	} else {
		tempMap["region"] = ipinfo[4]
	}
	if ipinfo[5] == "" {
		tempMap["city"] = "*"
	} else {
		tempMap["city"] = ipinfo[5]
	}
	if ipinfo[6] == "" {
		tempMap["isp"] = "*"
	} else {
		tempMap["isp"] = ipinfo[6]
	}
	area := strings.TrimSuffix(ipinfo[7], "\n")
	if area == "" {
		tempMap["area"] = "*"
	} else {
		tempMap["area"] = area
	}
	tempMap["len"] = strconv.FormatInt(InetAtonInt(ipinfo[1])-InetAtonInt(ipinfo[0])+1, 10)

	return tempMap
}

func GetDetectedIpInfoSlice(filename string, log *logger.Logger) []map[string]string {
	fp, err := os.Open(filename)
	if err != nil {
		fmt.Println("open ipinfo file failed")
		return nil
	}
	defer fp.Close()
	infoMap := make(map[string]interface{})
	infoList := make([]map[string]string, 0)
	br := bufio.NewReader(fp)
	for {
		bline, err := br.ReadString('\n')
		if err != nil {
			log.Debug("reach end of file")
			break
		}

		tempMap := ConstrucIpMapFromStr(bline)
		if tempMap == nil {
			continue
		}
		if tempMap["country"] != "" {
			exMap, exists := infoMap[tempMap["ip"]]
			if !exists {
				infoList = append(infoList, tempMap)
				infoMap[tempMap["ip"]] = tempMap
			} else {
				exMap1 := exMap.(map[string]string)
				curlen, _ := strconv.Atoi(tempMap["len"])
				exlen, _ := strconv.Atoi(exMap1["len"])
				if curlen < exlen {
					infoMap[tempMap["ip"]] = tempMap
				} else {
					log.ErrorF("ip %s is repeated and range %d big", tempMap["ip"], curlen-exlen)
				}
			}
		} else {
			log.DebugF("no country %s", bline)
		}
	}

	log.InfoF("total key %d", len(infoList))
	return infoList
}

func GetDetectedIpInfo(log *logger.Logger, filename string, infoMap map[string]interface{}) {
	fp, err := os.Open(filename)
	if err != nil {
		log.CriticalF("open ipinfo file %s failed", filename)
		return
	}
	defer fp.Close()
	br := bufio.NewReader(fp)
	for {
		bline, err := br.ReadString('\n')
		if err != nil {
			log.DebugF("reach end of file: %s", filename)
			break
		}
		tempMap := ConstrucIpMapFromStr(bline)
		if tempMap == nil {
			continue
		}

		if tempMap["country"] != "" && tempMap["isp"] != "*" && tempMap["region"] != "*" {
			_, exists := infoMap[tempMap["ip"]]
			if !exists {
				infoMap[tempMap["ip"]] = tempMap
				infoMap[tempMap["end"]] = tempMap
			}
		} else if tempMap["country"] != "中国" { //foreign country
			_, exists := infoMap[tempMap["ip"]]
			if !exists {
				infoMap[tempMap["ip"]] = tempMap
				infoMap[tempMap["end"]] = tempMap
			}
		} else {
			log.InfoF("no country or isp or region %s", bline)
		}
	}

	log.InfoF("total key %d", len(infoMap))

}

func QualifiedIpAtLevel(level string, mipinfoMap, ipstartMap, ipendMap map[string]string) string {
	ipm := mipinfoMap[level]
	start := ipstartMap[level]
	end := ipendMap[level]
	if ipm == start && ipm == end {
		return ipconfig.Goon
	} else if ipm == start && ipm != end {
		return ipconfig.Leftmove
	} else if ipm != start && ipm == end {
		return ipconfig.Rightmove
	}
	return ipconfig.Morenetwork
}

func QualifiedIpAtRegion(mipinfoMap, ipstartMap, ipendMap map[string]string) string {
	country := QualifiedIpAtLevel("country", mipinfoMap, ipstartMap, ipendMap)
	switch country {
	case ipconfig.Goon:
		isp := QualifiedIpAtLevel("isp", mipinfoMap, ipstartMap, ipendMap)
		switch isp {
		case ipconfig.Goon:
			region := QualifiedIpAtLevel("region", mipinfoMap, ipstartMap, ipendMap)
			return region
			//switch region {
			//case ipconfig.Goon:
			//	return QualifiedIpAtLevel("city", mipinfoMap, ipstartMap, ipendMap)
			//default:
			//	return region
			//}
		default:
			return isp
		}
	default:
		return country
	}
}
