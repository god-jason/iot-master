package amap

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/god-jason/iot-master/pkg/config"
)

// GetWeather 实时天气查询
// city可以是城市名称（如"北京"）或城市编码（如"110000"）
// 参考文档：https://lbs.amap.com/api/webservice/guide/api/weatherinfo
//
// 使用示例：
//
//	// 查询北京实时天气
//	response, err := amap.GetWeather("北京", "base")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if len(response.Lives) > 0 {
//	    live := response.Lives[0]
//	    fmt.Printf("%s天气: %s, 温度: %s°C\n", live.City, live.Weather, live.Temperature)
//	}
//
//	// 查询北京天气预报
//	response, err = amap.GetWeather("北京", "all")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if len(response.Forecasts) > 0 {
//	    for _, cast := range response.Forecasts[0].Casts {
//	        fmt.Printf("%s: %s ~ %s°C\n", cast.Date, cast.DayWeather, cast.DayTemp)
//	    }
//	}
func GetWeather(city string, extensions string) (*WeatherResponse, error) {
	// 获取配置
	server := config.GetString(MODULE, "server")
	key := config.GetString(MODULE, "key")

	// 检查API密钥是否配置
	if key == "" {
		return nil, fmt.Errorf("AMap API key not configured")
	}

	// 检查城市参数
	if city == "" {
		return nil, fmt.Errorf("city is empty")
	}

	// 设置默认extensions
	if extensions == "" {
		extensions = "base" // base: 实时天气, all: 实时+预报
	}

	// 构建请求参数
	params := url.Values{}
	params.Set("key", key)
	params.Set("city", city)
	params.Set("extensions", extensions)

	// 构建请求URL
	urlStr := fmt.Sprintf("%s/v3/weather/weatherInfo?%s", server, params.Encode())

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
	var response WeatherResponse
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