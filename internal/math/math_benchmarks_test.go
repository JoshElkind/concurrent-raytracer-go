package math

import (
	"math"
	"testing"
	"time"
)

func BenchmarkMathOperations(b *testing.B) {
	vec1 := Vec3{X: 1.0, Y: 2.0, Z: 3.0}
	vec2 := Vec3{X: 4.0, Y: 5.0, Z: 6.0}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_ = vec1.Add(vec2)
		_ = vec1.Sub(vec2)
		_ = vec1.Mul(vec2)
		_ = vec1.DivScalar(2.0)
		_ = vec1.Dot(vec2)
		_ = vec1.Cross(vec2)
		_ = vec1.Length()
		_ = vec1.LengthSquared()
		_ = vec1.Normalize()
		_ = vec1.Reflect(vec2)
		_ = vec1.Refract(vec2, 1.5)
	}
}

func BenchmarkFastMath(b *testing.B) {
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		x := float64(i) * 0.1
		_ = FastSqrt(x)
		_ = FastExp(x)
		_ = FastLog(x + 1)
		_ = FastSin(x)
		_ = FastCos(x)
		_ = FastAtan(x)
		_ = FastPow(x, 2.0)
		_ = FastLerp(0.0, 1.0, x)
		_ = FastSmoothstep(0.0, 1.0, x)
	}
}

func BenchmarkNoise(b *testing.B) {
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		x := float64(i) * 0.1
		y := float64(i) * 0.2
		z := float64(i) * 0.3
		
		_ = FastNoise1D(x)
		_ = FastNoise2D(x, y)
		_ = FastNoise3D(x, y, z)
	}
}

func BenchmarkVectorOps(b *testing.B) {
	vec1 := Vec3{X: 1.0, Y: 2.0, Z: 3.0}
	vec2 := Vec3{X: 4.0, Y: 5.0, Z: 6.0}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_ = FastVec3Add(vec1, vec2)
		_ = FastVec3Sub(vec1, vec2)
		_ = FastVec3Mul(vec1, vec2)
		_ = FastVec3Div(vec1, vec2)
		_ = FastVec3Dot(vec1, vec2)
		_ = FastVec3Cross(vec1, vec2)
		_ = FastVec3Length(vec1)
		_ = FastVec3Normalize(vec1)
		_ = FastVec3Reflect(vec1, vec2)
		_ = FastVec3Refract(vec1, vec2, 1.5)
	}
}

func BenchmarkRandom(b *testing.B) {
	rng := NewFastRandom(12345)
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_ = rng.Next()
		_ = rng.Float64()
		_ = rng.Float64Range(0.0, 1.0)
	}
}

func TestMathematicalAccuracy(t *testing.T) {
	tests := []struct {
		name     string
		fastFunc func(float64) float64
		stdFunc  func(float64) float64
		inputs   []float64
		tolerance float64
	}{
		{
			name:     "FastSqrt",
			fastFunc: FastSqrt,
			stdFunc:  math.Sqrt,
			inputs:   []float64{0.1, 1.0, 10.0, 100.0},
			tolerance: 1e-2,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, input := range tt.inputs {
				fast := tt.fastFunc(input)
				std := tt.stdFunc(input)
				
				diff := math.Abs(fast - std)
				if diff > tt.tolerance {
					t.Errorf("%s(%.2f) = %.6f, want %.6f (diff: %.6f)", 
						tt.name, input, fast, std, diff)
				}
			}
		})
	}
}

func TestVectorMathAccuracy(t *testing.T) {
	vec1 := Vec3{X: 1.0, Y: 2.0, Z: 3.0}
	vec2 := Vec3{X: 4.0, Y: 5.0, Z: 6.0}
	
	result := vec1.Add(vec2)
	expected := Vec3{X: 5.0, Y: 7.0, Z: 9.0}
	if !vec3Equal(result, expected) {
		t.Errorf("Add: got %v, want %v", result, expected)
	}
	
	result = vec1.Sub(vec2)
	expected = Vec3{X: -3.0, Y: -3.0, Z: -3.0}
	if !vec3Equal(result, expected) {
		t.Errorf("Sub: got %v, want %v", result, expected)
	}
	
	result = vec1.Mul(vec2)
	expected = Vec3{X: 4.0, Y: 10.0, Z: 18.0}
	if !vec3Equal(result, expected) {
		t.Errorf("Mul: got %v, want %v", result, expected)
	}
	
	dot := vec1.Dot(vec2)
	expectedDot := 32.0
	if math.Abs(dot-expectedDot) > 1e-10 {
		t.Errorf("Dot: got %.6f, want %.6f", dot, expectedDot)
	}
	
	cross := vec1.Cross(vec2)
	expectedCross := Vec3{X: -3.0, Y: 6.0, Z: -3.0}
	if !vec3Equal(cross, expectedCross) {
		t.Errorf("Cross: got %v, want %v", cross, expectedCross)
	}
	
	length := vec1.Length()
	expectedLength := math.Sqrt(14.0)
	if math.Abs(length-expectedLength) > 1e-10 {
		t.Errorf("Length: got %.6f, want %.6f", length, expectedLength)
	}
}

func TestPerformanceComparison(t *testing.T) {
	iterations := 1000000
	
	start := time.Now()
	for i := 0; i < iterations; i++ {
		x := float64(i) * 0.001
		_ = math.Sqrt(x)
		_ = math.Exp(x)
		_ = math.Sin(x)
	}
	stdTime := time.Since(start)
	
	start = time.Now()
	for i := 0; i < iterations; i++ {
		x := float64(i) * 0.001
		_ = FastSqrt(x)
		_ = FastExp(x)
		_ = FastSin(x)
	}
	fastTime := time.Since(start)
	
	if fastTime >= stdTime {
		t.Logf("Fast math time: %v, Standard math time: %v", fastTime, stdTime)
		t.Log("Note: Fast math may not always be faster due to compiler optimizations")
	}
}

func vec3Equal(a, b Vec3) bool {
	tolerance := 1e-10
	return math.Abs(a.X-b.X) < tolerance &&
		math.Abs(a.Y-b.Y) < tolerance &&
		math.Abs(a.Z-b.Z) < tolerance
}

func BenchmarkMemoryUsage(b *testing.B) {
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		vectors := make([]Vec3, 1000)
		for j := 0; j < 1000; j++ {
			vectors[j] = Vec3{
				X: float64(j),
				Y: float64(j + 1),
				Z: float64(j + 2),
			}
		}
		
		for j := 0; j < len(vectors)-1; j++ {
			_ = vectors[j].Add(vectors[j+1])
			_ = vectors[j].Dot(vectors[j+1])
			_ = vectors[j].Cross(vectors[j+1])
		}
	}
}

func BenchmarkCachePerformance(b *testing.B) {
	size := 10000
	vectors := make([]Vec3, size)
	for i := 0; i < size; i++ {
		vectors[i] = Vec3{
			X: float64(i),
			Y: float64(i + 1),
			Z: float64(i + 2),
		}
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		sum := Vec3{}
		for j := 0; j < size; j++ {
			sum = sum.Add(vectors[j])
		}
		_ = sum
	}
}

func BenchmarkParallelOperations(b *testing.B) {
	vec1 := Vec3{X: 1.0, Y: 2.0, Z: 3.0}
	vec2 := Vec3{X: 4.0, Y: 5.0, Z: 6.0}
	
	b.ResetTimer()
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = vec1.Add(vec2)
			_ = vec1.Sub(vec2)
			_ = vec1.Mul(vec2)
			_ = vec1.Dot(vec2)
			_ = vec1.Cross(vec2)
		}
	})
} 