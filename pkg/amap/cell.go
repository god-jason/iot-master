package amap

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/god-jason/iot-master/pkg/config"
)

// ==================== 基站定位接口 ====================
// 参考文档：https://lbs.amap.com/api/webservice/guide/api-advanced/hardware-location
// V1接口：https://apilocate.amap.com/position

// V1LocationResult V1版本定位结果结构
type V1LocationResult struct {
	Type     string `json:"type"`     // 定位类型，0表示未定位到，其他表示定位成功
	Location string `json:"location"` // 定位结果，格式：经度,纬度
	Radius   string `json:"radius"`   // 定位精度半径，单位：米
	Desc     string `json:"desc"`     // 位置描述
	Country  string `json:"country"`  // 国家
	Province string `json:"province"` // 省份
	City     string `json:"city"`     // 城市
	Citycode string `json:"citycode"` // 城市编码
	Adcode   string `json:"adcode"`   // 区域编码
	Road     string `json:"road"`     // 道路名
	Poi      string `json:"poi"`      // 定位附近的POI名称
}

// V1LocationResponse V1版本API响应结构
type V1LocationResponse struct {
	Status   string            `json:"status"`   // 状态码，1表示成功，0表示失败
	Info     string            `json:"info"`     // 状态信息
	InfoCode string            `json:"infocode"` // 信息码
	Result   *V1LocationResult `json:"result"`   // 定位结果
}

// GetLocationLBS 根据基站信息获取位置（V1版本）
// scell: 当前连接的基站
// cells: 附近基站列表（可选）
// 使用V1接口：https://apilocate.amap.com/position
func GetLocationLBS(scell CellTower, cells []CellTower) (*Location, error) {
	// 获取配置
	key := config.GetString(MODULE, "key")
	apiURL := config.GetString(MODULE, "apilocate_url")

	// 检查API密钥是否配置
	if key == "" {
		return nil, fmt.Errorf("高德地图API密钥未配置")
	}

	// 构建请求参数
	params := url.Values{}
	params.Set("key", key)
	params.Set("output", "json")  // 返回格式
	params.Set("accesstype", "0") // 0表示移动接入网络（基站定位）
	params.Set("cdma", "0")       // 0表示非CDMA网络
	params.Set("network", "GSM")  // 网络类型

	// 拼接当前基站信息
	// 格式：mcc,mnc,lac,cellid,signal
	signal := scell.Signal
	if signal == 0 {
		signal = -1
	}
	btsStr := fmt.Sprintf("%d,%d,%d,%d,%d", scell.MCC, scell.MNC, scell.LAC, scell.CI, signal)
	params.Set("bts", btsStr)

	// 拼接附近基站信息（nearbts参数）
	if len(cells) > 0 {
		var nearBtsData []string
		for _, cell := range cells {
			cellSignal := cell.Signal
			if cellSignal == 0 {
				cellSignal = -1
			}
			cellStr := fmt.Sprintf("%d,%d,%d,%d,%d", cell.MCC, cell.MNC, cell.LAC, cell.CI, cellSignal)
			nearBtsData = append(nearBtsData, cellStr)
		}
		params.Set("nearbts", strings.Join(nearBtsData, "|"))
	}

	// 构建请求URL（参数在URL中）
	requestURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	// 发送HTTP GET请求
	resp, err := http.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析JSON响应
	var response V1LocationResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("解析响应失败: %v, 原始响应: %s, 请求URL: %s", err, string(body), requestURL)
	}

	// 检查API返回状态
	if response.Status != "1" {
		return nil, fmt.Errorf("API错误: %s (%s), 请求URL: %s, 原始响应: %s", response.Info, response.InfoCode, requestURL, string(body))
	}

	// 构建返回结果
	location := &Location{
		Longitude: 0,
		Latitude:  0,
		Address:   "",
	}

	// 解析定位结果
	if response.Result != nil {
		// 检查定位类型，type=0表示未定位到
		if response.Result.Type == "0" {
			return nil, fmt.Errorf("未获取到有效定位结果, 请求URL: %s, 原始响应: %s", requestURL, string(body))
		}

		// 解析经纬度
		if response.Result.Location != "" {
			parts := strings.Split(response.Result.Location, ",")
			if len(parts) >= 2 {
				location.Longitude, _ = strconv.ParseFloat(parts[0], 64)
				location.Latitude, _ = strconv.ParseFloat(parts[1], 64)
			}
		}

		// 解析定位精度
		if response.Result.Radius != "" {
			location.Accuracy, _ = strconv.Atoi(response.Result.Radius)
		}

		// 解析地址描述
		location.Address = response.Result.Desc
	}

	// 检查是否获取到有效定位
	if location.Longitude == 0 && location.Latitude == 0 {
		return nil, fmt.Errorf("未获取到有效定位结果, 请求URL: %s, 原始响应: %s", requestURL, string(body))
	}

	// 将GCJ-02坐标转换为WGS-84坐标
	location.ConvertToWGS84()

	return location, nil
}

