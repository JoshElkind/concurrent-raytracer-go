package output

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"raytraceGo/internal/math"
)

func SavePPM(img *image.RGBA, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()
	
	fmt.Fprintf(file, "P3\n%d %d\n255\n", width, height)
	
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			fmt.Fprintf(file, "%d %d %d ", r, g, b)
		}
		fmt.Fprintln(file)
	}
	
	return nil
}

func SavePPMFromVec3(pixels [][]math.Vec3, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	
	height := len(pixels)
	if height == 0 {
		return fmt.Errorf("empty pixel data")
	}
	width := len(pixels[0])
	
	fmt.Fprintf(file, "P3\n%d %d\n255\n", width, height)
	
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			color := pixels[y][x]
			r, g, b := color.ToRGB()
			fmt.Fprintf(file, "%d %d %d ", r, g, b)
		}
		fmt.Fprintln(file)
	}
	
	return nil
}

func SavePPMFromFloat64(pixels [][]float64, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	
	height := len(pixels)
	if height == 0 {
		return fmt.Errorf("empty pixel data")
	}
	width := len(pixels[0])
	
	fmt.Fprintf(file, "P2\n%d %d\n255\n", width, height)
	
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			value := int(pixels[y][x] * 255)
			if value > 255 {
				value = 255
			}
			if value < 0 {
				value = 0
			}
			fmt.Fprintf(file, "%d ", value)
		}
		fmt.Fprintln(file)
	}
	
	return nil
}

func SavePPMFromRGBA(pixels [][]color.RGBA, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	
	height := len(pixels)
	if height == 0 {
		return fmt.Errorf("empty pixel data")
	}
	width := len(pixels[0])
	
	fmt.Fprintf(file, "P3\n%d %d\n255\n", width, height)
	
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixel := pixels[y][x]
			fmt.Fprintf(file, "%d %d %d ", pixel.R, pixel.G, pixel.B)
		}
		fmt.Fprintln(file)
	}
	
	return nil
}

func SavePPMFromVec3WithGamma(pixels [][]math.Vec3, filename string, gamma float64) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	
	height := len(pixels)
	if height == 0 {
		return fmt.Errorf("empty pixel data")
	}
	width := len(pixels[0])
	
	fmt.Fprintf(file, "P3\n%d %d\n255\n", width, height)
	
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			color := pixels[y][x]
			
			r := math.FastPow(color.X, 1.0/gamma)
			g := math.FastPow(color.Y, 1.0/gamma)
			b := math.FastPow(color.Z, 1.0/gamma)
			
			r = math.FastClamp(r, 0.0, 1.0)
			g = math.FastClamp(g, 0.0, 1.0)
			b = math.FastClamp(b, 0.0, 1.0)
			
			ri := int(r * 255)
			gi := int(g * 255)
			bi := int(b * 255)
			
			fmt.Fprintf(file, "%d %d %d ", ri, gi, bi)
		}
		fmt.Fprintln(file)
	}
	
	return nil
}

func SavePPMFromVec3WithToneMapping(pixels [][]math.Vec3, filename string, exposure float64) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	
	height := len(pixels)
	if height == 0 {
		return fmt.Errorf("empty pixel data")
	}
	width := len(pixels[0])
	
	fmt.Fprintf(file, "P3\n%d %d\n255\n", width, height)
	
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			color := pixels[y][x]
			
			color = color.MulScalar(exposure)
			
			r := 1.0 - math.FastExp(-color.X)
			g := 1.0 - math.FastExp(-color.Y)
			b := 1.0 - math.FastExp(-color.Z)
			
			r = math.FastClamp(r, 0.0, 1.0)
			g = math.FastClamp(g, 0.0, 1.0)
			b = math.FastClamp(b, 0.0, 1.0)
			
			ri := int(r * 255)
			gi := int(g * 255)
			bi := int(b * 255)
			
			fmt.Fprintf(file, "%d %d %d ", ri, gi, bi)
		}
		fmt.Fprintln(file)
	}
	
	return nil
}

func SavePPMFromVec3WithReinhardToneMapping(pixels [][]math.Vec3, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	
	height := len(pixels)
	if height == 0 {
		return fmt.Errorf("empty pixel data")
	}
	width := len(pixels[0])
	
	fmt.Fprintf(file, "P3\n%d %d\n255\n", width, height)
	
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			color := pixels[y][x]
			
			r := color.X / (1.0 + color.X)
			g := color.Y / (1.0 + color.Y)
			b := color.Z / (1.0 + color.Z)
			
			r = math.FastClamp(r, 0.0, 1.0)
			g = math.FastClamp(g, 0.0, 1.0)
			b = math.FastClamp(b, 0.0, 1.0)
			
			ri := int(r * 255)
			gi := int(g * 255)
			bi := int(b * 255)
			
			fmt.Fprintf(file, "%d %d %d ", ri, gi, bi)
		}
		fmt.Fprintln(file)
	}
	
	return nil
} 