package effects

import (
	stdmath "math"
	"raytraceGo/internal/math"
)

type FireEffect struct {
	Enabled     bool
	Intensity   float64
	Color       math.Vec3
	Height      float64
	Width       float64
	Turbulence  float64
	Time        float64
}

func NewFireEffect(intensity float64, color math.Vec3, height, width, turbulence float64) *FireEffect {
	return &FireEffect{
		Enabled:    true,
		Intensity:  intensity,
		Color:      color,
		Height:     height,
		Width:      width,
		Turbulence: turbulence,
		Time:       0.0,
	}
}

func (f *FireEffect) CalculateFire(point math.Vec3, time float64) math.Vec3 {
	if !f.Enabled {
		return math.Vec3{}
	}
	
	noise := f.fireNoise(point, time)
	flame := f.calculateFlameShape(point, time)
	
	fireIntensity := noise * flame * f.Intensity
	
	heightFactor := point.Y / f.Height
	baseColor := f.Color
	hotColor := math.Vec3{X: 1.0, Y: 0.3, Z: 0.0} // Orange-red
	fireColor := baseColor.Lerp(hotColor, heightFactor)
	
	return fireColor.MulScalar(fireIntensity)
}

func (f *FireEffect) fireNoise(point math.Vec3, time float64) float64 {
	noise1 := stdmath.Sin(point.X*2.0 + time*3.0) * stdmath.Cos(point.Z*2.0 + time*2.0)
	noise2 := stdmath.Sin(point.X*4.0 + time*5.0) * stdmath.Cos(point.Z*4.0 + time*4.0) * 0.5
	noise3 := stdmath.Sin(point.X*8.0 + time*7.0) * stdmath.Cos(point.Z*8.0 + time*6.0) * 0.25
	
	return (noise1 + noise2 + noise3) * f.Turbulence
}

func (f *FireEffect) calculateFlameShape(point math.Vec3, time float64) float64 {
	heightFactor := point.Y / f.Height
	widthFactor := stdmath.Abs(point.X) / f.Width
	
	flameShape := stdmath.Exp(-heightFactor * 3.0) * stdmath.Exp(-widthFactor * 2.0)
	
	movement := stdmath.Sin(time * 2.0) * 0.1
	flameShape *= (1.0 + movement)
	
	return stdmath.Max(0.0, flameShape)
}

type ExplosionEffect struct {
	Enabled     bool
	Intensity   float64
	Radius      float64
	Particles   int
	Time        float64
	Duration    float64
}

func NewExplosionEffect(intensity, radius float64, particles int, duration float64) *ExplosionEffect {
	return &ExplosionEffect{
		Enabled:   true,
		Intensity: intensity,
		Radius:    radius,
		Particles: particles,
		Time:      0.0,
		Duration:  duration,
	}
}

func (e *ExplosionEffect) CalculateExplosion(point math.Vec3, time float64) math.Vec3 {
	if !e.Enabled || time > e.Duration {
		return math.Vec3{}
	}
	
	distance := point.Length()
	waveRadius := e.Radius * (time / e.Duration)
	
	if distance > waveRadius {
		return math.Vec3{}
	}
	
	intensity := e.Intensity * (1.0 - time/e.Duration)
	
	centerColor := math.Vec3{X: 1.0, Y: 0.8, Z: 0.0} // Bright yellow
	edgeColor := math.Vec3{X: 0.8, Y: 0.2, Z: 0.0}   // Orange-red
	
	distanceFactor := distance / waveRadius
	explosionColor := centerColor.Lerp(edgeColor, distanceFactor)
	
	return explosionColor.MulScalar(intensity)
}

type LightningEffect struct {
	Enabled     bool
	Intensity   float64
	Branches    int
	Time        float64
	Duration    float64
}

func NewLightningEffect(intensity float64, branches int, duration float64) *LightningEffect {
	return &LightningEffect{
		Enabled:   true,
		Intensity: intensity,
		Branches:  branches,
		Time:      0.0,
		Duration:  duration,
	}
}

func (l *LightningEffect) CalculateLightning(point math.Vec3, time float64) math.Vec3 {
	if !l.Enabled || time > l.Duration {
		return math.Vec3{}
	}
	
	intensity := l.Intensity * stdmath.Sin(time * 50.0) * (1.0 - time/l.Duration)
	
	lightningColor := math.Vec3{X: 0.8, Y: 0.9, Z: 1.0} // Electric blue
	
	return lightningColor.MulScalar(intensity)
}

type AuroraEffect struct {
	Enabled     bool
	Intensity   float64
	Color       math.Vec3
	Height      float64
	Width       float64
	Time        float64
}

func NewAuroraEffect(intensity float64, color math.Vec3, height, width float64) *AuroraEffect {
	return &AuroraEffect{
		Enabled:   true,
		Intensity: intensity,
		Color:     color,
		Height:    height,
		Width:     width,
		Time:      0.0,
	}
}

