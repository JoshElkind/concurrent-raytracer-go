package renderer

import (
	stdmath "math"
	"raytraceGo/internal/geometry"
	"raytraceGo/internal/math"
	"raytraceGo/internal/scene"
)

func (r *ParallelRenderer) calculateSoftShadows(hit *geometry.HitRecord, hittables []geometry.Hittable, lights []scene.Light) math.Vec3 {
	shadowColor := math.Vec3{X: 1, Y: 1, Z: 1}
	
	for _, light := range lights {
		lightDir := light.Position.Sub(hit.Point).Normalize()
		lightDistance := light.Position.Sub(hit.Point).Length()
		
		shadowRay := geometry.NewRay(hit.Point, lightDir)
		
		shadowHit, hit := r.hitWorld(shadowRay, hittables, 0.001, lightDistance)
		if hit && shadowHit.T < lightDistance {
			shadowFactor := 0.3
			shadowColor = shadowColor.MulScalar(shadowFactor)
		}
	}
	
	return shadowColor
}

func (r *ParallelRenderer) applyDepthOfField(ray geometry.Ray, camera *scene.Camera) geometry.Ray {
	if !r.depthOfField {
		return ray
	}
	
	lensRadius := 0.1
	focusDistance := 10.0
	
	rd := math.RandomVec3InUnitDisk().MulScalar(lensRadius)
	offset := camera.Up.MulScalar(rd.X).Add(camera.LookAt.Cross(camera.Up).Normalize().MulScalar(rd.Y))
	
	origin := ray.Origin.Add(offset)
	direction := ray.Direction.MulScalar(focusDistance).Sub(offset).Normalize()
	
	return geometry.NewRay(origin, direction)
}

func (r *ParallelRenderer) calculateFresnel(cosTheta, eta float64) float64 {
	r0 := (1 - eta) / (1 + eta)
	r0 = r0 * r0
	return r0 + (1-r0)*stdmath.Pow(1-cosTheta, 5)
}

func (r *ParallelRenderer) calculateSchlick(cosine, refIdx float64) float64 {
	r0 := (1 - refIdx) / (1 + refIdx)
	r0 = r0 * r0
	return r0 + (1-r0)*stdmath.Pow(1-cosine, 5)
}

func (r *ParallelRenderer) calculateAtmosphericScattering(ray geometry.Ray, distance float64) math.Vec3 {
	beta := math.Vec3{X: 0.1, Y: 0.1, Z: 0.3}
	transmittance := math.Vec3{
		X: stdmath.Exp(-beta.X * distance),
		Y: stdmath.Exp(-beta.Y * distance),
		Z: stdmath.Exp(-beta.Z * distance),
	}
	return transmittance
}

func (r *ParallelRenderer) calculateMotionBlur(ray geometry.Ray, time float64) geometry.Ray {
	if time <= 0 {
		return ray
	}
	
	velocity := math.Vec3{X: 0.1, Y: 0, Z: 0}
	offset := velocity.MulScalar(time)
	
	newOrigin := ray.Origin.Add(offset)
	return geometry.NewRay(newOrigin, ray.Direction)
}

func (r *ParallelRenderer) calculateCaustics(hit *geometry.HitRecord, lights []scene.Light) math.Vec3 {
	causticColor := math.Vec3{}
	
	for _, light := range lights {
		lightDir := light.Position.Sub(hit.Point).Normalize()
		causticIntensity := stdmath.Max(0, hit.Normal.Dot(lightDir))
		causticColor = causticColor.Add(light.Color.MulScalar(causticIntensity))
	}
	
	return causticColor
}

func (r *ParallelRenderer) calculateSubsurfaceScattering(hit *geometry.HitRecord, depth int) math.Vec3 {
	if depth > 2 {
		return math.Vec3{}
	}
	
	scatterColor := math.Vec3{X: 0.8, Y: 0.6, Z: 0.6}
	
	return scatterColor.MulScalar(0.5)
}

func (r *ParallelRenderer) calculateVolumetricLighting(ray geometry.Ray, hittables []geometry.Hittable, lights []scene.Light) math.Vec3 {
	volumetricColor := math.Vec3{}
	
	for _, light := range lights {
		lightDir := light.Position.Sub(ray.Origin).Normalize()
		intensity := stdmath.Max(0, ray.Direction.Dot(lightDir))
		volumetricColor = volumetricColor.Add(light.Color.MulScalar(intensity * 0.1))
	}
	
	return volumetricColor
}

func (r *ParallelRenderer) calculateBumpMapping(hit *geometry.HitRecord) math.Vec3 {
	u := hit.Point.X * 10
	v := hit.Point.Y * 10
	
	bumpScale := 0.1
	bumpU := stdmath.Sin(u * 10) * bumpScale
	bumpV := stdmath.Cos(v * 10) * bumpScale
	
	normal := hit.Normal
	normal = normal.Add(math.Vec3{X: bumpU, Y: bumpV, Z: 0}).Normalize()
	
	return normal
}

func (r *ParallelRenderer) calculateProceduralTexture(hit *geometry.HitRecord) math.Vec3 {
	u := hit.Point.X * 10
	v := hit.Point.Y * 10
	
	noise := stdmath.Sin(u*20) * stdmath.Cos(v*20)
	pattern := stdmath.Sin(u*50) * stdmath.Sin(v*50)
	
	color := math.Vec3{
		X: (noise + 1) / 2,
		Y: (pattern + 1) / 2,
		Z: (noise * pattern + 1) / 2,
	}
	
	return color
} 