package amap

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/Icemap/coordinate"
	"github.com/god-jason/iot-master/pkg/config"
)

// Geocode 地理编码：地址转坐标
// 将中文地址转换为经纬度坐标
// 参考文档：https://lbs.amap.com/api/webservice/guide/api/georegeo
//
// 使用示例：
//
//	location, err := amap.Geocode("北京市朝阳区望京SOHO")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("经度: %f, 纬度: %f\n", location.Longitude, location.Latitude)
func Geocode(address string) (*Location, error) {
	// 获取配置
	server := config.GetString(MODULE, "server")
	key := config.GetString(MODULE, "key")

	// 检查API密钥是否配置
	if key == "" {
		return nil, fmt.Errorf("AMap API key not configured")
	}

	// 检查地址是否为空
	if address == "" {
		return nil, fmt.Errorf("address is empty")
	}

	// 构建请求参数
	params := url.Values{}
	params.Set("key", key)
	params.Set("address", address)

	// 构建请求URL
	urlStr := fmt.Sprintf("%s/v3/geocode/geo?%s", server, params.Encode())

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
	var response GeocodeResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("parse response failed: %v", err)
	}

	// 检查API返回状态
	if response.Status != "1" {
		return nil, fmt.Errorf("API error: %s (%s)", response.Info, response.Infocode)
	}

	// 检查是否有结果
	if len(response.Geocodes) == 0 {
		return nil, fmt.Errorf("no geocode results found")
	}

	// 构建返回结果
	geocode := response.Geocodes[0]
	location := &Location{
		Longitude: geocode.Longitude,
		Latitude:  geocode.Latitude,
		Address:   geocode.FormattedAddress,
	}

	// 将GCJ-02坐标转换为WGS-84坐标
	location.ConvertToWGS84()

	return location, nil
}

// ReverseGeocode 逆地理编码：坐标转地址
// 将经纬度坐标转换为中文地址
// 参考文档：https://lbs.amap.com/api/webservice/guide/api/georegeo
//
// 使用示例：
//
//	// 天安门坐标
//	location, err := amap.ReverseGeocode(116.403874, 39.914889)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("地址: %s\n", location.Address)
func ReverseGeocode(lng, lat float64) (*Location, error) {
	// 获取配置
	server := config.GetString(MODULE, "server")
	key := config.GetString(MODULE, "key")

	// 检查API密钥是否配置
	if key == "" {
		return nil, fmt.Errorf("AMap API key not configured")
	}

	// 将WGS-84坐标转换为GCJ-02坐标（高德使用GCJ-02）
	wgs84Point := coordinate.Coordinate{X: lng, Y: lat}
	gcj02Point, err := coordinate.Convert(coordinate.WGS84, coordinate.GCJ02, wgs84Point)
	if err != nil {
		return nil, fmt.Errorf("坐标转换失败: %v", err)
	}

	// 构建请求参数
	params := url.Values{}
	params.Set("key", key)
	params.Set("location", fmt.Sprintf("%f,%f", gcj02Point.X, gcj02Point.Y))

	// 构建请求URL
	urlStr := fmt.Sprintf("%s/v3/geocode/regeo?%s", server, params.Encode())

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
	var response RegeoResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("parse response failed: %v", err)
	}

	// 检查API返回状态
	if response.Status != "1" {
		return nil, fmt.Errorf("API error: %s (%s)", response.Info, response.Infocode)
	}

	// 构建返回结果（保持原始WGS-84坐标）
	return &Location{
		Longitude: lng,
		Latitude:  lat,
		Address:   response.Regeocode.FormattedAddress.String(),
	}, nil
}
