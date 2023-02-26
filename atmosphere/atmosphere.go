package atmosphere

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"math"
	"os"
	"strconv"

	"github.com/andewx/dieselfluid/math/common"
	"github.com/andewx/dieselfluid/math/polar"
	"github.com/andewx/dieselfluid/math/vector"
	"github.com/andewx/dieselfluid/render/light"
	"github.com/andewx/sampler"
)

/*
Atmosphere environment lighting model.

This model generates a sun position based on lat/long and solar times and then
we simulate atmospheric scattering processes via Rayleigh/Mie Scattering.
*/
const (
	PI                 = 3.141529
	AXIAL_TILT         = 23.5      //Degrees
	RAYLEIGH_SAMPLES   = 25        //RAYLEIGH sampling
	LIGHT_PATH_SAMPLES = 25        //PATH SAMPLES
	AU                 = 150000000 //SUN
	HM                 = 1500      //AEROSOL MIE SCATTER HEIGHT
	HR                 = 8000      //RAYLEIGH SCATTER HEIGHT

)

type Domain struct {
	min float32
	max float32
}

// Maps a value with an associated domain to this Domain
func (m *Domain) Map(value float32, domain Domain) float32 {
	r := domain.max - domain.min
	norm := (r - value) / r
	return m.min + norm*(m.max-m.min)
}

// Atmosphere Environment
type Atmosphere struct {
	Light light.Light
	Spd   light.Spectrum
	Sun   polar.Polar //Orbtial Solar System Earth 2 Sun Polar Coordinate
	Earth *EarthCoords
	Day   float32
	Dir   vector.Vec //Euclidian Sun Directio
}

// Allocates Default Data Structure and Solar Coords Structs
func NewAtmosphere(lat float32, long float32) *Atmosphere {
	sky := Atmosphere{}
	sky.Earth = NewEarth(65.0, 0)
	sky.Sun = polar.NewPolar(-AU)
	sky.Light = light.Directional{vector.Vec{0, 0, 0}, vector.Vec{0, 0, 0}, light.Source{vector.Vec{1, 1, 1}, 20.5, light.WATTS}}
	sky.Spd = light.InitSunlight(20)
	sky.InitPosition(1.5, common.DEG2RAD*45.0)
	return &sky
}

// Initialize sun positional reference coordinates in terms of polar coordinates
func (sky *Atmosphere) InitPosition(absDay float32, inclinationOffset float32) error {
	var err error
	theta := (1 / 2 * PI) * absDay
	pos, _ := polar.Vec2Sphere(vector.Vec{1.0, theta, PI - inclinationOffset})
	sky.Sun = pos
	sky.Dir, err = polar.Sphere2Vec(sky.Sun)
	sky.Dir = sky.Dir.Norm()
	sky.Light.SetDir(sky.Dir)
	return err
}

// Updates sun positional coordinates by rotating Azimuth & Polar coords by delta degrees
func (sky *Atmosphere) UpdatePosition(delta float32) error {
	var err error
	theta := common.DEG2RAD * delta
	//sky.Sun.AddAzimuth(theta)
	sky.Sun.AddPolar(theta)
	sky.Dir, err = polar.Sphere2Vec(sky.Sun)
	sky.Dir = sky.Dir.Norm()
	sky.Light.SetDir(sky.Dir)
	return err
}

// Creates texture from from computed atmosphere, non-clamping allows for HDR storage
func (sky *Atmosphere) CreateTexture(width int, height int, clamp bool, filename string) {
	fmt.Printf("Processing file:%s\n", filename)
	rgbs := sky.ComputeAtmosphere(width, height) //pre-normalized (non-hdr)
	ImageFromPixels(rgbs, width, height, clamp, 0xff, filename)
}

// Creates 6-textures from from computed atmosphere, non-clamping allows for HDR storage
// width and height are the per texture width height values of the image size
func (sky *Atmosphere) CreateEnvBox(width int, height int, clamp bool) {
	wd, _ := os.Getwd()
	fmt.Printf("Working Dir: %s\n", wd)
	if height != width || height <= 0 || width <= 0 {
		fmt.Printf("Computed region must be to a square texture. Parameters don't pass safeguard\n")
		return
	}
	if height%4 != 0 || width%4 != 0 {
		fmt.Printf("Computed region must be modulo 4\n")
		return
	}
	//Generate each side face
	region_width := int(width / 2)
	region_height := int(height / 2)
	x_corner := 0
	y_corner := 0

	//Faces loop
	for i := 0; i < 4; i++ {
		rgbs := sky.ComputeRegion(width, height, x_corner, y_corner, region_width, region_height)
		ImageFromPixels(rgbs, region_width, region_height, clamp, 0xff, "ENVBOX_"+strconv.FormatInt(int64(i), 10)+".png")

		//Move computed x,y region
		if x_corner < region_width {
			x_corner += region_width
		} else {
			if y_corner < region_height {
				y_corner += region_height
			}
		}
	}

	//Generate top and bottom dregion
	rgbs := sky.ComputeRegion(width, height, region_width/2, region_height/2, region_width, region_height)
	ImageFromPixels(rgbs, region_width, region_height, clamp, 0xff, "ENVBOX_4.png")
	ImageFromPixels(rgbs, region_width, region_height, clamp, 0x44, "ENVBOX_5.png")

}

