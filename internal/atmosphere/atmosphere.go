package atmosphere

import (
	stdmath "math"
	"raytraceGo/internal/math"
)

type AtmosphereConfig struct {
	SkyColorTop    math.Vec3
	SkyColorBottom math.Vec3
	
	SunDirection math.Vec3
	SunColor     math.Vec3
	SunIntensity float64
	SunSize      float64
	
	RayleighScattering math.Vec3
	MieScattering      math.Vec3
	AtmosphericDepth   float64
	
	FogDensity    float64
	FogColor      math.Vec3
	HazeIntensity float64
	
	TimeOfDay float64
}

func NewDefaultAtmosphere() *AtmosphereConfig {
	return &AtmosphereConfig{
		SkyColorTop:    math.Vec3{X: 0.6, Y: 0.8, Z: 1.0},
		SkyColorBottom: math.Vec3{X: 0.9, Y: 0.95, Z: 1.0},
		SunDirection:   math.Vec3{X: 0.0, Y: 0.8, Z: -0.6},
		SunColor:       math.Vec3{X: 1.0, Y: 0.98, Z: 0.95},
		SunIntensity:   1.2,
		SunSize:        0.015,
		RayleighScattering: math.Vec3{X: 0.6, Y: 0.8, Z: 1.0},
		MieScattering:      math.Vec3{X: 1.0, Y: 0.98, Z: 0.95},
		AtmosphericDepth:   0.3,
		FogDensity:         0.0,
		FogColor:           math.Vec3{X: 0.9, Y: 0.92, Z: 0.95},
		HazeIntensity:      0.05,
		TimeOfDay:          0.6,
	}
}

func NewWhiteAtmosphere() *AtmosphereConfig {
	return &AtmosphereConfig{
		SkyColorTop:    math.Vec3{X: 0.98, Y: 0.98, Z: 1.0},
		SkyColorBottom: math.Vec3{X: 0.92, Y: 0.92, Z: 0.95},
		SunDirection:   math.Vec3{X: 0.0, Y: 0.8, Z: -0.6},
		SunColor:       math.Vec3{X: 1.0, Y: 0.99, Z: 0.97},
		SunIntensity:   0.8,
		SunSize:        0.012,
		RayleighScattering: math.Vec3{X: 0.9, Y: 0.9, Z: 0.95},
		MieScattering:      math.Vec3{X: 0.95, Y: 0.95, Z: 0.98},
		AtmosphericDepth:   0.2,
		FogDensity:         0.0,
		FogColor:           math.Vec3{X: 0.95, Y: 0.95, Z: 0.98},
		HazeIntensity:      0.02,
		TimeOfDay:          0.6,
	}
}

func NewSunsetAtmosphere() *AtmosphereConfig {
	return &AtmosphereConfig{
		SkyColorTop:    math.Vec3{X: 1.0, Y: 0.4, Z: 0.2},
		SkyColorBottom: math.Vec3{X: 1.0, Y: 0.8, Z: 0.6},
		SunDirection:   math.Vec3{X: 0.0, Y: 0.3, Z: -0.9},
		SunColor:       math.Vec3{X: 1.0, Y: 0.6, Z: 0.3},
		SunIntensity:   1.2,
		SunSize:        0.03,
		RayleighScattering: math.Vec3{X: 1.0, Y: 0.4, Z: 0.2},
		MieScattering:      math.Vec3{X: 1.0, Y: 0.8, Z: 0.6},
		AtmosphericDepth:   0.8,
		FogDensity:         0.1,
		FogColor:           math.Vec3{X: 1.0, Y: 0.8, Z: 0.6},
		HazeIntensity:      0.3,
		TimeOfDay:          0.8,
	}
}

func NewNightAtmosphere() *AtmosphereConfig {
	return &AtmosphereConfig{
		SkyColorTop:    math.Vec3{X: 0.1, Y: 0.1, Z: 0.3},
		SkyColorBottom: math.Vec3{X: 0.2, Y: 0.2, Z: 0.4},
		SunDirection:   math.Vec3{X: 0.0, Y: -0.7, Z: -0.7},
		SunColor:       math.Vec3{X: 0.8, Y: 0.8, Z: 1.0},
		SunIntensity:   0.3,
		SunSize:        0.005,
		RayleighScattering: math.Vec3{X: 0.1, Y: 0.1, Z: 0.3},
		MieScattering:      math.Vec3{X: 0.8, Y: 0.8, Z: 1.0},
		AtmosphericDepth:   0.2,
		FogDensity:         0.0,
		FogColor:           math.Vec3{X: 0.1, Y: 0.1, Z: 0.2},
		HazeIntensity:      0.0,
		TimeOfDay:          0.0,
	}
}

func (a *AtmosphereConfig) GetSkyColor(rayDirection math.Vec3) math.Vec3 {
	unitDirection := math.FastVec3Normalize(rayDirection)
	
	t := 0.5 * (unitDirection.Y + 1.0)
	skyColor := math.FastVec3Lerp(a.SkyColorBottom, a.SkyColorTop, t)
	
	depth := stdmath.Max(0.0, unitDirection.Y)
	atmospheric := stdmath.Exp(-depth * a.AtmosphericDepth)
	scatteringColor := math.FastVec3Lerp(a.RayleighScattering, a.MieScattering, atmospheric)
	skyColor = math.FastVec3Lerp(skyColor, scatteringColor, 0.25) // Balanced atmospheric effect
	
	sunDot := math.FastVec3Dot(unitDirection, a.SunDirection)
	if sunDot > (1.0 - a.SunSize) {
		sunIntensity := stdmath.Pow((sunDot-(1.0-a.SunSize))/a.SunSize, 1.5)
		sunIntensity = stdmath.Min(sunIntensity, 1.0)
		skyColor = math.FastVec3Lerp(skyColor, a.SunColor, sunIntensity*a.SunIntensity*0.9)
	}
	
	timeFactor := a.TimeOfDay
	if timeFactor > 0.5 {
		timeFactor = 1.0 - timeFactor
	}
	timeFactor *= 2.0 // 0 to 1 range
	
	darkness := 1.0 - timeFactor*0.3 // Minimal darkening for clarity
	skyColor = math.FastVec3MulScalar(skyColor, darkness)
	
	if a.FogDensity > 0.0 {
		fogFactor := stdmath.Exp(-a.FogDensity)
		skyColor = math.FastVec3Lerp(a.FogColor, skyColor, fogFactor)
	}
	
	skyColor = skyColor.Clamp(0.1, 0.98) // Ensure minimum brightness for visibility
	
	return skyColor
}

func (a *AtmosphereConfig) GetAtmosphericAttenuation(distance float64) float64 {
	rayleighAttenuation := stdmath.Exp(-distance * 0.1)
	
	mieAttenuation := stdmath.Exp(-distance * 0.05)
	
	return rayleighAttenuation * mieAttenuation
} 