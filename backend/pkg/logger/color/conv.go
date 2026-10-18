package color

type RGB struct {
	Red   uint8
	Green uint8
	Blue  uint8
}

func NewRGB(r, g, b uint8) RGB { return RGB{r, g, b} }

type HSL struct {
	Hue        uint
	Saturation uint8
	Lightness  uint8
}

func NewHSL(h uint, s, l uint8) HSL {
	if h > 360 || s > 100 || l > 100 {
		panic("hsl: invalid input")
	}
	return HSL{
		h, s, l,
	}
}

func (h HSL) ToRGB() RGB {
	hue := float64(h.Hue % 360)
	sat := float64(h.Saturation) / 100
	light := float64(h.Lightness) / 100

	c := (1 - abs(2*light-1)) * sat
	x := c * (1 - abs(float64(int(hue)/60%2)-1))
	m := light - c/2

	var r, g, b float64

	switch {
	case hue < 60:
		r, g, b = c, x, 0
	case hue < 120:
		r, g, b = x, c, 0
	case hue < 180:
		r, g, b = 0, c, x
	case hue < 240:
		r, g, b = 0, x, c
	case hue < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	return RGB{
		uint8((r + m) * 255),
		uint8((g + m) * 255),
		uint8((b + m) * 255),
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
