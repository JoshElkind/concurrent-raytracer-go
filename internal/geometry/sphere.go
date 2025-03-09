package geometry

import (
	stdmath "math"
	"raytraceGo/internal/math"
)

type Sphere struct {
	Center   math.Vec3
	Radius   float64
	Material interface{}
}

func NewSphere(center math.Vec3, radius float64, material interface{}) *Sphere {
	return &Sphere{
		Center:   center,
		Radius:   radius,
		Material: material,
	}
}

func (s *Sphere) Hit(ray Ray, tMin, tMax float64) (*HitRecord, bool) {
	oc := ray.Origin.Sub(s.Center)
	a := ray.Direction.LengthSquared()
	halfB := oc.Dot(ray.Direction)
	c := oc.LengthSquared() - s.Radius*s.Radius
	
	discriminant := halfB*halfB - a*c
	if discriminant < 0 {
		return nil, false
	}
	
	sqrtd := stdmath.Sqrt(discriminant)
	root := (-halfB - sqrtd) / a
	if root < tMin || tMax < root {
		root = (-halfB + sqrtd) / a
		if root < tMin || tMax < root {
			return nil, false
		}
	}
	
	t := root
	point := ray.At(t)
	outwardNormal := point.Sub(s.Center).DivScalar(s.Radius)
	
	frontFace := ray.Direction.Dot(outwardNormal) < 0
	normal := outwardNormal
	if !frontFace {
		normal = outwardNormal.MulScalar(-1)
	}
	
	return &HitRecord{
		T:         t,
		Point:     point,
		Normal:    normal,
		FrontFace: frontFace,
		Material:  s.Material,
	}, true
}

func (s *Sphere) GetCenter() math.Vec3 {
	return s.Center
}

func (s *Sphere) GetRadius() float64 {
	return s.Radius
}

func (s *Sphere) GetMaterial() interface{} {
	return s.Material
}

func (s *Sphere) GetBoundingBox() (min, max math.Vec3) {
	radiusVec := math.Vec3{X: s.Radius, Y: s.Radius, Z: s.Radius}
	min = s.Center.Sub(radiusVec)
	max = s.Center.Add(radiusVec)
	return min, max
}

func (s *Sphere) GetSurfaceArea() float64 {
	return 4.0 * stdmath.Pi * s.Radius * s.Radius
}

func (s *Sphere) GetVolume() float64 {
	return (4.0 / 3.0) * stdmath.Pi * s.Radius * s.Radius * s.Radius
}

func (s *Sphere) ContainsPoint(point math.Vec3) bool {
	distanceSquared := point.Sub(s.Center).LengthSquared()
	return distanceSquared <= s.Radius*s.Radius
}

func (s *Sphere) GetClosestPoint(point math.Vec3) math.Vec3 {
	direction := point.Sub(s.Center).Normalize()
	return s.Center.Add(direction.MulScalar(s.Radius))
}

func (s *Sphere) GetDistanceToPoint(point math.Vec3) float64 {
	distance := point.Sub(s.Center).Length()
	return stdmath.Max(0, distance-s.Radius)
}

func (s *Sphere) GetNormalAtPoint(point math.Vec3) math.Vec3 {
	return point.Sub(s.Center).Normalize()
}

func (s *Sphere) IntersectsWith(other *Sphere) bool {
	distance := s.Center.Sub(other.Center).Length()
	return distance <= s.Radius+other.Radius
}

func (s *Sphere) GetIntersectionVolume(other *Sphere) float64 {
	if !s.IntersectsWith(other) {
		return 0.0
	}
	
	distance := s.Center.Sub(other.Center).Length()
	if distance >= s.Radius+other.Radius {
		return 0.0
	}
	
	if distance <= stdmath.Abs(s.Radius-other.Radius) {
		smallerRadius := stdmath.Min(s.Radius, other.Radius)
		return (4.0 / 3.0) * stdmath.Pi * smallerRadius * smallerRadius * smallerRadius
	}
	
	h := (s.Radius + other.Radius - distance) / 2.0
	volume := stdmath.Pi * h * h * (3.0*(s.Radius+other.Radius) - h) / 3.0
	return volume
} 