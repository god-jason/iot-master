package amap

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/god-jason/iot-master/pkg/config"
)

// District 查询行政区划信息
// keywords: 查询关键字，可选，如"北京"、"朝阳区"
// subdistrict: 子级行政区级数，0-不返回下级，1-返回下一级，以此类推
// 参考文档：https://lbs.amap.com/api/webservice/guide/api/district
//
// 使用示例：
//
//	// 查询全国行政区划
//	response, err := amap.District("", 1)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, district := range response.Districts {
//	    fmt.Printf("%s - %s\n", district.Name, district.Adcode)
//	}
//
//	// 查询北京市行政区划
//	response, err = amap.District("北京", 2)
//	if err != nil {
//	    log.Fatal(err)
//	}
func District(keywords string, subdistrict int) (*DistrictResponse, error) {
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
	if keywords != "" {
		params.Set("keywords", keywords)
	}
	params.Set("subdistrict", fmt.Sprintf("%d", subdistrict))

	// 构建请求URL
	urlStr := fmt.Sprintf("%s/v3/config/district?%s", server, params.Encode())

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
	var response DistrictResponse
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

// GetProvinces 获取全国省份列表
func GetProvinces() ([]DistrictItem, error) {
	response, err := District("", 1)
	if err != nil {
		return nil, err
	}
	if len(response.Districts) > 0 {
		return response.Districts[0].SubDistricts, nil
	}
	return nil, nil
}

// GetCities 获取指定省份的城市列表
// provinceCode: 省份编码
func GetCities(provinceCode string) ([]DistrictItem, error) {
	response, err := District(provinceCode, 1)
	if err != nil {
		return nil, err
	}
	if len(response.Districts) > 0 {
		return response.Districts[0].SubDistricts, nil
	}
	return nil, nil
}

// GetDistricts 获取指定城市的区县列表
// cityCode: 城市编码
func GetDistricts(cityCode string) ([]DistrictItem, error) {
	response, err := District(cityCode, 1)
	if err != nil {
		return nil, err
	}
	if len(response.Districts) > 0 {
		return response.Districts[0].SubDistricts, nil
	}
	return nil, nil
}