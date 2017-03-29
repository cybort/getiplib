package iputil

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"ipconfig"
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

func ParseUrlToMap(url string) (map[string]string, bool) {
	t0 := time.Now()
	response, _ := http.Get(url)
	t2 := time.Now()
	fmt.Printf("http get took %v to run\n", t2.Sub(t0))
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("!!!!!get panic info, recoverit", r)
		}
	}()
	body, _ := ioutil.ReadAll(response.Body)
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

func GetDetectedIpInfo(filename string) map[string]interface{} {
	fp, err := os.Open(filename)
	if err != nil {
		fmt.Println("open ipinfo file failed")
		return nil
	}
	defer fp.Close()
	infoMap := make(map[string]interface{}, 1)
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

		infoMap[tempMap["ip"]] = tempMap
	}

	fmt.Println("total key ", len(infoMap))
	return infoMap
}
func QualifiedIpAtLevel(level string, mipinfoMap, ipstartMap, ipendMap map[string]string) string {
	ipl := mipinfoMap[level]
	start := ipstartMap[level]
	end := ipendMap[level]
	if ipl == start && ipl == end {
		return ipconfig.Goon
	} else if ipl == start && ipl != end {
		return ipconfig.Leftmove
	} else if ipl != start && ipl == end {
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
		default:
			return isp
		}
	default:
		return country
	}
}
