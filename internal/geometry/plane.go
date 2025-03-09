package geometry

import (
	"raytraceGo/internal/math"
)

type Plane struct {
	Point    math.Vec3
	Normal   math.Vec3
	Material interface{}
}

func NewPlane(point, normal math.Vec3, material interface{}) *Plane {
	return &Plane{
		Point:    point,
		Normal:   normal.Normalize(),
		Material: material,
	}
}

func (p *Plane) Hit(ray Ray, tMin, tMax float64) (*HitRecord, bool) {
	denominator := ray.Direction.Dot(p.Normal)
	if denominator == 0 {
		return nil, false
	}
	
	t := p.Point.Sub(ray.Origin).Dot(p.Normal) / denominator
	if t < tMin || t > tMax {
		return nil, false
	}
	
	point := ray.At(t)
	frontFace := ray.Direction.Dot(p.Normal) < 0
	normal := p.Normal
	if !frontFace {
		normal = p.Normal.MulScalar(-1)
	}
	
	return &HitRecord{
		T:         t,
		Point:     point,
		Normal:    normal,
		FrontFace: frontFace,
		Material:  p.Material,
	}, true
}

func (p *Plane) GetPoint() math.Vec3 {
	return p.Point
}

func (p *Plane) GetNormal() math.Vec3 {
	return p.Normal
}

func (p *Plane) GetMaterial() interface{} {
	return p.Material
}

func (p *Plane) GetDistanceToPoint(point math.Vec3) float64 {
	return point.Sub(p.Point).Dot(p.Normal)
}

func (p *Plane) GetClosestPoint(point math.Vec3) math.Vec3 {
	distance := p.GetDistanceToPoint(point)
	return point.Sub(p.Normal.MulScalar(distance))
}

func (p *Plane) IsPointOnPlane(point math.Vec3, tolerance float64) bool {
	distance := p.GetDistanceToPoint(point)
	return distance <= tolerance
} 