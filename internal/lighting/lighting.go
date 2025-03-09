package lighting

import (
	stdmath "math"
	"raytraceGo/internal/math"
)

type Light interface {
	GetIntensity(point math.Vec3) float64
	GetDirection(point math.Vec3) math.Vec3
	GetColor() math.Vec3
	GetPosition() math.Vec3
	GetType() string
	GetShadowRay(point math.Vec3) (math.Vec3, float64)
}

type PointLight struct {
	Position  math.Vec3
	Color     math.Vec3
	Intensity float64
	Attenuation Attenuation
}

type Attenuation struct {
	Constant  float64
	Linear    float64
	Quadratic float64
}

func NewPointLight(position math.Vec3, color math.Vec3, intensity float64) *PointLight {
	return &PointLight{
		Position:  position,
		Color:     color,
		Intensity: intensity,
		Attenuation: Attenuation{
			Constant:  1.0,
			Linear:    0.09,
			Quadratic: 0.032,
		},
	}
}

func (pl *PointLight) GetIntensity(point math.Vec3) float64 {
	distance := point.Sub(pl.Position).Length()
	attenuation := pl.Attenuation.Constant + 
		pl.Attenuation.Linear*distance + 
		pl.Attenuation.Quadratic*distance*distance
	return pl.Intensity / attenuation
}

func (pl *PointLight) GetDirection(point math.Vec3) math.Vec3 {
	return pl.Position.Sub(point).Normalize()
}

func (pl *PointLight) GetColor() math.Vec3 {
	return pl.Color
}

func (pl *PointLight) GetPosition() math.Vec3 {
	return pl.Position
}

func (pl *PointLight) GetType() string {
	return "point"
}

func (pl *PointLight) GetShadowRay(point math.Vec3) (math.Vec3, float64) {
	direction := pl.GetDirection(point)
	distance := point.Sub(pl.Position).Length()
	return direction, distance
}

type DirectionalLight struct {
	Direction math.Vec3
	Color     math.Vec3
	Intensity float64
}

func NewDirectionalLight(direction math.Vec3, color math.Vec3, intensity float64) *DirectionalLight {
	return &DirectionalLight{
		Direction: direction.Normalize(),
		Color:     color,
		Intensity: intensity,
	}
}

func (dl *DirectionalLight) GetIntensity(point math.Vec3) float64 {
	return dl.Intensity
}

func (dl *DirectionalLight) GetDirection(point math.Vec3) math.Vec3 {
	return dl.Direction.MulScalar(-1)
}

func (dl *DirectionalLight) GetColor() math.Vec3 {
	return dl.Color
}

func (dl *DirectionalLight) GetPosition() math.Vec3 {
	return math.Vec3{} // Directional lights don't have a position
}

func (dl *DirectionalLight) GetType() string {
	return "directional"
}

func (dl *DirectionalLight) GetShadowRay(point math.Vec3) (math.Vec3, float64) {
	return dl.GetDirection(point), stdmath.Inf(1)
}

type AreaLight struct {
	Position  math.Vec3
	Color     math.Vec3
	Intensity float64
	Size      float64
	Samples   int
}

func NewAreaLight(position math.Vec3, color math.Vec3, intensity, size float64) *AreaLight {
	return &AreaLight{
		Position:  position,
		Color:     color,
		Intensity: intensity,
		Size:      size,
		Samples:   16,
	}
}

func (al *AreaLight) GetIntensity(point math.Vec3) float64 {
	distance := point.Sub(al.Position).Length()
	attenuation := 1.0 + 0.09*distance + 0.032*distance*distance
	return al.Intensity / attenuation
}

func (al *AreaLight) GetDirection(point math.Vec3) math.Vec3 {
	return al.Position.Sub(point).Normalize()
}

func (al *AreaLight) GetColor() math.Vec3 {
	return al.Color
}

func (al *AreaLight) GetPosition() math.Vec3 {
	return al.Position
}

func (al *AreaLight) GetType() string {
	return "area"
}

func (al *AreaLight) GetShadowRay(point math.Vec3) (math.Vec3, float64) {
	direction := al.GetDirection(point)
	distance := point.Sub(al.Position).Length()
	return direction, distance
}

type SpotLight struct {
	Position  math.Vec3
	Direction math.Vec3
	Color     math.Vec3
	Intensity float64
	CutOff    float64
	OuterCutOff float64
}

func NewSpotLight(position, direction math.Vec3, color math.Vec3, intensity, cutOff, outerCutOff float64) *SpotLight {
	return &SpotLight{
		Position:     position,
		Direction:    direction.Normalize(),
		Color:        color,
		Intensity:    intensity,
		CutOff:       cutOff,
		OuterCutOff:  outerCutOff,
	}
}

func (sl *SpotLight) GetIntensity(point math.Vec3) float64 {
	direction := sl.GetDirection(point)
	cosTheta := direction.Dot(sl.Direction)
	
	if cosTheta < sl.OuterCutOff {
		return 0.0
	}
	
	if cosTheta > sl.CutOff {
		return sl.Intensity
	}
	
	epsilon := sl.CutOff - sl.OuterCutOff
	intensity := (cosTheta - sl.OuterCutOff) / epsilon
	return sl.Intensity * intensity
}

func (sl *SpotLight) GetDirection(point math.Vec3) math.Vec3 {
	return sl.Position.Sub(point).Normalize()
}

