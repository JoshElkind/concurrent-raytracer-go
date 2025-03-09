package geometry

import (
	"raytraceGo/internal/math"
)

type Triangle struct {
	Vertices [3]math.Vec3
	Normals  [3]math.Vec3
	Material interface{}
}

func NewTriangle(v0, v1, v2 math.Vec3, material interface{}) *Triangle {
	normal := calculateNormal(v0, v1, v2)
	return &Triangle{
		Vertices: [3]math.Vec3{v0, v1, v2},
		Normals:  [3]math.Vec3{normal, normal, normal},
		Material: material,
	}
}

func NewTriangleWithNormals(v0, v1, v2 math.Vec3, n0, n1, n2 math.Vec3, material interface{}) *Triangle {
	return &Triangle{
		Vertices: [3]math.Vec3{v0, v1, v2},
		Normals:  [3]math.Vec3{n0, n1, n2},
		Material: material,
	}
}

func calculateNormal(v0, v1, v2 math.Vec3) math.Vec3 {
	edge1 := v1.Sub(v0)
	edge2 := v2.Sub(v0)
	return edge1.Cross(edge2).Normalize()
}

func (t *Triangle) Hit(ray Ray, tMin, tMax float64) (*HitRecord, bool) {
	edge1 := t.Vertices[1].Sub(t.Vertices[0])
	edge2 := t.Vertices[2].Sub(t.Vertices[0])
	h := ray.Direction.Cross(edge2)
	a := edge1.Dot(h)
	
	if a > -1e-6 && a < 1e-6 {
		return nil, false
	}
	
	f := 1.0 / a
	s := ray.Origin.Sub(t.Vertices[0])
	u := f * s.Dot(h)
	
	if u < 0.0 || u > 1.0 {
		return nil, false
	}
	
	q := s.Cross(edge1)
	v := f * ray.Direction.Dot(q)
	
	if v < 0.0 || u+v > 1.0 {
		return nil, false
	}
	
	t_val := f * edge2.Dot(q)
	
	if t_val < tMin || t_val > tMax {
		return nil, false
	}
	
	point := ray.At(t_val)
	
	normal := t.calculateInterpolatedNormal(u, v)
	frontFace := ray.Direction.Dot(normal) < 0
	if !frontFace {
		normal = normal.MulScalar(-1)
	}
	
	return &HitRecord{
		T:         t_val,
		Point:     point,
		Normal:    normal,
		FrontFace: frontFace,
		Material:  t.Material,
	}, true
}

func (t *Triangle) calculateInterpolatedNormal(u, v float64) math.Vec3 {
	w := 1.0 - u - v
	normal := t.Normals[0].MulScalar(w).Add(t.Normals[1].MulScalar(u)).Add(t.Normals[2].MulScalar(v))
	return normal.Normalize()
}

func (t *Triangle) GetVertices() [3]math.Vec3 {
	return t.Vertices
}

func (t *Triangle) GetNormals() [3]math.Vec3 {
	return t.Normals
}

func (t *Triangle) GetMaterial() interface{} {
	return t.Material
}

func (t *Triangle) GetBoundingBox() (min, max math.Vec3) {
	min = t.Vertices[0]
	max = t.Vertices[0]
	
	for i := 1; i < 3; i++ {
		vertex := t.Vertices[i]
		if vertex.X < min.X {
			min.X = vertex.X
		}
		if vertex.Y < min.Y {
			min.Y = vertex.Y
		}
		if vertex.Z < min.Z {
			min.Z = vertex.Z
		}
		if vertex.X > max.X {
			max.X = vertex.X
		}
		if vertex.Y > max.Y {
			max.Y = vertex.Y
		}
		if vertex.Z > max.Z {
			max.Z = vertex.Z
		}
	}
	
	return min, max
}

func (t *Triangle) GetArea() float64 {
	edge1 := t.Vertices[1].Sub(t.Vertices[0])
	edge2 := t.Vertices[2].Sub(t.Vertices[0])
	cross := edge1.Cross(edge2)
	return cross.Length() / 2.0
}

func (t *Triangle) GetCentroid() math.Vec3 {
	return t.Vertices[0].Add(t.Vertices[1]).Add(t.Vertices[2]).DivScalar(3.0)
}

func (t *Triangle) ContainsPoint(point math.Vec3) bool {
	edge1 := t.Vertices[1].Sub(t.Vertices[0])
	edge2 := t.Vertices[2].Sub(t.Vertices[0])
	
	v0 := t.Vertices[0].Sub(point)
	v1 := edge1
	v2 := edge2
	
	dot00 := v0.Dot(v0)
	dot01 := v0.Dot(v1)
	dot02 := v0.Dot(v2)
	dot11 := v1.Dot(v1)
	dot12 := v1.Dot(v2)
	
	invDenom := 1.0 / (dot00*dot11 - dot01*dot01)
	u := (dot11*dot02 - dot01*dot12) * invDenom
	v := (dot00*dot12 - dot01*dot02) * invDenom
	
	return u >= 0 && v >= 0 && u+v <= 1
}

func (t *Triangle) GetClosestPoint(point math.Vec3) math.Vec3 {
	edge1 := t.Vertices[1].Sub(t.Vertices[0])
	edge2 := t.Vertices[2].Sub(t.Vertices[0])
	normal := edge1.Cross(edge2).Normalize()
	
	planePoint := t.Vertices[0]
	planeNormal := normal
	
	distance := point.Sub(planePoint).Dot(planeNormal)
	closestOnPlane := point.Sub(planeNormal.MulScalar(distance))
	
	if t.ContainsPoint(closestOnPlane) {
		return closestOnPlane
	}
	
	closest := t.Vertices[0]
	minDistance := point.Sub(closest).LengthSquared()
	
	for i := 1; i < 3; i++ {
		vertex := t.Vertices[i]
		distance := point.Sub(vertex).LengthSquared()
		if distance < minDistance {
			minDistance = distance
			closest = vertex
		}
	}
	
	return closest
}

func (t *Triangle) GetDistanceToPoint(point math.Vec3) float64 {
	closestPoint := t.GetClosestPoint(point)
	return point.Sub(closestPoint).Length()
} 