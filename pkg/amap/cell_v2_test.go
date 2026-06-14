package amap

import (
	"strings"
	"testing"
)

// ==================== V1 版本单元测试 ====================

// TestGetLocationLBS_SingleCell 测试单个基站定位（使用真实基站数据）
func TestGetLocationLBS_SingleCell(t *testing.T) {
	SetKey("0ce21bcb510ac0e1b978eb5c4a9525fb")
	// 使用用户提供的真实基站数据
	cell := CellTower{MCC: 460, MNC: 0, LAC: 21200, CI: 173806401, Signal: -61}
	location, err := GetLocationLBS(cell, nil)
	if err != nil {
		if !strings.Contains(err.Error(), "API 错误") {
			t.Errorf("GetLocationLBS 失败，错误信息：%v", err)
		} else {
			t.Errorf("API 返回错误：%v", err)
		}
		return
	}
	t.Logf("定位成功 - 经度：%f, 纬度：%f, 精度：%d米，地址：%s",
		location.Longitude, location.Latitude, location.Accuracy, location.Address)
}

// TestGetLocationLBS_WithCells 测试当前基站 + 附近基站定位（使用真实基站数据）
func TestGetLocationLBS_WithCells(t *testing.T) {
	SetKey("0ce21bcb510ac0e1b978eb5c4a9525fb")
	// 使用用户提供的真实基站数据
	scell := CellTower{MCC: 460, MNC: 0, LAC: 21200, CI: 173806401, Signal: -61}
	cells := []CellTower{
		{MCC: 460, MNC: 0, LAC: 21200, CI: 173806401, Signal: -61},
		{MCC: 460, MNC: 15, LAC: 21200, CI: 173806401, Signal: -61},
	}
	location, err := GetLocationLBS(scell, cells)
	if err != nil {
		if !strings.Contains(err.Error(), "API 错误") {
			t.Errorf("GetLocationLBS 失败，错误信息：%v", err)
		} else {
			t.Errorf("API 返回错误：%v", err)
		}
		return
	}
	t.Logf("定位成功 - 经度：%f, 纬度：%f, 精度：%d米，地址：%s",
		location.Longitude, location.Latitude, location.Accuracy, location.Address)
}

// TestGetLocationLBS_NewCellData 测试新的基站数据（用户提供的基站数据）
// 数据来源：[{"mnc":0,"dlbandwidth":5,"tdd":1,"earfcn":38400,"ulbandwidth":5,"band":39,"mcc":460,"pci":476,"rsrp":-77,"tac":21200,"rssi":-46,"rsrq":-11,"snr":14,"cid":232802050},
// {"mnc":15,"earfcn":38400,"pci":476,"rsrp":-77,"tac":21200,"mcc":460,"rsrq":-11,"snr":14,"cid":232802050},
// {"mnc":0,"earfcn":3590,"pci":158,"rsrp":-90,"tac":21200,"mcc":460,"rsrq":-12,"snr":5,"cid":173806401},
// {"mnc":0,"earfcn":1302,"pci":331,"rsrp":-102,"tac":21200,"mcc":460,"rsrq":-12,"snr":0,"cid":107741891}]
func TestGetLocationLBS_NewCellData(t *testing.T) {
	SetKey("0ce21bcb510ac0e1b978eb5c4a9525fb")
	// 当前连接的基站（有 rssi 字段的那个）
	scell := CellTower{MCC: 460, MNC: 0, LAC: 21200, CI: 232802050, Signal: -46}
	// 周边基站列表
	cells := []CellTower{
		{MCC: 460, MNC: 15, LAC: 21200, CI: 232802050, Signal: -77}, // 使用 rsrp 作为信号强度
		{MCC: 460, MNC: 0, LAC: 21200, CI: 173806401, Signal: -90},  // 使用 rsrp 作为信号强度
		{MCC: 460, MNC: 0, LAC: 21200, CI: 107741891, Signal: -102}, // 使用 rsrp 作为信号强度
	}
	location, err := GetLocationLBS(scell, cells)
	if err != nil {
		if !strings.Contains(err.Error(), "API 错误") {
			t.Errorf("GetLocationLBS 失败，错误信息：%v", err)
		} else {
			t.Errorf("API 返回错误：%v", err)
		}
		return
	}
	t.Logf("定位成功 - 经度：%f, 纬度：%f, 精度：%d米，地址：%s",
		location.Longitude, location.Latitude, location.Accuracy, location.Address)
}

