package material

import (
	stdmath "math"
	"raytraceGo/internal/geometry"
	"raytraceGo/internal/math"
)

type Glass struct {
	RefractionIndex float64
	Color           math.Vec3
}

func NewGlass(refractionIndex float64, color math.Vec3) *Glass {
	return &Glass{
		RefractionIndex: refractionIndex,
		Color:           color,
	}
}

func (g *Glass) Scatter(ray geometry.Ray, hit *geometry.HitRecord) (geometry.Ray, math.Vec3, bool) {
	attenuation := g.Color
	
	var refractionRatio float64
	if hit.FrontFace {
		refractionRatio = 1.0 / g.RefractionIndex
	} else {
		refractionRatio = g.RefractionIndex
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

func (g *Glass) Emitted() math.Vec3 {
	return math.Vec3{}
}

func (g *Glass) GetAlbedo() math.Vec3 {
	return g.Color
}

func (g *Glass) GetRoughness() float64 {
	return 0.0
}

func (g *Glass) GetMetallic() float64 {
	return 0.0
}

func (g *Glass) GetSpecular() float64 {
	return 1.0
}

type Mirror struct {
	Color math.Vec3
	Roughness float64
}

func NewMirror(color math.Vec3, roughness float64) *Mirror {
	return &Mirror{
		Color:     color,
		Roughness: stdmath.Min(roughness, 1.0),
	}
}

func (m *Mirror) Scatter(ray geometry.Ray, hit *geometry.HitRecord) (geometry.Ray, math.Vec3, bool) {
	reflected := ray.Direction.Reflect(hit.Normal)
	
	if m.Roughness > 0 {
		reflected = reflected.Add(math.RandomVec3InUnitSphere().MulScalar(m.Roughness))
	}
	
	scattered := geometry.NewRay(hit.Point, reflected)
	return scattered, m.Color, scattered.Direction.Dot(hit.Normal) > 0
}

func (m *Mirror) Emitted() math.Vec3 {
	return math.Vec3{}
}

func (m *Mirror) GetAlbedo() math.Vec3 {
	return m.Color
}

func (m *Mirror) GetRoughness() float64 {
	return m.Roughness
}

func (m *Mirror) GetMetallic() float64 {
	return 1.0
}

func (m *Mirror) GetSpecular() float64 {
	return 1.0
}

type PerfectMirror struct {
	Color     math.Vec3
	Roughness float64
	IOR       float64
}

func NewPerfectMirror(color math.Vec3, roughness float64) *PerfectMirror {
	return &PerfectMirror{
		Color:     color,
		Roughness: stdmath.Min(roughness, 1.0),
		IOR:       2.0,
	}
}

func (pm *PerfectMirror) Scatter(ray geometry.Ray, hit *geometry.HitRecord) (geometry.Ray, math.Vec3, bool) {
	reflected := ray.Direction.Reflect(hit.Normal)
	
	if pm.Roughness > 0.001 {
		perturbation := math.RandomVec3InUnitSphere().MulScalar(pm.Roughness)
		reflected = reflected.Add(perturbation).Normalize()
	}
	
	cosTheta := stdmath.Abs(ray.Direction.Dot(hit.Normal))
	fresnel := pm.calculateFresnel(cosTheta)
	
	enhancedColor := math.Vec3{
		X: pm.Color.X * (1.0 - 0.9) + fresnel.X * 0.9,
		Y: pm.Color.Y * (1.0 - 0.9) + fresnel.Y * 0.9,
		Z: pm.Color.Z * (1.0 - 0.9) + fresnel.Z * 0.9,
	}
	
	scattered := geometry.NewRay(hit.Point, reflected)
	return scattered, enhancedColor, true
}

func (pm *PerfectMirror) calculateFresnel(cosTheta float64) math.Vec3 {
	f0 := stdmath.Pow((pm.IOR-1.0)/(pm.IOR+1.0), 2.0)
	schlick := f0 + (1.0-f0)*stdmath.Pow(1.0-cosTheta, 5)
	
	return math.Vec3{X: schlick, Y: schlick, Z: schlick}
}

func (pm *PerfectMirror) Emitted() math.Vec3 {
	return math.Vec3{}
}

func (pm *PerfectMirror) GetAlbedo() math.Vec3 {
	return pm.Color
}

func (pm *PerfectMirror) GetRoughness() float64 {
	return pm.Roughness
}

func (pm *PerfectMirror) GetMetallic() float64 {
	return 1.0
}

func (pm *PerfectMirror) GetSpecular() float64 {
	return 1.0
}

type ProceduralTexture struct {
	BaseMaterial Material
	Scale        float64
	Octaves      int
	Persistence  float64
	Lacunarity   float64
}

func NewProceduralTexture(base Material, scale, persistence, lacunarity float64, octaves int) *ProceduralTexture {
	return &ProceduralTexture{
		BaseMaterial: base,
		Scale:        scale,
		Octaves:      octaves,
		Persistence:  persistence,
		Lacunarity:   lacunarity,
	}
}

func (pt *ProceduralTexture) Scatter(ray geometry.Ray, hit *geometry.HitRecord) (geometry.Ray, math.Vec3, bool) {
	return pt.BaseMaterial.Scatter(ray, hit)
}

func (pt *ProceduralTexture) Emitted() math.Vec3 {
	return pt.BaseMaterial.Emitted()
}

func (pt *ProceduralTexture) calculateNoise(point math.Vec3) math.Vec3 {
	noise := pt.simplexNoise(point)
	return math.Vec3{
		X: noise,
		Y: noise,
		Z: noise,
	}
}

func (pt *ProceduralTexture) simplexNoise(point math.Vec3) float64 {
	return math.RandomFloat()
}

type SubsurfaceScattering struct {
	BaseColor     math.Vec3
	ScatteringRadius float64
	Absorption    math.Vec3
	PhaseFunction float64
}

func NewSubsurfaceScattering(baseColor math.Vec3, scatteringRadius, phaseFunction float64, absorption math.Vec3) *SubsurfaceScattering {
	return &SubsurfaceScattering{
		BaseColor:     baseColor,
		ScatteringRadius: scatteringRadius,
		Absorption:    absorption,
		PhaseFunction: phaseFunction,
	}
}

func (sss *SubsurfaceScattering) Scatter(ray geometry.Ray, hit *geometry.HitRecord) (geometry.Ray, math.Vec3, bool) {
	scatterDirection := math.RandomVec3InUnitSphere()
	
	scatterDirection = scatterDirection.MulScalar(sss.PhaseFunction)
	
	scatterColor := sss.BaseColor
	
	absorption := sss.Absorption.MulScalar(sss.ScatteringRadius)
	scatterColor = scatterColor.Mul(absorption)
	
	scattered := geometry.NewRay(hit.Point, scatterDirection)
	return scattered, scatterColor, true
}

func (sss *SubsurfaceScattering) Emitted() math.Vec3 {
	return math.Vec3{}
}

type Anisotropic struct {
	BaseColor    math.Vec3
	Roughness    float64
	Anisotropy   float64
	Direction    math.Vec3
}

func NewAnisotropic(baseColor math.Vec3, roughness, anisotropy float64, direction math.Vec3) *Anisotropic {
	return &Anisotropic{
		BaseColor:    baseColor,
		Roughness:    roughness,
		Anisotropy:   anisotropy,
		Direction:    direction.Normalize(),
	}
}

func (a *Anisotropic) Scatter(ray geometry.Ray, hit *geometry.HitRecord) (geometry.Ray, math.Vec3, bool) {
	reflected := ray.Direction.Reflect(hit.Normal)
	
	anisotropicRoughness := a.Roughness * (1.0 + a.Anisotropy*a.Direction.Dot(hit.Normal))
	
	if anisotropicRoughness > 0 {
		reflected = reflected.Add(math.RandomVec3InUnitSphere().MulScalar(anisotropicRoughness))
		reflected = reflected.Normalize()
	}
	
	scattered := geometry.NewRay(hit.Point, reflected)
	return scattered, a.BaseColor, true
}

func (a *Anisotropic) Emitted() math.Vec3 {
	return math.Vec3{}
}

type Clearcoat struct {
	BaseMaterial Material
	ClearcoatStrength float64
	ClearcoatRoughness float64
	IOR               float64
}

func NewClearcoat(baseMaterial Material, strength, roughness, ior float64) *Clearcoat {
	return &Clearcoat{
		BaseMaterial:      baseMaterial,
		ClearcoatStrength: strength,
		ClearcoatRoughness: roughness,
		IOR:               ior,
	}
}

func (cc *Clearcoat) Scatter(ray geometry.Ray, hit *geometry.HitRecord) (geometry.Ray, math.Vec3, bool) {
	baseScattered, baseAttenuation, baseHit := cc.BaseMaterial.Scatter(ray, hit)
	
	_, clearcoatAttenuation, clearcoatHit := cc.scatterClearcoat(ray, hit)
	
	if baseHit && clearcoatHit {
		blend := cc.ClearcoatStrength
		finalAttenuation := math.Vec3{
			X: baseAttenuation.X*(1.0-blend) + clearcoatAttenuation.X*blend,
			Y: baseAttenuation.Y*(1.0-blend) + clearcoatAttenuation.Y*blend,
			Z: baseAttenuation.Z*(1.0-blend) + clearcoatAttenuation.Z*blend,
		}
		return baseScattered, finalAttenuation, true
	}
	
	return baseScattered, baseAttenuation, baseHit
}

func (cc *Clearcoat) scatterClearcoat(ray geometry.Ray, hit *geometry.HitRecord) (geometry.Ray, math.Vec3, bool) {
	reflected := ray.Direction.Reflect(hit.Normal)
	
	if cc.ClearcoatRoughness > 0 {
		reflected = reflected.Add(math.RandomVec3InUnitSphere().MulScalar(cc.ClearcoatRoughness))
		reflected = reflected.Normalize()
	}
	
	cosTheta := stdmath.Abs(ray.Direction.Dot(hit.Normal))
	f0 := stdmath.Pow((cc.IOR-1.0)/(cc.IOR+1.0), 2.0)
	schlick := f0 + (1.0-f0)*stdmath.Pow(1.0-cosTheta, 5)
	
	clearcoatAttenuation := math.Vec3{X: schlick, Y: schlick, Z: schlick}
	
	scattered := geometry.NewRay(hit.Point, reflected)
	return scattered, clearcoatAttenuation, true
}

type Sheen struct {
	BaseColor    math.Vec3
	SheenColor   math.Vec3
	SheenRoughness float64
	SheenTint    float64
}

func NewSheen(baseColor, sheenColor math.Vec3, sheenRoughness, sheenTint float64) *Sheen {
	return &Sheen{
		BaseColor:    baseColor,
		SheenColor:   sheenColor,
		SheenRoughness: sheenRoughness,
		SheenTint:    sheenTint,
	}
}

func (s *Sheen) Scatter(ray geometry.Ray, hit *geometry.HitRecord) (geometry.Ray, math.Vec3, bool) {
	sheenColor := math.Vec3{
		X: s.SheenColor.X * (1.0 - s.SheenTint) + s.BaseColor.X * s.SheenTint,
		Y: s.SheenColor.Y * (1.0 - s.SheenTint) + s.BaseColor.Y * s.SheenTint,
		Z: s.SheenColor.Z * (1.0 - s.SheenTint) + s.BaseColor.Z * s.SheenTint,
	}
	
	reflected := ray.Direction.Reflect(hit.Normal)
	
	if s.SheenRoughness > 0 {
		reflected = reflected.Add(math.RandomVec3InUnitSphere().MulScalar(s.SheenRoughness))
		reflected = reflected.Normalize()
	}
	
	scattered := geometry.NewRay(hit.Point, reflected)
	return scattered, sheenColor, true
}

func (s *Sheen) Emitted() math.Vec3 {
	return math.Vec3{}
}

type Emission struct {
	Color        math.Vec3
	Intensity    float64
	EmissionType EmissionType
	Falloff      float64
}

type EmissionType int

const (
	EmissionPoint EmissionType = iota
	EmissionDirectional
	EmissionArea
)

func NewEmission(color math.Vec3, intensity float64, emissionType EmissionType, falloff float64) *Emission {
	return &Emission{
		Color:        color,
		Intensity:    intensity,
		EmissionType: emissionType,
		Falloff:      falloff,
	}
}

func (e *Emission) Emit(hit *geometry.HitRecord) math.Vec3 {
	switch e.EmissionType {
	case EmissionPoint:
		return e.Color.MulScalar(e.Intensity)
	case EmissionDirectional:
		emissionDirection := math.Vec3{Y: 1.0}
		cosTheta := hit.Normal.Dot(emissionDirection)
		if cosTheta > 0 {
			return e.Color.MulScalar(e.Intensity * cosTheta)
		}
		return math.Vec3{}
	case EmissionArea:
		return e.Color.MulScalar(e.Intensity)
	default:
		return e.Color.MulScalar(e.Intensity)
	}
}

func (e *Emission) Emitted() math.Vec3 {
	return e.Color.MulScalar(e.Intensity)
}

type NoiseTexture struct {
	Scale       float64
	Octaves     int
	Persistence float64
	Lacunarity  float64
	Amplitude   float64
}

func NewNoiseTexture(scale float64, octaves int, persistence, lacunarity, amplitude float64) *NoiseTexture {
	return &NoiseTexture{
		Scale:       scale,
		Octaves:     octaves,
		Persistence: persistence,
		Lacunarity:  lacunarity,
		Amplitude:   amplitude,
	}
}

func (nt *NoiseTexture) Value(point math.Vec3) float64 {
	noise := nt.simplexNoise(point.MulScalar(nt.Scale))
	return noise * nt.Amplitude
}

func (nt *NoiseTexture) simplexNoise(point math.Vec3) float64 {
	return math.RandomFloat()
}

type MarbleTexture struct {
	BaseColor   math.Vec3
	VeinColor   math.Vec3
	Scale       float64
	Turbulence  float64
	Sharpness   float64
}

func NewMarbleTexture(baseColor, veinColor math.Vec3, scale, turbulence, sharpness float64) *MarbleTexture {
	return &MarbleTexture{
		BaseColor:   baseColor,
		VeinColor:   veinColor,
		Scale:       scale,
		Turbulence:  turbulence,
		Sharpness:   sharpness,
	}
}

func (mt *MarbleTexture) Value(point math.Vec3) math.Vec3 {
	marbleValue := stdmath.Sin(point.X*mt.Scale + point.Y*mt.Scale*0.5 + point.Z*mt.Scale*0.25)
	marbleValue = (marbleValue + 1.0) / 2.0
	
	marbleValue = stdmath.Pow(marbleValue, mt.Sharpness)
	
	finalColor := math.Vec3{
		X: mt.BaseColor.X*(1.0-marbleValue) + mt.VeinColor.X*marbleValue,
		Y: mt.BaseColor.Y*(1.0-marbleValue) + mt.VeinColor.Y*marbleValue,
		Z: mt.BaseColor.Z*(1.0-marbleValue) + mt.VeinColor.Z*marbleValue,
	}
	
	return finalColor
}

type WoodTexture struct {
	BaseColor   math.Vec3
	RingColor   math.Vec3
	Scale       float64
	Turbulence  float64
	RingWidth   float64
}

func NewWoodTexture(baseColor, ringColor math.Vec3, scale, turbulence, ringWidth float64) *WoodTexture {
	return &WoodTexture{
		BaseColor:   baseColor,
		RingColor:   ringColor,
		Scale:       scale,
		Turbulence:  turbulence,
		RingWidth:   ringWidth,
	}
}

func (wt *WoodTexture) Value(point math.Vec3) math.Vec3 {
	ringValue := stdmath.Sin(point.X*wt.Scale + point.Y*wt.Scale*0.5)
	ringValue = stdmath.Abs(ringValue)
	
	if ringValue < wt.RingWidth {
		return wt.RingColor
	}
	
	return wt.BaseColor
}

type CheckerboardTexture struct {
	Color1      math.Vec3
	Color2      math.Vec3
	Scale       float64
}

func NewCheckerboardTexture(color1, color2 math.Vec3, scale float64) *CheckerboardTexture {
	return &CheckerboardTexture{
		Color1:      color1,
		Color2:      color2,
		Scale:       scale,
	}
}

func (ct *CheckerboardTexture) Value(point math.Vec3) math.Vec3 {
	checker := stdmath.Floor(point.X*ct.Scale) + stdmath.Floor(point.Y*ct.Scale) + stdmath.Floor(point.Z*ct.Scale)
	if int(checker)%2 == 0 {
		return ct.Color1
	}
	return ct.Color2
}

type GradientTexture struct {
	Color1      math.Vec3
	Color2      math.Vec3
	Direction   math.Vec3
}

func NewGradientTexture(color1, color2 math.Vec3, direction math.Vec3) *GradientTexture {
	return &GradientTexture{
		Color1:      color1,
		Color2:      color2,
		Direction:   direction.Normalize(),
	}
}

func (gt *GradientTexture) Value(point math.Vec3) math.Vec3 {
	t := point.Dot(gt.Direction)
	t = (t + 1.0) / 2.0
	
	return math.Vec3{
		X: gt.Color1.X*(1.0-t) + gt.Color2.X*t,
		Y: gt.Color1.Y*(1.0-t) + gt.Color2.Y*t,
		Z: gt.Color1.Z*(1.0-t) + gt.Color2.Z*t,
	}
}

type PerlinNoiseTexture struct {
	Scale       float64
	Octaves     int
	Persistence float64
	Lacunarity  float64
}

func NewPerlinNoiseTexture(scale float64, octaves int, persistence, lacunarity float64) *PerlinNoiseTexture {
	return &PerlinNoiseTexture{
		Scale:       scale,
		Octaves:     octaves,
		Persistence: persistence,
		Lacunarity:  lacunarity,
	}
}

func (pnt *PerlinNoiseTexture) Value(point math.Vec3) float64 {
	noise := pnt.simplexNoise(point.MulScalar(pnt.Scale))
	return noise
}

func (pnt *PerlinNoiseTexture) simplexNoise(point math.Vec3) float64 {
	return math.RandomFloat()
}

type VoronoiTexture struct {
	Scale       float64
	Points      int
	DistanceType VoronoiDistanceType
}

type VoronoiDistanceType int

const (
	VoronoiEuclidean VoronoiDistanceType = iota
	VoronoiManhattan
	VoronoiChebyshev
)

func NewVoronoiTexture(scale float64, points int, distanceType VoronoiDistanceType) *VoronoiTexture {
	return &VoronoiTexture{
		Scale:       scale,
		Points:      points,
		DistanceType: distanceType,
	}
}

func (vt *VoronoiTexture) Value(point math.Vec3) float64 {
	minDistance := stdmath.Inf(1)
	
	for i := 0; i < vt.Points; i++ {
		randomPoint := math.Vec3{
			X: math.RandomFloat() * 2.0 - 1.0,
			Y: math.RandomFloat() * 2.0 - 1.0,
			Z: math.RandomFloat() * 2.0 - 1.0,
		}
		
		distance := vt.calculateDistance(point, randomPoint)
		if distance < minDistance {
			minDistance = distance
		}
	}
	
	return minDistance
}

func (vt *VoronoiTexture) calculateDistance(p1, p2 math.Vec3) float64 {
	switch vt.DistanceType {
	case VoronoiEuclidean:
		return p1.Sub(p2).Length()
	case VoronoiManhattan:
		diff := p1.Sub(p2)
		return stdmath.Abs(diff.X) + stdmath.Abs(diff.Y) + stdmath.Abs(diff.Z)
	case VoronoiChebyshev:
		diff := p1.Sub(p2)
		return stdmath.Max(stdmath.Max(stdmath.Abs(diff.X), stdmath.Abs(diff.Y)), stdmath.Abs(diff.Z))
	default:
		return p1.Sub(p2).Length()
	}
}

var noiseTexture *NoiseTexture

func init() {
	noiseTexture = NewNoiseTexture(1.0, 4, 0.5, 2.0, 1.0)
} 