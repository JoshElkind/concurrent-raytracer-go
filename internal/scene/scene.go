package scene

import (
	"encoding/json"
	"fmt"
	"os"
	"raytraceGo/internal/geometry"
	"raytraceGo/internal/material"
	"raytraceGo/internal/math"
)

type Scene struct {
	Camera  Camera   `json:"camera"`
	Objects []Object `json:"objects"`
	Lights  []Light  `json:"lights"`
}

type Camera struct {
	Position    math.Vec3 `json:"position"`
	LookAt      math.Vec3 `json:"lookAt"`
	Up          math.Vec3 `json:"up"`
	FOV         float64   `json:"fov"`
	AspectRatio float64   `json:"aspectRatio"`
}

type Object struct {
	Type     string                 `json:"type"`
	Position math.Vec3             `json:"position"`
	Size     math.Vec3             `json:"size,omitempty"`
	Radius   float64               `json:"radius,omitempty"`
	Material map[string]interface{} `json:"material"`
}

type Light struct {
	Type      string    `json:"type"`
	Position  math.Vec3 `json:"position"`
	Color     math.Vec3 `json:"color"`
	Intensity float64   `json:"intensity"`
}

type Hittable interface {
	Hit(ray geometry.Ray, tMin, tMax float64) (*geometry.HitRecord, bool)
}

func LoadFromFile(filename string) (*Scene, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}
	
	var scene Scene
	if err := json.Unmarshal(data, &scene); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}
	
	return &scene, nil
}

func (s *Scene) GetHittables() []geometry.Hittable {
	var hittables []geometry.Hittable
	
	fmt.Println("Creating hittables from", len(s.Objects), "scene objects...")
	
	for i, obj := range s.Objects {
		fmt.Printf("  Processing object %d: Type=%s, Material=%s\n", i+1, obj.Type, obj.Material["type"])
		
		var hittable geometry.Hittable
		
		switch obj.Type {
		case "sphere":
			sphereMaterial := createMaterial(obj.Material)
			hittable = geometry.NewSphere(obj.Position, obj.Radius, sphereMaterial)
			fmt.Printf("    Created sphere at %v with radius %.1f\n", obj.Position, obj.Radius)
			
		case "cube":
			cubeMaterial := createMaterial(obj.Material)
			hittable = createCube(obj.Position, obj.Size, cubeMaterial)
			fmt.Printf("    Created cube at %v with size %v\n", obj.Position, obj.Size)
			
		default:
			fmt.Printf("    Unknown object type: %s\n", obj.Type)
			continue
		}
		
		hittables = append(hittables, hittable)
	}
	
	fmt.Printf("Created %d hittables total\n", len(hittables))
	return hittables
}

func (s *Scene) GetLights() []Light {
	return s.Lights
}

func (s *Scene) GetCamera() Camera {
	return s.Camera
}

func (s *Scene) GetSceneName() string {
	return "demo_scene"
}

func createMaterial(materialData map[string]interface{}) material.Material {
	materialType := materialData["type"].(string)
	
	switch materialType {
	case "lambertian":
		color := parseVec3(materialData["color"].([]interface{}))
		return material.NewLambertian(color)
		
	case "metal":
		color := parseVec3(materialData["color"].([]interface{}))
		roughness := getFloat(materialData, "roughness", 0.0)
		metallic := getFloat(materialData, "metallic", 1.0)
		specular := getFloat(materialData, "specular", 1.0)
		return material.NewMetal(color, roughness, metallic, specular)
		
	case "shiny":
		color := parseVec3(materialData["color"].([]interface{}))
		roughness := getFloat(materialData, "roughness", 0.0)
		metallic := getFloat(materialData, "metallic", 0.0)
		specular := getFloat(materialData, "specular", 1.0)
		return material.NewShinyMaterial(color, roughness, metallic, specular)
		
	case "perfectmirror":
		color := parseVec3(materialData["color"].([]interface{}))
		roughness := getFloat(materialData, "roughness", 0.0)
		return material.NewPerfectMirror(color, roughness)
		
	case "glass":
		color := parseVec3(materialData["color"].([]interface{}))
		refractionIndex := getFloat(materialData, "refractionIndex", 1.5)
		return material.NewGlass(refractionIndex, color)
		
	case "dielectric":
		refractionIndex := getFloat(materialData, "refractionIndex", 1.5)
		return material.NewDielectric(refractionIndex)
		
	case "diffuselight":
		color := parseVec3(materialData["color"].([]interface{}))
		return material.NewDiffuseLight(color)
		
	default:
		color := parseVec3(materialData["color"].([]interface{}))
		return material.NewLambertian(color)
	}
}

func createCube(position, size math.Vec3, material interface{}) geometry.Hittable {
	halfSize := size.DivScalar(2.0)
	
	vertices := []math.Vec3{
		position.Add(math.Vec3{X: -halfSize.X, Y: -halfSize.Y, Z: -halfSize.Z}),
		position.Add(math.Vec3{X: halfSize.X, Y: -halfSize.Y, Z: -halfSize.Z}),
		position.Add(math.Vec3{X: halfSize.X, Y: halfSize.Y, Z: -halfSize.Z}),
		position.Add(math.Vec3{X: -halfSize.X, Y: halfSize.Y, Z: -halfSize.Z}),
		position.Add(math.Vec3{X: -halfSize.X, Y: -halfSize.Y, Z: halfSize.Z}),
		position.Add(math.Vec3{X: halfSize.X, Y: -halfSize.Y, Z: halfSize.Z}),
		position.Add(math.Vec3{X: halfSize.X, Y: halfSize.Y, Z: halfSize.Z}),
		position.Add(math.Vec3{X: -halfSize.X, Y: halfSize.Y, Z: halfSize.Z}),
	}
	
	faces := [][]int{
		{0, 1, 2, 3},
		{1, 5, 6, 2},
		{5, 4, 7, 6},
		{4, 0, 3, 7},
		{3, 2, 6, 7},
		{4, 5, 1, 0},
	}
	
	var triangles []geometry.Hittable
	
	for _, face := range faces {
		v0 := vertices[face[0]]
		v1 := vertices[face[1]]
		v2 := vertices[face[2]]
		v3 := vertices[face[3]]
		
		triangle1 := geometry.NewTriangle(v0, v1, v2, material)
		triangle2 := geometry.NewTriangle(v0, v2, v3, material)
		
		triangles = append(triangles, triangle1, triangle2)
	}
	
	return &Mesh{
		Triangles: triangles,
	}
}

type Mesh struct {
	Triangles []geometry.Hittable
}

func (m *Mesh) Hit(ray geometry.Ray, tMin, tMax float64) (*geometry.HitRecord, bool) {
	var closestHit *geometry.HitRecord
	closestT := tMax
	
	for _, triangle := range m.Triangles {
		hitRecord, hit := triangle.Hit(ray, tMin, closestT)
		if hit {
			closestT = hitRecord.T
			closestHit = hitRecord
		}
	}
	
	return closestHit, closestHit != nil
}

func parseVec3(data []interface{}) math.Vec3 {
	return math.Vec3{
		X: data[0].(float64),
		Y: data[1].(float64),
		Z: data[2].(float64),
	}
}

func getFloat(data map[string]interface{}, key string, defaultValue float64) float64 {
	if value, exists := data[key]; exists {
		return value.(float64)
	}
	return defaultValue
}