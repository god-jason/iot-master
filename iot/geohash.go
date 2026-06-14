package iot

import "math"

const geoHashBase32Alphabet = "0123456789bcdefghjkmnpqrstuvwxyz"

// GeoHashEncodeBase32 encodes WGS84 latitude/longitude into a base32 geohash.
// Precision is the number of characters to output (commonly 1..12).
func GeoHashEncodeBase32(latitude, longitude float64, precision int) string {
	if precision <= 0 {
		return ""
	}

	lat := clamp(latitude, -90, 90)
	lon := clamp(longitude, -180, 180)

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

		hash = append(hash, geoHashBase32Alphabet[ch])
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

