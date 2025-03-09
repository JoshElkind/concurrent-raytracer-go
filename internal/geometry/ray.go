package geometry

import (
	"raytraceGo/internal/math"
)

type HitRecord struct {
	T         float64
	Point     math.Vec3
	Normal    math.Vec3
	FrontFace bool
	Material  interface{}
}

type Hittable interface {
	Hit(ray Ray, tMin, tMax float64) (*HitRecord, bool)
}

type AABB struct {
	Min math.Vec3
	Max math.Vec3
}

type Ray struct {
	Origin    math.Vec3
	Direction math.Vec3
}

func NewRay(origin, direction math.Vec3) Ray {
	return Ray{
		Origin:    origin,
		Direction: direction,
	}
}

func (r Ray) At(t float64) math.Vec3 {
	return r.Origin.Add(r.Direction.MulScalar(t))
}

func (r Ray) PointAtParameter(t float64) math.Vec3 {
	return r.At(t)
}

func (r Ray) GetOrigin() math.Vec3 {
	return r.Origin
}

func (r Ray) GetDirection() math.Vec3 {
	return r.Direction
}

func (r Ray) GetParameter(t float64) math.Vec3 {
	return r.At(t)
}

func (r Ray) GetPointAtDistance(distance float64) math.Vec3 {
	return r.Origin.Add(r.Direction.Normalize().MulScalar(distance))
}

func (r Ray) GetDistanceToPoint(point math.Vec3) float64 {
	toPoint := point.Sub(r.Origin)
	projection := toPoint.Dot(r.Direction.Normalize())
	return projection
}

func (r Ray) GetClosestPointOnRay(point math.Vec3) math.Vec3 {
	toPoint := point.Sub(r.Origin)
	projection := toPoint.Dot(r.Direction.Normalize())
	return r.Origin.Add(r.Direction.Normalize().MulScalar(projection))
}

func (r Ray) GetDistanceToPointSquared(point math.Vec3) float64 {
	closestPoint := r.GetClosestPointOnRay(point)
	return point.Sub(closestPoint).LengthSquared()
}

func (r Ray) IsPointOnRay(point math.Vec3, tolerance float64) bool {
	distance := r.GetDistanceToPointSquared(point)
	return distance <= tolerance*tolerance
}

func (r Ray) GetReflectionDirection(normal math.Vec3) math.Vec3 {
	return r.Direction.Reflect(normal)
}

func (r Ray) GetRefractionDirection(normal math.Vec3, refractionIndex float64) math.Vec3 {
	return r.Direction.Refract(normal, refractionIndex)
}

func (r Ray) Transform(transformation func(math.Vec3) math.Vec3) Ray {
	return Ray{
		Origin:    transformation(r.Origin),
		Direction: transformation(r.Direction).Sub(transformation(math.Vec3{})).Normalize(),
	}
}

func (r Ray) Translate(offset math.Vec3) Ray {
	return Ray{
		Origin:    r.Origin.Add(offset),
		Direction: r.Direction,
	}
}

func (r Ray) Scale(factor float64) Ray {
	return Ray{
		Origin:    r.Origin.MulScalar(factor),
		Direction: r.Direction.Normalize(),
	}
}

func (r Ray) Rotate(axis math.Vec3, angle float64) Ray {
	cos := math.FastCos(angle)
	sin := math.FastSin(angle)
	
	rotationMatrix := func(v math.Vec3) math.Vec3 {
		return math.Vec3{
			X: v.X*(cos+axis.X*axis.X*(1-cos)) + v.Y*(axis.X*axis.Y*(1-cos)-axis.Z*sin) + v.Z*(axis.X*axis.Z*(1-cos)+axis.Y*sin),
			Y: v.X*(axis.Y*axis.X*(1-cos)+axis.Z*sin) + v.Y*(cos+axis.Y*axis.Y*(1-cos)) + v.Z*(axis.Y*axis.Z*(1-cos)-axis.X*sin),
			Z: v.X*(axis.Z*axis.X*(1-cos)-axis.Y*sin) + v.Y*(axis.Z*axis.Y*(1-cos)+axis.X*sin) + v.Z*(cos+axis.Z*axis.Z*(1-cos)),
		}
	}
	
	return r.Transform(rotationMatrix)
}

func (r Ray) GetBoundingBox() (min, max math.Vec3) {
	min = r.Origin
	max = r.Origin.Add(r.Direction)
	
	if min.X > max.X {
		min.X, max.X = max.X, min.X
	}
	if min.Y > max.Y {
		min.Y, max.Y = max.Y, min.Y
	}
	if min.Z > max.Z {
		min.Z, max.Z = max.Z, min.Z
	}
	
	return min, max
}

func (r Ray) GetLength() float64 {
	return r.Direction.Length()
}

func (r Ray) GetLengthSquared() float64 {
	return r.Direction.LengthSquared()
}

func (r Ray) IsValid() bool {
	return !r.Origin.NearZero() && !r.Direction.NearZero()
}

func (r Ray) IsParallel(other Ray) bool {
	cross := r.Direction.Cross(other.Direction)
	return cross.NearZero()
}

func (r Ray) IsPerpendicular(other Ray) bool {
	dot := r.Direction.Dot(other.Direction)
	return math.FastAbs(dot) < 1e-6
}

func (r Ray) GetAngle(other Ray) float64 {
	dot := r.Direction.Normalize().Dot(other.Direction.Normalize())
	dot = math.FastClamp(dot, -1.0, 1.0)
	return math.FastAcos(dot)
}

func (r Ray) GetDistanceToRay(other Ray) float64 {
	cross := r.Direction.Cross(other.Direction)
	if cross.NearZero() {
		return r.GetDistanceToPoint(other.Origin)
	}
	
	toOther := other.Origin.Sub(r.Origin)
	denominator := cross.LengthSquared()
	
	t1 := toOther.Cross(other.Direction).Dot(cross) / denominator
	t2 := toOther.Cross(r.Direction).Dot(cross) / denominator
	
	point1 := r.At(t1)
	point2 := other.At(t2)
	
	return point1.Sub(point2).Length()
} 