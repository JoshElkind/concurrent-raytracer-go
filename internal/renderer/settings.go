package renderer

func (r *ParallelRenderer) SetSamples(samples int) {
	r.samples = samples
}

func (r *ParallelRenderer) SetMaxDepth(maxDepth int) {
	r.maxDepth = maxDepth
}

func (r *ParallelRenderer) SetAntiAliasing(antiAliasing bool) {
	r.antiAliasing = antiAliasing
}

func (r *ParallelRenderer) SetRecursiveReflections(recursiveReflections bool) {
	r.recursiveReflections = recursiveReflections
}

func (r *ParallelRenderer) SetSoftShadows(softShadows bool) {
	r.softShadows = softShadows
}

func (r *ParallelRenderer) SetDepthOfField(depthOfField bool) {
	r.depthOfField = depthOfField
}

func (r *ParallelRenderer) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"workers":              r.numWorkers,
		"samples":              r.samples,
		"maxDepth":             r.maxDepth,
		"antiAliasing":         r.antiAliasing,
		"recursiveReflections": r.recursiveReflections,
		"softShadows":          r.softShadows,
		"depthOfField":         r.depthOfField,
	}
} 