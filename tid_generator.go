package twitterinternalapi

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"
)

const totalTime = 4096.0

// TransactionIDGenerator generates X-Client-Transaction-ID headers
type TransactionIDGenerator struct {
	frames [][][]int
}

// NewTransactionIDGenerator creates a new generator with frame data
func NewTransactionIDGenerator(frames [][][]int) *TransactionIDGenerator {
	rand.Seed(time.Now().UnixNano())
	return &TransactionIDGenerator{
		frames: frames,
	}
}

// GenerateHeader generates an X-Client-Transaction-ID header
func (g *TransactionIDGenerator) GenerateHeader(path, method, key string) string {
	if len(g.frames) == 0 {
		return ""
	}

	keyBytes := []int{}
	decodedKey := atob(key)
	for i := 0; i < len(decodedKey); i++ {
		keyBytes = append(keyBytes, charCodeAt(decodedKey, i))
	}

	timeNow := uint32((time.Now().UnixMilli() - 1682924400*1000) / 1000)
	timeNowBytes := timeToBytes(timeNow)

	row := g.frames[keyBytes[5]%4][keyBytes[2]%16]
	targetTime := float64(keyBytes[12]%16*(keyBytes[14]%16)*(keyBytes[7]%16)) / totalTime
	fromColor := []float64{float64(row[0]), float64(row[1]), float64(row[2]), 1.0}
	toColor := []float64{float64(row[3]), float64(row[4]), float64(row[5]), 1.0}

	fromRotation := []float64{0.0}
	toRotation := []float64{normalizeSVGNumber(float64(row[6]), 60.0, 360.0)}
	row = row[7:]
	curves := [4]float64{}
	for i := 0; i < len(row); i++ {
		curves[i] = normalizeSVGNumber(float64(row[i]), svgBValue(i), 1.0)
	}

	c := &cubic{Curves: curves}
	val := c.getValue(targetTime)
	color := interpolate(fromColor, toColor, val)
	rotation := interpolate(fromRotation, toRotation, val)
	matrix := convertRotationToMatrix(rotation[0])

	strArr := []string{}
	for i := 0; i < len(color)-1; i++ {
		strArr = append(strArr, hex.EncodeToString([]byte{byte(math.Round(color[i]))}))
	}
	for i := 0; i < len(matrix)-2; i++ {
		rounded := toFixed(matrix[i], 2)
		if rounded < 0 {
			rounded = -rounded
		}
		strArr = append(strArr, "0"+strings.ToLower(floatToHex(rounded)[1:]))
	}
	strArr = append(strArr, "0", "0")

	hash := sha256.Sum256([]byte(fmt.Sprintf(`%s!%s!%vbird%s`, method, path, timeNow, strings.Join(strArr, ""))))
	hashBytes := []int{}
	for i := 0; i < len(hash)-16; i++ {
		hashBytes = append(hashBytes, int(hash[i]))
	}

	xorByte := rand.Intn(256)
	bytes := []int{xorByte}
	bytes = append(bytes, keyBytes...)
	bytes = append(bytes, timeNowBytes...)
	bytes = append(bytes, hashBytes...)
	bytes = append(bytes, 1)

	out := []byte{}
	for i := 0; i < len(bytes); i++ {
		if i == 0 {
			out = append(out, byte(bytes[i]))
			continue
		}
		out = append(out, byte(bytes[i]^xorByte))
	}

	return strings.ReplaceAll(btoa(out), "=", "")
}

// cubic bezier curve implementation
type cubic struct {
	Curves [4]float64
}

func (c *cubic) getValue(time float64) float64 {
	startGradient := 0.0
	endGradient := 0.0

	if time <= 0.0 {
		if c.Curves[0] > 0.0 {
			startGradient = c.Curves[1] / c.Curves[0]
		} else if c.Curves[1] == 0.0 && c.Curves[2] > 0.0 {
			startGradient = c.Curves[3] / c.Curves[2]
		}
		return startGradient * time
	}

	if time >= 1.0 {
		if c.Curves[2] < 1.0 {
			endGradient = (c.Curves[3] - 1.0) / (c.Curves[2] - 1.0)
		} else if c.Curves[2] == 1.0 && c.Curves[0] < 1.0 {
			endGradient = (c.Curves[1] - 1.0) / (c.Curves[0] - 1.0)
		}
		return 1.0 + endGradient*(time-1.0)
	}

	start := 0.0
	end := 1.0
	mid := 0.0
	for start < end {
		mid = (start + end) / 2
		xEst := cubicF(c.Curves[0], c.Curves[2], mid)
		if abs(time-xEst) < 0.00001 {
			return cubicF(c.Curves[1], c.Curves[3], mid)
		}
		if xEst < time {
			start = mid
		} else {
			end = mid
		}
	}
	return cubicF(c.Curves[1], c.Curves[3], mid)
}

func cubicF(a, b, m float64) float64 {
	return 3.0*a*(1-m)*(1-m)*m + 3.0*b*(1-m)*m*m + m*m*m
}

func abs(in float64) float64 {
	if in < 0 {
		return -in
	}
	return in
}

// interpolation functions
func interpolate(from, to []float64, f float64) []float64 {
	out := []float64{}
	for i := 0; i < len(from); i++ {
		out = append(out, interpolateNum(from[i], to[i], f))
	}
	return out
}

func interpolateNum(from, to, f float64) float64 {
	return from*(1-f) + to*f
}

// rotation matrix conversion
func convertRotationToMatrix(degrees float64) []float64 {
	radians := degrees * math.Pi / 180
	c := math.Cos(radians)
	s := math.Sin(radians)
	return []float64{c, s, -s, c, 0, 0}
}

// utility functions
func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

func floatToHex(x float64) string {
	var result []byte
	quotient := int(x)
	fraction := x - float64(quotient)

	for quotient > 0 {
		quotient = int(x / 16)
		remainder := int(x - (float64(quotient) * 16))

		if remainder > 9 {
			result = append([]byte{byte(remainder + 55)}, result...)
		} else {
			result = append([]byte{byte(rune(remainder + 48))}, result...)
		}

		x = float64(quotient)
	}

	if fraction == 0 {
		return string(result)
	}

	result = append(result, '.')

	for fraction > 0 {
		fraction = fraction * 16
		integer := int(fraction)
		fraction = fraction - float64(integer)

		if integer > 9 {
			result = append(result, byte(integer+55))
		} else {
			result = append(result, byte(integer+48))
		}
	}

	return string(result)
}

func normalizeSVGNumber(b, c, d float64) float64 {
	return b*(d-c)/255 + c
}

func svgBValue(a int) float64 {
	if a%2 == 1 {
		return -1.0
	}
	return 0.0
}

func btoa(str []byte) string {
	return base64.StdEncoding.EncodeToString(str)
}

func timeToBytes(val uint32) []int {
	r := make([]int, 4)
	for i := uint32(0); i < 4; i++ {
		r[i] = int((val >> (8 * i)) & 0xff)
	}
	return r
}

func atob(input string) string {
	data, err := base64.RawStdEncoding.DecodeString(input)
	if err != nil {
		return ""
	}
	return string(data)
}

func charCodeAt(a string, i int) int {
	if i < len(a) {
		return int(a[i])
	}
	return 0
}
