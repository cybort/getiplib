package iputil

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"ipconfig"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

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

func InetNtoaStr(intip int64) string {
	addr_net := InetNtoa(intip)
	return addr_net.String()
}

func Gen_end_ip(start_ip string, span int64) string {
	addr := net.ParseIP(start_ip)
	ip_int := InetAton(addr)
	ip_int += span - 1
	addr_net := InetNtoa(ip_int)
	return addr_net.String()
}
func DeepCopy(value interface{}) interface{} {
	if valueMap, ok := value.(map[string]interface{}); ok {
		newMap := make(map[string]interface{})
		for k, v := range valueMap {
			newMap[k] = DeepCopy(v)
		}

		return newMap
	} else if valueSlice, ok := value.([]interface{}); ok {
		newSlice := make([]interface{}, len(valueSlice))
		for k, v := range valueSlice {
			newSlice[k] = DeepCopy(v)
		}

		return newSlice
	}

	return value
}
func ParseUrlToMap(ip string) (map[string]string, bool) {
	t0 := time.Now()
	url := fmt.Sprintf("http://%s%s%s", ipconfig.Taobaoip[rand.Intn(10000)%2], ipconfig.UrlSuffix, ip)
	//url := fmt.Sprintf("%s%s", ipconfig.Taobao_url, ip)
	req, _ := http.NewRequest("GET", url, nil)
	req.Host = ipconfig.TaobaoHost
	resp, _ := http.DefaultClient.Do(req)
	time.Sleep(200 * time.Millisecond)
	t2 := time.Now()
	fmt.Printf("%s get took %v to run\n", url, t2.Sub(t0))
	defer resp.Body.Close()
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("!!!!!get panic info, recoverit", r)
		}
	}()
	body, _ := ioutil.ReadAll(resp.Body)
	var dat map[string]interface{}
	if err := json.Unmarshal(body, &dat); err == nil {
		md, ok := dat["data"].(map[string]interface{})
		if ok {
			rtnValue := make(map[string]string)
			for k, v := range md {
				rtnValue[k] = v.(string)
			}
			return rtnValue, true
		}

		return nil, false
	}
	return nil, false
}
func UsefulInfoForPrint(md map[string]string) string {
	address := fmt.Sprintf("%s|%s|%s|%s|%s", md["ip"], md["end"], md["country_id"], md["isp_id"], md["region_id"])
	return address
}
func Format_to_output(md map[string]string) string {
	address := fmt.Sprintf("%s|%s:%s|%s:%s|%s:%s|%s:%s|%s:%s", md["ip"], md["country"], md["country_id"], md["isp"], md["isp_id"], md["area"], md["area_id"], md["city"], md["city_id"], md["region"], md["region_id"])
	return address
}

func AllKeyInfoFormat_to_output(md map[string]string) string {
	address := fmt.Sprintf("%s|%s|%s|%s|%s:%s|%s:%s|%s:%s|%s:%s|%s:%s", md["ip"], md["end"], md["len"], md["ip"], md["country"], md["country_id"], md["isp"], md["isp_id"], md["area"], md["area_id"], md["city"], md["city_id"], md["region"], md["region_id"])
	return address
}

func ConstrucIpMapFromStr(ipinfoline string) map[string]string {
	ipinfo := strings.Split(ipinfoline, "|")
	if len(ipinfo) < 9 {
		return nil
	}
	var tempMap map[string]string = make(map[string]string, 0)

	tempMap["ip"] = ipinfo[0]
	tempMap["end"] = ipinfo[1]
	tempMap["len"] = ipinfo[2]

	tempMap["country"] = strings.Split(ipinfo[4], ":")[0]
	tempMap["isp"] = strings.Split(ipinfo[5], ":")[0]
	tempMap["area"] = strings.Split(ipinfo[6], ":")[0]
	tempMap["city"] = strings.Split(ipinfo[7], ":")[0]
	tempMap["region"] = strings.Split(strings.TrimSuffix(ipinfo[8], "\n"), ":")[0]
	tempMap["country_id"] = strings.Split(ipinfo[4], ":")[1]
	tempMap["isp_id"] = strings.Split(ipinfo[5], ":")[1]
	tempMap["area_id"] = strings.Split(ipinfo[6], ":")[1]
	tempMap["city_id"] = strings.Split(ipinfo[7], ":")[1]
	tempMap["region_id"] = strings.Split(strings.TrimSuffix(ipinfo[8], "\n"), ":")[1]

	return tempMap
}

