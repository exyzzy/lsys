package lsys

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"strings"

	"github.com/exyzzy/lsys/drawing"
)

// LSys - axiom: beginning string, rules: rewrite rules, level: number of rewrite iterations
func LSys(axiom string, rules map[string]string, level int) (result string, err error) {
	s := axiom
	for i := 0; i < level; i++ {
		ns := ""
		for _, v := range s {
			switch v {
			case '-', '+', '[', ']':
				ns = ns + string(v)
			case ' ':
				//ignore space
			default:
				r, ok := rules[string(v)]
				if !ok {
					err = errors.New("no rule for: " + string(v))
					return
				}
				ns = ns + r
			}
		}
		s = ns
	}
	result = s
	return
}

type StackItem struct {
	Point drawing.FPoint
	Theta float64
}

// DrawLsys - drw: map point paths to draw into, lSys: the complete lsys string to draw, theta: the beginning angle (orientation), color: the RGBA color to use for all paths, onePath: force a single path for the entire fractal
func DrawLSys(drw *drawing.Drawing, lSys string, theta float64, angle float64, color color.RGBA, onePath bool) {
	var stack []StackItem
	p := drawing.FPoint{X: 0, Y: 0}
	drw.MoveTo(p, color)
	for _, v := range lSys {
		switch v {
		case 'F': // draw forward
			p = drawing.PointFromTheta(p, theta, 1.0)
			drw.LineTo(p)
		case '-': // turn left by angle
			theta -= angle
		case '+': // turn right by angle
			theta += angle
		case 'f': // move forward without drawing
			p = drawing.PointFromTheta(p, theta, 1.0)
			if !onePath {
				drw.MoveTo(p, color)
			}
		case '[': // push current location and direction onto stack
			var se = StackItem{Point: p, Theta: theta}
			stack = append(stack, se)
		case ']': // pop last location and direction from stack
			n := len(stack) - 1
			se := stack[n]
			p = se.Point
			if !onePath {
				drw.MoveTo(p, color)
			}
			theta = se.Theta
			stack = stack[:n]
		}
	}
}

func LsysByName(name string) (LFractal, error) {
	for _, f := range fractals {
		if f.Name == name {
			return f, nil
		}
	}
	return LFractal{}, errors.New("No fractal by name: " + name)
}

func RenderLsys(t io.Writer, fractal LFractal, color color.RGBA, rect image.Rectangle, vector bool) error {
	var drw drawing.Drawing
	s, err := LSys(fractal.Axiom, fractal.Rules, fractal.Levels)
	if err != nil {
		return err
	}
	DrawLSys(&drw, s, fractal.Theta, fractal.Angle, color, fractal.OnePath)
	var str string

	if vector {
		str, err = drw.RenderSvg(&rect, "images/"+fractal.Name+".svg")
	} else {
		str, err = drw.RenderPng(&rect, "images/"+fractal.Name+".png")
	}
	fmt.Fprintln(t, str)
	return err
}

func RenderAllLsys(t io.Writer) error {
	for _, f := range fractals {
		fmt.Fprintln(t, strings.TrimRight(strings.Replace(fmt.Sprintf("== %s == Angle: %v, Axiom: %v, Rules: %v", f.Name, f.Angle, f.Axiom, f.Rules), "map[", "", 1), "]"))
		var r = image.Rectangle{Min: image.Point{X: 0, Y: 0}, Max: image.Point{X: 1024, Y: 1024}}
		err := RenderLsys(t, f, drawing.ColorBLACK, r, true)
		if err != nil {
			return err
		}
		err = RenderLsys(t, f, drawing.ColorBLACK, r, false)
		if err != nil {
			return err
		}
	}
	return nil
}
