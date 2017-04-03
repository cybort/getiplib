package main

import (
	"bufio"
	"fmt"
	"io"
	"ipconfig"
	"iputil"
	"math/rand"
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

	rand.Seed(time.Now().UnixNano())
	ipinfoMap := make(map[string]interface{})
	iputil.GetDetectedIpInfo(fIpInfo, ipinfoMap)
	iputil.GetDetectedIpInfo(fMiddle, ipinfoMap)

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

		//fmt.Printf("last detect to lineno: %d\n ", last_fileno)
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
			//time.Sleep(1 * time.Second)
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
		CalcuAndSplit(startip, endip, ipinfoMap, resultFP, middleFP, 1)
		result <- fileline
	}(startip, endip, ipinfoMap, resultFP, middleFP, fileline)

	return result
}

func CalcuAndSplit(startip, endip string, ipinfoMap map[string]interface{}, resultFP, middleresultFP *os.File, depth int) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("get panic when detected", r, startip, endip)
		}
	}()
	var startipMap map[string]string
	var endipMap map[string]string

	var prefix string
	for i := 0; i <= depth; i++ {
		prefix = prefix + "+"
	}
	prefix = prefix + "|" + strconv.Itoa(depth)
	fmt.Println(prefix + "|startip|endip|" + startip + "|" + endip)

	info1, b1 := ipinfoMap[startip]
	if b1 == false {
		startipMap, _ = iputil.ParseUrlToMap(startip)
		ipinfoMap[startip] = startipMap
		result1 := iputil.Format_to_output(startipMap)
		middleresultFP.WriteString(startip + "|" + startip + "|1|" + result1 + "\n")
	} else {
		startipMap = info1.(map[string]string)
	}

	if startip == endip {
		SaveSameNetwork(startip, endip, startipMap, resultFP)
		return
	}

	info2, b2 := ipinfoMap[endip]
	if b2 == false {
		endipMap, _ = iputil.ParseUrlToMap(endip)
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
		mip := iputil.InetNtoaStr(m)
		mip_left := iputil.InetNtoaStr(m - 1)
		mip_rfirst := iputil.InetNtoaStr(m + 1)
		fmt.Println(prefix+"|start|middle-ip|end", startip, mip, endip)
		var mipinfoMap map[string]string

		mipInfo1, exist1 := ipinfoMap[mip]
		if exist1 == false {
			mipinfoMap, _ = iputil.ParseUrlToMap(mip)
			ipinfoMap[mip] = mipinfoMap
			/*store middle detect result*/
			result1 := iputil.Format_to_output(mipinfoMap)
			middleresultFP.WriteString(mip + "|" + mip + "|1|" + result1 + "\n")
			middleresultFP.Sync()
		} else {
			mipinfoMap = mipInfo1.(map[string]string)
		}

		startinfo := iputil.UsefulInfoForPrint(startipMap)
		midinfo := iputil.UsefulInfoForPrint(mipinfoMap)
		endinfo := iputil.UsefulInfoForPrint(endipMap)
		fmt.Println("--start|mid|end--", startinfo, midinfo, endinfo)

		finded := iputil.QualifiedIpAtRegion(mipinfoMap, startipMap, endipMap)
		fmt.Println("[detected result:]", finded)
		switch finded {
		case ipconfig.Goon:
			SaveSameNetwork(startip, endip, startipMap, resultFP)
		case ipconfig.Leftmove:
			SaveSameNetwork(startip, mip, startipMap, resultFP)
			CalcuAndSplit(mip_rfirst, endip, ipinfoMap, resultFP, middleresultFP, depth+1)
		case ipconfig.Rightmove:
			SaveSameNetwork(mip, endip, mipinfoMap, resultFP)
			CalcuAndSplit(startip, mip_left, ipinfoMap, resultFP, middleresultFP, depth+1)
		case ipconfig.Morenetwork:
			CalcuAndSplit(startip, mip, ipinfoMap, resultFP, middleresultFP, depth+1)
			CalcuAndSplit(mip_rfirst, endip, ipinfoMap, resultFP, middleresultFP, depth+1)
		}

	} else {
		fmt.Println("ip1 > ip2 , network split failed!!!")
	}
}

func SaveSameNetwork(startip, endip string, ipinfoMap interface{}, fileFP *os.File) {
	fmt.Println("---!!!this is same network!!!---:", startip, endip)
	ipmap := ipinfoMap.(map[string]string)
	if ipmap == nil {
		ipmap, _ = iputil.ParseUrlToMap(endip)
	}
	lens := iputil.InetAtonInt(endip) - iputil.InetAtonInt(startip) + 1
	result := iputil.Format_to_output(ipmap)
	fileFP.WriteString(startip + "|" + endip + "|" + strconv.Itoa(int(lens)) + "|" + result + "\n")
	fileFP.Sync()
}
