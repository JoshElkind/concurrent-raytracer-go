package optimization

import (
	"raytraceGo/internal/geometry"
	"raytraceGo/internal/math"
	"sync"
)

type BVH struct {
	Left, Right *BVH
	Box         geometry.AABB
	Object      geometry.Hittable
	IsLeaf      bool
}

func NewBVH(objects []geometry.Hittable, start, end int) *BVH {
	if end-start == 1 {
		return &BVH{
			Object: objects[start],
			IsLeaf: true,
			Box:    objects[start].BoundingBox(),
		}
	}
	
	box := objects[start].BoundingBox()
	for i := start + 1; i < end; i++ {
		box = surroundingBox(box, objects[i].BoundingBox())
	}
	
	_ = longestAxis(box)
	
	
	mid := (start + end) / 2
	
	bvh := &BVH{
		Box: box,
	}
	
	bvh.Left = NewBVH(objects, start, mid)
	bvh.Right = NewBVH(objects, mid, end)
	
	return bvh
}

func (bvh *BVH) Hit(ray geometry.Ray, tMin, tMax float64) (*geometry.HitRecord, bool) {
	if !bvh.Box.Hit(ray, tMin, tMax) {
		return nil, false
	}
	
	if bvh.IsLeaf {
		return bvh.Object.Hit(ray, tMin, tMax)
	}
	
	hitLeftRec, hitLeftOk := bvh.Left.Hit(ray, tMin, tMax)
	hitRightRec, hitRightOk := bvh.Right.Hit(ray, tMin, tMax)
	
	if hitLeftOk && hitRightOk {
		if hitLeftRec.T < hitRightRec.T {
			return hitLeftRec, true
		}
		return hitRightRec, true
	} else if hitLeftOk {
		return hitLeftRec, true
	} else if hitRightOk {
		return hitRightRec, true
	}
	
	return nil, false
}

func (bvh *BVH) BoundingBox() geometry.AABB {
	return bvh.Box
}

type Octree struct {
	Center   math.Vec3
	Size     float64
	Children [8]*Octree
	Objects  []geometry.Hittable
	MaxDepth int
	MaxObjects int
}

func NewOctree(center math.Vec3, size float64, maxDepth, maxObjects int) *Octree {
	return &Octree{
		Center:     center,
		Size:       size,
		MaxDepth:   maxDepth,
		MaxObjects: maxObjects,
	}
}

func (ot *Octree) Insert(object geometry.Hittable) {
	ot.insertRecursive(object, 0)
}

func (ot *Octree) insertRecursive(object geometry.Hittable, depth int) {
	if depth >= ot.MaxDepth || len(ot.Objects) < ot.MaxObjects {
		ot.Objects = append(ot.Objects, object)
		return
	}
	
	if ot.Children[0] == nil {
		ot.subdivide()
	}
	
	childIndex := ot.getChildIndex(object.BoundingBox())
	ot.Children[childIndex].insertRecursive(object, depth+1)
}

func (ot *Octree) subdivide() {
	halfSize := ot.Size / 2.0
	
	for i := 0; i < 8; i++ {
		childCenter := ot.Center.Add(math.Vec3{
			X: (float64(i&1) - 0.5) * halfSize,
			Y: (float64(i&2) - 1.0) * halfSize,
			Z: (float64(i&4) - 2.0) * halfSize,
		})
		ot.Children[i] = NewOctree(childCenter, halfSize, ot.MaxDepth, ot.MaxObjects)
	}
}

func (ot *Octree) getChildIndex(box geometry.AABB) int {
	childIndex := 0
	center := ot.Center
	
	if box.Max.X > center.X {
		childIndex |= 1
	}
	if box.Max.Y > center.Y {
		childIndex |= 2
	}
	if box.Max.Z > center.Z {
		childIndex |= 4
	}
	
	return childIndex
}

func (ot *Octree) Hit(ray geometry.Ray, tMin, tMax float64) (*geometry.HitRecord, bool) {
	octreeBox := geometry.AABB{
		Min: ot.Center.Sub(math.Vec3{X: ot.Size, Y: ot.Size, Z: ot.Size}),
		Max: ot.Center.Add(math.Vec3{X: ot.Size, Y: ot.Size, Z: ot.Size}),
	}
	
	if !octreeBox.Hit(ray, tMin, tMax) {
		return nil, false
	}
	
	var closestHit *geometry.HitRecord
	closestT := tMax
	
	for _, object := range ot.Objects {
		if hitRecord, hit := object.Hit(ray, tMin, closestT); hit {
			closestT = hitRecord.T
			closestHit = hitRecord
		}
	}
	
	for _, child := range ot.Children {
		if child != nil {
			if hitRecord, hit := child.Hit(ray, tMin, closestT); hit {
				closestT = hitRecord.T
				closestHit = hitRecord
			}
		}
	}
	
	return closestHit, closestHit != nil
}

