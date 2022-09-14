package converters

import (
	"encoding/hex"
	"fmt"
)

const (
	Infinity              = "Infinity"
	InvalidInputNumberMsg = "Invalid Input Number %s "
)

var (
	int64MaxByte = []byte{202, 10, 23, 34, 73, 4, 69, 55, 78, 59, 8}
	int64MinByte = []byte{53, 92, 79, 68, 29, 98, 33, 47, 24, 43, 93, 102}
)

type Number []byte

func (n Number) String() string {
	v, err := numToString(n)
	if err != nil {
		panic(err)
	}
	return v
}

func (n Number) Uint() uint64 {
	return uint64(n.Int())
}

func (n Number) Int() int64 {
	v, err := toInt64([]byte(n))
	if err != nil {
		panic(err)
	}
	return v
}

func (n Number) Float() float64 {
	return DecodeDouble(n)
}

func toInt64(data []byte) (int64, error) {
	positive := isPositive(data)
	var (
		b1 int
		b2 int
		b3 int
	)
	if positive {
		b1 = int(data[0]&0x7f - 65)
	} else {
		b1 = int(((data[0] ^ 0xFF) & 0x7F) - 65)
	}
	if positive || (len(data) == 21 && data[20] != 102) {
		b2 = len(data) - 1
	} else {
		b2 = len(data) - 2
	}
	if b2 > b1+1 {
		b3 = b1 + 1
	} else {
		b3 = b2
	}
	var l int64
	if positive {
		for i := 0; i < b3; i++ {
			l = l*100 + (int64(data[i+1]) - 1)
		}
	} else {
		for i := 0; i < b3; i++ {
			l = l*100 + (101 - int64(data[i+1]))
		}
	}
	for i := b1 - b2; i >= 0; i-- {
		l *= 100
	}
	if positive {
		return l, nil
	}
	return -l, nil
}

func isPositive(p []byte) bool {
	return p[0]&128 != 0
}

func numToString(b []byte) (string, error) {
	if isZero(b) {
		return "0", nil
	} else if isPosInf(b) {
		return Infinity, nil
	} else if isNegInf(b) {
		return Infinity, nil
	} else if !isValid(b) {
		return "", fmt.Errorf(InvalidInputNumberMsg, hex.EncodeToString(b))
	}
	var (
		pos     = 0
		dataLen int // data length for after convert
	)
	// convert normal byte
	data := fromLnxFmt(b)
	exponent := int(data[0]) // Unsigned integers do not appear negative
	if exponent > 127 {
		exponent = exponent - 256
	}
	realDataLen := len(data) - 1
	k := exponent - (realDataLen - 1)
	if k >= 0 {
		dataLen = 2*(exponent+1) + 1
	} else if exponent >= 0 {
		dataLen = 2 * (realDataLen + 1)
	} else {
		dataLen = 2*(realDataLen-exponent) + 3
	}

	result := make([]byte, dataLen)
	if !isPositive(b) {
		result[pos] = '-'
		pos++
	}
	var b1 int
	if k >= 0 {
		pos += byteToChars(data[1], result, pos)
		for b1 = 2; b1 <= realDataLen; exponent-- {
			byteTo2Chars(data[b1], result, pos)
			pos += 2
			b1++
		}
		if exponent > 0 {
			for exponent > 0 {
				result[pos] = '0'
				pos++
				result[pos] = '0'
				pos++
				exponent--
			}
		}
	} else {
		n := realDataLen + k
		if n > 0 {
			pos += byteToChars(data[1], result, pos)
			if n == 1 {
				result[pos] = '.'
				pos++
			}
			for b1 = 2; b1 < realDataLen; b1++ {
				byteTo2Chars(data[b1], result, pos)
				pos += 2
				if n == b1 {
					result[pos] = '.'
					pos++
				}
			}
			if data[b1]%10 == 0 {
				pos += byteToChars(data[b1]/10, result, pos)
			} else {
				byteTo2Chars(data[b1], result, pos)
				pos += 2
			}
		} else {
			result[pos] = '0'
			pos++
			result[pos] = '.'
			pos++
			for n < 0 {
				n++
				result[pos] = '0'
				pos++
				result[pos] = '0'
				pos++
			}

			for b1 = 1; b1 < realDataLen; b1++ {
				byteTo2Chars(data[b1], result, pos)
				pos += 2
			}

			if data[b1]%10 == 0 {
				pos += byteToChars(data[b1]/10, result, pos)
			} else {
				byteTo2Chars(data[b1], result, pos)
				pos += 2
			}
		}
	}
	return string(result[:pos]), nil
}

