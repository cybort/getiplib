package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/apsdehal/go-logger"
	"io"
	"ipconfig"
	"iputil"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
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

func signalListen(c chan os.Signal) {
	signal.Notify(c, syscall.SIGINT)
	signal.Notify(c, syscall.SIGTERM)
	for {
		s := <-c
		//收到信号后的处理，这里只是输出信号内容，可以做一些更有意思的事
		fmt.Println("get signal:", s)
		break
	}
}
func main() {
	log, err := logger.New("test", 1, os.Stdout)
	if err != nil {
		panic(err) // Check for error
	}
	_ipinfoMap := make(map[string]interface{})
	iputil.GetDetectedIpInfo(log, ipconfig.FileMiddleDetectedResult, _ipinfoMap)
	iputil.GetDetectedIpInfo(log, ipconfig.FileVerifiedSameIpsection, _ipinfoMap)

	var ipinfoMap MySafeMap = MySafeMap{}
	ipinfoMap.infoMap = _ipinfoMap

	resultFP, err := os.OpenFile(ipconfig.FileVerifiedSameIpsection, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("open result file failed")
		return
	}
	defer resultFP.Close()

	middleresultFP, err := os.OpenFile(ipconfig.FileMiddleDetectedResult, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("open middle file failed")
		return
	}

	notsameFP, err := os.OpenFile(ipconfig.FileNotSameIpsection, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("open not same file failed")
		return
	}
	defer notsameFP.Close()

	fileFP, err := os.Open(ipconfig.FileNeedIpsectionCheck)
	if err != nil {
		fmt.Println("file not exists")
		return
	}
	defer fileFP.Close()

	breakpointFP, lastFileNo := GetBreakpointInfo(ipconfig.FileBreakpoint)
	if breakpointFP == nil {
		log.Error("open and create breakpoint file failed")
		return
	}

	defer breakpointFP.Close()

	batchControl := make(chan int, ipconfig.BATCHNUM)

	fileno := 0
	br := bufio.NewReader(fileFP)
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
		/*if fileno less than lastFileNo, mean those line already detected*/
		if fileno <= lastFileNo {
			continue
		}

		line := string(bline)
		arr := strings.Split(line, "|")
		startip := arr[0]
		endip := strings.TrimSuffix(arr[1], "\n")

		/*control detect concurrent number*/
		batchControl <- 1
		startRecurseDetect(startip, endip, &ipinfoMap, resultFP, middleresultFP, notsameFP, batchControl, log)
		/*record current detcted line no*/
		breakpointFP.Truncate(0)
		breakpointFP.Seek(0, 0)
		fmt.Fprintf(breakpointFP, "%d", fileno)
		breakpointFP.Sync()
		log.DebugF("line no--- %d", fileno)
	}

	log.Warning("process will exit in 3 minutes...")
	time.Sleep(3 * 60 * time.Second)
}

