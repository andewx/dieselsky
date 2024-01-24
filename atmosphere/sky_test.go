package atmosphere

import (
	"strconv"
	"testing"
)

func TestPNGCreate(t *testing.T) {
	mSky := NewAtmosphere(45.0, 0.0)
	base := "sky_"
	mSky.UpdatePosition((365 / 48) * 36)
	for i := 46; i < 47; i++ {
		filename := base + strconv.FormatInt(int64(i), 10) + ".jpg"
		mSky.UpdatePosition(365 / 48)
		mSky.CreateTexture(4096, 4096, true, filename)
	}
}

/*
func TestEnvBox(t *testing.T) {
	mSky := NewAtmosphere(45.0, 0.0)
	mSky.UpdatePosition(1.5)
	mSky.CreateEnvBox(128, 128, true)
}
*/
