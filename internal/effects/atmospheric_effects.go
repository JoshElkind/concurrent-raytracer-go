package effects

import (
	stdmath "math"
	"raytraceGo/internal/math"
)

type AtmosphericScattering struct {
	Enabled     bool
	Density     float64
	Scattering  float64
	Absorption  float64
	PhaseFunction float64
	Height      float64
}

func NewAtmosphericScattering(density, scattering, absorption, height float64) *AtmosphericScattering {
	return &AtmosphericScattering{
		Enabled:        true,
		Density:        density,
		Scattering:     scattering,
		Absorption:     absorption,
		PhaseFunction:  0.9,
		Height:         height,
	}
}

func (as *AtmosphericScattering) CalculateScattering(ray math.Vec3, sunDirection math.Vec3, distance float64) math.Vec3 {
	if !as.Enabled {
		return math.Vec3{}
	}
	
	rayleighScattering := as.calculateRayleighScattering(ray, sunDirection, distance)
	
	mieScattering := as.calculateMieScattering(ray, sunDirection, distance)
	
	return rayleighScattering.Add(mieScattering)
}

func (as *AtmosphericScattering) calculateRayleighScattering(ray math.Vec3, sunDirection math.Vec3, distance float64) math.Vec3 {
	rayleighCoeff := math.Vec3{
		X: 5.802,  // Red
		Y: 13.558, // Green  
		Z: 33.1,   // Blue
	}
	
	cosTheta := ray.Dot(sunDirection)
	phase := 3.0 / (16.0 * stdmath.Pi) * (1.0 + cosTheta*cosTheta)
	
	density := as.getAtmosphericDensity(distance)
	
	scattering := rayleighCoeff.MulScalar(as.Scattering * density * phase * distance)
	
	return scattering
}

func (as *AtmosphericScattering) calculateMieScattering(ray math.Vec3, sunDirection math.Vec3, distance float64) math.Vec3 {
	mieCoeff := 21.0
	
	cosTheta := ray.Dot(sunDirection)
	g := as.PhaseFunction
	phase := (1.0 - g*g) / stdmath.Pow(1.0 + g*g - 2.0*g*cosTheta, 1.5)
	
	density := as.getAtmosphericDensity(distance)
	
	scattering := math.Vec3{X: mieCoeff, Y: mieCoeff, Z: mieCoeff}.MulScalar(as.Scattering * density * phase * distance)
	
	return scattering
}

func (as *AtmosphericScattering) getAtmosphericDensity(height float64) float64 {
	return as.Density * stdmath.Exp(-height / as.Height)
}

type VolumetricLighting struct {
	Enabled     bool
	Density     float64
	Scattering  float64
	Absorption  float64
	Steps       int
	MaxDistance float64
}

func NewVolumetricLighting(density, scattering, absorption, maxDistance float64) *VolumetricLighting {
	return &VolumetricLighting{
		Enabled:     true,
		Density:     density,
		Scattering:  scattering,
		Absorption:  absorption,
		Steps:       64,
		MaxDistance: maxDistance,
	}
}

func (vl *VolumetricLighting) CalculateVolumetricLighting(ray math.Vec3, lightDirection math.Vec3, distance float64) math.Vec3 {
	if !vl.Enabled {
		return math.Vec3{}
	}
	
	stepSize := distance / float64(vl.Steps)
	accumulatedLight := math.Vec3{}
	transmittance := 1.0
	
	for i := 0; i < vl.Steps; i++ {
		stepDistance := float64(i) * stepSize
		samplePoint := ray.MulScalar(stepDistance)
		
		density := vl.getVolumetricDensity(samplePoint)
		
		scattering := vl.calculateScattering(ray, lightDirection, density)
		
		accumulatedLight = accumulatedLight.Add(scattering.MulScalar(transmittance * stepSize))
		
		extinction := vl.Absorption + vl.Scattering
		transmittance *= stdmath.Exp(-extinction * density * stepSize)
	}
	
	return accumulatedLight
}