// TestGetLocationLBS_NewCellDataV2 测试最新基站数据（用户提供）
// 数据来源：
// [
//
//	{"mnc":0,"dlbandwidth":5,"tdd":1,"earfcn":38950,"ulbandwidth":5,"band":40,"mcc":460,"pci":383,"rsrp":-97,"tac":20952,"rssi":-70,"rsrq":-7,"snr":8,"cid":219069203},
//	{"mnc":0,"earfcn":1400,"pci":3,"rsrp":-90,"tac":20952,"mcc":460,"rsrq":-11,"snr":12,"cid":230437906},
//	{"mnc":15,"earfcn":38950,"pci":383,"rsrp":-97,"tac":20952,"mcc":460,"rsrq":-7,"snr":8,"cid":219069203},
//	{"mnc":0,"earfcn":3590,"pci":408,"rsrp":-98,"tac":20952,"mcc":460,"rsrq":-16,"snr":3,"cid":10020231},
//	{"mnc":0,"earfcn":39148,"pci":287,"rsrp":-110,"tac":20952,"mcc":460,"rsrq":-26,"snr":4,"cid":219069221}
//
// ]
func TestGetLocationLBS_NewCellDataV2(t *testing.T) {
	SetKey("0ce21bcb510ac0e1b978eb5c4a9525fb")
	// 当前连接的基站（有 rssi=-70 的那个）
	scell := CellTower{MCC: 460, MNC: 0, LAC: 20952, CI: 219069203, Signal: -70}
	// 周边基站列表（使用 rsrp 作为信号强度）
	cells := []CellTower{
		{MCC: 460, MNC: 0, LAC: 20952, CI: 230437906, Signal: -90},
		{MCC: 460, MNC: 15, LAC: 20952, CI: 219069203, Signal: -97},
		{MCC: 460, MNC: 0, LAC: 20952, CI: 10020231, Signal: -98},
		{MCC: 460, MNC: 0, LAC: 20952, CI: 219069221, Signal: -110},
	}
	location, err := GetLocationLBS(scell, cells)
	if err != nil {
		if !strings.Contains(err.Error(), "API 错误") {
			t.Errorf("GetLocationLBS 失败，错误信息：%v", err)
		} else {
			t.Errorf("API 返回错误：%v", err)
		}
		return
	}
	t.Logf("定位成功 - 经度：%f, 纬度：%f, 精度：%d米，地址：%s",
		location.Longitude, location.Latitude, location.Accuracy, location.Address)
}

// TestGetLocationWifi_MultipleWifis 测试多个WiFi定位（使用真实WiFi数据）
func TestGetLocationWifi_MultipleWifis(t *testing.T) {
	SetKey("0ce21bcb510ac0e1b978eb5c4a9525fb")
	// 使用用户提供的真实WiFi数据
	wifis := []Wifi{
		{MAC: "8c:5a:c1:6a:54:e1", Signal: -59, SSID: ""},
		{MAC: "b0:30:55:8a:02:87", Signal: -60, SSID: "CMCC-2fuY"},
		{MAC: "20:6b:e7:a2:b2:51", Signal: -60, SSID: "TP-LINK_B251"},
		{MAC: "18:d9:8f:0c:13:40", Signal: -60, SSID: "ZS"},
		{MAC: "9c:6f:52:8d:0d:10", Signal: -60, SSID: "CMCC-3f3c"},
	}
	location, err := GetLocationWifi(wifis)
	if err != nil {
		if !strings.Contains(err.Error(), "API错误") {
			t.Errorf("GetLocationWifi失败，错误信息: %v", err)
		} else {
			t.Errorf("API返回错误: %v", err)
		}
		return
	}
	t.Logf("定位成功 - 经度: %f, 纬度: %f, 精度: %d米, 地址: %s",
		location.Longitude, location.Latitude, location.Accuracy, location.Address)
}