// Utility function creates an image from a set of pixels
func ImageFromPixels(pixels []vector.Vec, width int, height int, clamp bool, alpha uint8, filename string) {
	corner := image.Point{0, 0}
	bottom := image.Point{width, height}
	img := image.NewRGBA(image.Rectangle{corner, bottom})
	index := 0
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			pixel := pixels[index]
			r_pixel := float64(pixel[0])
			g_pixel := float64(pixel[1])
			b_pixel := float64(pixel[2])
			if clamp {
				//pixel.Clamp(0, 1)
				g := 0.5
				b := -1.0

				if r_pixel < 1.31 {
					r_pixel = (math.Log(r_pixel + 1.0))
				} else {
					r_pixel = (1 / (1 + math.Exp(-r_pixel*g+b)))
				}

				if g_pixel < 1.31 {
					g_pixel = (math.Log(g_pixel + 1.0))
				} else {
					g_pixel = (1 / (1 + math.Exp(-g_pixel*g+b)))
				}

				if b_pixel < 1.31 {
					b_pixel = (math.Log(b_pixel + 1.0))
				} else {
					b_pixel = (1 / (1 + math.Exp(-b_pixel*g+b)))
				}

			}
			r := uint8(float32(r_pixel) * 255)
			g := uint8(float32(g_pixel) * 255)
			b := uint8(float32(b_pixel) * 255)
			img.Set(x, y, color.RGBA{r, g, b, alpha})
			index++
		}
	}

	//Encode as JPG
	options := jpeg.Options{80}
	f, _ := os.Create(filename)
	err := jpeg.Encode(f, img, &options)

	if err != nil {
		fmt.Printf("Error writing image to %s\n", filename)
	}
}

// Maps texel coordinates to spherical coordinate sampler values (-1,1) and stores
// resultant map in single texture.
func (sky *Atmosphere) ComputeAtmosphere(uSampleDomain int, vSampleDomain int) []vector.Vec {
	sizeT := uSampleDomain * vSampleDomain
	tex := make([]vector.Vec, sizeT+1)
	index := 0

	for x := 0; x < uSampleDomain; x++ {
		u := 2.0*(float64(x)+0.5)/(float64(uSampleDomain-1)) - 1.0
		for y := 0; y < vSampleDomain; y++ {
			v := 2.0*(float64(y)+0.5)/(float64(vSampleDomain-1)) - 1.0
			z2 := u*u + v*v
			phi := math.Atan2(v, u)
			theta := math.Acos(1 - z2)
			sample := vector.Vec{float32(math.Sin(theta) * math.Cos(phi)), float32(math.Sin(theta) * math.Sin(phi)), float32(math.Cos(theta))}
			if index < len(tex) {
				tex[index] = sky.VolumetricScatterRay(sample, vector.Vec{0, 1, 0})
			}
			index++
		}
	}
	return tex
}

// Maps texel coordinates to spherical coordinate sampler values (-1,1) and stores
func (sky *Atmosphere) ComputeRegion(uSampleDomain int, vSampleDomain int, x_corner int, y_corner int, width int, height int) []vector.Vec {
	sizeT := width * height
	tex := make([]vector.Vec, sizeT+1)
	index := 0
	for x := x_corner; x < x_corner+width; x++ {
		u := (2.0*(float64(x)+0.5)/(float64(uSampleDomain-1)) - 1.0)
		for y := y_corner; y < y_corner+height; y++ {
			v := (2.0*(float64(y)+0.5)/(float64(vSampleDomain-1)) - 1.0)
			z2 := u*u + v*v
			phi := math.Atan2(v, u)
			theta := math.Acos(1 - z2)
			sample := vector.Vec{float32(math.Sin(theta) * math.Cos(phi)), float32(math.Sin(theta) * math.Sin(phi)), float32(math.Cos(theta))}
			if index < len(tex) {
				tex[index] = sky.VolumetricScatterRay(sample, vector.Vec{0, 1, 0})
			}
			index++
		}
	}
	return tex
}

