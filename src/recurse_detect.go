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
	"sync"
	"time"
)

var fileMutex sync.Mutex

type MySafeMap struct {
	infoMap map[string]interface{}
	Lock    sync.Mutex
}

func (msm MySafeMap) Get(key string) (interface{}, bool) {
	msm.Lock.Lock()
	defer msm.Lock.Unlock()
	vinfo, exists := msm.infoMap[key]
	return iputil.DeepCopy(vinfo), exists
}
func (msm MySafeMap) Set(key string, value interface{}) {
	msm.Lock.Lock()
	msm.infoMap[key] = value
	msm.Lock.Unlock()
}

func main() {

	rand.Seed(time.Now().UnixNano())
	_ipinfoMap := make(map[string]interface{})
	iputil.GetDetectedIpInfo(ipconfig.F_Middle, _ipinfoMap)
	iputil.GetDetectedIpInfo(ipconfig.F_already_IpInfo, _ipinfoMap)
	iputil.GetDetectedIpInfo(ipconfig.F_same_ip, _ipinfoMap)

	var ipinfoMap MySafeMap = MySafeMap{}
	ipinfoMap.infoMap = _ipinfoMap

	resultFP, err := os.OpenFile(ipconfig.F_same_ip, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("open result file failed")
		return
	}
	defer resultFP.Close()

	middleFP, err := os.OpenFile(ipconfig.F_Middle, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("open middle file failed")
		return
	}
	defer middleFP.Close()

	fileFP, err := os.Open(ipconfig.F_ip_section_file)
	if err != nil {
		fmt.Println("file not exists")
		return
	}
	defer fileFP.Close()
	br := bufio.NewReader(fileFP)

	var fileno int = 0
	var last_fileno int = 0
	breakpoint_file, temp_err := os.OpenFile(ipconfig.F_breakpoint_file, os.O_RDWR, 0666)
	if temp_err != nil {
		breakpoint_file, temp_err = os.OpenFile(ipconfig.F_breakpoint_file, os.O_RDWR|os.O_CREATE, 0666)
		if temp_err != nil {
			fmt.Printf("create file %s failed\n", ipconfig.F_breakpoint_file)
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

	var batch_iplist [ipconfig.BATCH_NUM]map[string]string
	var detect_count int = 0
	for {
		bline, isPrefix, err := br.ReadLine()
		if err != nil {
			if err == io.EOF {
				fmt.Println("reach EOF, detect completed")
			}
			fmt.Println("read line at end")
			break
		}

		if isPrefix != false {
			return
		}

		fileno += 1
		if fileno <= last_fileno {
			continue
		}

		line := string(bline)
		arr := strings.Split(line, "|")
		startip := arr[0]
		endip := strings.TrimSuffix(arr[1], "\n")
		linenum := strconv.Itoa(fileno)
		tmp := map[string]string{
			"startip":  startip,
			"endip":    endip,
			"fileline": linenum,
		}
		batch_iplist[detect_count] = tmp
		if detect_count == ipconfig.BATCH_NUM-1 {
			batch_detect(batch_iplist[0:], ipinfoMap, resultFP, middleFP, breakpoint_file)
			detect_count = 0
			//time.Sleep(1 * time.Second)
		} else {
			detect_count += 1
		}
	}

	if detect_count > 0 {
		batch_detect(batch_iplist[0:detect_count], ipinfoMap, resultFP, middleFP, breakpoint_file)
	}

}

func batch_detect(linelist []map[string]string, ipinfoMap MySafeMap, resultFP, middleresultFP, breakpoint_file *os.File) int {

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
		breakpoint_file.Truncate(0)
		breakpoint_file.Seek(0, 0)
		fmt.Fprintf(breakpoint_file, "%d", fileno)
		breakpoint_file.Sync()
	}
	fmt.Println("-----------------batch end-----------")

	return fileno
}

func start_recurse_detect(startip, endip string, ipinfoMap MySafeMap, resultFP, middleFP *os.File, fileline int) chan int {
	result := make(chan int, 1)
	go func(startip, endip string, ipinfoMap MySafeMap, resultFP, middleFP *os.File, fileline int) {
		CalcuAndSplit(startip, endip, ipinfoMap, resultFP, middleFP, 1)
		result <- fileline
	}(startip, endip, ipinfoMap, resultFP, middleFP, fileline)

	return result
}

func CalcuAndSplit(startip, endip string, ipinfoMap MySafeMap, resultFP, middleresultFP *os.File, depth int) {
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

	info1, b1 := ipinfoMap.Get(startip)
	if b1 == false {
		startipMap, _ = iputil.ParseUrlToMap(startip)
		fileMutex.Lock()
		WriteIpinfoToFile(middleresultFP, startip, startip, 1, startipMap)
		fileMutex.Unlock()
		ipinfoMap.Set(startip, startipMap)
	} else {
		startipMap = info1.(map[string]string)
	}

	if startip == endip {
		SaveSameNetwork(startip, endip, startipMap, resultFP)
		return
	}

	info2, b2 := ipinfoMap.Get(endip)
	if b2 == false {
		endipMap, _ = iputil.ParseUrlToMap(endip)
		fileMutex.Lock()
		WriteIpinfoToFile(middleresultFP, endip, endip, 1, endipMap)
		fileMutex.Unlock()
		ipinfoMap.Set(endip, endipMap)
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

		mipInfo1, exist1 := ipinfoMap.Get(mip)
		if exist1 == false {
			mipinfoMap, _ = iputil.ParseUrlToMap(mip)
			/*store middle detect result*/
			fileMutex.Lock()
			WriteIpinfoToFile(middleresultFP, mip, mip, 1, mipinfoMap)
			fileMutex.Unlock()
			middleresultFP.Sync()
			ipinfoMap.Set(mip, mipinfoMap)
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

func SaveSameNetwork(startip, endip string, oneipMap interface{}, fileFP *os.File) {
	fmt.Println("---!!!this is same network!!!---:", startip, endip)
	ipmap := oneipMap.(map[string]string)
	if ipmap == nil {
		ipmap, _ = iputil.ParseUrlToMap(endip)
	}
	lens := iputil.InetAtonInt(endip) - iputil.InetAtonInt(startip) + 1
	fileMutex.Lock()
	WriteIpinfoToFile(fileFP, startip, endip, int(lens), ipmap)
	fileMutex.Unlock()
	fileFP.Sync()
}
func WriteIpinfoToFile(fp *os.File, startip, endip string, len int, ipmap map[string]string) {
	result := iputil.Format_to_output(ipmap)
	lenstr := strconv.Itoa(len)
	fp.WriteString(startip + "|" + endip + "|" + lenstr + "|" + result + "\n")
}
