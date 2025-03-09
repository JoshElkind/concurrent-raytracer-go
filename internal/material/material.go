package material

import (
	"raytraceGo/internal/geometry"
	"raytraceGo/internal/math"
	stdmath "math"
)

type Material interface {
	Scatter(ray geometry.Ray, hit *geometry.HitRecord) (geometry.Ray, math.Vec3, bool)
	Emitted() math.Vec3
	GetAlbedo() math.Vec3
	GetRoughness() float64
	GetMetallic() float64
	GetSpecular() float64
}

type Lambertian struct {
	Albedo math.Vec3
}

func NewLambertian(albedo math.Vec3) *Lambertian {
	return &Lambertian{Albedo: albedo}
}

func (l *Lambertian) Scatter(ray geometry.Ray, hit *geometry.HitRecord) (geometry.Ray, math.Vec3, bool) {
	scatterDirection := hit.Normal.Add(math.RandomVec3InUnitSphere())
	if scatterDirection.NearZero() {
		scatterDirection = hit.Normal
	}
	scatterDirection = scatterDirection.Normalize()
	
	scattered := geometry.NewRay(hit.Point, scatterDirection)
	return scattered, l.Albedo, true
}

func (l *Lambertian) Emitted() math.Vec3 {
	return math.Vec3{}
}

func (l *Lambertian) GetAlbedo() math.Vec3 {
	return l.Albedo
}

func (l *Lambertian) GetRoughness() float64 {
	return 1.0
}

func (l *Lambertian) GetMetallic() float64 {
	return 0.0
}

func (l *Lambertian) GetSpecular() float64 {
	return 0.0
}

type Metal struct {
	Albedo    math.Vec3
	Roughness float64
	Metallic  float64
	Specular  float64
	IOR       float64
}

func NewMetal(albedo math.Vec3, roughness, metallic, specular float64) *Metal {
	return &Metal{
		Albedo:    albedo,
		Roughness: stdmath.Min(roughness, 1.0),
		Metallic:  stdmath.Min(metallic, 1.0),
		Specular:  stdmath.Min(specular, 1.0),
		IOR:       1.5,
	}
}

func (m *Metal) Scatter(ray geometry.Ray, hit *geometry.HitRecord) (geometry.Ray, math.Vec3, bool) {
	reflected := ray.Direction.Reflect(hit.Normal)
	
	if m.Roughness > 0.001 {
		perturbation := math.RandomVec3InUnitSphere().MulScalar(m.Roughness)
		reflected = reflected.Add(perturbation).Normalize()
	}
	
	albedo := m.Albedo
	
	cosTheta := stdmath.Abs(ray.Direction.Dot(hit.Normal))
	fresnel := m.calculateFresnel(cosTheta)
	
	fresnelStrength := 0.6 + m.Metallic * 0.4
	
	enhancedAlbedo := math.Vec3{
		X: albedo.X * (1.0 - fresnelStrength) + fresnel.X * fresnelStrength,
		Y: albedo.Y * (1.0 - fresnelStrength) + fresnel.Y * fresnelStrength,
		Z: albedo.Z * (1.0 - fresnelStrength) + fresnel.Z * fresnelStrength,
	}
	
	enhancedAlbedo = math.Vec3{
		X: stdmath.Max(0.0, stdmath.Min(1.0, enhancedAlbedo.X)),
		Y: stdmath.Max(0.0, stdmath.Min(1.0, enhancedAlbedo.Y)),
		Z: stdmath.Max(0.0, stdmath.Min(1.0, enhancedAlbedo.Z)),
	}
	
	if m.Metallic > 0.8 {
		metallicFresnel := 0.4 + m.Metallic * 0.5
		enhancedAlbedo = math.Vec3{
			X: enhancedAlbedo.X * (1.0 - metallicFresnel) + fresnel.X * metallicFresnel,
			Y: enhancedAlbedo.Y * (1.0 - metallicFresnel) + fresnel.Y * metallicFresnel,
			Z: enhancedAlbedo.Z * (1.0 - metallicFresnel) + fresnel.Z * metallicFresnel,
		}
	}
	
	scattered := geometry.NewRay(hit.Point, reflected)
	return scattered, enhancedAlbedo, true
}

func (m *Metal) calculateFresnel(cosTheta float64) math.Vec3 {
	f0 := math.Vec3{
		X: stdmath.Pow((m.IOR-1.0)/(m.IOR+1.0), 2.0),
		Y: stdmath.Pow((m.IOR-1.0)/(m.IOR+1.0), 2.0),
		Z: stdmath.Pow((m.IOR-1.0)/(m.IOR+1.0), 2.0),
	}
	
	schlick := math.Vec3{
		X: f0.X + (1.0-f0.X)*stdmath.Pow(1.0-cosTheta, 5),
		Y: f0.Y + (1.0-f0.Y)*stdmath.Pow(1.0-cosTheta, 5),
		Z: f0.Z + (1.0-f0.Z)*stdmath.Pow(1.0-cosTheta, 5),
	}
	
	return schlick
}

func (m *Metal) Emitted() math.Vec3 {
	return math.Vec3{}
}

func (m *Metal) GetAlbedo() math.Vec3 {
	return m.Albedo
}

func (m *Metal) GetRoughness() float64 {
	return m.Roughness
}

func (m *Metal) GetMetallic() float64 {
	return m.Metallic
}

func (m *Metal) GetSpecular() float64 {
	return m.Specular
}

type ShinyMaterial struct {
	Albedo    math.Vec3
	Roughness float64
	Metallic  float64
	Specular  float64
	IOR       float64
}

