package main

func GG32LoginHash(password string, seed uint32) uint32 {
	var x, y, z uint32

	y = seed

	for i := 0; i < len(password); i++ {
		x = (x & 0xffffff00) | uint32(password[i])
		y ^= x
		y += x
		x <<= 8
		y ^= x
		x <<= 8
		y -= x
		x <<= 8
		y ^= x

		z = y & 0x1f
		y = (y << z) | (y >> (32 - z))
	}

	return y
}