type KDTree struct {
	Left, Right *KDTree
	Axis        int
	Split       float64
	Object      geometry.Hittable
	IsLeaf      bool
	Box         geometry.AABB
}

func NewKDTree(objects []geometry.Hittable, depth int) *KDTree {
	if len(objects) == 0 {
		return nil
	}
	
	if depth > 20 || len(objects) == 1 {
		if len(objects) == 1 {
			return &KDTree{
				Object: objects[0],
				IsLeaf: true,
				Box:    objects[0].BoundingBox(),
			}
		}
		return &KDTree{
			Object: objects[0], // Just use first object as representative
			IsLeaf: true,
			Box:    objects[0].BoundingBox(),
		}
	}
	
	box := objects[0].BoundingBox()
	for _, obj := range objects[1:] {
		box = surroundingBox(box, obj.BoundingBox())
	}
	
	axis := depth % 3
	var split float64
	switch axis {
	case 0:
		split = (box.Min.X + box.Max.X) / 2.0
	case 1:
		split = (box.Min.Y + box.Max.Y) / 2.0
	case 2:
		split = (box.Min.Z + box.Max.Z) / 2.0
	}
	
	left, right := partitionObjects(objects, axis, split)
	
	if len(left) == len(objects) || len(right) == len(objects) {
		return &KDTree{
			Object: objects[0],
			IsLeaf: true,
			Box:    box,
		}
	}
	
	kd := &KDTree{
		Axis:  axis,
		Split: split,
		Box:   box,
	}
	
	if len(left) > 0 {
		kd.Left = NewKDTree(left, depth+1)
	}
	if len(right) > 0 {
		kd.Right = NewKDTree(right, depth+1)
	}
	
	return kd
}

func (kd *KDTree) Hit(ray geometry.Ray, tMin, tMax float64) (*geometry.HitRecord, bool) {
	if !kd.Box.Hit(ray, tMin, tMax) {
		return nil, false
	}
	
	if kd.IsLeaf {
		return kd.Object.Hit(ray, tMin, tMax)
	}
	
	var hitLeft, hitRight *geometry.HitRecord
	var hitLeftOk, hitRightOk bool
	
	var rayOrigin, rayDir float64
	switch kd.Axis {
	case 0:
		rayOrigin = ray.Origin.X
		rayDir = ray.Direction.X
	case 1:
		rayOrigin = ray.Origin.Y
		rayDir = ray.Direction.Y
	case 2:
		rayOrigin = ray.Origin.Z
		rayDir = ray.Direction.Z
	}
	
	if rayDir != 0 {
		t := (kd.Split - rayOrigin) / rayDir
		if t >= 0 {
			if rayDir > 0 {
				hitLeft, hitLeftOk = kd.Left.Hit(ray, tMin, t)
				hitRight, hitRightOk = kd.Right.Hit(ray, t, tMax)
			} else {
				hitRight, hitRightOk = kd.Right.Hit(ray, tMin, t)
				hitLeft, hitLeftOk = kd.Left.Hit(ray, t, tMax)
			}
		} else {
			hitLeft, hitLeftOk = kd.Left.Hit(ray, tMin, tMax)
			hitRight, hitRightOk = kd.Right.Hit(ray, tMin, tMax)
		}
	} else {
		hitLeft, hitLeftOk = kd.Left.Hit(ray, tMin, tMax)
		hitRight, hitRightOk = kd.Right.Hit(ray, tMin, tMax)
	}
	
	if hitLeftOk && hitRightOk {
		if hitLeft.T < hitRight.T {
			return hitLeft, true
		}
		return hitRight, true
	} else if hitLeftOk {
		return hitLeft, true
	} else if hitRightOk {
		return hitRight, true
	}
	
	return nil, false
}

func (kd *KDTree) BoundingBox() geometry.AABB {
	return kd.Box
}

type SIMDVector struct {
	X, Y, Z, W [4]float64 // 4-wide SIMD
}

func NewSIMDVector(v math.Vec3) SIMDVector {
	return SIMDVector{
		X: [4]float64{v.X, v.X, v.X, v.X},
		Y: [4]float64{v.Y, v.Y, v.Y, v.Y},
		Z: [4]float64{v.Z, v.Z, v.Z, v.Z},
		W: [4]float64{0, 0, 0, 0},
	}
}

