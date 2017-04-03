package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

var fResult string = "all/middle_result_store.txt.again"
var ContryFile string = "result/country.txt"
var IspFile string = "result/isp.txt"
var AreaFile string = "result/area.txt"
var CityFile string = "result/City.txt"
var RegionFile string = "result/Region.txt"

func main() {
	resultFP, err := os.Open(fResult)
	if err != nil {
		return
	}
	defer resultFP.Close()

	br := bufio.NewReader(resultFP)
	country := make(map[string]string)
	isp := make(map[string]string)
	city := make(map[string]string)
	area := make(map[string]string)
	region := make(map[string]string)
	for {
		bline, isPrefix, err := br.ReadLine()
		if err != nil {
			if err == io.EOF {
				fmt.Println("reach EOF")
				break
			}
		}

		if isPrefix != false {
			return
		}

		line := string(bline)
		arr := strings.Split(line, "|")
		if len(arr) < 9 {
			fmt.Println("iplib format error ", line)
		}
		country_name := strings.Split(arr[4], ":")[0]
		country_id := strings.Split(arr[4], ":")[1]
		isp_name := strings.Split(arr[5], ":")[0]
		isp_id := strings.Split(arr[5], ":")[1]
		area_name := strings.Split(arr[6], ":")[0]
		area_id := strings.Split(arr[6], ":")[1]
		city_name := strings.Split(arr[7], ":")[0]
		city_id := strings.Split(arr[7], ":")[1]
		region_name := strings.Split(arr[8], ":")[0]
		region_id := strings.Split(arr[8], ":")[1]
		set_info_map(country, country_id, country_name)
		set_info_map(isp, isp_id, isp_name)
		set_info_map(area, area_id, area_name)
		set_info_map(city, city_id, city_name)
		set_info_map(region, region_id, region_name)

	}
	write_iplib_info(ContryFile, country)
	write_iplib_info(IspFile, isp)
	write_iplib_info(CityFile, city)
	write_iplib_info(AreaFile, area)
	write_iplib_info(RegionFile, region)

}

func write_iplib_info(filename string, infomap map[string]string) {
	resultFP, err := os.OpenFile(filename, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return
	}
	defer resultFP.Close()
	sortedList := SortMap(infomap)
	for _, v := range sortedList {
		info := v["key"] + "|" + v["value"] + "\n"
		resultFP.WriteString(info)
	}
}

func set_info_map(infomap map[string]string, infoid, infoname string) {
	_, existsinfo := infomap[infoid]
	if existsinfo == false {
		if infoid != "" {
			infomap[infoid] = infoname
		}
	}
}

func SortMap(infomap map[string]string) []map[string]string {
	keysSlice := make([]string, 0, len(infomap))
	for k, _ := range infomap {
		keysSlice = append(keysSlice, k)
	}
	sort.Strings(keysSlice)
	sortedList := make([]map[string]string, 0, len(keysSlice))
	for _, k := range keysSlice {
		elem := map[string]string{
			"key":   k,
			"value": infomap[k],
		}
		sortedList = append(sortedList, elem)
	}
	return sortedList
}
