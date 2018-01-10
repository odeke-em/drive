package blen

func Blen(v uint64) int {
	l := int(0)

	for {
		if v <= 0xFF {
			break
		}
		v >>= 8
		l += 8
	}

	if v&0x80 != 0 {
		l += 8
	} else if v&0x40 != 0 {
		l += 7
	} else if v&0x20 != 0 {
		l += 6
	} else if v&0x10 != 0 {
		l += 5
	} else if v&0x8 != 0 {
		l += 4
	} else if v&0x4 != 0 {
		l += 3
	} else if v&0x2 != 0 {
		l += 2
	} else {
		l += 1
	}

	return l
}
