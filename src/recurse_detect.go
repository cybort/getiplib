package main

import (
	"bufio"
	"fmt"
	"io"
	"ipconfig"
	"iputil"
	"os"
	"strconv"
	"strings"
	"time"
)

var fileSuffix string = ".again"
var fProblemIP string = "all/merge_break_ip.txt"

//var fResult string = "all/network_after_split.txt" + fileSuffix
var fMiddle string = "all/middle_result_store.txt" + fileSuffix
var f_breakpoint_file string = "all/breakpoint_recurse_again.txt"

var fIpInfo string = "all/same_network.txt"

//var fIpInfo string = ipconfig.F_ip_result_detected

var taobaoURL string = ipconfig.Taobao_url

const BATCH_NUM = 5

func main() {

	ipinfoMap := iputil.GetDetectedIpInfo(fIpInfo)

	resultFP, err := os.OpenFile(fIpInfo, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("open result file failed")
		return
	}
	defer resultFP.Close()

	middleFP, err := os.OpenFile(fMiddle, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("open middle file failed")
		return
	}
	defer middleFP.Close()

	fileFP, err := os.Open(fProblemIP)
	if err != nil {
		fmt.Println("file not exists")
		return
	}
	defer fileFP.Close()
	br := bufio.NewReader(fileFP)

	var fileno int = 0
	var last_fileno int = 0
	breakpoint_file, temp_err := os.OpenFile(f_breakpoint_file, os.O_RDWR, 0666)
	if temp_err != nil {
		breakpoint_file, temp_err = os.OpenFile(f_breakpoint_file, os.O_RDWR|os.O_CREATE, 0666)
		if temp_err != nil {
			fmt.Printf("create file %s failed\n", f_breakpoint_file)
			return
		}
		last_fileno = 0
		fmt.Printf("break point file not exists, create it and set lineno =0\n")
	} else {
		buf := make([]byte, 512)
		n, _ := breakpoint_file.Read(buf)
		s := string(buf[:n])
		s = strings.TrimRight(s, "\n")
		var e error
		last_fileno, e = strconv.Atoi(s)
		if e != nil {
			println("eorr atoi", e)
		}

		fmt.Printf("last detect to lineno: %d\n ", last_fileno)
	}

	defer breakpoint_file.Close()

	var batch_iplist [BATCH_NUM]map[string]string
	var detect_count int = 0
	for {
		bline, isPrefix, err := br.ReadLine()
		if err != nil {
			if err == io.EOF {
				fmt.Println("reach EOF, detect completed")
			}
			fmt.Println("read line failed")
			return
		}

		if isPrefix != false {
			return
		}

		fileno += 1
		fmt.Println("fileno", fileno)
		if fileno <= last_fileno {
			continue
		}

		line := string(bline)
		arr := strings.Split(line, "|")
		startip := arr[0]
		endip := strings.TrimSuffix(arr[1], "\n")
		//CalcuAndSplit(startip, endip, ipinfoMap, resultFP, middleFP)
		linenum := strconv.Itoa(fileno)
		tmp := map[string]string{
			"startip":  startip,
			"endip":    endip,
			"fileline": linenum,
		}
		batch_iplist[detect_count] = tmp
		if detect_count == BATCH_NUM-1 {
			batch_detect(batch_iplist[0:], ipinfoMap, resultFP, middleFP)
			detect_count = 0
			breakpoint_file.Truncate(0)
			breakpoint_file.Seek(0, 0)
			fmt.Fprintf(breakpoint_file, "%d", fileno)
			time.Sleep(1 * time.Second)
		} else {
			detect_count += 1
		}
	}

	if detect_count > 0 {
		lineno := batch_detect(batch_iplist[0:detect_count], ipinfoMap, resultFP, middleFP)
		breakpoint_file.Truncate(0)
		breakpoint_file.Seek(0, 0)
		fmt.Fprintf(breakpoint_file, "%d", lineno)
	}

}

