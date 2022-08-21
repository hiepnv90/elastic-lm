package common

import (
	"math/big"
)

func GetSqrtRatioAtTick(tick int) *big.Int {
	var absTick uint
	if tick < 0 {
		absTick = uint(-tick)
	} else {
		absTick = uint(tick)
	}

	var ratio *big.Int
	if absTick&0x1 != 0 {
		ratio = NewBigIntFromHex("fffcb933bd6fad37aa2d162d1a594001")
	} else {
		ratio = NewBigIntFromHex("100000000000000000000000000000000")
	}

	if absTick&0x2 != 0 {
		ratio = BigShiftRight(BigMul(ratio, NewBigIntFromHex("fff97272373d413259a46990580e213a")), 128)
	}
	if absTick&0x4 != 0 {
		ratio = BigShiftRight(BigMul(ratio, NewBigIntFromHex("fff2e50f5f656932ef12357cf3c7fdcc")), 128)
	}
	if absTick&0x8 != 0 {
		ratio = BigShiftRight(BigMul(ratio, NewBigIntFromHex("ffe5caca7e10e4e61c3624eaa0941cd0")), 128)
	}
	if absTick&0x10 != 0 {
		ratio = BigShiftRight(BigMul(ratio, NewBigIntFromHex("ffcb9843d60f6159c9db58835c926644")), 128)
	}
	if absTick&0x20 != 0 {
		ratio = BigShiftRight(BigMul(ratio, NewBigIntFromHex("ff973b41fa98c081472e6896dfb254c0")), 128)
	}
	if absTick&0x40 != 0 {
		ratio = BigShiftRight(BigMul(ratio, NewBigIntFromHex("ff2ea16466c96a3843ec78b326b52861")), 128)
	}
	if absTick&0x80 != 0 {
		ratio = BigShiftRight(BigMul(ratio, NewBigIntFromHex("fe5dee046a99a2a811c461f1969c3053")), 128)
	}
	if absTick&0x100 != 0 {
		ratio = BigShiftRight(BigMul(ratio, NewBigIntFromHex("fcbe86c7900a88aedcffc83b479aa3a4")), 128)
	}
	if absTick&0x200 != 0 {
		ratio = BigShiftRight(BigMul(ratio, NewBigIntFromHex("f987a7253ac413176f2b074cf7815e54")), 128)
	}
	if absTick&0x400 != 0 {
		ratio = BigShiftRight(BigMul(ratio, NewBigIntFromHex("f3392b0822b70005940c7a398e4b70f3")), 128)
	}
	if absTick&0x800 != 0 {
		ratio = BigShiftRight(BigMul(ratio, NewBigIntFromHex("e7159475a2c29b7443b29c7fa6e889d9")), 128)
	}
	if absTick&0x1000 != 0 {
		ratio = BigShiftRight(BigMul(ratio, NewBigIntFromHex("d097f3bdfd2022b8845ad8f792aa5825")), 128)
	}
	if absTick&0x2000 != 0 {
		ratio = BigShiftRight(BigMul(ratio, NewBigIntFromHex("a9f746462d870fdf8a65dc1f90e061e5")), 128)
	}
	if absTick&0x4000 != 0 {
		ratio = BigShiftRight(BigMul(ratio, NewBigIntFromHex("70d869a156d2a1b890bb3df62baf32f7")), 128)
	}
	if absTick&0x8000 != 0 {
		ratio = BigShiftRight(BigMul(ratio, NewBigIntFromHex("31be135f97d08fd981231505542fcfa6")), 128)
	}
	if absTick&0x10000 != 0 {
		ratio = BigShiftRight(BigMul(ratio, NewBigIntFromHex("9aa508b5b7a84e1c677de54f3e99bc9")), 128)
	}
	if absTick&0x20000 != 0 {
		ratio = BigShiftRight(BigMul(ratio, NewBigIntFromHex("5d6af8dedb81196699c329225ee604")), 128)
	}
	if absTick&0x40000 != 0 {
		ratio = BigShiftRight(BigMul(ratio, NewBigIntFromHex("2216e584f5fa1ea926041bedfe98")), 128)
	}
	if absTick&0x80000 != 0 {
		ratio = BigShiftRight(BigMul(ratio, NewBigIntFromHex("48a170391f7dc42444e8fa2")), 128)
	}

	if tick > 0 {
		ratio = BigDiv(MaxUint256, ratio)
	}

	remain := Big0
	if !BigIsZero(BigMod(ratio, big.NewInt(1<<32))) {
		remain = Big1
	}
	return BigAdd(BigShiftRight(ratio, 32), remain)
}
