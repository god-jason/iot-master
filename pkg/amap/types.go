package amap

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/Icemap/coordinate"
)

// Location 位置坐标结构体
// 表示一个地理坐标点
type Location struct {
	Longitude float64 `json:"longitude"` // 经度
	Latitude  float64 `json:"latitude"`  // 纬度
	Accuracy  int     `json:"accuracy"`  // 定位精度（米）
	Address   string  `json:"address"`   // 地址描述
}

// ConvertToWGS84 将坐标转换为WGS-84（原地修改）
func (l *Location) ConvertToWGS84() {
	if l == nil {
		return
	}
	gcj02Point := coordinate.Coordinate{X: l.Longitude, Y: l.Latitude}
	wgs84Point, err := coordinate.Convert(coordinate.GCJ02, coordinate.WGS84, gcj02Point)
	if err == nil {
		l.Longitude = wgs84Point.X
		l.Latitude = wgs84Point.Y
	}
}

// ConvertToGCJ02 将坐标转换为GCJ-02（原地修改）
func (l *Location) ConvertToGCJ02() {
	if l == nil {
		return
	}
	wgs84Point := coordinate.Coordinate{X: l.Longitude, Y: l.Latitude}
	gcj02Point, err := coordinate.Convert(coordinate.WGS84, coordinate.GCJ02, wgs84Point)
	if err == nil {
		l.Longitude = gcj02Point.X
		l.Latitude = gcj02Point.Y
	}
}

// ParseCoordinate 解析字符串格式的坐标
// 支持格式："lng,lat" 或 "lng, lat"
func ParseCoordinate(coord string) (lng, lat float64, err error) {
	parts := strings.Split(strings.ReplaceAll(coord, " ", ""), ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("无法解析坐标: %s", coord)
	}

	lng, err = strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("无法解析坐标: %s", coord)
	}

	lat, err = strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("无法解析坐标: %s", coord)
	}

	return lng, lat, nil
}

// CellTower 基站信息结构体
// 参考文档：https://lbs.amap.com/api/webservice/guide/api-advanced/hardware-location
type CellTower struct {
	MCC      int    `json:"mcc"`      // Mobile Country Code，移动国家代码，中国为460
	MNC      int    `json:"mnc"`      // Mobile Network Code，移动网络代码
	LAC      int    `json:"lac"`      // Location Area Code，位置区码
	CI       int    `json:"ci"`       // Cell Identity，基站小区ID
	Signal   int    `json:"signal"`   // Signal Strength，信号强度（dBm），可选参数
	Operator string `json:"operator"` // Operator Name，运营商名称，可选参数
}

// Wifi WiFi热点信息结构体
type Wifi struct {
	MAC    string `json:"mac"`    // MAC地址，格式如：00:11:22:33:44:55
	Signal int    `json:"signal"` // 信号强度（dBm）
	SSID   string `json:"ssid"`   // 无线网络名称，可选参数
}

// StringOrArray 自定义类型，用于处理API返回的字符串或数组
type StringOrArray []interface{}

// UnmarshalJSON 自定义反序列化，支持字符串或数组
func (s *StringOrArray) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		*s = StringOrArray{}
		return nil
	}

	// 尝试作为数组解析
	var arr []interface{}
	if err := json.Unmarshal(data, &arr); err == nil {
		*s = StringOrArray(arr)
		return nil
	}

	// 尝试作为字符串解析
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*s = StringOrArray{str}
		return nil
	}

	return fmt.Errorf("cannot unmarshal %s into StringOrArray", string(data))
}

// String 返回字符串表示，如果是数组则返回第一个元素或空字符串
func (s StringOrArray) String() string {
	if len(s) == 0 {
		return ""
	}
	switch v := s[0].(type) {
	case string:
		return v
	default:
		return ""
	}
}

// IPResponse IP定位API响应结构体
type IPResponse struct {
	Status    string        `json:"status"`    // 返回状态码，1表示成功，0表示失败
	Info      string        `json:"info"`      // 返回状态说明
	Infocode  string        `json:"infocode"`  // 状态码说明
	Province  StringOrArray `json:"province"`  // 省份
	City      StringOrArray `json:"city"`      // 城市
	CityCode  StringOrArray `json:"cityCode"`  // 城市编码
	District  StringOrArray `json:"district"`  // 区县
	Adcode    StringOrArray `json:"adcode"`    // 行政区划代码
	Longitude StringOrArray `json:"longitude"` // 经度
	Latitude  StringOrArray `json:"latitude"`  // 纬度
}

