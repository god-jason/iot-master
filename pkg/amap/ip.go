package amap

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/god-jason/iot-master/pkg/config"
)

// GetLocationByIP 根据IP地址获取位置信息
// ip参数为空时，使用请求来源IP
// 参考文档：https://lbs.amap.com/api/webservice/guide/api/ipconfig
//
// 使用示例：
//
//	// 使用当前请求IP定位
//	location, err := amap.GetLocationByIP("")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("位置: %s, 坐标: %f, %f\n", location.Address, location.Longitude, location.Latitude)
//
//	// 指定IP地址定位
//	location, err = amap.GetLocationByIP("114.114.114.114")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("位置: %s\n", location.Address)
func GetLocationByIP(ip string) (*Location, error) {
	// 获取配置
	server := config.GetString(MODULE, "server")
	key := config.GetString(MODULE, "key")

	// 检查API密钥是否配置
	if key == "" {
		return nil, fmt.Errorf("AMap API key not configured")
	}

	// 构建请求参数
	params := url.Values{}
	params.Set("key", key)
	if ip != "" {
		params.Set("ip", ip)
	}

	// 构建请求URL
	urlStr := fmt.Sprintf("%s/v3/ip?%s", server, params.Encode())

	// 发送HTTP GET请求
	resp, err := http.Get(urlStr)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %v", err)
	}

	// 解析JSON响应
	var response IPResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("parse response failed: %v", err)
	}

	// 检查API返回状态
	if response.Status != "1" {
		return nil, fmt.Errorf("API error: %s (%s)", response.Info, response.Infocode)
	}

	// 构建返回结果
	location := &Location{
		Address: fmt.Sprintf("%s%s%s", response.Province.String(), response.City.String(), response.District.String()),
	}

	// 解析经纬度
	if response.Longitude.String() != "" {
		location.Longitude, _ = strconv.ParseFloat(response.Longitude.String(), 64)
	}
	if response.Latitude.String() != "" {
		location.Latitude, _ = strconv.ParseFloat(response.Latitude.String(), 64)
	}

	// 将GCJ-02坐标转换为WGS-84坐标
	location.ConvertToWGS84()

	return location, nil
}