func (vl *VolumetricLighting) getVolumetricDensity(point math.Vec3) float64 {
	return vl.Density * stdmath.Exp(-point.Length() / 10.0)
}

func (vl *VolumetricLighting) calculateScattering(ray math.Vec3, lightDirection math.Vec3, density float64) math.Vec3 {
	scatteringCoeff := math.Vec3{X: 1.0, Y: 1.0, Z: 1.0}
	return scatteringCoeff.MulScalar(vl.Scattering * density)
}

type Fog struct {
	Enabled     bool
	Density     float64
	Color       math.Vec3
	Height      float64
	FogType     FogType
}

type FogType int

const (
	LinearFog FogType = iota
	ExponentialFog
	ExponentialSquaredFog
)

func NewFog(density float64, color math.Vec3, height float64, fogType FogType) *Fog {
	return &Fog{
		Enabled: true,
		Density: density,
		Color:   color,
		Height:  height,
		FogType: fogType,
	}
}

func (f *Fog) CalculateFogFactor(distance float64) float64 {
	if !f.Enabled {
		return 0.0
	}
	
	switch f.FogType {
	case LinearFog:
		return stdmath.Min(1.0, distance / f.Height)
	case ExponentialFog:
		return 1.0 - stdmath.Exp(-f.Density * distance)
	case ExponentialSquaredFog:
		return 1.0 - stdmath.Exp(-f.Density * f.Density * distance * distance)
	default:
		return 0.0
	}
}

func (f *Fog) ApplyFog(color math.Vec3, distance float64) math.Vec3 {
	fogFactor := f.CalculateFogFactor(distance)
	return color.Lerp(f.Color, fogFactor)
}

type MotionBlur struct {
	Enabled     bool
	ShutterTime float64
	Velocity    math.Vec3
}

func NewMotionBlur(shutterTime float64, velocity math.Vec3) *MotionBlur {
	return &MotionBlur{
		Enabled:     true,
		ShutterTime: shutterTime,
		Velocity:    velocity,
	}
}

func (mb *MotionBlur) ApplyMotionBlur(position math.Vec3, time float64) math.Vec3 {
	if !mb.Enabled {
		return position
	}
	
	motionOffset := mb.Velocity.MulScalar(time * mb.ShutterTime)
	return position.Add(motionOffset)
}

type DepthOfField struct {
	Enabled     bool
	FocusDistance float64
	Aperture     float64
	FocalLength  float64
}

func NewDepthOfField(focusDistance, aperture, focalLength float64) *DepthOfField {
	return &DepthOfField{
		Enabled:       true,
		FocusDistance: focusDistance,
		Aperture:     aperture,
		FocalLength:  focalLength,
	}
}

func (dof *DepthOfField) CalculateCircleOfConfusion(distance float64) float64 {
	if !dof.Enabled {
		return 0.0
	}
	
	focusDistance := dof.FocusDistance
	focalLength := dof.FocalLength
	aperture := dof.Aperture
	
	if stdmath.Abs(distance - focusDistance) < 0.001 {
		return 0.0
	}
	
	coc := (stdmath.Abs(distance - focusDistance) / distance) * (focalLength * focalLength) / (focusDistance * aperture)
	return coc
}

func (dof *DepthOfField) GetDefocusDiskRadius(distance float64) float64 {
	return dof.CalculateCircleOfConfusion(distance)
}

type LensFlare struct {
	Enabled     bool
	Intensity   float64
	Color       math.Vec3
	Size        float64
	Elements    []LensFlareElement
}

type LensFlareElement struct {
	Position float64
	Size     float64
	Color    math.Vec3
	Intensity float64
}

func NewLensFlare(intensity float64, color math.Vec3, size float64) *LensFlare {
	return &LensFlare{
		Enabled:   true,
		Intensity: intensity,
		Color:     color,
		Size:      size,
		Elements:  []LensFlareElement{
			{Position: 0.0, Size: 0.1, Color: color, Intensity: 1.0},
			{Position: 0.3, Size: 0.05, Color: color, Intensity: 0.7},
			{Position: 0.6, Size: 0.08, Color: color, Intensity: 0.5},
			{Position: 0.9, Size: 0.03, Color: color, Intensity: 0.3},
		},
	}
}