// GeoLocationResponse 硬件定位API响应结构体(V3版本)
type GeoLocationResponse struct {
	Status    string  `json:"status"`    // 返回状态码
	Info      string  `json:"info"`      // 返回状态说明
	Infocode  string  `json:"infocode"`  // 状态码说明
	Longitude float64 `json:"longitude"` // 经度
	Latitude  float64 `json:"latitude"`  // 纬度
	Accuracy  string  `json:"accuracy"`  // 定位精度
	Address   string  `json:"address"`   // 地址描述
}

// IoTLocationResponse IoT定位API响应结构体(V5版本)
// 参考文档：https://lbs.amap.com/api/webservice/guide/api-advanced/hardware-location
type IoTLocationResponse struct {
	Status    string  `json:"status"`    // 返回状态码，1表示成功，0表示失败
	Info      string  `json:"info"`      // 返回状态说明
	Infocode  string  `json:"infocode"`  // 状态码说明
	Longitude float64 `json:"longitude"` // 经度
	Latitude  float64 `json:"latitude"`  // 纬度
	Accuracy  string  `json:"accuracy"`  // 定位精度（单位：米）
	Address   string  `json:"address"`   // 地址描述
}

// GeocodeResponse 地理编码API响应结构体
type GeocodeResponse struct {
	Status   string        `json:"status"`   // 返回状态码
	Info     string        `json:"info"`     // 返回状态说明
	Infocode string        `json:"infocode"` // 状态码说明
	Geocodes []GeocodeItem `json:"geocodes"` // 地理编码结果列表
}

// GeocodeItem 地理编码结果项结构体
type GeocodeItem struct {
	FormattedAddress string        `json:"formatted_address"` // 格式化地址
	Province         string        `json:"province"`          // 省份
	City             string        `json:"city"`              // 城市
	District         string        `json:"district"`          // 区县
	Adcode           string        `json:"adcode"`            // 行政区划代码
	Street           StringOrArray `json:"street"`            // 街道
	StreetNumber     StringOrArray `json:"number"`            // 门牌号
	Longitude        float64       `json:"longitude,string"`  // 经度
	Latitude         float64       `json:"latitude,string"`   // 纬度
}

// RegeoResponse 逆地理编码API响应结构体
type RegeoResponse struct {
	Status    string        `json:"status"`    // 返回状态码
	Info      string        `json:"info"`      // 返回状态说明
	Infocode  string        `json:"infocode"`  // 状态码说明
	Regeocode RegeocodeData `json:"regeocode"` // 逆地理编码数据
}

// RegeocodeData 逆地理编码数据结构体
type RegeocodeData struct {
	AddressComponent AddressComponent `json:"addressComponent"`  // 地址组成部分
	FormattedAddress StringOrArray    `json:"formatted_address"` // 格式化地址
}

// StreetNumber 街道门牌号结构体
type StreetNumber struct {
	Street    StringOrArray `json:"street"`    // 街道
	Number    StringOrArray `json:"number"`    // 门牌号
	Location  string        `json:"location"`  // 坐标
	Direction StringOrArray `json:"direction"` // 方向
	Distance  StringOrArray `json:"distance"`  // 距离
}

// BuildingInfo 建筑物信息结构体
type BuildingInfo struct {
	Name StringOrArray `json:"name"` // 建筑物名称
	Type StringOrArray `json:"type"` // 建筑物类型
}

// NeighborhoodInfo 社区信息结构体
type NeighborhoodInfo struct {
	Name StringOrArray `json:"name"` // 社区名称
	Type StringOrArray `json:"type"` // 社区类型
}

// BusinessArea 商圈信息结构体
type BusinessArea struct {
	Location string        `json:"location"` // 商圈中心点坐标
	Name     StringOrArray `json:"name"`     // 商圈名称
	ID       StringOrArray `json:"id"`       // 商圈ID
}

