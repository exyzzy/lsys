// drawing is a set of 2d line drawing functions for working with lines as an array of paths
// where a path is assumed to be a set of connected Float64 points with the same color

package drawing

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"

	"github.com/StephaneBunel/bresenham"
)

// Color* is a set of standard image colors for rendering, used also in paths
var ColorRED = color.RGBA{255, 0, 0, 255}
var ColorGREEN = color.RGBA{0, 255, 0, 255}
var ColorBLUE = color.RGBA{0, 0, 255, 255}
var ColorWHITE = color.RGBA{255, 255, 255, 255}
var ColorBLACK = color.RGBA{0, 0, 0, 255}

// FPoint is a floating point 2d point
type FPoint struct {
	X float64
	Y float64
}

// Path is a single path of connected points with RGBA color
type Path struct {
	Points []FPoint
	Color  color.RGBA
}

// Drawing is an array of paths
type Drawing struct {
	Paths []Path
}

// FRect is an FPoint min/max rectangle
type FRect struct {
	Min FPoint
	Max FPoint
}

// RenderPng renders a drawing centered as a png with given filepath and given size (rect)
// If rect is nil size defaults to 2kx2k
func (drawing *Drawing) RenderPng(rect *image.Rectangle, filePath string) (string, error) {
	var img *image.RGBA
	if rect == nil {
		rect = &image.Rectangle{}
		*rect = image.Rect(0, 0, 2000, 2000)
	}
	img = image.NewRGBA(*rect)
	ib := RectBounds(*rect)
	// ib := ImageBounds(img)
	drawing.Flip(true)
	drawing.CenterWithMargin(ib, FPoint{X: 0.1, Y: 0.1}) //add a 10% of size margin
	drawing.DrawToImage(img)
	// flipImg := ImageFlipV(img)
	toimg, err := os.Create(filePath)
	if err != nil {
		return filePath, err
	}
	defer toimg.Close()
	err = png.Encode(toimg, img)
	if err != nil {
		return filePath, err
	}
	return fmt.Sprintf("%s: %v paths", filePath, len(drawing.Paths)), nil
}

// RenderSvg renders a drawing centered as a svg with given filepath and given size (rect)
// If rect is nil size defaults to 2kx2k
func (drawing *Drawing) RenderSvg(rect *image.Rectangle, filePath string) (string, error) {

	if rect == nil {
		rect = &image.Rectangle{}
		*rect = image.Rect(0, 0, 2000, 2000)
	}
	ib := RectBounds(*rect)
	drawing.Flip(true)
	drawing.CenterWithMargin(ib, FPoint{X: 0.1, Y: 0.1}) //add a 10% of size margin
	fSvg, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer fSvg.Close()
	drawing.DrawToSvg(fSvg, *rect)
	if err != nil {
		return filePath, err
	}
	return fmt.Sprintf("%s: %v paths", filePath, len(drawing.Paths)), nil
}

// MoveTo starts a new Path set with the Path color, and moves to the first point
func (drawing *Drawing) MoveTo(point FPoint, color color.RGBA) {
	// fmt.Println("   MoveTo: ", point)
	var path Path
	path.Points = append(path.Points, point)
	path.Color = color
	drawing.Paths = append(drawing.Paths, path)
}

// LineTo adds a new point in the current Path with the current Path color
func (drawing *Drawing) LineTo(point FPoint) {
	// fmt.Println("   LineTo: ", point)
	last := len(drawing.Paths) - 1
	drawing.Paths[last].Points = append(drawing.Paths[last].Points, point)
}

// RectBounds returns the Floating point FRect of an image rectangle.
// effectively converting image coordinates to drawing coordinates
func RectBounds(rect image.Rectangle) (pb FRect) {
	// fmt.Println(">>RectBounds: ")
	pb.Min.X = float64(rect.Min.X)
	pb.Max.X = float64(rect.Max.X - 1)
	pb.Min.Y = float64(rect.Min.Y)
	pb.Max.Y = float64(rect.Max.Y - 1)
	// fmt.Println("  ...", pb)
	return
}

// Path or Point function type for Traverse
type fn func(s ...interface{})