func (lf *LensFlare) CalculateLensFlare(sunScreenPos math.Vec3, screenCenter math.Vec3) math.Vec3 {
	if !lf.Enabled {
		return math.Vec3{}
	}
	
	direction := sunScreenPos.Sub(screenCenter).Normalize()
	
	flareColor := math.Vec3{}
	
	for _, element := range lf.Elements {
		elementPos := screenCenter.Add(direction.MulScalar(element.Position * lf.Size))
		
		distance := elementPos.Sub(sunScreenPos).Length()
		
		intensity := element.Intensity * lf.Intensity * (1.0 - distance/lf.Size)
		intensity = stdmath.Max(0.0, intensity)
		
		flareColor = flareColor.Add(element.Color.MulScalar(intensity))
	}
	
	return flareColor
}

type Bloom struct {
	Enabled     bool
	Threshold   float64
	Intensity   float64
	Radius      float64
}

func NewBloom(threshold, intensity, radius float64) *Bloom {
	return &Bloom{
		Enabled:   true,
		Threshold: threshold,
		Intensity: intensity,
		Radius:    radius,
	}
}

func (b *Bloom) CalculateBloomIntensity(color math.Vec3) float64 {
	if !b.Enabled {
		return 0.0
	}
	
	luminance := 0.299*color.X + 0.587*color.Y + 0.114*color.Z
	
	if luminance > b.Threshold {
		return (luminance - b.Threshold) * b.Intensity
	}
	
	return 0.0
}

func (b *Bloom) ApplyBloom(color math.Vec3) math.Vec3 {
	bloomIntensity := b.CalculateBloomIntensity(color)
	return color.Add(color.MulScalar(bloomIntensity))
}

type ChromaticAberration struct {
	Enabled     bool
	RedOffset   float64
	GreenOffset float64
	BlueOffset  float64
}

func NewChromaticAberration(redOffset, greenOffset, blueOffset float64) *ChromaticAberration {
	return &ChromaticAberration{
		Enabled:     true,
		RedOffset:   redOffset,
		GreenOffset: greenOffset,
		BlueOffset:  blueOffset,
	}
}

func (ca *ChromaticAberration) ApplyChromaticAberration(color math.Vec3, uv math.Vec3) math.Vec3 {
	if !ca.Enabled {
		return color
	}
	
	_ = uv.Add(math.Vec3{X: ca.RedOffset, Y: ca.RedOffset})
	_ = uv.Add(math.Vec3{X: ca.GreenOffset, Y: ca.GreenOffset})
	_ = uv.Add(math.Vec3{X: ca.BlueOffset, Y: ca.BlueOffset})
	
	redColor := color.X
	greenColor := color.Y
	blueColor := color.Z
	
	return math.Vec3{X: redColor, Y: greenColor, Z: blueColor}
}

type Vignette struct {
	Enabled     bool
	Intensity   float64
	Radius      float64
	Softness    float64
}

func NewVignette(intensity, radius, softness float64) *Vignette {
	return &Vignette{
		Enabled:   true,
		Intensity: intensity,
		Radius:    radius,
		Softness:  softness,
	}
}

func (v *Vignette) CalculateVignette(uv math.Vec3) float64 {
	if !v.Enabled {
		return 1.0
	}
	
	center := math.Vec3{X: 0.5, Y: 0.5}
	distance := uv.Sub(center).Length()
	
	factor := 1.0 - (distance / v.Radius)
	factor = stdmath.Max(0.0, factor)
	factor = stdmath.Pow(factor, v.Softness)
	
	return 1.0 - (v.Intensity * (1.0 - factor))
}

func (v *Vignette) ApplyVignette(color math.Vec3, uv math.Vec3) math.Vec3 {
	vignetteFactor := v.CalculateVignette(uv)
	return color.MulScalar(vignetteFactor)
} 