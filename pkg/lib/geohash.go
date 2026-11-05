package lib

import "math"

// GeoHashEncodeBase32 编码 WGS84 经纬度为 Base32 GeoHash 字符串
// precision 是输出字符数，常用 8（~38m 精度）或 12（~3m 精度）
func GeoHashEncodeBase32(latitude, longitude float64, precision int) string {
	if precision <= 0 {
		return ""
	}

	lat := clamp(latitude, -90, 90)
	lon := clamp(longitude, -180, 180)

	const alphabet = "0123456789bcdefghjkmnpqrstuvwxyz"
	var (
		latMin, latMax = -90.0, 90.0
		lonMin, lonMax = -180.0, 180.0

		bit      = 0
		ch       = 0
		evenBit  = true
		hash     = make([]byte, 0, precision)
		bitsMask = []int{16, 8, 4, 2, 1}
	)

	for len(hash) < precision {
		if evenBit {
			mid := (lonMin + lonMax) / 2
			if lon >= mid {
				ch |= bitsMask[bit]
				lonMin = mid
			} else {
				lonMax = mid
			}
		} else {
			mid := (latMin + latMax) / 2
			if lat >= mid {
				ch |= bitsMask[bit]
				latMin = mid
			} else {
				latMax = mid
			}
		}

		evenBit = !evenBit
		if bit < 4 {
			bit++
			continue
		}

		hash = append(hash, alphabet[ch])
		bit = 0
		ch = 0
	}

	return string(hash)
}

func clamp(v, min, max float64) float64 {
	if math.IsNaN(v) {
		return min
	}
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
