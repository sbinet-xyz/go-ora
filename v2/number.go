package go_ora

import "github.com/sijms/go-ora/v2/converters"

type Num []byte

func (n Num) Uint() uint64 {
	return uint64(n.Int())
}

func (n Num) Int() int64 {
	v, err := toInt64([]byte(n))
	if err != nil {
		panic(err)
	}
	return v
}

func (n Num) Float() float64 {
	return converters.DecodeDouble(n)
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
