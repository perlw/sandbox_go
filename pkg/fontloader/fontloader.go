package fontloader

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"os"

	"github.com/golang/freetype"
	"golang.org/x/image/font"
)

type Glyph struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type Charset struct {
	Image *image.RGBA
}

func (c *Charset) Save(filepath string) error {
	outFile, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("could not create file \"%s\": %w", filepath, err)
	}
	defer outFile.Close()

	b := bufio.NewWriter(outFile)
	err = png.Encode(b, c.Image)
	if err != nil {
		return fmt.Errorf("could not png encode \"%s\": %w", filepath, err)
	}

	err = b.Flush()
	if err != nil {
		return fmt.Errorf("could not flush file \"%s\": %w", filepath, err)
	}
	return nil
}

func LoadTTF(filepath string) (*Charset, error) {
	fontBytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("could not read font \"%s\": %w", filepath, err)
	}
	ft, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return nil, fmt.Errorf("could not parse font \"%s\": %w", filepath, err)
	}
	size := 16
	fg, bg := image.White, image.Black
	rgba := image.NewGray(image.Rect(0, 0, 256, 256))
	draw.Draw(rgba, rgba.Bounds(), bg, image.ZP, draw.Src)
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(ft)
	c.SetFontSize(float64(size))
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)
	c.SetSrc(fg)
	c.SetHinting(font.HintingNone)

	gw := 8
	gh := 16

	var gi int
	var x, y int
	var gstart, gend rune = 32, 255
	glyphs := make([]Glyph, gend-gstart+1)
	for r := gstart; r <= gend; r++ {
		glyphs[gi].X = x
		glyphs[gi].Y = y - gh
		glyphs[gi].Width = gw
		glyphs[gi].Height = gh
		pt := freetype.Pt(x, y-2)
		c.DrawString(string(r), pt)

		if gi%32 == 0 {
			x = 0
			y += gh
		} else {
			x += gw
		}

		gi++
	}

	fontImg := image.NewRGBA(rgba.Bounds())
	for y := 0; y < rgba.Bounds().Dy(); y++ {
		for x := 0; x < rgba.Bounds().Dx(); x++ {
			curr := rgba.At(x, y).(color.Gray)
			fontImg.Set(x, y, color.RGBA{R: 255, G: 255, B: 255, A: curr.Y})
			if x == 0 || y == 0 {
				continue
			}

			prev := rgba.At(x-1, y-1).(color.Gray)
			if prev.Y >= 128 && curr.Y < 128 {
				fontImg.Set(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 255})
			}
		}
	}

	return &Charset{
		Image: fontImg,
	}, nil
}