// Given a sampling vector and a viewing direction calculate RGB stimulus return
// Based on the Attenuation/Mie Phase Scatter/RayleighScatter Terms
func (sky *Atmosphere) VolumetricScatterRay(sample vector.Vec, view vector.Vec) vector.Vec {

	if sample[2] < 0.0 {
		return vector.Vec{0, 0, 0}
	}

	//Declare volumetric scatter ray vars
	intersects := polar.RaySphereIntersect(sample, sky.Earth.GetPosition(), sky.Earth.GreaterSphere)
	viewRay := vector.Scale(sample, polar.Priority(intersects))
	viewRayMag := viewRay.Mag()

	rgb := vector.Vec{0, 0, 0} //Pixel Output

	betaR := vector.Vec{0.0000058, .0000135, 0.0000331}
	betaM := vector.Vec{0.00210, 0.0021, 0.0021}
	sumR := vector.Vec{0, 0, 0} //rayleigh
	sumM := vector.Vec{0, 0, 0} //mie

	//Compute rayleight/mie coefficients from sample vector and sun direction
	u := vector.Dot(sample, sky.Dir)

	mu := float64(u)
	phaseR := float32(3.0 / (16.0 * PI) * (1.0 + mu*mu))
	g := 0.76
	phaseM := float32(3.0 / (8.0 * PI) * ((1.0 - g*g) * (1.0 + mu*mu)))
	phaseM = phaseM / float32((2+g*g)*math.Pow((1+g*g-2*g*mu), 1.1))
	var opticalDepthR, opticalDepthM float32
	var vmag0, vmag1, vds float32
	var lfactor = float32(1.0)

	//Aysmptotic Scaling
	scale_factor := float32(7.0)
	if sky.Dir[2] < 1.0 && sky.Dir[2] >= 0 {
		z := math.Abs(float64(sky.Dir[2]))
		lfactor = 1 / (float32(z))
		if lfactor > scale_factor {
			lfactor = scale_factor
		}
	} else if sky.Dir[2] <= 0 {
		lfactor = scale_factor
	}

	//Rayleigh Scatter Computation
	sampleStep := float32(1.0/RAYLEIGH_SAMPLES) * viewRayMag
	for i := 1; i <= RAYLEIGH_SAMPLES && len(intersects) > 0; i++ {

		//Generate Sample Rays along sample view ray path- assume sample ray is normalized
		w := float32(1.0)
		sampleScale := sampler.Ease(float32(i)*sampleStep, w)

		viewSample := vector.Scale(viewRay, sampleScale)
		depth := sky.Earth.GetSampleDepth(viewSample)

		//Compute the view ray ds parameters
		vmag1 = viewRayMag * sampleScale
		vds = vmag1 - vmag0
		vmag0 = vmag1

		//Get optical depth
		hr := float32(math.Exp(float64(-depth/HR))) * vds
		hm := float32(math.Exp(float64(-depth/HM))) * vds
		opticalDepthR += hr
		opticalDepthM += hm

		//Constructs Light Path Rays from Viewer Sample Positions and Calculates
		viewSampleOrigin := vector.Add(viewSample, sky.Earth.GetPosition())
		lightIntersects := polar.RaySphereIntersect(vector.Scale(sky.Dir, -1.0), viewSampleOrigin, sky.Earth.GreaterSphere)
		viRay := viewSampleOrigin.Sub(sky.Dir)
		if len(lightIntersects) == 0 {
			return vector.Vec{0, 0, 0}
		}

		//Light Path Transmittance + Attenutation
		lightRay := vector.Scale(viRay, polar.Priority(lightIntersects))
		lightRayMag := lightRay.Mag()
		lightPathSampleStep := float32(1.0/LIGHT_PATH_SAMPLES) * lightRayMag

		//Light Transmittance
		ds := float32(0.0)   //differential magnitude
		mag0 := float32(0.0) //for calculating differential
		mag1 := float32(0.0) //for calculating differntial

		var opticalDepthLightM, opticalDepthLightR float32

		//Compute light path to sample position - and light path segment length
		for j := 0; j < LIGHT_PATH_SAMPLES; j++ {
			pathScale := sampler.Ease(lightPathSampleStep*float32(j), w)
			lightPath := vector.Scale(lightRay, pathScale)
			mag1 = lightPath.Mag()
			ds = mag1 - mag0
			mag0 = mag1
			lightPathSamplePosition := vector.Add(viewSample, lightPath)
			lightPathDepth := sky.Earth.GetSampleDepth(lightPathSamplePosition)

			/*	if lightPathDepth < 0 {
					break
				}
			*/

			//Accumlate Light Path Transmittance
			opticalDepthLightR += float32(math.Exp(float64(-lightPathDepth/HR))) * ds
			opticalDepthLightM += float32(math.Exp(float64(-lightPathDepth/HM))) * ds
		}
		//Compute Contributions and Accumulate (Add constant to the rayleigh scale)
		tau := betaR.Scale(lfactor * (opticalDepthR + opticalDepthLightR)).Add(betaM.Scale(1.25).Scale(opticalDepthM + opticalDepthLightM))
		attenuation := vector.Vec{float32(math.Exp(float64(-tau[0]))), float32(math.Exp(float64(-tau[1]))), float32(math.Exp(float64(-tau[2])))}
		sumR = sumR.Add(attenuation.Scale(hr))
		sumM = sumM.Add(attenuation.Scale(hm))
	}
	rayliegh := sumR.Mul(betaR).Scale(phaseR)
	mie := sumM.Mul(betaM).Scale(phaseM)
	rgb = rayliegh.Add(mie).Scale(sky.Light.Lx().Flux).Mul(sky.Light.Lx().RGB)

	return rgb
}