// AddressComponent 地址组成部分结构体
type AddressComponent struct {
	Province         StringOrArray    `json:"province"`         // 省份
	City             StringOrArray    `json:"city"`             // 城市
	District         StringOrArray    `json:"district"`         // 区县
	Street           StringOrArray    `json:"street"`           // 街道
	StreetNumber     StreetNumber     `json:"streetNumber"`     // 门牌号信息
	Adcode           StringOrArray    `json:"adcode"`           // 行政区划代码
	Citycode         StringOrArray    `json:"citycode"`         // 城市编码
	BusinessAreas    []BusinessArea   `json:"businessAreas"`    // 商圈
	Building         BuildingInfo     `json:"building"`         // 建筑物
	BuildingType     StringOrArray    `json:"buildingType"`     // 建筑物类型
	Neighborhood     NeighborhoodInfo `json:"neighborhood"`     // 社区
	NeighborhoodType StringOrArray    `json:"neighborhoodType"` // 社区类型
	Township         StringOrArray    `json:"township"`         // 乡镇
	Towncode         StringOrArray    `json:"towncode"`         // 乡镇编码
	Village          StringOrArray    `json:"village"`          // 村庄
	VillageType      StringOrArray    `json:"villageType"`      // 村庄类型
}

// WeatherResponse 天气查询API响应结构体
type WeatherResponse struct {
	Status    string            `json:"status"`    // 返回状态码
	Info      string            `json:"info"`      // 返回状态说明
	Infocode  string            `json:"infocode"`  // 状态码说明
	Lives     []WeatherLive     `json:"lives"`     // 实时天气数据列表
	Forecasts []WeatherForecast `json:"forecasts"` // 天气预报数据列表
}

// WeatherLive 实时天气数据结构体
type WeatherLive struct {
	Province      string `json:"province"`      // 省份
	City          string `json:"city"`          // 城市
	Adcode        string `json:"adcode"`        // 行政区划代码
	Weather       string `json:"weather"`       // 天气现象
	Temperature   string `json:"temperature"`   // 气温
	WindDirection string `json:"windDirection"` // 风向
	WindPower     string `json:"windPower"`     // 风力
	Humidity      string `json:"humidity"`      // 湿度
	ReportTime    string `json:"reportTime"`    // 数据发布时间
}

// WeatherForecast 天气预报数据结构体
type WeatherForecast struct {
	Province   string         `json:"province"`   // 省份
	City       string         `json:"city"`       // 城市
	Adcode     string         `json:"adcode"`     // 行政区划代码
	ReportTime string         `json:"reportTime"` // 数据发布时间
	Casts      []ForecastCast `json:"casts"`      // 预报数据列表
}

// ForecastCast 预报数据结构体
type ForecastCast struct {
	Date               string `json:"date"`               // 日期
	Week               string `json:"week"`               // 星期
	DayWeather         string `json:"dayweather"`         // 白天天气
	NightWeather       string `json:"nightweather"`       // 夜间天气
	DayTemp            string `json:"daytemp"`            // 白天温度
	NightTemp          string `json:"nighttemp"`          // 夜间温度
	DayWindDirection   string `json:"daywinddirection"`   // 白天风向
	NightWindDirection string `json:"nightwinddirection"` // 夜间风向
	DayWindPower       string `json:"daywindpower"`       // 白天风力
	NightWindPower     string `json:"nightwindpower"`     // 夜间风力
}

// POISearchResponse POI搜索API响应结构体
type POISearchResponse struct {
	Status   string    `json:"status"`   // 返回状态码
	Info     string    `json:"info"`     // 返回状态说明
	Infocode string    `json:"infocode"` // 状态码说明
	Count    string    `json:"count"`    // 结果总数
	Pois     []POIItem `json:"pois"`     // POI数据列表
}

// POIItem POI数据项结构体
type POIItem struct {
	ID            string        `json:"id"`            // POI唯一标识
	Name          string        `json:"name"`          // POI名称
	Type          string        `json:"type"`          // POI类型
	TypeCode      string        `json:"typecode"`      // POI类型编码
	Address       string        `json:"address"`       // 地址
	Location      string        `json:"location"`      // 经纬度
	Postcode      string        `json:"postcode"`      // 邮编
	Tel           string        `json:"tel"`           // 电话
	PoiWeight     StringOrArray `json:"poiweight"`     // POI权重
	Website       string        `json:"website"`       // 网址
	Email         string        `json:"email"`         // 邮箱
	Province      string        `json:"province"`      // 省份
	City          string        `json:"city"`          // 城市
	Citycode      string        `json:"citycode"`      // 城市编码
	District      string        `json:"district"`      // 区县
	Adcode        string        `json:"adcode"`        // 行政区划代码
	BusinessArea  StringOrArray `json:"businessarea"`  // 商圈
	Building      StringOrArray `json:"building"`      // 建筑物
	Floor         string        `json:"floor"`         // 楼层
	Room          string        `json:"room"`          // 房间
	Match         string        `json:"match"`         // 匹配结果
	Tag           string        `json:"tag"`           // 标签
	NaviPOIID     string        `json:"navipoiid"`     // 导航POI ID
	EnterLocation string        `json:"enterlocation"` // 入口坐标
	ExitLocation  string        `json:"exitlocation"`  // 出口坐标
	Direction     string        `json:"direction"`     // 方向
	Distance      StringOrArray `json:"distance"`      // 距离（单位：米）
}