func isZero(b []byte) bool {
	return b[0] == 128 && len(b) == 1
}

func isNegInf(b []byte) bool {
	return b[0] == 0 && len(b) == 1
}

func isPosInf(b []byte) bool {
	// -1 =255
	return len(b) == 2 && b[0] == 255 && b[1] == 101
}

func isInf(b []byte) bool {
	if (len(b) == 2 && b[0] == 255 && b[1] == 101) || (b[0] == 0 && len(b) == 1) {
		return true
	}
	return false
}

func isValid(b []byte) bool {
	var (
		dataLen  = len(b)
		tempByte byte
		pos      int
	)
	if isPositive(b) {
		if dataLen == 1 {
			return isZero(b)
		} else if b[0] == 255 && b[1] == 101 {
			return dataLen == 2
		} else if dataLen > 21 {
			return false
		} else if b[1] >= 2 && b[dataLen-1] >= 2 {
			for pos = 1; pos < dataLen; pos++ {
				tempByte = b[pos]
				if tempByte < 1 || tempByte > 100 {
					return false
				}
			}
			return true
		} else {
			return false
		}
	} else if dataLen < 3 {
		return isNegInf(b)
	} else if dataLen > 21 {
		return false
	} else {
		if b[dataLen-1] != 102 {
			if dataLen <= 20 {
				return false
			}
		} else {
			dataLen--
		}
		if b[1] <= 100 && b[dataLen-1] <= 100 {
			for pos = 1; pos < dataLen; pos++ {
				tempByte = b[pos]
				if tempByte < 2 || tempByte > 101 {
					return false
				}
			}
			return true
		} else {
			return false
		}
	}
}

func fromLnxFmt(b []byte) []byte {
	bLen := len(b)
	var newData []byte
	if isPositive(b) {
		newData = make([]byte, bLen)
		newData[0] = b[0]&0x7f - 65
		for i := 1; i < bLen; i++ {
			newData[i] = b[i] - 1
		}
	} else {
		if bLen-1 == 20 && b[bLen-1] != 102 {
			newData = make([]byte, bLen)
		} else {
			newData = make([]byte, bLen-1)
		}

		newData[0] = ((b[0] ^ 0xFF) & 0x7F) - 65
		for i := 1; i < len(newData); i++ {
			newData[i] = 101 - b[i]
		}
	}
	return newData
}

func byteToChars(paramByte byte, result []byte, pos int) int {
	if paramByte < 0 {
		return 0
	} else if paramByte < 10 {
		result[pos] = 48 + paramByte
		return 1
	} else if paramByte < 100 {
		result[pos] = 48 + paramByte/10
		result[pos+1] = 48 + paramByte%10
		return 2
	} else {
		result[pos] = '1'
		result[pos+1] = 48 + paramByte/10 - 10
		result[pos+2] = 48 + paramByte%10
		return 3
	}
}

func byteTo2Chars(paramByte byte, result []byte, pos int) {
	if paramByte < 0 {
		result[pos] = '0'
		result[pos+1] = '0'
	} else if paramByte < 10 {
		result[pos] = '0'
		result[pos+1] = 48 + paramByte
	} else if paramByte < 100 {
		result[pos] = 48 + paramByte/10
		result[pos+1] = 48 + paramByte%10
	} else {
		result[pos] = '0'
		result[pos+1] = '0'
	}
}
