package amap

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/god-jason/iot-master/pkg/config"
)

// RouteDriving 驾车路径规划
// origin: 起点，格式为"lng,lat"或地址
// destination: 终点，格式为"lng,lat"或地址
// 参考文档：https://lbs.amap.com/api/webservice/guide/api/direction
//
// 使用示例：
//
//	response, err := amap.RouteDriving("116.403874,39.914889", "116.481028,39.914383")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if len(response.Route.Paths) > 0 {
//	    path := response.Route.Paths[0]
//	    fmt.Printf("距离: %s米, 预计时间: %s秒\n", path.Distance, path.Duration)
//	}
func RouteDriving(origin, destination string) (*RouteResponse, error) {
	return route("driving", origin, destination, "")
}

// RouteWalking 步行路径规划
// origin: 起点，格式为"lng,lat"或地址
// destination: 终点，格式为"lng,lat"或地址
// 参考文档：https://lbs.amap.com/api/webservice/guide/api/direction
func RouteWalking(origin, destination string) (*RouteResponse, error) {
	return route("walking", origin, destination, "")
}

// RouteRiding 骑行路径规划
// origin: 起点，格式为"lng,lat"或地址
// destination: 终点，格式为"lng,lat"或地址
// 参考文档：https://lbs.amap.com/api/webservice/guide/api/direction
func RouteRiding(origin, destination string) (*RouteResponse, error) {
	return route("riding", origin, destination, "")
}

// RouteTransit 公交路径规划
// origin: 起点，格式为"lng,lat"或地址
// destination: 终点，格式为"lng,lat"或地址
// city: 城市名称，可选
// 参考文档：https://lbs.amap.com/api/webservice/guide/api/direction
func RouteTransit(origin, destination, city string) (*RouteResponse, error) {
	return route("transit", origin, destination, city)
}

// route 通用路径规划函数
func route(routeType, origin, destination, city string) (*RouteResponse, error) {
	// 获取配置
	server := config.GetString(MODULE, "server")
	key := config.GetString(MODULE, "key")

	// 检查API密钥是否配置
	if key == "" {
		return nil, fmt.Errorf("AMap API key not configured")
	}

	// 检查参数
	if origin == "" {
		return nil, fmt.Errorf("origin is empty")
	}
	if destination == "" {
		return nil, fmt.Errorf("destination is empty")
	}

	// 构建请求参数
	params := url.Values{}
	params.Set("key", key)
	params.Set("origin", origin)
	params.Set("destination", destination)
	if city != "" {
		params.Set("city", city)
	}

	// 构建请求URL
	urlStr := fmt.Sprintf("%s/v3/direction/%s?%s", server, routeType, params.Encode())

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
	var response RouteResponse
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