func (simd *SIMDVector) Add(other SIMDVector) SIMDVector {
	result := SIMDVector{}
	for i := 0; i < 4; i++ {
		result.X[i] = simd.X[i] + other.X[i]
		result.Y[i] = simd.Y[i] + other.Y[i]
		result.Z[i] = simd.Z[i] + other.Z[i]
	}
	return result
}

func (simd *SIMDVector) Sub(other SIMDVector) SIMDVector {
	result := SIMDVector{}
	for i := 0; i < 4; i++ {
		result.X[i] = simd.X[i] - other.X[i]
		result.Y[i] = simd.Y[i] - other.Y[i]
		result.Z[i] = simd.Z[i] - other.Z[i]
	}
	return result
}

func (simd *SIMDVector) MulScalar(scalar float64) SIMDVector {
	result := SIMDVector{}
	for i := 0; i < 4; i++ {
		result.X[i] = simd.X[i] * scalar
		result.Y[i] = simd.Y[i] * scalar
		result.Z[i] = simd.Z[i] * scalar
	}
	return result
}

type ObjectPool struct {
	rayPool      *sync.Pool
	hitPool      *sync.Pool
	vectorPool   *sync.Pool
	boundingBoxPool *sync.Pool
}

func NewObjectPool() *ObjectPool {
	return &ObjectPool{
		rayPool: &sync.Pool{
			New: func() interface{} {
				return &geometry.Ray{}
			},
		},
		hitPool: &sync.Pool{
			New: func() interface{} {
				return &geometry.HitRecord{}
			},
		},
		vectorPool: &sync.Pool{
			New: func() interface{} {
				return &math.Vec3{}
			},
		},
		boundingBoxPool: &sync.Pool{
			New: func() interface{} {
				return &geometry.AABB{}
			},
		},
	}
}

func (op *ObjectPool) GetRay() *geometry.Ray {
	return op.rayPool.Get().(*geometry.Ray)
}

func (op *ObjectPool) PutRay(ray *geometry.Ray) {
	op.rayPool.Put(ray)
}

func (op *ObjectPool) GetHitRecord() *geometry.HitRecord {
	return op.hitPool.Get().(*geometry.HitRecord)
}

func (op *ObjectPool) PutHitRecord(hit *geometry.HitRecord) {
	op.hitPool.Put(hit)
}

func (op *ObjectPool) GetVector() *math.Vec3 {
	return op.vectorPool.Get().(*math.Vec3)
}

func (op *ObjectPool) PutVector(vec *math.Vec3) {
	op.vectorPool.Put(vec)
}

func (op *ObjectPool) GetBoundingBox() *geometry.AABB {
	return op.boundingBoxPool.Get().(*geometry.AABB)
}

func (op *ObjectPool) PutBoundingBox(box *geometry.AABB) {
	op.boundingBoxPool.Put(box)
}

func surroundingBox(box0, box1 geometry.AABB) geometry.AABB {
	return geometry.AABB{
		Min: math.Vec3{
			X: math.FastMin(box0.Min.X, box1.Min.X),
			Y: math.FastMin(box0.Min.Y, box1.Min.Y),
			Z: math.FastMin(box0.Min.Z, box1.Min.Z),
		},
		Max: math.Vec3{
			X: math.FastMax(box0.Max.X, box1.Max.X),
			Y: math.FastMax(box0.Max.Y, box1.Max.Y),
			Z: math.FastMax(box0.Max.Z, box1.Max.Z),
		},
	}
}

func longestAxis(box geometry.AABB) int {
	extent := box.Max.Sub(box.Min)
	if extent.X > extent.Y && extent.X > extent.Z {
		return 0
	} else if extent.Y > extent.Z {
		return 1
	}
	return 2
}

func partitionObjects(objects []geometry.Hittable, axis int, split float64) ([]geometry.Hittable, []geometry.Hittable) {
	var left, right []geometry.Hittable
	
	for _, obj := range objects {
		box := obj.BoundingBox()
		var center float64
		switch axis {
		case 0:
			center = (box.Min.X + box.Max.X) / 2.0
		case 1:
			center = (box.Min.Y + box.Max.Y) / 2.0
		case 2:
			center = (box.Min.Z + box.Max.Z) / 2.0
		}
		
		if center < split {
			left = append(left, obj)
		} else {
			right = append(right, obj)
		}
	}
	
	return left, right
} 