// RouteResponse 路径规划API响应结构体
type RouteResponse struct {
	Status   string    `json:"status"`   // 返回状态码
	Info     string    `json:"info"`     // 返回状态说明
	Infocode string    `json:"infocode"` // 状态码说明
	Route    RouteData `json:"route"`    // 路径规划数据
}

// RouteData 路径规划数据结构体
type RouteData struct {
	Origin      string      `json:"origin"`      // 起点
	Destination string      `json:"destination"` // 终点
	Paths       []RoutePath `json:"paths"`       // 路径列表
}

// RoutePath 路径数据结构体
type RoutePath struct {
	Distance     string        `json:"distance"`      // 距离（单位：米）
	Duration     string        `json:"duration"`      // 预计时间（单位：秒）
	Steps        []RouteStep   `json:"steps"`         // 路径步骤
	Tolls        string        `json:"tolls"`         // 过路费（单位：元）
	TollDistance string        `json:"toll_distance"` // 收费路段距离
	Restriction  StringOrArray `json:"restriction"`   // 是否限行
}

// RouteStep 路径步骤结构体
type RouteStep struct {
	Instruction      string        `json:"instruction"`       // 导航指示
	Distance         string        `json:"distance"`          // 距离（单位：米）
	Duration         string        `json:"duration"`          // 预计时间（单位：秒）
	Polyline         string        `json:"polyline"`          // 路径坐标串
	Action           StringOrArray `json:"action"`            // 动作
	AssistanceAction StringOrArray `json:"assistance_action"` // 辅助动作
}

// DistanceResponse 距离测量API响应结构体
type DistanceResponse struct {
	Status   string           `json:"status"`   // 返回状态码
	Info     string           `json:"info"`     // 返回状态说明
	Infocode string           `json:"infocode"` // 状态码说明
	Results  []DistanceResult `json:"results"`  // 距离结果列表
}

// DistanceResult 距离结果结构体
type DistanceResult struct {
	Origin      string `json:"origin"`      // 起点
	Destination string `json:"destination"` // 终点
	Distance    string `json:"distance"`    // 距离（单位：米）
	Duration    string `json:"duration"`    // 预计时间（单位：秒）
}

// DistrictResponse 行政区划查询API响应结构体
type DistrictResponse struct {
	Status    string         `json:"status"`    // 返回状态码
	Info      string         `json:"info"`      // 返回状态说明
	Infocode  string         `json:"infocode"`  // 状态码说明
	Districts []DistrictItem `json:"districts"` // 行政区划列表
}

// DistrictItem 行政区划数据项结构体
type DistrictItem struct {
	Adcode       string         `json:"adcode"`     // 行政区划代码
	Name         string         `json:"name"`       // 名称
	Center       string         `json:"center"`     // 中心点坐标
	Level        string         `json:"level"`      // 级别（country/province/city/district/street）
	ParentCode   string         `json:"parentcode"` // 父级行政区划代码
	SubDistricts []DistrictItem `json:"districts"`  // 子级行政区划列表
}

// BatchGeocodeResponse 批量地理编码API响应结构体
type BatchGeocodeResponse struct {
	Status   string               `json:"status"`   // 返回状态码
	Info     string               `json:"info"`     // 返回状态说明
	Infocode string               `json:"infocode"` // 状态码说明
	Results  []BatchGeocodeResult `json:"results"`  // 批量结果列表
}

// BatchGeocodeResult 批量地理编码结果结构体
type BatchGeocodeResult struct {
	Status   string        `json:"status"`   // 单个请求状态码
	Info     string        `json:"info"`     // 状态说明
	Geocodes []GeocodeItem `json:"geocodes"` // 地理编码结果
}

// ErrorResponse 错误响应结构体
type ErrorResponse struct {
	Status   string `json:"status"`   // 返回状态码
	Info     string `json:"info"`     // 返回状态说明
	Infocode string `json:"infocode"` // 状态码说明
}
