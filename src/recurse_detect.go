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

type MySafeMap struct {
	infoMap map[string]interface{}
	Lock    sync.Mutex
}

type DetectManager struct {
	ResultFP        *os.File
	MiddleFP        *os.File
	NotsameFP       *os.File
	InputFP         *os.File
	BreakFP         *os.File
	InvalidResFP    *os.File
	BatchControl    chan int
	Log             *logger.Logger
	fileMutex       sync.Mutex
	MapMutex        sync.Mutex
	RunGoroutineNum int
	ipinfoMap       MySafeMap
}

func CreateManager(gorontineCount int) *DetectManager {
	dm := new(DetectManager)
	dm.BatchControl = make(chan int, gorontineCount)
	dm.RunGoroutineNum = 0
	log, err := logger.New("test", 1, os.Stdout)
	if err != nil {
		panic(err) // Check for error
	}
	dm.Log = log
	return dm
}

func (dm *DetectManager) Stop() {
	dm.Log.Info("stop detect manager...")
	if dm.ResultFP != nil {
		dm.ResultFP.Close()
	}
	if dm.MiddleFP != nil {
		dm.MiddleFP.Close()
	}
	if dm.NotsameFP != nil {
		dm.NotsameFP.Close()
	}
	if dm.InputFP != nil {
		dm.InputFP.Close()
	}
	if dm.BreakFP != nil {
		dm.BreakFP.Close()
	}
	if dm.InvalidResFP != nil {
		dm.InvalidResFP.Close()
	}
	if dm.BatchControl != nil {
		close(dm.BatchControl)
	}
}
func (dm *DetectManager) SafeAddGoNum() {
	dm.BatchControl <- 1
	dm.MapMutex.Lock()
	dm.RunGoroutineNum += 1
	dm.MapMutex.Unlock()
}
func (dm *DetectManager) SafeReduceGoNum() {
	dm.MapMutex.Lock()
	dm.RunGoroutineNum -= 1
	dm.MapMutex.Unlock()
	<-dm.BatchControl
}

func (dm *DetectManager) SaveBreakpoint(fileno int) {
	dm.BreakFP.Truncate(0)
	dm.BreakFP.Seek(0, 0)
	fmt.Fprintf(dm.BreakFP, "%d", fileno)
	dm.BreakFP.Sync()
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

func signalListen(dm *DetectManager, quit chan struct{}) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT)
	signal.Notify(c, syscall.SIGTERM)
	select {
	case <-c:
		dm.Log.Debug("get interrupt signal:")
	case <-quit:
		dm.Log.Debug("get quit signal")
	}
}
func main() {
	finishSig := make(chan struct{})
	dm := CreateManager(ipconfig.BATCHNUM)
	defer dm.Stop()
	go dm.DetectHandle(finishSig)
	signalListen(dm, finishSig)
}

