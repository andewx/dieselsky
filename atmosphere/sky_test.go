package atmosphere

import (
	"strconv"
	"testing"

	"github.com/andewx/dieselfluid/common"
)

func TestPNGCreate(t *testing.T) {
	mSky := NewAtmosphere(45.0, 0.0)
	base := "sky_"
	mSky.UpdatePosition((365 / 48) * 36)
	for i := 36; i < 48; i++ {
		filename := base + strconv.FormatInt(int64(i), 10) + ".jpg"
		mSky.UpdatePosition(365 / 48)
		mSky.CreateTexture(512, 512, true, common.ProjectRelativePath("etc/"+filename))
	}
}

/*
func TestEnvBox(t *testing.T) {
	mSky := NewAtmosphere(45.0, 0.0)
	mSky.UpdatePosition(1.5)
	mSky.CreateEnvBox(128, 128, true)
}
*/
