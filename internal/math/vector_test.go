package math

import (
	"math"
	"testing"
)

func TestVec3Add(t *testing.T) {
	v1 := Vec3{X: 1, Y: 2, Z: 3}
	v2 := Vec3{X: 4, Y: 5, Z: 6}
	result := v1.Add(v2)
	
	expected := Vec3{X: 5, Y: 7, Z: 9}
	if result != expected {
		t.Errorf("Add failed: got %v, want %v", result, expected)
	}
}

func TestVec3Sub(t *testing.T) {
	v1 := Vec3{X: 5, Y: 7, Z: 9}
	v2 := Vec3{X: 1, Y: 2, Z: 3}
	result := v1.Sub(v2)
	
	expected := Vec3{X: 4, Y: 5, Z: 6}
	if result != expected {
		t.Errorf("Sub failed: got %v, want %v", result, expected)
	}
}

func TestVec3Dot(t *testing.T) {
	v1 := Vec3{X: 1, Y: 2, Z: 3}
	v2 := Vec3{X: 4, Y: 5, Z: 6}
	result := v1.Dot(v2)
	
	expected := 1*4 + 2*5 + 3*6
	if result != float64(expected) {
		t.Errorf("Dot failed: got %f, want %f", result, float64(expected))
	}
}

func TestVec3Cross(t *testing.T) {
	v1 := Vec3{X: 1, Y: 0, Z: 0}
	v2 := Vec3{X: 0, Y: 1, Z: 0}
	result := v1.Cross(v2)
	
	expected := Vec3{X: 0, Y: 0, Z: 1}
	if result != expected {
		t.Errorf("Cross failed: got %v, want %v", result, expected)
	}
}

func TestVec3Length(t *testing.T) {
	v := Vec3{X: 3, Y: 4, Z: 0}
	result := v.Length()
	
	expected := 5.0
	if math.Abs(result-expected) > 1e-10 {
		t.Errorf("Length failed: got %f, want %f", result, expected)
	}
}

func TestVec3Normalize(t *testing.T) {
	v := Vec3{X: 3, Y: 4, Z: 0}
	result := v.Normalize()
	
	expected := Vec3{X: 0.6, Y: 0.8, Z: 0}
	if math.Abs(result.X-expected.X) > 1e-10 || 
	   math.Abs(result.Y-expected.Y) > 1e-10 || 
	   math.Abs(result.Z-expected.Z) > 1e-10 {
		t.Errorf("Normalize failed: got %v, want %v", result, expected)
	}
}

func TestVec3Reflect(t *testing.T) {
	v := Vec3{X: 1, Y: -1, Z: 0}
	normal := Vec3{X: 0, Y: 1, Z: 0}
	result := v.Reflect(normal)
	
	expected := Vec3{X: 1, Y: 1, Z: 0}
	if math.Abs(result.X-expected.X) > 1e-10 || 
	   math.Abs(result.Y-expected.Y) > 1e-10 || 
	   math.Abs(result.Z-expected.Z) > 1e-10 {
		t.Errorf("Reflect failed: got %v, want %v", result, expected)
	}
}

func TestVec3Clamp(t *testing.T) {
	v := Vec3{X: -1, Y: 0.5, Z: 2}
	result := v.Clamp(0, 1)
	
	expected := Vec3{X: 0, Y: 0.5, Z: 1}
	if result != expected {
		t.Errorf("Clamp failed: got %v, want %v", result, expected)
	}
}

func TestVec3ToRGB(t *testing.T) {
	v := Vec3{X: 0.5, Y: 0.25, Z: 1.0}
	r, g, b := v.ToRGB()
	
	expectedR, expectedG, expectedB := uint8(127), uint8(63), uint8(255)
	if r != expectedR || g != expectedG || b != expectedB {
		t.Errorf("ToRGB failed: got (%d, %d, %d), want (%d, %d, %d)", 
			r, g, b, expectedR, expectedG, expectedB)
	}
} 