// GetLocationWifi 根据WiFi热点信息获取位置（V1版本）
// wifis: WiFi热点列表（至少需要2个，最多30个）
// 使用V1接口：https://apilocate.amap.com/position
func GetLocationWifi(wifis []Wifi) (*Location, error) {
	if len(wifis) < 2 {
		return nil, fmt.Errorf("WiFi热点数量不足，至少需要2个")
	}

	if len(wifis) > 30 {
		return nil, fmt.Errorf("WiFi热点数量过多，最多支持30个")
	}

	// 获取配置
	key := config.GetString(MODULE, "key")
	apiURL := config.GetString(MODULE, "apilocate_url")

	// 检查API密钥是否配置
	if key == "" {
		return nil, fmt.Errorf("高德地图API密钥未配置")
	}

	// 构建请求参数
	params := url.Values{}
	params.Set("key", key)
	params.Set("output", "json")  // 返回格式
	params.Set("accesstype", "1") // 1表示WiFi接入网络

	// 拼接WiFi信息
	// 格式：mac,signal,ssid|mac,signal,ssid|...
	var macData []string
	for _, wifi := range wifis {
		macStr := fmt.Sprintf("%s,%d,%s", strings.ToLower(wifi.MAC), wifi.Signal, wifi.SSID)
		macData = append(macData, macStr)
	}
	params.Set("macs", strings.Join(macData, "|"))

	// 构建请求URL（参数在URL中）
	requestURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	// 发送HTTP GET请求
	resp, err := http.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析JSON响应
	var response V1LocationResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("解析响应失败: %v, 原始响应: %s, 请求URL: %s", err, string(body), requestURL)
	}

	// 检查API返回状态
	if response.Status != "1" {
		return nil, fmt.Errorf("API错误: %s (%s), 请求URL: %s, 原始响应: %s", response.Info, response.InfoCode, requestURL, string(body))
	}

	// 构建返回结果
	location := &Location{
		Longitude: 0,
		Latitude:  0,
		Address:   "",
	}

	// 解析定位结果
	if response.Result != nil {
		// 检查定位类型，type=0表示未定位到
		if response.Result.Type == "0" {
			return nil, fmt.Errorf("未获取到有效定位结果, 请求URL: %s, 原始响应: %s", requestURL, string(body))
		}

		// 解析经纬度
		if response.Result.Location != "" {
			parts := strings.Split(response.Result.Location, ",")
			if len(parts) >= 2 {
				location.Longitude, _ = strconv.ParseFloat(parts[0], 64)
				location.Latitude, _ = strconv.ParseFloat(parts[1], 64)
			}
		}

		// 解析定位精度
		if response.Result.Radius != "" {
			location.Accuracy, _ = strconv.Atoi(response.Result.Radius)
		}

		// 解析地址描述
		location.Address = response.Result.Desc
	}

	// 检查是否获取到有效定位
	if location.Longitude == 0 && location.Latitude == 0 {
		return nil, fmt.Errorf("未获取到有效定位结果, 请求URL: %s, 原始响应: %s", requestURL, string(body))
	}

	// 将GCJ-02坐标转换为WGS-84坐标
	location.ConvertToWGS84()

	return location, nil
}