func NewShinyMaterial(albedo math.Vec3, roughness, metallic, specular float64) *ShinyMaterial {
	return &ShinyMaterial{
		Albedo:    albedo,
		Roughness: stdmath.Min(roughness, 1.0),
		Metallic:  stdmath.Min(metallic, 1.0),
		Specular:  stdmath.Min(specular, 1.0),
		IOR:       1.5,
	}
}

func (s *ShinyMaterial) Scatter(ray geometry.Ray, hit *geometry.HitRecord) (geometry.Ray, math.Vec3, bool) {
	reflected := ray.Direction.Reflect(hit.Normal)
	
	if s.Roughness > 0 {
		reflected = reflected.Add(math.RandomVec3InUnitSphere().MulScalar(s.Roughness))
		reflected = reflected.Normalize()
	}
	
	cosTheta := stdmath.Abs(ray.Direction.Dot(hit.Normal))
	fresnel := s.calculateFresnel(cosTheta)
	
	fresnelStrength := 0.4 + s.Specular * 0.4
	enhancedAlbedo := math.Vec3{
		X: stdmath.Min(1.0, s.Albedo.X * (1.0 - fresnelStrength) + fresnel.X * fresnelStrength),
		Y: stdmath.Min(1.0, s.Albedo.Y * (1.0 - fresnelStrength) + fresnel.Y * fresnelStrength),
		Z: stdmath.Min(1.0, s.Albedo.Z * (1.0 - fresnelStrength) + fresnel.Z * fresnelStrength),
	}
	
	scattered := geometry.NewRay(hit.Point, reflected)
	return scattered, enhancedAlbedo, true
}

func (s *ShinyMaterial) calculateFresnel(cosTheta float64) math.Vec3 {
	f0 := math.Vec3{
		X: stdmath.Pow((s.IOR-1.0)/(s.IOR+1.0), 2.0),
		Y: stdmath.Pow((s.IOR-1.0)/(s.IOR+1.0), 2.0),
		Z: stdmath.Pow((s.IOR-1.0)/(s.IOR+1.0), 2.0),
	}
	
	schlick := math.Vec3{
		X: f0.X + (1.0-f0.X)*stdmath.Pow(1.0-cosTheta, 5),
		Y: f0.Y + (1.0-f0.Y)*stdmath.Pow(1.0-cosTheta, 5),
		Z: f0.Z + (1.0-f0.Z)*stdmath.Pow(1.0-cosTheta, 5),
	}
	
	return schlick
}

func (s *ShinyMaterial) Emitted() math.Vec3 {
	return math.Vec3{}
}

func (s *ShinyMaterial) GetAlbedo() math.Vec3 {
	return s.Albedo
}

func (s *ShinyMaterial) GetRoughness() float64 {
	return s.Roughness
}

func (s *ShinyMaterial) GetMetallic() float64 {
	return s.Metallic
}

func (s *ShinyMaterial) GetSpecular() float64 {
	return s.Specular
}

type Dielectric struct {
	RefractionIndex float64
}

func NewDielectric(refractionIndex float64) *Dielectric {
	return &Dielectric{RefractionIndex: refractionIndex}
}

func (d *Dielectric) Scatter(ray geometry.Ray, hit *geometry.HitRecord) (geometry.Ray, math.Vec3, bool) {
	attenuation := math.Vec3{X: 1.0, Y: 1.0, Z: 1.0}
	
	var refractionRatio float64
	if hit.FrontFace {
		refractionRatio = 1.0 / d.RefractionIndex
	} else {
		refractionRatio = d.RefractionIndex
	}
	
	unitDirection := ray.Direction.Normalize()
	cosTheta := stdmath.Min(unitDirection.MulScalar(-1).Dot(hit.Normal), 1.0)
	sinTheta := stdmath.Sqrt(1.0 - cosTheta*cosTheta)
	
	cannotRefract := refractionRatio*sinTheta > 1.0
	
	var direction math.Vec3
	if cannotRefract || reflectance(cosTheta, refractionRatio) > math.RandomFloat() {
		direction = unitDirection.Reflect(hit.Normal)
	} else {
		direction = unitDirection.Refract(hit.Normal, refractionRatio)
	}
	
	scattered := geometry.NewRay(hit.Point, direction)
	return scattered, attenuation, true
}

func (d *Dielectric) Emitted() math.Vec3 {
	return math.Vec3{}
}

func (d *Dielectric) GetAlbedo() math.Vec3 {
	return math.Vec3{X: 1.0, Y: 1.0, Z: 1.0}
}

func (d *Dielectric) GetRoughness() float64 {
	return 0.0
}

func (d *Dielectric) GetMetallic() float64 {
	return 0.0
}

func (d *Dielectric) GetSpecular() float64 {
	return 1.0
}

func reflectance(cosine, refIdx float64) float64 {
	r0 := (1 - refIdx) / (1 + refIdx)
	r0 = r0 * r0
	return r0 + (1-r0)*stdmath.Pow(1-cosine, 5)
}

type DiffuseLight struct {
	Emit math.Vec3
}

func NewDiffuseLight(emit math.Vec3) *DiffuseLight {
	return &DiffuseLight{Emit: emit}
}

func (dl *DiffuseLight) Scatter(ray geometry.Ray, hit *geometry.HitRecord) (geometry.Ray, math.Vec3, bool) {
	return geometry.Ray{}, math.Vec3{}, false
}

func (dl *DiffuseLight) Emitted() math.Vec3 {
	return dl.Emit
}

func (dl *DiffuseLight) GetAlbedo() math.Vec3 {
	return math.Vec3{}
}

func (dl *DiffuseLight) GetRoughness() float64 {
	return 1.0
}

func (dl *DiffuseLight) GetMetallic() float64 {
	return 0.0
}

func (dl *DiffuseLight) GetSpecular() float64 {
	return 0.0
} 

 