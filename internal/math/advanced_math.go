package math

import (
	stdmath "math"
)

type FastRandom struct {
	state uint64
}

func NewFastRandom(seed uint64) *FastRandom {
	return &FastRandom{state: seed}
}

func (fr *FastRandom) Next() uint64 {
	fr.state ^= fr.state >> 12
	fr.state ^= fr.state << 25
	fr.state ^= fr.state >> 27
	return fr.state * 2685821657736338717
}

func (fr *FastRandom) Float64() float64 {
	return float64(fr.Next()) / float64(^uint64(0))
}

func (fr *FastRandom) Float64Range(min, max float64) float64 {
	return min + fr.Float64()*(max-min)
}

func (fr *FastRandom) IntRange(min, max int) int {
	return min + int(fr.Next())%(max-min+1)
}

func FastSin(x float64) float64 {
	return stdmath.Sin(x)
}

func FastCos(x float64) float64 {
	return stdmath.Cos(x)
}

func FastTan(x float64) float64 {
	return stdmath.Tan(x)
}

func FastSqrt(x float64) float64 {
	return stdmath.Sqrt(x)
}

func FastPow(x, y float64) float64 {
	return stdmath.Pow(x, y)
}

func FastExp(x float64) float64 {
	return stdmath.Exp(x)
}

func FastLog(x float64) float64 {
	return stdmath.Log(x)
}

func FastAbs(x float64) float64 {
	return stdmath.Abs(x)
}