func (a *AuroraEffect) CalculateAurora(point math.Vec3, time float64) math.Vec3 {
	if !a.Enabled {
		return math.Vec3{}
	}
	
	curtain := stdmath.Sin(point.X * 0.5 + time * 0.2) * stdmath.Cos(point.Z * 0.3 + time * 0.1)
	
	heightFactor := stdmath.Max(0.0, (point.Y - a.Height*0.5) / (a.Height * 0.5))
	
	auroraColor := a.Color
	auroraColor = auroraColor.Add(math.Vec3{
		X: stdmath.Sin(time * 0.5) * 0.2,
		Y: stdmath.Cos(time * 0.3) * 0.2,
		Z: stdmath.Sin(time * 0.7) * 0.2,
	})
	
	intensity := curtain * heightFactor * a.Intensity
	return auroraColor.MulScalar(intensity)
}

type HologramEffect struct {
	Enabled     bool
	Intensity   float64
	Color       math.Vec3
	ScanLines   bool
	Glitch      bool
	Time        float64
}

func NewHologramEffect(intensity float64, color math.Vec3) *HologramEffect {
	return &HologramEffect{
		Enabled:   true,
		Intensity: intensity,
		Color:     color,
		ScanLines: true,
		Glitch:    true,
		Time:      0.0,
	}
}

func (h *HologramEffect) CalculateHologram(point math.Vec3, time float64) math.Vec3 {
	if !h.Enabled {
		return math.Vec3{}
	}
	
	intensity := h.Intensity
	
	if h.ScanLines {
		scanLine := stdmath.Sin(point.Y * 50.0 + time * 2.0) * 0.5 + 0.5
		intensity *= scanLine
	}
	
	if h.Glitch {
		glitch := stdmath.Sin(time * 10.0) * stdmath.Sin(point.X * 100.0)
		if glitch > 0.8 {
			intensity *= 2.0
		}
	}
	
	hologramColor := h.Color
	hologramColor = hologramColor.Add(math.Vec3{
		X: stdmath.Sin(time * 0.5) * 0.1,
		Y: stdmath.Cos(time * 0.3) * 0.1,
		Z: stdmath.Sin(time * 0.7) * 0.1,
	})
	
	return hologramColor.MulScalar(intensity)
}

type PortalEffect struct {
	Enabled     bool
	Intensity   float64
	Color       math.Vec3
	Radius      float64
	SwirlSpeed  float64
	Time        float64
}

func NewPortalEffect(intensity float64, color math.Vec3, radius, swirlSpeed float64) *PortalEffect {
	return &PortalEffect{
		Enabled:    true,
		Intensity:  intensity,
		Color:      color,
		Radius:     radius,
		SwirlSpeed: swirlSpeed,
		Time:       0.0,
	}
}

func (p *PortalEffect) CalculatePortal(point math.Vec3, time float64) math.Vec3 {
	if !p.Enabled {
		return math.Vec3{}
	}
	
	distance := stdmath.Sqrt(point.X*point.X + point.Z*point.Z)
	
	if distance > p.Radius {
		return math.Vec3{}
	}
	
	angle := stdmath.Atan2(point.Z, point.X)
	swirl := stdmath.Sin(angle*3.0 + time*p.SwirlSpeed)
	
	radialIntensity := 1.0 - (distance / p.Radius)
	
	portalColor := p.Color
	portalColor = portalColor.Add(math.Vec3{
		X: stdmath.Sin(time * 2.0) * 0.3,
		Y: stdmath.Cos(time * 1.5) * 0.3,
		Z: stdmath.Sin(time * 2.5) * 0.3,
	})
	
	intensity := radialIntensity * swirl * p.Intensity
	return portalColor.MulScalar(intensity)
}

type EnergyFieldEffect struct {
	Enabled     bool
	Intensity   float64
	Color       math.Vec3
	Radius      float64
	Frequency   float64
	Time        float64
}

func NewEnergyFieldEffect(intensity float64, color math.Vec3, radius, frequency float64) *EnergyFieldEffect {
	return &EnergyFieldEffect{
		Enabled:   true,
		Intensity: intensity,
		Color:     color,
		Radius:    radius,
		Frequency: frequency,
		Time:      0.0,
	}
}

func (ef *EnergyFieldEffect) CalculateEnergyField(point math.Vec3, time float64) math.Vec3 {
	if !ef.Enabled {
		return math.Vec3{}
	}
	
	distance := point.Length()
	
	if distance > ef.Radius {
		return math.Vec3{}
	}
	
	pulse := stdmath.Sin(time * ef.Frequency) * 0.5 + 0.5
	
	fieldIntensity := (1.0 - distance/ef.Radius) * pulse * ef.Intensity
	
	energyColor := ef.Color
	energyColor = energyColor.Add(math.Vec3{
		X: stdmath.Sin(time * 3.0) * 0.2,
		Y: stdmath.Cos(time * 2.0) * 0.2,
		Z: stdmath.Sin(time * 4.0) * 0.2,
	})
	
	return energyColor.MulScalar(fieldIntensity)
}

