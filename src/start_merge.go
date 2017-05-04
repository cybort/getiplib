package main

import (
	"bufio"
	"fmt"
	"ipconfig"
	"iputil"
	"os"
	"sort"
	"strconv"
	"strings"
)

var detectedIpFile string = ipconfig.F_same_ip

var sortedFile string = "all/sorted_ip.txt"

var mergedFile string = "all/merge_result_1.txt"
var breakFile string = "all/merge_break_ip.txt"

type ByIP []map[string]string

func (obj ByIP) Len() int {
	return len(obj)
}

func (obj ByIP) Swap(i, j int) {
	obj[i], obj[j] = obj[j], obj[i]
}
func (obj ByIP) Less(i, j int) bool {
	len1 := iputil.InetAtonInt(obj[i]["ip"])
	len2 := iputil.InetAtonInt(obj[j]["ip"])
	if len1 == len2 {
		len3 := iputil.InetAtonInt(obj[i]["end"])
		len4 := iputil.InetAtonInt(obj[j]["end"])
		return len3 < len4
	}
	return len1 < len2
}

func sortNetwork(filename string, sortedFile string) {
	iplist := iputil.GetDetectedIpInfoSlice(filename)
	sort.Sort(ByIP(iplist))
	resultFP, _ := os.Create(sortedFile)
	defer resultFP.Close()
	for _, v := range iplist {
		info := iputil.AllKeyInfoFormat_to_output(v)
		resultFP.WriteString(info + "\n")
	}
}

func EqualOfTwoNetwork(ipMap1, ipMap2 map[string]string) bool {
	var br bool
	if ipMap1["country"] == ipMap2["country"] && ipMap1["isp"] == ipMap2["isp"] && ipMap1["region"] == ipMap2["region"] {
		br = true
	} else {

		br = false
	}
	return br
}

func MergeIP(filename, mergedFile, breakFile string) bool {
	mergeFP, err := os.Create(mergedFile)
	if err != nil {
		fmt.Println("open file failed")
		return false
	}
	defer mergeFP.Close()

	breakFP, err := os.Create(breakFile)
	if err != nil {
		fmt.Println("open file failed")
		return false
	}
	defer breakFP.Close()

	iplist := iputil.GetDetectedIpInfoSlice(filename)
	bIntegrity := true
	current := iplist[0]
	for i, ipMap := range iplist {
		if i == 0 {
			continue
		}
		if current == nil {
			current = ipMap
			continue
		}
		if ipMap["end"] == "" {
			//ipMap["end"] = ipMap["ip"]
			//ipMap["len"] = "1"
			continue
		}
		testip1 := iputil.InetAtonInt(current["end"])
		testip2 := iputil.InetAtonInt(ipMap["ip"])
		testip1_beg := iputil.InetAtonInt(current["ip"])
		testip2_end := iputil.InetAtonInt(ipMap["end"])
		if testip1 == testip2 {
			if testip1_beg == testip1 {

				if EqualOfTwoNetwork(current, ipMap) == true {
					current = ipMap
				} else {
					newgInfo := iputil.AllKeyInfoFormat_to_output(current)
					mergeFP.WriteString(newgInfo + "\n")

					newip := iputil.InetAtonInt(ipMap["ip"]) + 1
					ipMap["ip"] = iputil.InetNtoaStr(newip)
					newlen, _ := strconv.Atoi(ipMap["len"])
					ipMap["len"] = strconv.Itoa(newlen - 1)
					current = ipMap
				}

			} else if testip2 == testip2_end {
				if EqualOfTwoNetwork(current, ipMap) == true {
					continue
				} else {

					newip := iputil.InetAtonInt(current["end"]) - 1
					current["end"] = iputil.InetNtoaStr(newip)
					newlen, _ := strconv.Atoi(ipMap["len"])
					current["len"] = strconv.Itoa(newlen - 1)

					newgInfo := iputil.AllKeyInfoFormat_to_output(current)
					mergeFP.WriteString(newgInfo + "\n")

					current = ipMap

				}

			}
		} else if testip1 < testip2 {

			newgInfo := iputil.AllKeyInfoFormat_to_output(current)
			mergeFP.WriteString(newgInfo + "\n")
			if testip1+1 < testip2 {
				sip1 := iputil.InetNtoaStr(testip1 + 1)
				sip2 := iputil.InetNtoaStr(testip2 - 1)
				breakFP.WriteString(sip1 + "|" + sip2 + "\n")

				bIntegrity = false
			}
			current = ipMap
		} else {
			if testip1 > testip2 {
				// 5, 10
				// 6 + 8
				// 6 + 12
				testip0 := iputil.InetAtonInt(current["ip"])
				ip22 := testip2
				if testip0 <= ip22-1 {
					sip1 := iputil.InetNtoaStr(testip0)
					sip2 := iputil.InetNtoaStr(ip22 - 1)
					breakFP.WriteString(sip1 + "|" + sip2 + "\n")
					bIntegrity = false
				}
				current = ipMap
			}
		}
	}
	return bIntegrity
}

func main() {
	sortNetwork(detectedIpFile, sortedFile)
	m := MergeIP(sortedFile, mergedFile, breakFile)
	fmt.Println("intergrity is: ", m)
}