func (sl *SpotLight) GetColor() math.Vec3 {
	return sl.Color
}

func (sl *SpotLight) GetPosition() math.Vec3 {
	return sl.Position
}

func (sl *SpotLight) GetType() string {
	return "spot"
}

func (sl *SpotLight) GetShadowRay(point math.Vec3) (math.Vec3, float64) {
	direction := sl.GetDirection(point)
	distance := point.Sub(sl.Position).Length()
	return direction, distance
}

type LightingModel interface {
	CalculateLighting(hit *HitRecord, light Light, viewDirection math.Vec3) math.Vec3
}

type PhongModel struct {
	Ambient   float64
	Diffuse   float64
	Specular  float64
	Shininess float64
}

func NewPhongModel(ambient, diffuse, specular, shininess float64) *PhongModel {
	return &PhongModel{
		Ambient:   ambient,
		Diffuse:   diffuse,
		Specular:  specular,
		Shininess: shininess,
	}
}

func (pm *PhongModel) CalculateLighting(hit *HitRecord, light Light, viewDirection math.Vec3) math.Vec3 {
	lightDirection := light.GetDirection(hit.Point)
	lightIntensity := light.GetIntensity(hit.Point)
	lightColor := light.GetColor()
	
	ambient := lightColor.MulScalar(pm.Ambient * lightIntensity)
	
	diffuseFactor := stdmath.Max(0, hit.Normal.Dot(lightDirection))
	diffuse := lightColor.MulScalar(pm.Diffuse * diffuseFactor * lightIntensity)
	
	reflectDirection := lightDirection.Reflect(hit.Normal)
	specularFactor := stdmath.Pow(stdmath.Max(0, viewDirection.Dot(reflectDirection)), pm.Shininess)
	specular := lightColor.MulScalar(pm.Specular * specularFactor * lightIntensity)
	
	return ambient.Add(diffuse).Add(specular)
}

type BlinnPhongModel struct {
	Ambient   float64
	Diffuse   float64
	Specular  float64
	Shininess float64
}

func NewBlinnPhongModel(ambient, diffuse, specular, shininess float64) *BlinnPhongModel {
	return &BlinnPhongModel{
		Ambient:   ambient,
		Diffuse:   diffuse,
		Specular:  specular,
		Shininess: shininess,
	}
}

func (bpm *BlinnPhongModel) CalculateLighting(hit *HitRecord, light Light, viewDirection math.Vec3) math.Vec3 {
	lightDirection := light.GetDirection(hit.Point)
	lightIntensity := light.GetIntensity(hit.Point)
	lightColor := light.GetColor()
	
	ambient := lightColor.MulScalar(bpm.Ambient * lightIntensity)
	
	diffuseFactor := stdmath.Max(0, hit.Normal.Dot(lightDirection))
	diffuse := lightColor.MulScalar(bpm.Diffuse * diffuseFactor * lightIntensity)
	
	halfVector := lightDirection.Add(viewDirection).Normalize()
	specularFactor := stdmath.Pow(stdmath.Max(0, hit.Normal.Dot(halfVector)), bpm.Shininess)
	specular := lightColor.MulScalar(bpm.Specular * specularFactor * lightIntensity)
	
	return ambient.Add(diffuse).Add(specular)
}

type GlobalIllumination struct {
	AmbientOcclusion bool
	IndirectLighting  bool
	Caustics         bool
	VolumetricLighting bool
}

func NewGlobalIllumination() *GlobalIllumination {
	return &GlobalIllumination{
		AmbientOcclusion: true,
		IndirectLighting: true,
		Caustics:         false,
		VolumetricLighting: false,
	}
}

func (gi *GlobalIllumination) CalculateAmbientOcclusion(hit *HitRecord, scene Scene) float64 {
	if !gi.AmbientOcclusion {
		return 1.0
	}
	
	samples := 16
	occlusion := 0.0
	
	for i := 0; i < samples; i++ {
		randomDirection := math.RandomVec3InHemisphere(hit.Normal)
		ray := NewRay(hit.Point, randomDirection)
		
		if hit, _ := scene.Hit(ray, 0.001, 1.0); hit != nil {
			occlusion += 1.0
		}
	}
	
	return 1.0 - (occlusion / float64(samples))
}

func (gi *GlobalIllumination) CalculateIndirectLighting(hit *HitRecord, scene Scene, depth int) math.Vec3 {
	if !gi.IndirectLighting || depth > 3 {
		return math.Vec3{}
	}
	
	samples := 8
	indirectColor := math.Vec3{}
	
	for i := 0; i < samples; i++ {
		randomDirection := math.RandomVec3InHemisphere(hit.Normal)
		ray := NewRay(hit.Point, randomDirection)
		
		indirectColor = indirectColor.Add(traceRay(ray, scene, depth+1))
	}
	
	return indirectColor.DivScalar(float64(samples)).MulScalar(0.5)
}

type HitRecord struct {
	Point  math.Vec3
	Normal math.Vec3
	Material interface{}
}

type Scene interface {
	Hit(ray Ray, tMin, tMax float64) (*HitRecord, bool)
}

type Ray struct {
	Origin    math.Vec3
	Direction math.Vec3
}

func NewRay(origin, direction math.Vec3) Ray {
	return Ray{Origin: origin, Direction: direction}
}

func traceRay(ray Ray, scene Scene, depth int) math.Vec3 {
	return math.Vec3{}
} 