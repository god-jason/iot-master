package amap

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/god-jason/iot-master/pkg/config"
)

// Distance 距离测量
// origins: 起点列表，格式为"lng1,lat1|lng2,lat2|..."
// destinations: 终点列表，格式为"lng1,lat1|lng2,lat2|..."
// type: 计算类型，0-直线距离，1-驾车距离，2-公交距离，3-步行距离
// 参考文档：https://lbs.amap.com/api/webservice/guide/api/distance
//
// 使用示例：
//
//	response, err := amap.Distance("116.403874,39.914889", "116.481028,39.914383", 0)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if len(response.Results) > 0 {
//	    fmt.Printf("距离: %s米\n", response.Results[0].Distance)
//	}
func Distance(origins, destinations string, distanceType int) (*DistanceResponse, error) {
	// 获取配置
	server := config.GetString(MODULE, "server")
	key := config.GetString(MODULE, "key")

	// 检查API密钥是否配置
	if key == "" {
		return nil, fmt.Errorf("AMap API key not configured")
	}

	// 检查参数
	if origins == "" {
		return nil, fmt.Errorf("origins is empty")
	}
	if destinations == "" {
		return nil, fmt.Errorf("destinations is empty")
	}

	// 构建请求参数
	params := url.Values{}
	params.Set("key", key)
	params.Set("origins", origins)
	params.Set("destinations", destinations)
	params.Set("type", fmt.Sprintf("%d", distanceType))

	// 构建请求URL
	urlStr := fmt.Sprintf("%s/v3/distance?%s", server, params.Encode())

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
	var response DistanceResponse
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

// DistanceSingle 单起点到单终点距离测量
// origin: 起点，格式为"lng,lat"
// destination: 终点，格式为"lng,lat"
// distanceType: 计算类型
func DistanceSingle(origin, destination string, distanceType int) (*DistanceResult, error) {
	response, err := Distance(origin, destination, distanceType)
	if err != nil {
		return nil, err
	}
	if len(response.Results) > 0 {
		return &response.Results[0], nil
	}
	return nil, fmt.Errorf("no results")
}

// DistanceBatch 批量距离测量
// origins: 起点坐标数组
// destinations: 终点坐标数组
// distanceType: 计算类型
func DistanceBatch(origins, destinations []string, distanceType int) (*DistanceResponse, error) {
	return Distance(strings.Join(origins, "|"), strings.Join(destinations, "|"), distanceType)
}