func GetDetectedIpInfoSlice(filename string) []map[string]string {
	fp, err := os.Open(filename)
	if err != nil {
		fmt.Println("open ipinfo file failed")
		return nil
	}
	defer fp.Close()
	infoList := make([]map[string]string, 0)
	br := bufio.NewReader(fp)
	for {
		bline, err := br.ReadString('\n')
		if err != nil {
			fmt.Println("reach end of file")
			break
		}

		tempMap := ConstrucIpMapFromStr(bline)
		if tempMap == nil {
			continue
		}
		infoList = append(infoList, tempMap)
	}

	fmt.Println("total key ", len(infoList))
	return infoList
}

func GetDetectedIpInfo(filename string, infoMap map[string]interface{}) {
	fp, err := os.Open(filename)
	if err != nil {
		fmt.Println("open ipinfo file failed")
		return
	}
	defer fp.Close()
	//infoMap := make(map[string]interface{}, 1)
	br := bufio.NewReader(fp)
	for {
		bline, err := br.ReadString('\n')
		if err != nil {
			fmt.Println("reach end of file")
			break
		}
		tempMap := ConstrucIpMapFromStr(bline)
		if tempMap == nil {
			continue
		}

		if tempMap["country_id"] != "" {
			infoMap[tempMap["ip"]] = tempMap
			infoMap[tempMap["end"]] = tempMap
			//alreay1, bexist := infoMap[tempMap["ip"]]
			//if bexist == false {
			//	infoMap[tempMap["ip"]] = tempMap
			//} else {
			//	alreay := alreay1.(map[string]string)
			//	correct_ipinfomap(infoMap, alreay, tempMap)
			//}
			//alreay2, bexist2 := infoMap[tempMap["end"]]
			//if bexist2 == false {
			//	infoMap[tempMap["end"]] = tempMap
			//} else {
			//	already := alreay2.(map[string]string)
			//	correct_ipinfomap(infoMap, already, tempMap)
			//}
		} else {
			fmt.Println("no country_id", bline)
		}
	}

	fmt.Println("total key ", len(infoMap))

}

func correct_ipinfomap(infoMap map[string]interface{}, alreay, tempMap map[string]string) {
	if !same_ipmap(alreay, tempMap) {
		ip := tempMap["ip"]
		wireMap, _ := ParseUrlToMap(ip)
		if same_ipmap(tempMap, wireMap) {
			infoMap[ip] = tempMap
			fmt.Println("[alreay not correct] tempMap:", UsefulInfoForPrint(tempMap))
			fmt.Println("[alreay not correct] alreay:", UsefulInfoForPrint(alreay))
		} else if same_ipmap(alreay, wireMap) {
			//infoMap[ip] = wireMap
			fmt.Println("[==================] alreay with wire:")

		} else {
			fmt.Println("[11111]alreay in map:", UsefulInfoForPrint(alreay))
			fmt.Println("[22222]still ip found:", UsefulInfoForPrint(tempMap))
			fmt.Println("[33333]wire in map:", UsefulInfoForPrint(wireMap))
		}
	}

}
func same_ipmap(map1, map2 map[string]string) bool {
	//if map1["country_id"] == map2["country_id"] and map1["isp_id"] == map2["isp_id"] and map1["region_id"] == map2["region_id"] {

	if map1["country_id"] == map2["country_id"] {
		return true
	}
	return false
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
	country := QualifiedIpAtLevel("country_id", mipinfoMap, ipstartMap, ipendMap)
	switch country {
	case ipconfig.Goon:
		isp := QualifiedIpAtLevel("isp_id", mipinfoMap, ipstartMap, ipendMap)
		switch isp {
		case ipconfig.Goon:
			region := QualifiedIpAtLevel("region_id", mipinfoMap, ipstartMap, ipendMap)
			return region
		default:
			return isp
		}
	default:
		return country
	}
}