// Traverse is a general purpose traverse that applies functions to each Path and/or Point in a drawing
// fpa: Function to apply at each Path, s[0] is path
// fpt: Function to apply at each Point in a Path, s[0] is point
// s[1:] all params as interfaces
// either fn can be nil if nothing to do at path or point
func (drawing *Drawing) Traverse(fpa fn, fpt fn, s ...interface{}) {
	// fmt.Println("Traverse")
	// fmt.Println("Paths: ", len(drawing.Paths))
	for path := 0; path < len(drawing.Paths); path++ {
		if fpa != nil {
			// fmt.Println("Path: ", path)
			// fmt.Println("Points: ", len(drawing.Paths[path].Points))
			//prepend path as interface s[0]: path, so that fpa can access it
			var p interface{} = &(drawing.Paths[path])
			spa := append([]interface{}{p}, s...)
			fpa(spa...)
		}
		for pt := 0; pt < len(drawing.Paths[path].Points); pt++ {
			if fpt != nil {
				// fmt.Println("Point: ", pt)
				//prepend point as interface s[0]: point, so that fpt can access it
				var q interface{} = &(drawing.Paths[path].Points[pt])
				spt := append([]interface{}{q}, s...)
				fpt(spt...)
			}
		}
	}
}

// == Use Traverse to extract the boundary of all points in a drawing

// Bounds Point function
func BoundsPt(s ...interface{}) {
	// fmt.Println("BoundsPt: ", *(s[0].(*FPoint)))
	(*(s[1].(*FRect))).Min.X = math.Min((*(s[1].(*FRect))).Min.X, (*(s[0].(*FPoint))).X)
	(*(s[1].(*FRect))).Max.X = math.Max((*(s[1].(*FRect))).Max.X, (*(s[0].(*FPoint))).X)
	(*(s[1].(*FRect))).Min.Y = math.Min((*(s[1].(*FRect))).Min.Y, (*(s[0].(*FPoint))).Y)
	(*(s[1].(*FRect))).Max.Y = math.Max((*(s[1].(*FRect))).Max.Y, (*(s[0].(*FPoint))).Y)
}

// Bounds extracts bounds from all points in a drawing
func (drawing *Drawing) Bounds() (pb FRect) {
	// fmt.Println(">>Bounds: ")
	pb.Min.X = drawing.Paths[0].Points[0].X
	pb.Max.X = drawing.Paths[0].Points[0].X
	pb.Min.Y = drawing.Paths[0].Points[0].Y
	pb.Max.Y = drawing.Paths[0].Points[0].Y
	drawing.Traverse(nil, BoundsPt, &pb)
	// fmt.Println("  ...", pb)
	return
}

// == Use Traverse to translate all points in a drawing

// Translate Point function
func TranslatePt(s ...interface{}) {
	// fmt.Println("TranslatePt: ", *(s[0].(*FPoint)))
	(*(s[0].(*FPoint))).X = (*(s[0].(*FPoint))).X + (*(s[1].(*FPoint))).X
	(*(s[0].(*FPoint))).Y = (*(s[0].(*FPoint))).Y + (*(s[1].(*FPoint))).Y
}

// Translate all points by delta x and y
func (drawing *Drawing) Translate(delta FPoint) {
	// fmt.Println(">>Translate: ", delta)
	drawing.Traverse(nil, TranslatePt, &delta)
}

// == Use Traverse to scale all points in a drawing

// Scale Point function
func ScalePt(s ...interface{}) {
	// fmt.Println("ScalePt: ", *(s[0].(*FPoint)))
	(*(s[0].(*FPoint))).X = (*(s[0].(*FPoint))).X * *(s[1].(*float64))
	(*(s[0].(*FPoint))).Y = (*(s[0].(*FPoint))).Y * *(s[1].(*float64))
}

// Scale all points by scalar
func (drawing *Drawing) Scale(scalar float64) {
	// fmt.Println(">>Scale: ", scalar)
	drawing.Traverse(nil, ScalePt, &scalar)
}

// == Use Traverse to rotate all points in a drawing about the origin

// Rotate Point function
func RotatePt(s ...interface{}) {
	// fmt.Println("RotatePt: ", *(s[0].(*FPoint)))
	rotx := ((*(s[0].(*FPoint))).X * *(s[1].(*float64))) - ((*(s[0].(*FPoint))).Y * *(s[2].(*float64)))
	roty := ((*(s[0].(*FPoint))).X * *(s[2].(*float64))) + ((*(s[0].(*FPoint))).Y * *(s[1].(*float64)))
	// fmt.Println("  Cos, sin: ", *(s[1].(*float64)), *(s[2].(*float64)))
	// fmt.Println("  Rotx: ", rotx)
	// fmt.Println("  Roty: ", roty)
	(*(s[0].(*FPoint))).X = rotx
	(*(s[0].(*FPoint))).Y = roty
}