func (dm *DetectManager) DetectHandle(quit chan struct{}) {

	defer func(dm *DetectManager, quit chan struct{}) {
		for {
			if dm.RunGoroutineNum == 0 {
				close(quit)
				break
			} else {
				time.Sleep(1 * time.Second)
			}
		}
	}(dm, quit)

	dm.ipinfoMap = MySafeMap{infoMap: make(map[string]interface{})}
	iputil.GetDetectedIpInfo(dm.Log, ipconfig.FileMiddleDetectedResult, dm.ipinfoMap.infoMap)
	iputil.GetDetectedIpInfo(dm.Log, ipconfig.FileVerifiedSameIpsection, dm.ipinfoMap.infoMap)

	dm.ResultFP = createFileIfNotExist(ipconfig.FileVerifiedSameIpsection)

	dm.MiddleFP = createFileIfNotExist(ipconfig.FileMiddleDetectedResult)

	dm.NotsameFP = createFileIfNotExist(ipconfig.FileNotSameIpsection)

	dm.InvalidResFP = createFileIfNotExist(ipconfig.FileInvalidResIpsection)

	fileFP, err := os.Open(ipconfig.FileNeedIpsectionCheck)
	if err != nil {
		fmt.Println("file not exists")
		return
	}
	dm.InputFP = fileFP

	breakpointFP, lastFileNo := GetBreakpointInfo(ipconfig.FileBreakpoint)
	dm.BreakFP = breakpointFP

	fileno := 0
	br := bufio.NewReader(fileFP)
	for {
		bline, isPrefix, err := br.ReadLine()
		if err != nil {
			if err == io.EOF {
				dm.Log.Info("reach EOF, detect completed")
			}
			dm.Log.Info("read line at end")
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
		dm.SafeAddGoNum()
		startRecurseDetect(startip, endip, dm)
		/*record current detcted line no*/
		dm.SaveBreakpoint(fileno)
		dm.Log.DebugF("line no--- %d", fileno)
	}

	dm.Log.Info("all ip is detected, quit normally...")
}

func startRecurseDetect(startip, endip string, dm *DetectManager) {
	go func(startip, endip string, dm *DetectManager) {
		CalcuAndSplit(startip, endip, dm, 1)
		dm.SafeReduceGoNum()
	}(startip, endip, dm)

}
func GetAndSet(ipstr string, dm *DetectManager) (map[string]string, error) {
	var _ms bool
	var ipMap map[string]string
	info1, b1 := dm.ipinfoMap.Get(ipstr)
	if b1 == false {
		t0 := time.Now()
		dm.Log.DebugF("http get %s", ipstr)
		ipMap, _ms = iputil.ParseUrlToMap(dm.Log, ipconfig.TaobaoUrl, ipstr)
		if _ms {
			if ipMap["country"] == "中国" && (ipMap["region"] == "*" || ipMap["isp"] == "*") {
				dm.Log.DebugF("ip %s response is not sufficient: %s", ipstr, iputil.UsefulInfoForPrint(ipMap))
				WriteIpinfoToFile(dm.fileMutex, dm.InvalidResFP, ipstr, ipstr, 1, ipMap)
				return nil, errors.New("response is not sufficient")
			}
			t2 := time.Now()
			dm.Log.DebugF("http get %s took %v seconds", ipstr, t2.Sub(t0))
			WriteIpinfoToFile(dm.fileMutex, dm.MiddleFP, ipstr, ipstr, 1, ipMap)
			dm.ipinfoMap.Set(ipstr, ipMap)
		} else {
			return nil, errors.New("get ipinfo from taobao failed")
		}
	}
	dm.MapMutex.Lock()
	defer dm.MapMutex.Unlock()
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
func CalcuAndSplit(startip, endip string, dm *DetectManager, depth int) {
	defer func() {
		if r := recover(); r != nil {
			dm.Log.WarningF("get panic when detected %s, %s, %s", r, startip, endip)
		}
	}()
	/*prefix just for print recursive detect depth*/
	var prefix string
	for i := 0; i <= depth; i++ {
		prefix = prefix + "+"
	}
	prefix = prefix + "|" + strconv.Itoa(depth)
	dm.Log.Debug(prefix + "|startip|endip|" + startip + "|" + endip)

	startipMap, err := GetAndSet(startip, dm)
	if err != nil {
		dm.Log.ErrorF("%s", err)
		return
	}
	if startip == endip {
		SaveSameNetwork(startip, endip, startipMap, dm.ResultFP, dm)
		return
	}

	endipMap, err := GetAndSet(endip, dm)
	if err != nil {
		dm.Log.ErrorF("%s", err)
		return
	}
	ip1 := iputil.InetAtonInt(startip)
	ip2 := iputil.InetAtonInt(endip)

	if ip1 < ip2 {
		m := (ip1 + ip2) / 2
		mip := iputil.InetNtoaStr(m)
		mipinfoMap, err := GetAndSet(mip, dm)
		if err != nil {
			dm.Log.ErrorF("%s", err)
			return
		}

		startinfo := iputil.UsefulInfoForPrint(startipMap)
		//midinfo := iputil.UsefulInfoForPrint(mipinfoMap)
		endinfo := iputil.UsefulInfoForPrint(endipMap)
		dm.Log.DebugF(prefix+"@@@@-start:%s-info:%+v", startip, startipMap)
		dm.Log.DebugF(prefix+"@@@@-middle:%s-info:%+v", mip, mipinfoMap)
		dm.Log.DebugF(prefix+"@@@@-end:%s-info:%+v", endip, endipMap)

		finded := iputil.QualifiedIpAtRegion(mipinfoMap, startipMap, endipMap)
		dm.Log.NoticeF("[detected result: %s, %s|%s|%s|%s|%s]", finded, startip, mip, endip, startinfo, endinfo)
		switch finded {
		case ipconfig.Goon:
			SaveSameNetwork(startip, endip, startipMap, dm.ResultFP, dm)
		default:
			dm.Log.ErrorF("network %s|%s not in same view", startip, endip)
			dm.NotsameFP.WriteString(startip + "|" + endip + "|" + startinfo + "-" + endinfo + "\n")
			//CalcuAndSplit(startip, mip, dm, depth+1)
			//CalcuAndSplit(mip_rfirst, endip, dm, depth+1 )
		}

	} else {
		dm.Log.Critical("ip1 > ip2 , network split failed!!!")
	}
}

func SaveSameNetwork(startip, endip string, ipmap map[string]string, fileFP *os.File, dm *DetectManager) {
	lens := iputil.InetAtonInt(endip) - iputil.InetAtonInt(startip) + 1
	WriteIpinfoToFile(dm.fileMutex, fileFP, startip, endip, int(lens), ipmap)
	dm.Log.NoticeF("---!!!same network %s|%s!!!---", startip, endip)
}
func WriteIpinfoToFile(fileMutex sync.Mutex, fileFP *os.File, startip, endip string, len int, ipmap map[string]string) {
	lenstr := strconv.Itoa(len)
	result := iputil.Format2Output(ipmap)
	fileMutex.Lock()
	fileFP.WriteString(startip + "|" + endip + "|" + lenstr + "|" + result + "\n")
	fileMutex.Unlock()
	fileFP.Sync()
}
func GetBreakpointInfo(breakpointFilename string) (*os.File, int) {
	lastFileNo := 0
	breakpointFP, temp_err := os.OpenFile(breakpointFilename, os.O_RDWR, 0666)
	if temp_err != nil {
		breakpointFP, temp_err = os.OpenFile(breakpointFilename, os.O_RDWR|os.O_CREATE, 0666)
		if temp_err != nil {
			fmt.Printf("create file %s failed\n", breakpointFilename)
			panic("breakpint file failed")
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
func createFileIfNotExist(filename string) *os.File {
	FP, err := os.OpenFile(filename, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("open and create file failed")
		panic("open file failed")
	}
	return FP
}
