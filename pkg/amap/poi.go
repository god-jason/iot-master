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

// POISearch POI搜索
// keywords: 搜索关键词
// city: 城市名称或编码
// types: POI类型，可选
// page: 页码，从1开始
// offset: 每页数量，默认20，最大50
// 参考文档：https://lbs.amap.com/api/webservice/guide/api/search
//
// 使用示例：
//
//	response, err := amap.POISearch("餐厅", "北京", "餐饮服务", 1, 10)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("找到 %s 个结果\n", response.Count)
//	for _, poi := range response.Pois {
//	    fmt.Printf("%s: %s\n", poi.Name, poi.Address)
//	}
func POISearch(keywords, city, types string, page, offset int) (*POISearchResponse, error) {
	// 获取配置
	server := config.GetString(MODULE, "server")
	key := config.GetString(MODULE, "key")

	// 检查API密钥是否配置
	if key == "" {
		return nil, fmt.Errorf("AMap API key not configured")
	}

	// 检查关键词
	if keywords == "" {
		return nil, fmt.Errorf("keywords is empty")
	}

	// 设置默认值
	if page <= 0 {
		page = 1
	}
	if offset <= 0 || offset > 50 {
		offset = 20
	}

	// 构建请求参数
	params := url.Values{}
	params.Set("key", key)
	params.Set("keywords", keywords)
	if city != "" {
		params.Set("city", city)
	}
	if types != "" {
		params.Set("types", types)
	}
	params.Set("page", fmt.Sprintf("%d", page))
	params.Set("offset", fmt.Sprintf("%d", offset))

	// 构建请求URL
	urlStr := fmt.Sprintf("%s/v3/place/text?%s", server, params.Encode())

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
	var response POISearchResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("parse response failed: %v", err)
	}

	// 检查API返回状态
	if response.Status != "1" {
		return nil, fmt.Errorf("API error: %s (%s)", response.Info, response.Infocode)
	}

	return &response, nil
}

// POISearchNearby 周边POI搜索
// location: 中心点坐标，格式为"lng,lat"（WGS-84坐标）
// radius: 搜索半径（米），默认1000
// keywords: 搜索关键词，可选
// types: POI类型，可选
// 参考文档：https://lbs.amap.com/api/webservice/guide/api/search
//
// 使用示例：
//
//	response, err := amap.POISearchNearby("116.403874,39.914889", 5000, "酒店", "")
//	if err != nil {
//	    log.Fatal(err)
//	}
func POISearchNearby(location string, radius int, keywords, types string) (*POISearchResponse, error) {
	// 获取配置
	server := config.GetString(MODULE, "server")
	key := config.GetString(MODULE, "key")

	// 检查API密钥是否配置
	if key == "" {
		return nil, fmt.Errorf("AMap API key not configured")
	}

	// 检查位置参数
	if location == "" {
		return nil, fmt.Errorf("location is empty")
	}

	// 解析坐标（WGS-84格式）
	lng, lat, err := ParseCoordinate(location)
	if err != nil {
		return nil, fmt.Errorf("解析坐标失败: %v", err)
	}

	// 将WGS-84坐标转换为GCJ-02坐标（高德使用GCJ-02）
	wgs84Point := coordinate.Coordinate{X: lng, Y: lat}
	gcj02Point, err := coordinate.Convert(coordinate.WGS84, coordinate.GCJ02, wgs84Point)
	if err != nil {
		return nil, fmt.Errorf("坐标转换失败: %v", err)
	}
	gcjLocation := fmt.Sprintf("%f,%f", gcj02Point.X, gcj02Point.Y)

	// 设置默认半径
	if radius <= 0 {
		radius = 1000
	}

	// 构建请求参数
	params := url.Values{}
	params.Set("key", key)
	params.Set("location", gcjLocation)
	params.Set("radius", fmt.Sprintf("%d", radius))
	if keywords != "" {
		params.Set("keywords", keywords)
	}
	if types != "" {
		params.Set("types", types)
	}

	// 构建请求URL
	urlStr := fmt.Sprintf("%s/v3/place/around?%s", server, params.Encode())

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
	var response POISearchResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("parse response failed: %v", err)
	}

	// 检查API返回状态
	if response.Status != "1" {
		return nil, fmt.Errorf("API error: %s (%s)", response.Info, response.Infocode)
	}

	// 将返回的POI坐标从GCJ-02转换为WGS-84
	for i := range response.Pois {
		if response.Pois[i].Location != "" {
			poiLng, poiLat, err := ParseCoordinate(response.Pois[i].Location)
			if err == nil {
				gcj02Point := coordinate.Coordinate{X: poiLng, Y: poiLat}
				wgs84Point, err := coordinate.Convert(coordinate.GCJ02, coordinate.WGS84, gcj02Point)
				if err == nil {
					response.Pois[i].Location = fmt.Sprintf("%f,%f", wgs84Point.X, wgs84Point.Y)
				}
			}
		}
	}

	return &response, nil
}