// Rotate all points by angle
func (drawing *Drawing) Rotate(angle float64) {
	// fmt.Println(">>Rotate: ", angle)
	cos := math.Cos(ToRadians(angle))
	sin := math.Sin(ToRadians(angle))
	// fmt.Println("  Cos, sin: ", cos, sin)
	drawing.Traverse(nil, RotatePt, &cos, &sin)
}

// == Use Traverse to flip all points in a drawing either Horizontally or Vertically

// Vertical Flip Point function
func VFlipPt(s ...interface{}) {
	(*(s[0].(*FPoint))).Y = (s[1].(*FRect)).Max.Y - (*(s[0].(*FPoint))).Y + (s[1].(*FRect)).Min.Y
}

// Horizontal Flip Point function
func HFlipPt(s ...interface{}) {
	(*(s[0].(*FPoint))).X = (s[1].(*FRect)).Max.X - (*(s[0].(*FPoint))).X + (s[1].(*FRect)).Min.X
}

// Flip all points on vertical axis if vert == true, else horizontal
func (drawing *Drawing) Flip(vert bool) {
	// fmt.Println(">>Rotate: ", angle)
	db := drawing.Bounds()
	if vert {
		drawing.Traverse(nil, VFlipPt, &db)
	} else {
		drawing.Traverse(nil, HFlipPt, &db)
	}
}

// CenterWithMargin will scale and center a drawing to tb the tb FRect with the margin tm.x, tm.y in % of total bounds. eg 0.1 = 10% boundary on each edge. So in 11x17 this is 1.1" margin in Y and 1.7" margin in X
func (drawing *Drawing) CenterWithMargin(tb FRect, tm FPoint) {
	// fmt.Println(">>CenterWithMargin: ", tb)
	db := drawing.Bounds()
	// fmt.Println("Drawing Bounds: ", db)
	scale := math.Min(((tb.Max.X-tb.Min.X)-(tm.X*2*(tb.Max.X-tb.Min.X)))/(db.Max.X-db.Min.X), ((tb.Max.Y-tb.Min.Y)-(tm.Y*2*(tb.Max.Y-tb.Min.Y)))/(db.Max.Y-db.Min.Y))
	// fmt.Println("Points before scale: ", drawing.Paths)
	// fmt.Println("Scale: ", scale)
	drawing.Scale(scale)
	// fmt.Println("Points after scale: ", drawing.Paths)
	db.Min.X = db.Min.X * scale
	db.Max.X = db.Max.X * scale
	db.Min.Y = db.Min.Y * scale
	db.Max.Y = db.Max.Y * scale
	var delta FPoint
	//just delta.X = tb.Max.X - db-Max.X (and Y)?
	delta.X = ((tb.Max.X + tb.Min.X) / 2.0) - ((db.Max.X + db.Min.X) / 2.0)
	delta.Y = ((tb.Max.Y + tb.Min.Y) / 2.0) - ((db.Max.Y + db.Min.Y) / 2.0)
	// fmt.Println("Delta: ", delta)
	drawing.Translate(delta)
	// fmt.Println("Points after Translate: ", drawing.Paths)
}

// == Use Traverse to render a drawing to a png

// DrawToImage Path function
func DrawToImagePa(s ...interface{}) {
	// fmt.Println("DrawToImgPa: ", *(s[2].(*color.RGBA)))
	pa := *(s[0].(*Path))            // get current path
	*(s[2].(*color.RGBA)) = pa.Color // pass color to point function
	*(s[3].(*FPoint)) = pa.Points[0] // pass fromPt to point function
}

// DrawToImage Point function
func DrawToImagePt(s ...interface{}) {
	// fmt.Println("DrawToImgPt: ", *(s[0].(*FPoint)))
	p0 := *(s[3].(*FPoint)) // po is the last point (fromPt)
	p1 := *(s[0].(*FPoint)) // p1 is the current point
	// db := *(s[4].(*FRect))
	// bresenham.Bresenham(*(s[1].(*draw.Image)), int(p0.X), int(db.Max.Y-p0.Y+db.Min.Y), int(p1.X), int(db.Max.Y-p1.Y+db.Min.Y), *(s[2].(*color.RGBA)))
	bresenham.Bresenham(*(s[1].(*draw.Image)), int(p0.X), int(p0.Y), int(p1.X), int(p1.Y), *(s[2].(*color.RGBA)))
	*(s[3].(*FPoint)) = p1 // reset the fromPt
}