type PlasmaEffect struct {
	Enabled     bool
	Intensity   float64
	Temperature float64
	Time        float64
}

func NewPlasmaEffect(intensity, temperature float64) *PlasmaEffect {
	return &PlasmaEffect{
		Enabled:     true,
		Intensity:   intensity,
		Temperature: temperature,
		Time:        0.0,
	}
}

func (p *PlasmaEffect) CalculatePlasma(point math.Vec3, time float64) math.Vec3 {
	if !p.Enabled {
		return math.Vec3{}
	}
	
	noise1 := stdmath.Sin(point.X*2.0 + time*3.0) * stdmath.Cos(point.Y*2.0 + time*2.0)
	noise2 := stdmath.Sin(point.X*4.0 + time*5.0) * stdmath.Cos(point.Y*4.0 + time*4.0) * 0.5
	noise3 := stdmath.Sin(point.X*8.0 + time*7.0) * stdmath.Cos(point.Y*8.0 + time*6.0) * 0.25
	
	plasmaNoise := (noise1 + noise2 + noise3) / 3.0
	
	hotColor := math.Vec3{X: 1.0, Y: 0.2, Z: 0.0}   // Red-hot
	warmColor := math.Vec3{X: 1.0, Y: 0.6, Z: 0.0}  // Orange
	coolColor := math.Vec3{X: 0.2, Y: 0.4, Z: 1.0}  // Blue
	
	temperature := p.Temperature + plasmaNoise*0.5
	
	var plasmaColor math.Vec3
	if temperature > 0.7 {
		plasmaColor = hotColor
	} else if temperature > 0.3 {
		plasmaColor = warmColor
	} else {
		plasmaColor = coolColor
	}
	
	intensity := p.Intensity * (0.5 + plasmaNoise*0.5)
	return plasmaColor.MulScalar(intensity)
}

type CrystalEffect struct {
	Enabled     bool
	Intensity   float64
	Color       math.Vec3
	Facets      int
	Refraction  float64
	Time        float64
}

func NewCrystalEffect(intensity float64, color math.Vec3, facets int, refraction float64) *CrystalEffect {
	return &CrystalEffect{
		Enabled:    true,
		Intensity:  intensity,
		Color:      color,
		Facets:     facets,
		Refraction: refraction,
		Time:       0.0,
	}
}

func (c *CrystalEffect) CalculateCrystal(point math.Vec3, time float64) math.Vec3 {
	if !c.Enabled {
		return math.Vec3{}
	}
	
	facetPattern := stdmath.Sin(point.X*float64(c.Facets)) * stdmath.Cos(point.Y*float64(c.Facets))
	
	internalStructure := stdmath.Sin(point.X*10.0 + time) * stdmath.Cos(point.Y*10.0 + time*0.5)
	
	crystalColor := c.Color
	crystalColor = crystalColor.Add(math.Vec3{
		X: stdmath.Sin(time * 0.3) * 0.1,
		Y: stdmath.Cos(time * 0.2) * 0.1,
		Z: stdmath.Sin(time * 0.4) * 0.1,
	})
	
	intensity := c.Intensity * (0.5 + facetPattern*0.3 + internalStructure*0.2)
	return crystalColor.MulScalar(intensity)
}

type NebulaEffect struct {
	Enabled     bool
	Intensity   float64
	Colors      []math.Vec3
	Scale       float64
	Time        float64
}

func NewNebulaEffect(intensity float64, colors []math.Vec3, scale float64) *NebulaEffect {
	return &NebulaEffect{
		Enabled:   true,
		Intensity: intensity,
		Colors:    colors,
		Scale:     scale,
		Time:      0.0,
	}
}

func (n *NebulaEffect) CalculateNebula(point math.Vec3, time float64) math.Vec3 {
	if !n.Enabled {
		return math.Vec3{}
	}
	
	noise1 := stdmath.Sin(point.X*n.Scale + time*0.1) * stdmath.Cos(point.Y*n.Scale + time*0.2)
	noise2 := stdmath.Sin(point.X*n.Scale*2.0 + time*0.3) * stdmath.Cos(point.Y*n.Scale*2.0 + time*0.4) * 0.5
	noise3 := stdmath.Sin(point.X*n.Scale*4.0 + time*0.5) * stdmath.Cos(point.Y*n.Scale*4.0 + time*0.6) * 0.25
	
	nebulaNoise := (noise1 + noise2 + noise3) / 3.0
	
	if len(n.Colors) == 0 {
		return math.Vec3{}
	}
	
	colorIndex := int((nebulaNoise + 1.0) * 0.5 * float64(len(n.Colors)-1))
	colorIndex = int(stdmath.Max(0, stdmath.Min(float64(len(n.Colors)-1), float64(colorIndex))))
	
	nebulaColor := n.Colors[colorIndex]
	intensity := n.Intensity * (0.3 + nebulaNoise*0.7)
	
	return nebulaColor.MulScalar(intensity)
} 