// TestGetLocationLBS_NewCellDataV3 测试更多基站数据（用户提供，8个基站）
// 数据来源：
// [
//
//	{"mnc":0,"dlbandwidth":3,"tdd":0,"earfcn":1400,"ulbandwidth":3,"band":3,"mcc":460,"pci":3,"rsrp":-93,"tac":20952,"rssi":-67,"rsrq":-9,"snr":9,"cid":230437906},
//	{"mnc":0,"earfcn":39148,"pci":287,"rsrp":-91,"tac":20952,"mcc":460,"rsrq":-7,"snr":6,"cid":219069221},
//	{"mnc":15,"earfcn":1400,"pci":3,"rsrp":-93,"tac":20952,"mcc":460,"rsrq":-9,"snr":9,"cid":230437906},
//	{"mnc":0,"earfcn":3590,"pci":408,"rsrp":-96,"tac":20952,"mcc":460,"rsrq":-15,"snr":4,"cid":10020231},
//	{"mnc":0,"earfcn":3590,"pci":52,"rsrp":-98,"tac":20952,"mcc":460,"rsrq":-16,"snr":5,"cid":220699908},
//	{"mnc":0,"earfcn":38950,"pci":383,"rsrp":-99,"tac":20952,"mcc":460,"rsrq":-8,"snr":14,"cid":219069203},
//	{"mnc":0,"earfcn":1276,"pci":13,"rsrp":-102,"tac":25299,"mcc":460,"rsrq":-12,"snr":0,"cid":9983167},
//	{"mnc":0,"earfcn":1309,"pci":489,"rsrp":-103,"tac":20952,"mcc":460,"rsrq":-20,"snr":4,"cid":230546698}
//
// ]
func TestGetLocationLBS_NewCellDataV3(t *testing.T) {
	SetKey("0ce21bcb510ac0e1b978eb5c4a9525fb")
	// 当前连接的基站（有 rssi=-67 的那个）
	scell := CellTower{MCC: 460, MNC: 0, LAC: 20952, CI: 230437906, Signal: -67}
	// 周边基站列表（使用 rsrp 作为信号强度）
	cells := []CellTower{
		{MCC: 460, MNC: 0, LAC: 20952, CI: 219069221, Signal: -91},
		{MCC: 460, MNC: 15, LAC: 20952, CI: 230437906, Signal: -93},
		{MCC: 460, MNC: 0, LAC: 20952, CI: 10020231, Signal: -96},
		{MCC: 460, MNC: 0, LAC: 20952, CI: 220699908, Signal: -98},
		{MCC: 460, MNC: 0, LAC: 20952, CI: 219069203, Signal: -99},
		{MCC: 460, MNC: 0, LAC: 25299, CI: 9983167, Signal: -102},
		{MCC: 460, MNC: 0, LAC: 20952, CI: 230546698, Signal: -103},
	}
	location, err := GetLocationLBS(scell, cells)
	if err != nil {
		if !strings.Contains(err.Error(), "API 错误") {
			t.Errorf("GetLocationLBS 失败，错误信息：%v", err)
		} else {
			t.Errorf("API 返回错误：%v", err)
		}
		return
	}
	t.Logf("定位成功 - 经度：%f, 纬度：%f, 精度：%d米，地址：%s",
		location.Longitude, location.Latitude, location.Accuracy, location.Address)
}