func batch_detect(linelist []map[string]string, ipinfoMap map[string]interface{}, resultFP, middleresultFP *os.File) int {

	resultlist := make([]chan int, 0, len(linelist))
	fmt.Println("-----------------batch start-----------")
	for _, ipinfo := range linelist {
		startip := ipinfo["startip"]
		endip := ipinfo["endip"]
		fileline, _ := strconv.Atoi(ipinfo["fileline"])
		result := start_recurse_detect(startip, endip, ipinfoMap, resultFP, middleresultFP, fileline)
		resultlist = append(resultlist, result)
	}

	var fileno int
	for _, result_ch := range resultlist {
		fileno = <-result_ch
		fmt.Println(" detect line no -----> ", fileno)
	}
	fmt.Println("-----------------batch end-----------")

	return fileno
}

func start_recurse_detect(startip, endip string, ipinfoMap map[string]interface{}, resultFP, middleFP *os.File, fileline int) chan int {
	result := make(chan int, 1)
	go func(startip, endip string, ipinfoMap map[string]interface{}, resultFP, middleFP *os.File, fileline int) {
		CalcuAndSplit(startip, endip, ipinfoMap, resultFP, middleFP)
		result <- fileline
	}(startip, endip, ipinfoMap, resultFP, middleFP, fileline)

	return result
}

