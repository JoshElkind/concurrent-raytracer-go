package math

import (
	"encoding/json"
	"fmt"
	"math"
)

type Vec3 struct {
	X, Y, Z float64
}

func NewVec3(x, y, z float64) Vec3 {
	return Vec3{X: x, Y: y, Z: z}
}

func (v Vec3) Add(other Vec3) Vec3 {
	return Vec3{X: v.X + other.X, Y: v.Y + other.Y, Z: v.Z + other.Z}
}

func (v Vec3) Sub(other Vec3) Vec3 {
	return Vec3{X: v.X - other.X, Y: v.Y - other.Y, Z: v.Z - other.Z}
}

func (v Vec3) Mul(other Vec3) Vec3 {
	return Vec3{X: v.X * other.X, Y: v.Y * other.Y, Z: v.Z * other.Z}
}

func (v Vec3) MulScalar(scalar float64) Vec3 {
	return Vec3{X: v.X * scalar, Y: v.Y * scalar, Z: v.Z * scalar}
}

func (v Vec3) DivScalar(scalar float64) Vec3 {
	return Vec3{X: v.X / scalar, Y: v.Y / scalar, Z: v.Z / scalar}
}

func (v Vec3) Dot(other Vec3) float64 {
	return v.X*other.X + v.Y*other.Y + v.Z*other.Z
}

func (v Vec3) Cross(other Vec3) Vec3 {
	return Vec3{
		X: v.Y*other.Z - v.Z*other.Y,
		Y: v.Z*other.X - v.X*other.Z,
		Z: v.X*other.Y - v.Y*other.X,
	}
}

func (v Vec3) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

func (v Vec3) FastLength() float64 {
	return FastSqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

func (v Vec3) LengthSquared() float64 {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z
}

func (v Vec3) Normalize() Vec3 {
	length := v.Length()
	if length == 0 {
		return Vec3{}
	}
	return v.DivScalar(length)
}

func (v Vec3) FastNormalize() Vec3 {
	length := v.FastLength()
	if length == 0 {
		return Vec3{}
	}
	return v.DivScalar(length)
}

func (v Vec3) Reflect(normal Vec3) Vec3 {
	return v.Sub(normal.MulScalar(2 * v.Dot(normal)))
}

func (v Vec3) Refract(normal Vec3, eta float64) Vec3 {
	cosTheta := v.Dot(normal)
	if cosTheta > 0 {
		normal = normal.MulScalar(-1)
		eta = 1 / eta
		cosTheta = -cosTheta
	}

	sinTheta2 := eta * eta * (1 - cosTheta*cosTheta)
	if sinTheta2 > 1 {
		return v.Reflect(normal)
	}

	cosTheta2 := math.Sqrt(1 - sinTheta2)
	return v.MulScalar(eta).Sub(normal.MulScalar(eta*cosTheta + cosTheta2))
}

func (v Vec3) Clamp(min, max float64) Vec3 {
	return Vec3{
		X: math.Max(min, math.Min(max, v.X)),
		Y: math.Max(min, math.Min(max, v.Y)),
		Z: math.Max(min, math.Min(max, v.Z)),
	}
}

func (v Vec3) ToRGB() (r, g, b uint8) {
	clamped := v.Clamp(0, 1)
	return uint8(clamped.X * 255), uint8(clamped.Y * 255), uint8(clamped.Z * 255)
}

func (v Vec3) NearZero() bool {
	const s = 1e-8
	return math.Abs(v.X) < s && math.Abs(v.Y) < s && math.Abs(v.Z) < s
}

func (v Vec3) Lerp(other Vec3, t float64) Vec3 {
	return v.Add(other.Sub(v).MulScalar(t))
}

func Lerp(a, b Vec3, t float64) Vec3 {
	return a.MulScalar(1-t).Add(b.MulScalar(t))
}

func RandomVec3() Vec3 {
	return Vec3{
		X: RandomFloat(),
		Y: RandomFloat(),
		Z: RandomFloat(),
	}
}

func RandomVec3InUnitSphere() Vec3 {
	for {
		p := RandomVec3().MulScalar(2).Sub(Vec3{X: 1, Y: 1, Z: 1})
		if p.LengthSquared() < 1 {
			return p
		}
	}
}

func RandomVec3InUnitDisk() Vec3 {
	for {
		p := Vec3{
			X: RandomFloat()*2 - 1,
			Y: RandomFloat()*2 - 1,
			Z: 0,
		}
		if p.LengthSquared() < 1 {
			return p
		}
	}
}

func RandomUnitVector() Vec3 {
	return RandomVec3InUnitSphere().Normalize()
}

func RandomVec3InHemisphere(normal Vec3) Vec3 {
	inUnitSphere := RandomVec3InUnitSphere()
	if inUnitSphere.Dot(normal) > 0 {
		return inUnitSphere
	}
	return inUnitSphere.MulScalar(-1)
}

func Vec3Distance(a, b Vec3) float64 {
	diff := a.Sub(b)
	return diff.Length()
}

func FastVec3Distance(a, b Vec3) float64 {
	diff := a.Sub(b)
	return diff.FastLength()
}

func (v *Vec3) UnmarshalJSON(data []byte) error {
	var arr []float64
	if err := json.Unmarshal(data, &arr); err == nil {
		if len(arr) != 3 {
			return fmt.Errorf("expected 3 elements for Vec3, got %d", len(arr))
		}
		v.X, v.Y, v.Z = arr[0], arr[1], arr[2]
		return nil
	}
	var obj struct {
		X, Y, Z float64
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	v.X, v.Y, v.Z = obj.X, obj.Y, obj.Z
	return nil
}

func (v Vec3) MarshalJSON() ([]byte, error) {
	return json.Marshal([]float64{v.X, v.Y, v.Z})
} 