// DrawToImage draws drawing to image
func (drawing *Drawing) DrawToImage(img draw.Image) {
	// fmt.Println(">>DrawToImage")
	db := drawing.Bounds()
	var color color.RGBA
	var fromPt FPoint
	// We pass in the Path function and Point function, and either pass in or save room for other vars in the s args
	drawing.Traverse(DrawToImagePa, DrawToImagePt, &img, &color, &fromPt, &db)
}

// ImageFlipV flips an image vertically
func ImageFlipV(img image.Image) *image.RGBA {
	bnds := img.Bounds()
	var newImg = image.NewRGBA(bnds)
	for j := bnds.Min.Y; j < bnds.Max.Y; j++ {
		for i := bnds.Min.X; i < bnds.Max.X; i++ {
			c := img.At(i, j)
			newImg.Set(i, bnds.Max.Y-j-1, c)
		}
	}
	return newImg
}

// == Use Traverse to render a drawing to an svg

// DrawToSvg Path function
// s[0] current path
// s[1] fSvg
func DrawToSvgPa(s ...interface{}) {
	pa := *(s[0].(*Path))
	p := pa.Points[0]

	fSvg := *(s[1].(*os.File))
	str := fmt.Sprintf("\" />\n<polyline fill=\"none\" stroke=\"#%02x%02x%02x%02x\" stroke-width=\"2\" points=\"%v,%v", pa.Color.R, pa.Color.G, pa.Color.B, pa.Color.A, p.X, p.Y)
	_, err := fSvg.WriteString(str)
	if err != nil {
		panic(err)
	}
}

// DrawToSvg Point function
// s[0] current point
// s[1] fSvg
func DrawToSvgPt(s ...interface{}) {
	p := *(s[0].(*FPoint))
	fSvg := *(s[1].(*os.File))
	str := fmt.Sprintf(" %v,%v", p.X, p.Y)
	_, err := fSvg.WriteString(str)
	if err != nil {
		panic(err)
	}
}

// DrawToSvg draws drawing to svg file
func (drawing *Drawing) DrawToSvg(fSvg *os.File, rect image.Rectangle) {
	str := fmt.Sprintf("<?xml version=\"1.0\" standalone=\"no\"?>\n<svg width=\"%d\" height=\"%d\"\nxmlns=\"http://www.w3.org/2000/svg\" version=\"1.1\">\n<rect x=\"1\" y=\"1\" width=\"%v\" height=\"%v\"\nfill=\"none\" stroke=\"black\" stroke-width=\"1", rect.Max.X, rect.Max.Y, rect.Max.X, rect.Max.Y)
	_, err := fSvg.WriteString(str)
	if err != nil {
		panic(err)
	}
	drawing.Traverse(DrawToSvgPa, DrawToSvgPt, fSvg)
	_, err = fSvg.WriteString("\" />\n</svg>")
	if err != nil {
		panic(err)
	}
}

// General functions

func ToRadians(degrees float64) float64 {
	return degrees * (math.Pi / 180.0)
}

func ToDegrees(radians float64) float64 {
	return radians * (180.0 / math.Pi)
}

// return length of p0 to p1
func Length(p0, p1 FPoint) (length float64) {
	return math.Sqrt((p1.X-p0.X)*(p1.X-p0.X) + (p1.Y-p0.Y)*(p1.Y-p0.Y))
}

// Return angle (degrees) from 2 points p0->p1
func ThetaFromPoint(p0, p1 FPoint) (theta float64) {
	if (p1.X - p0.X) == 0 {
		if p1.Y > p0.Y {
			theta = 90
		} else {
			theta = 270
		}
	} else {
		theta = ToDegrees(math.Atan((p1.Y - p0.Y) / (p1.X - p0.X)))
	}
	if p0.X > p1.X {
		theta = theta + 180
	}
	return
}

// return point from point, angle (degrees) and length p0->p1
func PointFromTheta(p0 FPoint, theta float64, length float64) (p1 FPoint) {
	p1.X = length*math.Cos(ToRadians(theta)) + p0.X
	p1.Y = length*math.Sin(ToRadians(theta)) + p0.Y
	return
}
