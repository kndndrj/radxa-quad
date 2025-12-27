package main

import (
	"image"
	"log"
	"time"

	"github.com/warthog618/go-gpiocdev"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/devices/v3/ssd1306"
	"periph.io/x/devices/v3/ssd1306/image1bit"
	"periph.io/x/host/v3"
)

// Hello future me! This demo shows you how to display some text on the oled. The most important bit
// is the toggle of the reset line - without it, rpi with nixos doesn't work.
func demo() {
	reset, err := gpiocdev.RequestLine("gpiochip0", 23, gpiocdev.AsOutput(1))
	if err != nil {
		panic(err)
	}

	if err := reset.SetValue(0); err != nil {
		panic(err)
	}

	time.Sleep(10 * time.Millisecond)

	if err := reset.SetValue(1); err != nil {
		panic(err)
	}
	_ = reset.Close()

	// Load all the drivers:
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Open a handle to the first available I²C bus:
	bus, err := i2creg.Open("")
	if err != nil {
		log.Fatal(err)
	}

	// Open a handle to a ssd1306 connected on the I²C bus:
	dev, err := ssd1306.NewI2C(bus, &ssd1306.Opts{
		W:          128,
		H:          32,
		Rotated:    false,
		Sequential: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	img := image1bit.NewVerticalLSB(dev.Bounds())

	f := basicfont.Face7x13
	drawer := font.Drawer{
		Dst:  img,
		Src:  &image.Uniform{image1bit.On},
		Face: f,
		Dot:  fixed.P(0, 15), // 1st row.
	}
	drawer.DrawString("Hello")

	// Move to 2nd row.
	drawer.Dot.Y = fixed.I(30)
	drawer.Dot.X = fixed.I(0)

	drawer.DrawString("   there!")

	if err := dev.Draw(dev.Bounds(), img, image.Point{}); err != nil {
		log.Fatal("draw:", err)
	}
}