func startRecurseDetect(startip, endip string, ipinfoMap *MySafeMap, resultFP, middleFP, notsameFP *os.File, batchControl chan int, log *logger.Logger) {
	go func(startip, endip string, ipinfoMap *MySafeMap, resultFP, middleFP, notsameFP *os.File, batchControl chan int, log *logger.Logger) {
		CalcuAndSplit(startip, endip, ipinfoMap, resultFP, middleFP, notsameFP, 1, log)
		<-batchControl
	}(startip, endip, ipinfoMap, resultFP, middleFP, notsameFP, batchControl, log)

}
func GetAndSet(log *logger.Logger, ipinfoMap *MySafeMap, ipstr string, middleFP *os.File) (map[string]string, error) {
	var _ms bool
	var ipMap map[string]string
	info1, b1 := ipinfoMap.Get(ipstr)
	if b1 == false {
		ipMap, _ms = iputil.ParseUrlToMap(log, ipconfig.TaobaoUrl, ipstr)
		if _ms {
			t0 := time.Now()
			WriteIpinfoToFile(middleFP, ipstr, ipstr, 1, ipMap)
			t2 := time.Now()
			log.DebugF("http get %s took %v seconds", ipstr, t2.Sub(t0))
			ipinfoMap.Set(ipstr, ipMap)
		} else {
			return nil, errors.New("get ipinfo from taobao failed")
		}
	}
	mapMutex.Lock()
	defer mapMutex.Unlock()
	if b1 {
		ipMap = info1.(map[string]string)
	}
	rtnv := make(map[string]string)
	e := iputil.DeepCopy(&rtnv, ipMap)
	if e != nil {
		fmt.Println("Deepcopy failed", ipMap)
		return nil, errors.New("Deepcopy failed")
	}
	return rtnv, nil
}
func CalcuAndSplit(startip, endip string, ipinfoMap *MySafeMap, resultFP, middleresultFP, notsameFP *os.File, depth int, log *logger.Logger) {
	defer func() {
		if r := recover(); r != nil {
			log.WarningF("get panic when detected %s, %s, %s", r, startip, endip)
		}
	}()
	/*prefix just for print recursive detect depth*/
	var prefix string
	for i := 0; i <= depth; i++ {
		prefix = prefix + "+"
	}
	prefix = prefix + "|" + strconv.Itoa(depth)
	log.Debug(prefix + "|startip|endip|" + startip + "|" + endip)

	startipMap, err := GetAndSet(log, ipinfoMap, startip, middleresultFP)
	if err != nil {
		log.ErrorF("%s", err)
		return
	}
	if startip == endip {
		SaveSameNetwork(startip, endip, startipMap, resultFP, log)
		return
	}

	endipMap, err := GetAndSet(log, ipinfoMap, endip, middleresultFP)
	if err != nil {
		log.ErrorF("%s", err)
		return
	}
	ip1 := iputil.InetAtonInt(startip)
	ip2 := iputil.InetAtonInt(endip)

	if ip1 < ip2 {
		m := (ip1 + ip2) / 2
		mip := iputil.InetNtoaStr(m)
		mipinfoMap, err := GetAndSet(log, ipinfoMap, mip, middleresultFP)
		if err != nil {
			log.ErrorF("%s", err)
			return
		}

		startinfo := iputil.UsefulInfoForPrint(startipMap)
		//midinfo := iputil.UsefulInfoForPrint(mipinfoMap)
		endinfo := iputil.UsefulInfoForPrint(endipMap)
		log.DebugF(prefix+"@@@@-start:%s-info:%+v", startip, startipMap)
		log.DebugF(prefix+"@@@@-middle:%s-info:%+v", mip, mipinfoMap)
		log.DebugF(prefix+"@@@@-end:%s-info:%+v", endip, endipMap)

		finded := iputil.QualifiedIpAtRegion(mipinfoMap, startipMap, endipMap)
		log.NoticeF("[detected result: %s, %s|%s|%s|%s|%s]", finded, startip, mip, endip, startinfo, endinfo)
		switch finded {
		case ipconfig.Goon:
			SaveSameNetwork(startip, endip, startipMap, resultFP, log)
		default:
			log.ErrorF("network %s|%s not in same view", startip, endip)
			notsameFP.WriteString(startip + "|" + endip + "|" + startinfo + "-" + endinfo + "\n")
			//CalcuAndSplit(startip, mip, ipinfoMap, resultFP, middleresultFP, notsameFP, depth+1, log)
			//CalcuAndSplit(mip_rfirst, endip, ipinfoMap, resultFP, middleresultFP, notsameFP, depth+1, log)
		}

	} else {
		log.Critical("ip1 > ip2 , network split failed!!!")
	}
}

func SaveSameNetwork(startip, endip string, ipmap map[string]string, fileFP *os.File, log *logger.Logger) {
	lens := iputil.InetAtonInt(endip) - iputil.InetAtonInt(startip) + 1
	WriteIpinfoToFile(fileFP, startip, endip, int(lens), ipmap)
	log.NoticeF("---!!!same network %s|%s!!!---", startip, endip)
}
func WriteIpinfoToFile(middleFP *os.File, startip, endip string, len int, ipmap map[string]string) {
	lenstr := strconv.Itoa(len)
	result := iputil.Format2Output(ipmap)
	fileMutex.Lock()
	middleFP.WriteString(startip + "|" + endip + "|" + lenstr + "|" + result + "\n")
	fileMutex.Unlock()
	middleFP.Sync()
}
func GetBreakpointInfo(breakpointFilename string) (*os.File, int) {
	lastFileNo := 0
	breakpointFP, temp_err := os.OpenFile(breakpointFilename, os.O_RDWR, 0666)
	if temp_err != nil {
		breakpointFP, temp_err = os.OpenFile(breakpointFilename, os.O_RDWR|os.O_CREATE, 0666)
		if temp_err != nil {
			fmt.Printf("create file %s failed\n", breakpointFilename)
			return nil, 0
		}
		lastFileNo = 0
		fmt.Printf("break point file not exists, create it and set lineno =0\n")
	} else {
		buf := make([]byte, 512)
		n, _ := breakpointFP.Read(buf)
		s := strings.TrimRight(string(buf[:n]), "\n")
		lastFileNo, _ = strconv.Atoi(s)
	}

	return breakpointFP, lastFileNo
}
