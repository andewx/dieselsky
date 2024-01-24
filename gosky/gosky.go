package main

import (
	"fmt"
	"os"
	"strconv"

	atm "github.com/andewx/dieselsky/atmosphere"
)

func main() {

	var width, height int
	var clamp bool
	var filename string
	var err error
	args := os.Args[1:]

	if len(args) != 4 {
		Usage()
		os.Exit(1)
	}

	width, err = strconv.Atoi(args[0])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	height, err = strconv.Atoi(args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	clamp, err = strconv.ParseBool(args[2])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	filename = args[3]

	mSky := atm.NewAtmosphere(45.0, 0.0)
	base := filename
	mSky.UpdatePosition((365 / 48) * 36)
	for i := 46; i < 47; i++ {
		filename := base + strconv.FormatInt(int64(i), 10) + ".jpg"
		mSky.UpdatePosition(365 / 48)
		mSky.CreateTexture(width, height, clamp, filename)
	}
}

func CreateTexture(width, height int, isHDR bool, filename string) {
	mSky := atm.NewAtmosphere(45.0, 0.0)
	mSky.UpdatePosition((365 / 48) * 36)
	mSky.UpdatePosition(365 / 48)
	mSky.CreateTexture(width, height, true, filename)
}

func Usage() {
	fmt.Printf("Usage: %s [options] <width> <height> <clampValues> <prefix>\n", os.Args[0])
}