func CalcuAndSplit(startip, endip string, ipinfoMap map[string]interface{}, resultFP, middleresultFP *os.File) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("get panic when detected", r)
		}
	}()
	var startipMap map[string]string
	var endipMap map[string]string

	fmt.Println("-------------------------------------------------")
	fmt.Println("startip|endip|" + startip + "|" + endip)
	fmt.Println("-------------------------------------------------")

	info1, b1 := ipinfoMap[startip]
	if b1 == false {
		url1 := fmt.Sprintf("%s%s", taobaoURL, startip)
		startipMap, _ = iputil.ParseUrlToMap(url1)
		ipinfoMap[startip] = startipMap
		result1 := iputil.Format_to_output(startipMap)
		middleresultFP.WriteString(startip + "|" + startip + "|1|" + result1 + "\n")
	} else {
		startipMap = info1.(map[string]string)
	}

	if startip == endip {
		SaveSameNetwork(startip, endip, ipinfoMap[startip], resultFP)
		return
	}

	info2, b2 := ipinfoMap[endip]
	if b2 == false {
		url2 := fmt.Sprintf("%s%s", taobaoURL, endip)
		endipMap, _ = iputil.ParseUrlToMap(url2)
		ipinfoMap[endip] = endipMap
		result2 := iputil.Format_to_output(endipMap)
		middleresultFP.WriteString(endip + "|" + endip + "|1|" + result2 + "\n")
	} else {
		endipMap = info2.(map[string]string)
	}

	ip1 := iputil.InetAtonInt(startip)
	ip2 := iputil.InetAtonInt(endip)

	if ip1 < ip2 {
		m := (ip1 + ip2) / 2
		ip1_str := iputil.InetNtoaStr(ip1)
		ip2_str := iputil.InetNtoaStr(ip2)
		mip := iputil.InetNtoaStr(m)
		mip_rfirst := iputil.InetNtoaStr(m + 1)
		fmt.Println("start|middle-ip|end", ip1_str, mip, ip2_str)
		url1 := fmt.Sprintf("%s%s", taobaoURL, mip)
		url2 := fmt.Sprintf("%s%s", taobaoURL, mip_rfirst)
		var mipinfo1 map[string]string

		mipInfo1, exist1 := ipinfoMap[mip]
		if exist1 == false {
			mipinfo1, _ = iputil.ParseUrlToMap(url1)
			ipinfoMap[mip] = mipinfo1
			/*store middle detect result*/
			result1 := iputil.Format_to_output(mipinfo1)
			middleresultFP.WriteString(mip + "|" + mip + "|1|" + result1 + "\n")
			middleresultFP.Sync()
		} else {
			mipinfo1 = mipInfo1.(map[string]string)
		}

		_, exist2 := ipinfoMap[mip_rfirst]
		if exist2 == false {
			mipinfo2, _ := iputil.ParseUrlToMap(url2)
			ipinfoMap[mip_rfirst] = mipinfo2
			result2 := iputil.Format_to_output(mipinfo2)
			middleresultFP.WriteString(mip_rfirst + "|" + mip_rfirst + "|1|" + result2 + "\n")
			middleresultFP.Sync()
		}

		var finded string
		//fmt.Println("+++ipinfo1", mipinfo1)
		//fmt.Println("+++startmap", startipMap)
		//fmt.Println("+++endipmap", endipMap)
		finded = iputil.QualifiedIpAtLevel("country", mipinfo1, startipMap, endipMap)
		//fmt.Println("country finded:", finded)
		switch finded {
		case ipconfig.Goon:
			finded = iputil.QualifiedIpAtLevel("isp", mipinfo1, startipMap, endipMap)
			//fmt.Println("isp finded:", finded)
			switch finded {
			case ipconfig.Goon:
				finded = iputil.QualifiedIpAtLevel("region", mipinfo1, startipMap, endipMap)
				//fmt.Println("province finded:", finded)
				switch finded {
				case ipconfig.Goon:
					fmt.Println("this is same network:", ip1_str, ip2_str)
					SaveSameNetwork(ip1_str, ip2_str, ipinfoMap[ip1_str], resultFP)
				case ipconfig.Leftmove:
					SaveSameNetwork(ip1_str, mip, ipinfoMap[ip1_str], resultFP)
					CalcuAndSplit(mip_rfirst, ip2_str, ipinfoMap, resultFP, middleresultFP)
				case ipconfig.Rightmove:
					SaveSameNetwork(mip_rfirst, ip2_str, ipinfoMap[mip_rfirst], resultFP)
					CalcuAndSplit(ip1_str, mip, ipinfoMap, resultFP, middleresultFP)
				case ipconfig.Morenetwork:
					CalcuAndSplit(ip1_str, mip, ipinfoMap, resultFP, middleresultFP)
					CalcuAndSplit(mip_rfirst, ip2_str, ipinfoMap, resultFP, middleresultFP)
				}

			case ipconfig.Leftmove:
				SaveSameNetwork(ip1_str, mip, ipinfoMap[ip1_str], resultFP)
				CalcuAndSplit(mip_rfirst, ip2_str, ipinfoMap, resultFP, middleresultFP)
			case ipconfig.Rightmove:
				SaveSameNetwork(mip_rfirst, ip2_str, ipinfoMap[mip_rfirst], resultFP)
				CalcuAndSplit(ip1_str, mip, ipinfoMap, resultFP, middleresultFP)
			case ipconfig.Morenetwork:
				CalcuAndSplit(ip1_str, mip, ipinfoMap, resultFP, middleresultFP)
				CalcuAndSplit(mip_rfirst, ip2_str, ipinfoMap, resultFP, middleresultFP)
			}

		case ipconfig.Leftmove:
			SaveSameNetwork(ip1_str, mip, ipinfoMap[ip1_str], resultFP)
			CalcuAndSplit(mip_rfirst, ip2_str, ipinfoMap, resultFP, middleresultFP)
		case ipconfig.Rightmove:
			SaveSameNetwork(mip_rfirst, ip2_str, ipinfoMap[mip_rfirst], resultFP)
			CalcuAndSplit(ip1_str, mip, ipinfoMap, resultFP, middleresultFP)
		case ipconfig.Morenetwork:
			CalcuAndSplit(ip1_str, mip, ipinfoMap, resultFP, middleresultFP)
			CalcuAndSplit(mip_rfirst, ip2_str, ipinfoMap, resultFP, middleresultFP)
		}

	}
}

func SaveSameNetwork(startip, endip string, ipinfoMap interface{}, fileFP *os.File) {
	ipmap := ipinfoMap.(map[string]string)
	if ipmap == nil {
		url := fmt.Sprintf("%s%s", taobaoURL, endip)
		ipmap, _ = iputil.ParseUrlToMap(url)
	}
	lens := iputil.InetAtonInt(endip) - iputil.InetAtonInt(startip) + 1
	result := iputil.Format_to_output(ipmap)
	fileFP.WriteString(startip + "|" + endip + "|" + strconv.Itoa(int(lens)) + "|" + result + "\n")
	fileFP.Sync()
}
