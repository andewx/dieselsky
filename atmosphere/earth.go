package atmosphere

import (
	"github.com/andewx/dieselfluid/math/common"
	"github.com/andewx/dieselfluid/math/polar"
	"github.com/andewx/dieselfluid/math/vector"
)

const (
	EARTH_RAD = 6370
)

//Earth Coordinates and Greater Earth will not be rotated for simplicity all
//Polar transformations are added to sun Polar coordinates as negative transforms
type EarthCoords struct {
	Latitude         float32     //Decimal Lat
	Longitude        float32     //Decimal Long
	PolarCoord       polar.Polar //Polar Axis offset
	StandardMeridian float32     //Longitude Standard Meridian For Local Time Offsets
	DomainOffset     [2]float32  //Domain offsets for polar sampling of sky depths
	GreaterSphere    polar.Polar //Atmospheric Polar Parameters Polar Parameters
}

//Declare New Sun Environment with Standard Merdian time set for NYC - Sky Functions handle sun rotation
func NewEarth(lat float32, long float32) *EarthCoords {
	myEarth := EarthCoords{}
	myEarth.Latitude = lat * common.DEG2RAD
	myEarth.Longitude = long * common.DEG2RAD
	myEarth.PolarCoord = polar.NewPolar(EARTH_RAD)
	myEarth.GreaterSphere = polar.NewPolar(EARTH_RAD + (HR))
	myEarth.getPolarSamplerDomain()

	return &myEarth
}

func (earth *EarthCoords) GetRadius() float32 {
	return earth.PolarCoord.Sphere[0]
}

func (earth *EarthCoords) GetPosition() vector.Vec {
	a, _ := polar.Sphere2Vec(earth.PolarCoord)
	return a
}

//Takes clamped [U,V] polar coordinates from [-1,1.0] and returns the ray depth
//Returns vector with magnitude to fixed point in sky in valid coordinates
func (earth *EarthCoords) GetSample(uv [2]float32) vector.Vec {
	uv[0] = common.Clamp1f(uv[0], -1.0, 1.0)
	uv[1] = common.Clamp1f(uv[1], -1.0, 1.0)
	atmosphereCoords := polar.NewPolar(EARTH_RAD + HR)
	atmosphereCoords.AddAzimuth(earth.PolarCoord.Sphere[0] + uv[0]*earth.DomainOffset[0])
	atmosphereCoords.AddPolar(earth.PolarCoord.Sphere[1] + uv[1]*earth.DomainOffset[1])
	rE_Vec, _ := polar.Sphere2Vec(earth.PolarCoord)
	rSK_Vec, _ := polar.Sphere2Vec(atmosphereCoords)
	return vector.Sub(rE_Vec, rSK_Vec)
}

//Approximation of sample depth based on Z coordinate
func (earth *EarthCoords) GetSampleDepth(sample vector.Vec) float32 {
	return sample[2]
}

//Gets when rotated earth tangent vector and tangent 2 atomospheric perion vectors are parallel
func (earth *EarthCoords) getPolarSamplerDomain() [2]float32 {
	samplerDomain := [2]float32{PI / 2, PI / 2}
	earth.DomainOffset = samplerDomain
	return samplerDomain
}