func FastMin(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func FastMax(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func FastClamp(x, min, max float64) float64 {
	return FastMax(min, FastMin(x, max))
}

func FastLerp(a, b, t float64) float64 {
	return a + t*(b-a)
}

func FastSmoothStep(edge0, edge1, x float64) float64 {
	t := FastClamp((x-edge0)/(edge1-edge0), 0.0, 1.0)
	return t * t * (3.0 - 2.0*t)
}

func FastStep(edge, x float64) float64 {
	if x < edge {
		return 0.0
	}
	return 1.0
}

func FastFract(x float64) float64 {
	return x - stdmath.Floor(x)
}

func FastMod(x, y float64) float64 {
	return x - y*stdmath.Floor(x/y)
}

func FastSign(x float64) float64 {
	if x > 0 {
		return 1.0
	}
	if x < 0 {
		return -1.0
	}
	return 0.0
}

func FastFloor(x float64) float64 {
	return stdmath.Floor(x)
}

func FastCeil(x float64) float64 {
	return stdmath.Ceil(x)
}

func FastRound(x float64) float64 {
	return stdmath.Round(x)
}

func FastTrunc(x float64) float64 {
	return stdmath.Trunc(x)
}

func FastIsNaN(x float64) bool {
	return stdmath.IsNaN(x)
}

func FastIsInf(x float64) bool {
	return stdmath.IsInf(x, 0)
}

func FastIsFinite(x float64) bool {
	return !stdmath.IsNaN(x) && !stdmath.IsInf(x, 0)
}

func FastDegToRad(degrees float64) float64 {
	return degrees * stdmath.Pi / 180.0
}

func FastRadToDeg(radians float64) float64 {
	return radians * 180.0 / stdmath.Pi
}

func FastAtan2(y, x float64) float64 {
	return stdmath.Atan2(y, x)
}

func FastAsin(x float64) float64 {
	return stdmath.Asin(x)
}

func FastAcos(x float64) float64 {
	return stdmath.Acos(x)
}

func FastAtan(x float64) float64 {
	return stdmath.Atan(x)
}

func FastSinh(x float64) float64 {
	return stdmath.Sinh(x)
}

func FastCosh(x float64) float64 {
	return stdmath.Cosh(x)
}

func FastTanh(x float64) float64 {
	return stdmath.Tanh(x)
}

func FastAsinh(x float64) float64 {
	return stdmath.Asinh(x)
}

func FastAcosh(x float64) float64 {
	return stdmath.Acosh(x)
}

func FastAtanh(x float64) float64 {
	return stdmath.Atanh(x)
}

func FastGamma(x float64) float64 {
	return stdmath.Gamma(x)
}

func FastLgamma(x float64) (lgamma float64, sign int) {
	return stdmath.Lgamma(x)
}

func FastErf(x float64) float64 {
	return stdmath.Erf(x)
}

func FastErfc(x float64) float64 {
	return stdmath.Erfc(x)
}

func FastJ0(x float64) float64 {
	return stdmath.J0(x)
}

func FastJ1(x float64) float64 {
	return stdmath.J1(x)
}

func FastJn(n int, x float64) float64 {
	return stdmath.Jn(n, x)
}

func FastY0(x float64) float64 {
	return stdmath.Y0(x)
}

func FastY1(x float64) float64 {
	return stdmath.Y1(x)
}

func FastYn(n int, x float64) float64 {
	return stdmath.Yn(n, x)
}

func FastHypot(p, q float64) float64 {
	return stdmath.Hypot(p, q)
}

func FastCbrt(x float64) float64 {
	return stdmath.Cbrt(x)
}

func FastLog10(x float64) float64 {
	return stdmath.Log10(x)
}

func FastLog2(x float64) float64 {
	return stdmath.Log2(x)
}

func FastLog1p(x float64) float64 {
	return stdmath.Log1p(x)
}

func FastExpm1(x float64) float64 {
	return stdmath.Expm1(x)
}

func FastCopysign(x, y float64) float64 {
	return stdmath.Copysign(x, y)
}

func FastDim(x, y float64) float64 {
	return stdmath.Dim(x, y)
}

func FastNextafter(x, y float64) float64 {
	return stdmath.Nextafter(x, y)
}

func FastNextafter32(x, y float32) float32 {
	return stdmath.Nextafter32(x, y)
}

func FastRemainder(x, y float64) float64 {
	return stdmath.Remainder(x, y)
}

func FastIlogb(x float64) int {
	return stdmath.Ilogb(x)
}

func FastLdexp(frac float64, exp int) float64 {
	return stdmath.Ldexp(frac, exp)
}

func FastFrexp(f float64) (frac float64, exp int) {
	return stdmath.Frexp(f)
}

func FastModf(f float64) (int float64, frac float64) {
	return stdmath.Modf(f)
}

func FastSincos(x float64) (sin, cos float64) {
	return stdmath.Sincos(x)
}

func FastSinhcosh(x float64) (sinh, cosh float64) {
	return stdmath.Sinh(x), stdmath.Cosh(x)
}

func FastPow10(n int) float64 {
	return stdmath.Pow10(n)
}

func FastSignbit(x float64) bool {
	return stdmath.Signbit(x)
}

func FastIsInfWithSign(x float64, sign int) bool {
	return stdmath.IsInf(x, sign)
}

func FastIsNormal(x float64) bool {
	return !stdmath.IsNaN(x) && !stdmath.IsInf(x, 0) && x != 0
}

func FastIsSubnormal(x float64) bool {
	return !stdmath.IsNaN(x) && !stdmath.IsInf(x, 0) && x != 0 && stdmath.Abs(x) < stdmath.SmallestNonzeroFloat64
}

func FastIsZero(x float64) bool {
	return x == 0
}

func FastIsPositive(x float64) bool {
	return x > 0
}

func FastIsNegative(x float64) bool {
	return x < 0
}

func FastIsNonNegative(x float64) bool {
	return x >= 0
}

func FastIsNonPositive(x float64) bool {
	return x <= 0
}

func FastIsInteger(x float64) bool {
	return x == stdmath.Trunc(x)
}

func FastIsEven(x float64) bool {
	return FastIsInteger(x) && int(x)%2 == 0
}

func FastIsOdd(x float64) bool {
	return FastIsInteger(x) && int(x)%2 == 1
}

func FastIsPowerOfTwo(x float64) bool {
	if x <= 0 || !FastIsInteger(x) {
		return false
	}
	n := int(x)
	return n > 0 && (n&(n-1)) == 0
}

func FastIsPerfectSquare(x float64) bool {
	if x < 0 || !FastIsInteger(x) {
		return false
	}
	n := int(x)
	root := int(FastSqrt(float64(n)))
	return root*root == n
}

func FastIsPerfectCube(x float64) bool {
	if !FastIsInteger(x) {
		return false
	}
	n := int(x)
	root := int(FastCbrt(float64(n)))
	return root*root*root == n
}

func FastIsPrime(x float64) bool {
	if x < 2 || !FastIsInteger(x) {
		return false
	}
	n := int(x)
	if n == 2 {
		return true
	}
	if n%2 == 0 {
		return false
	}
	for i := 3; i*i <= n; i += 2 {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func FastIsFibonacci(x float64) bool {
	if x < 0 || !FastIsInteger(x) {
		return false
	}
	n := int(x)
	if n == 0 || n == 1 {
		return true
	}
	
	a, b := 0, 1
	for b <= n {
		if b == n {
			return true
		}
		a, b = b, a+b
	}
	return false
}

func FastIsPalindrome(x float64) bool {
	if !FastIsInteger(x) {
		return false
	}
	n := int(x)
	if n < 0 {
		return false
	}
	
	original := n
	reversed := 0
	for n > 0 {
		reversed = reversed*10 + n%10
		n /= 10
	}
	return original == reversed
}

func FastIsArmstrong(x float64) bool {
	if !FastIsInteger(x) {
		return false
	}
	n := int(x)
	if n < 0 {
		return false
	}
	
	original := n
	digits := 0
	temp := n
	for temp > 0 {
		digits++
		temp /= 10
	}
	
	sum := 0
	temp = n
	for temp > 0 {
		digit := temp % 10
		sum += int(FastPow(float64(digit), float64(digits)))
		temp /= 10
	}
	
	return original == sum
}

func FastIsPerfect(x float64) bool {
	if x <= 0 || !FastIsInteger(x) {
		return false
	}
	n := int(x)
	sum := 0
	for i := 1; i < n; i++ {
		if n%i == 0 {
			sum += i
		}
	}
	return sum == n
}

func FastIsAbundant(x float64) bool {
	if x <= 0 || !FastIsInteger(x) {
		return false
	}
	n := int(x)
	sum := 0
	for i := 1; i < n; i++ {
		if n%i == 0 {
			sum += i
		}
	}
	return sum > n
}

func FastIsDeficient(x float64) bool {
	if x <= 0 || !FastIsInteger(x) {
		return false
	}
	n := int(x)
	sum := 0
	for i := 1; i < n; i++ {
		if n%i == 0 {
			sum += i
		}
	}
	return sum < n
}

func FastIsTriangular(x float64) bool {
	if x < 0 || !FastIsInteger(x) {
		return false
	}
	n := int(x)
	discriminant := 1 + 8*n
	root := int(FastSqrt(float64(discriminant)))
	return root*root == discriminant && (root-1)%2 == 0
}

func FastIsSquare(x float64) bool {
	if x < 0 {
		return false
	}
	root := int(FastSqrt(x))
	return float64(root*root) == x
}

func FastIsCube(x float64) bool {
	root := int(FastCbrt(x))
	return float64(root*root*root) == x
}

func FastIsPentagonal(x float64) bool {
	if x < 0 || !FastIsInteger(x) {
		return false
	}
	n := int(x)
	discriminant := 1 + 24*n
	root := int(FastSqrt(float64(discriminant)))
	return root*root == discriminant && (root+1)%6 == 0
}

func FastIsHexagonal(x float64) bool {
	if x < 0 || !FastIsInteger(x) {
		return false
	}
	n := int(x)
	discriminant := 1 + 8*n
	root := int(FastSqrt(float64(discriminant)))
	return root*root == discriminant && (root+1)%4 == 0
}

func FastIsHeptagonal(x float64) bool {
	if x < 0 || !FastIsInteger(x) {
		return false
	}
	n := int(x)
	discriminant := 9 + 40*n
	root := int(FastSqrt(float64(discriminant)))
	return root*root == discriminant && (root+3)%10 == 0
}

func FastIsOctagonal(x float64) bool {
	if x < 0 || !FastIsInteger(x) {
		return false
	}
	n := int(x)
	discriminant := 4 + 12*n
	root := int(FastSqrt(float64(discriminant)))
	return root*root == discriminant && (root+2)%6 == 0
}

func FastIsNonagonal(x float64) bool {
	if x < 0 || !FastIsInteger(x) {
		return false
	}
	n := int(x)
	discriminant := 25 + 56*n
	root := int(FastSqrt(float64(discriminant)))
	return root*root == discriminant && (root+5)%14 == 0
}

func FastIsDecagonal(x float64) bool {
	if x < 0 || !FastIsInteger(x) {
		return false
	}
	n := int(x)
	discriminant := 16 + 20*n
	root := int(FastSqrt(float64(discriminant)))
	return root*root == discriminant && (root+4)%10 == 0
} 