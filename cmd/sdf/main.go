// SDF Generator.
// Algorithm adapted from http://www.codersnotes.com/notes/signed-distance-fields/
package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"image"
	"image/color"
	"image/png"
)

func loadPNG(filepath string) (image.Image, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("file \"%s\" could not be opened: %w", filepath, err)
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("file \"%s\" could not be decoded: %w", filepath, err)
	}

	return img, nil
}

func savePNG(img image.Image, filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("file \"%s\" could not be created: %w", filepath, err)
	}
	defer file.Close()

	buf := bufio.NewWriter(file)
	if err := png.Encode(buf, img); err != nil {
		return fmt.Errorf("file \"%s\" could not be encoded: %w", filepath, err)
	}
	if err := buf.Flush(); err != nil {
		return fmt.Errorf("file \"%s\" could not be flushed: %w", filepath, err)
	}
	return nil
}

// Point is a single point offset
type Point struct {
	dx, dy int
}

// DistSq is the squared distance for the point
func (p *Point) DistSq() int {
	return p.dx*p.dx + p.dy*p.dy
}

// Grid is the intermediate distance field
type Grid struct {
	width, height int
	pts           []Point
}

func (g *Grid) compare(p Point, x, y, offsetx, offsety int) Point {
	if x+offsetx >= 0 && x+offsetx < g.width && y+offsety >= 0 && y+offsety < g.height {
		other := g.pts[((y+offsety)*g.width)+x+offsetx]
		other.dx += offsetx
		other.dy += offsety
		if other.DistSq() < p.DistSq() {
			return other
		}
		return p
	}
	return Point{dx: 9999, dy: 9999}
}

// Generate generates the SDF for the grid
func (g *Grid) Generate() {
	for y := 0; y < g.height; y++ {
		i := y * g.width
		for x := 0; x < g.width; x++ {
			p := g.pts[i+x]
			p = g.compare(p, x, y, -1, 0)
			p = g.compare(p, x, y, 0, -1)
			p = g.compare(p, x, y, -1, -1)
			p = g.compare(p, x, y, 1, -1)
			g.pts[i+x] = p
		}
		for x := g.width - 1; x >= 0; x-- {
			p := g.pts[i+x]
			p = g.compare(p, x, y, 1, 0)
			g.pts[i+x] = p
		}
	}

	for y := g.height - 1; y >= 0; y-- {
		i := y * g.width
		for x := g.width - 1; x >= 0; x-- {
			p := g.pts[i+x]
			p = g.compare(p, x, y, 1, 0)
			p = g.compare(p, x, y, 0, 1)
			p = g.compare(p, x, y, -1, 1)
			p = g.compare(p, x, y, 1, 1)
			g.pts[i+x] = p
		}
		for x := 0; x < g.width; x++ {
			p := g.pts[i+x]
			p = g.compare(p, x, y, -1, 0)
			g.pts[i+x] = p
		}
	}
}

func calcSDF(src image.Image) (image.Image, error) {
	srcWidth := src.Bounds().Dx()
	srcHeight := src.Bounds().Dy()
	grid1 := Grid{
		width:  srcWidth,
		height: srcHeight,
		pts:    make([]Point, srcWidth*srcHeight),
	}
	grid2 := Grid{
		width:  srcWidth,
		height: srcHeight,
		pts:    make([]Point, srcWidth*srcHeight),
	}

	for y := 0; y < srcHeight; y++ {
		i := y * srcWidth
		for x := 0; x < srcWidth; x++ {
			c := src.At(x, y)
			if r, _, _, _ := c.RGBA(); r < 128 {
				grid1.pts[i+x].dx = 0
				grid1.pts[i+x].dy = 0
				grid2.pts[i+x].dx = 9999
				grid2.pts[i+x].dy = 9999
			} else {
				grid1.pts[i+x].dx = 9999
				grid1.pts[i+x].dy = 9999
				grid2.pts[i+x].dx = 0
				grid2.pts[i+x].dy = 0
			}
		}
	}

	grid1.Generate()
	grid2.Generate()

	dest := image.NewGray(image.Rect(0, 0, srcWidth, srcHeight))
	for y := 0; y < srcHeight; y++ {
		i := y * srcWidth
		for x := 0; x < srcWidth; x++ {
			dist1 := int(math.Sqrt(float64(grid1.pts[i+x].DistSq())))
			dist2 := int(math.Sqrt(float64(grid2.pts[i+x].DistSq())))
			dist := dist1 - dist2

			c := (dist*3 + 128)
			if c < 0 {
				c = 0
			}
			if c > 255 {
				c = 255
			}
			dest.SetGray(x, y, color.Gray{Y: uint8(c)})
		}
	}
	return dest, nil
}

func main() {
	var inFile, outFile string
	flag.StringVar(&inFile, "in", "", "the png to calculate the sdf for")
	flag.StringVar(&outFile, "out", "", "the png to output sdf to")
	flag.Parse()

	if inFile == "" || outFile == "" {
		flag.PrintDefaults()
		os.Exit(-1)
	}

	if filepath.Ext(inFile) != ".png" || filepath.Ext(outFile) != ".png" {
		fmt.Println("in/out file should be png")
		os.Exit(-1)
	}

	srcPNG, err := loadPNG(inFile)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	outPNG, err := calcSDF(srcPNG)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	if err := savePNG(outPNG, outFile); err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
}
