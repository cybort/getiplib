package main

import (
	"bufio"
	"fmt"
	"github.com/apsdehal/go-logger"
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
var mapMutex sync.Mutex

type MySafeMap struct {
	infoMap map[string]interface{}
	Lock    sync.Mutex
}

func (msm *MySafeMap) Get(key string) (interface{}, bool) {
	msm.Lock.Lock()
	defer msm.Lock.Unlock()
	vinfo, exists := msm.infoMap[key]
	return vinfo, exists
}
func (msm *MySafeMap) Set(key string, value interface{}) {
	msm.Lock.Lock()
	msm.infoMap[key] = value
	msm.Lock.Unlock()
}

func main() {
	log, err := logger.New("test", 1, os.Stdout)
	if err != nil {
		panic(err) // Check for error
	}
	rand.Seed(time.Now().UnixNano())
	_ipinfoMap := make(map[string]interface{})
	iputil.GetDetectedIpInfo(log, ipconfig.F_Middle, _ipinfoMap)
	iputil.GetDetectedIpInfo(log, ipconfig.F_verified_same_ipsection, _ipinfoMap)
	iputil.GetDetectedIpInfo(log, ipconfig.F_need_check_ipsection, _ipinfoMap)

	var ipinfoMap MySafeMap = MySafeMap{}
	ipinfoMap.infoMap = _ipinfoMap

	resultFP, err := os.OpenFile(ipconfig.F_verified_same_ipsection, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("open result file failed")
		return
	}
	defer resultFP.Close()

	middleresultFP, err := os.OpenFile(ipconfig.F_Middle, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("open middle file failed")
		return
	}

	notsameFP, err := os.OpenFile(ipconfig.F_not_same_ipsection, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("open not same file failed")
		return
	}
	defer notsameFP.Close()

	fileFP, err := os.Open(ipconfig.F_need_check_ipsection)
	if err != nil {
		fmt.Println("file not exists")
		return
	}
	defer fileFP.Close()
	br := bufio.NewReader(fileFP)

	var fileno int = 0
	var last_fileno int = 0
	breakpointFP, temp_err := os.OpenFile(ipconfig.F_breakpoint_file, os.O_RDWR, 0666)
	if temp_err != nil {
		breakpointFP, temp_err = os.OpenFile(ipconfig.F_breakpoint_file, os.O_RDWR|os.O_CREATE, 0666)
		if temp_err != nil {
			fmt.Printf("create file %s failed\n", ipconfig.F_breakpoint_file)
			return
		}
		last_fileno = 0
		fmt.Printf("break point file not exists, create it and set lineno =0\n")
	} else {
		buf := make([]byte, 512)
		n, _ := breakpointFP.Read(buf)
		s := string(buf[:n])
		s = strings.TrimRight(s, "\n")
		var e error
		last_fileno, e = strconv.Atoi(s)
		if e != nil {
			println("eorr atoi", e)
		}

		//fmt.Printf("last detect to lineno: %d\n ", last_fileno)
	}

	defer breakpointFP.Close()

	batch_control := make(chan int, ipconfig.BATCH_NUM)
	for {
		bline, isPrefix, err := br.ReadLine()
		if err != nil {
			if err == io.EOF {
				log.Info("reach EOF, detect completed")
			}
			log.Info("read line at end")
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

		batch_control <- 1
		start_recurse_detect(startip, endip, &ipinfoMap, resultFP, middleresultFP, notsameFP, batch_control, log)
		breakpointFP.Truncate(0)
		breakpointFP.Seek(0, 0)
		log.DebugF("line no--- %d", fileno)
		fmt.Fprintf(breakpointFP, "%d", fileno)
		breakpointFP.Sync()
	}

	log.Warning("process will exit in 3 minutes...")
	time.Sleep(3 * 60 * time.Second)
}

func start_recurse_detect(startip, endip string, ipinfoMap *MySafeMap, resultFP, middleFP, notsameFP *os.File, batch_control chan int, log *logger.Logger) {
	go func(startip, endip string, ipinfoMap *MySafeMap, resultFP, middleFP, notsameFP *os.File, batch_control chan int, log *logger.Logger) {
		CalcuAndSplit(startip, endip, ipinfoMap, resultFP, middleFP, notsameFP, 1, log)
		<-batch_control
	}(startip, endip, ipinfoMap, resultFP, middleFP, notsameFP, batch_control, log)

}
func GetAndSet(log *logger.Logger, ipinfoMap *MySafeMap, ipstr string, middleFP *os.File) map[string]string {

	var ipMap map[string]string
	var _ms bool
	rtnv := make(map[string]string)
	info1, b1 := ipinfoMap.Get(ipstr)
	if b1 == false {
		ipMap, _ms = iputil.ParseUrlToMap(log, ipstr)
		if _ms {
			WriteIpinfoToFile(middleFP, ipstr, ipstr, 1, ipMap)
			ipinfoMap.Set(ipstr, ipMap)
		} else {
			log.ErrorF("http get %s from taobao iplib failed", ipstr)
			return nil
		}
	}
	mapMutex.Lock()
	defer mapMutex.Unlock()
	if b1 {
		ipMap = info1.(map[string]string)
	}
	e := iputil.DeepCopy(&rtnv, ipMap)
	if e != nil {
		fmt.Println("Deepcopy failed", ipMap)
		return nil
	}
	return rtnv
}
func CalcuAndSplit(startip, endip string, ipinfoMap *MySafeMap, resultFP, middleresultFP, notsameFP *os.File, depth int, log *logger.Logger) {
	defer func() {
		if r := recover(); r != nil {
			log.WarningF("get panic when detected %s, %s, %s", r, startip, endip)
		}
	}()
	var startipMap map[string]string
	var endipMap map[string]string

	var prefix string
	for i := 0; i <= depth; i++ {
		prefix = prefix + "+"
	}
	prefix = prefix + "|" + strconv.Itoa(depth)
	log.Debug(prefix + "|startip|endip|" + startip + "|" + endip)

	startipMap = GetAndSet(log, ipinfoMap, startip, middleresultFP)
	if startip == endip {
		SaveSameNetwork(startip, endip, startipMap, ipinfoMap, resultFP, log)
		return
	}

	endipMap = GetAndSet(log, ipinfoMap, endip, middleresultFP)
	ip1 := iputil.InetAtonInt(startip)
	ip2 := iputil.InetAtonInt(endip)

	if ip1 < ip2 {
		m := (ip1 + ip2) / 2
		mip := iputil.InetNtoaStr(m)
		var mipinfoMap map[string]string

		mipinfoMap = GetAndSet(log, ipinfoMap, mip, middleresultFP)

		startinfo := iputil.UsefulInfoForPrint(startipMap)
		midinfo := iputil.UsefulInfoForPrint(mipinfoMap)
		endinfo := iputil.UsefulInfoForPrint(endipMap)
		log.DebugF(prefix+"-start:%s-info:%s", startip, startinfo)
		log.DebugF(prefix+"-middl:%s-info:%s", mip, midinfo)
		log.DebugF(prefix+"---end:%s-info:%s", endip, endinfo)

		finded := iputil.QualifiedIpAtRegion(mipinfoMap, startipMap, endipMap)
		log.NoticeF("[detected result: %s, %s|%s|%s]", finded, startip, mip, endip)
		switch finded {
		case ipconfig.Goon:
			SaveSameNetwork(startip, endip, startipMap, ipinfoMap, resultFP, log)
		default:
			log.ErrorF("network %s|%s not in same view", startip, endip)
			notsameFP.WriteString(startinfo + "-" + endinfo + "\n")
			//CalcuAndSplit(startip, mip, ipinfoMap, resultFP, middleresultFP,notsameFP, depth+1)
			//CalcuAndSplit(mip_rfirst, endip, ipinfoMap, resultFP, middleresultFP,notsameFP, depth+1)
		}

	} else {
		log.Critical("ip1 > ip2 , network split failed!!!")
	}
}

func SaveSameNetwork(startip, endip string, oneipMap interface{}, ipinfoMap *MySafeMap, fileFP *os.File, log *logger.Logger) {
	ipmap := oneipMap.(map[string]string)
	//exists_s := ipmap["ip"] == startip
	//exists_e := ipmap["end"] == endip
	_, exists_s := ipinfoMap.Get(startip)
	_, exists_e := ipinfoMap.Get(endip)
	if exists_s && exists_e {
		lens := iputil.InetAtonInt(endip) - iputil.InetAtonInt(startip) + 1
		WriteIpinfoToFile(fileFP, startip, endip, int(lens), ipmap)
		log.NoticeF("---!!!this is same network %s|%s!!!---:", startip, endip)
	} else {
		if !exists_s {
			log.ErrorF("startip %s no location info", startip)
		}
		if !exists_e {
			log.ErrorF("endip %s no location info", endip)
		}
	}
}
func WriteIpinfoToFile(middleFP *os.File, startip, endip string, len int, ipmap map[string]string) {
	result := iputil.Format_to_output(ipmap)
	lenstr := strconv.Itoa(len)
	fileMutex.Lock()
	middleFP.WriteString(startip + "|" + endip + "|" + lenstr + "|" + result + "\n")
	fileMutex.Unlock()
	middleFP.Sync()
}
