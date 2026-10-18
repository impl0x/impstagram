package color

import "fmt"

var White = [10]RGB{
	NewHSL(0, 0, 100).ToRGB(), // #ffffff
	NewHSL(0, 0, 90).ToRGB(),  // ##e6e6e6
	NewHSL(0, 0, 80).ToRGB(),  // ##cccccc
	NewHSL(0, 0, 70).ToRGB(),  // ##b3b3b3
	NewHSL(0, 0, 60).ToRGB(),  // ##999999
	NewHSL(0, 0, 50).ToRGB(),
	NewHSL(0, 0, 40).ToRGB(),
	NewHSL(0, 0, 30).ToRGB(),
	NewHSL(0, 0, 20).ToRGB(),
	NewHSL(0, 0, 10).ToRGB(),
}

var Red = [10]RGB{
	NewHSL(0, 100, 100).ToRGB(),
	NewHSL(0, 100, 90).ToRGB(),
	NewHSL(0, 100, 80).ToRGB(),
	NewHSL(0, 100, 70).ToRGB(),
	NewHSL(0, 100, 60).ToRGB(),
	NewHSL(0, 100, 50).ToRGB(),
	NewHSL(0, 100, 40).ToRGB(),
	NewHSL(0, 100, 30).ToRGB(),
	NewHSL(0, 100, 20).ToRGB(),
	NewHSL(0, 100, 10).ToRGB(),
}

var Green = [10]RGB{
	NewHSL(120, 0, 100).ToRGB(),
	NewHSL(120, 0, 90).ToRGB(),
	NewHSL(120, 0, 80).ToRGB(),
	NewHSL(120, 0, 70).ToRGB(),
	NewHSL(120, 0, 60).ToRGB(),
	NewHSL(120, 0, 50).ToRGB(),
	NewHSL(120, 0, 40).ToRGB(),
	NewHSL(120, 0, 30).ToRGB(),
	NewHSL(120, 0, 20).ToRGB(),
	NewHSL(120, 0, 10).ToRGB(),
}

var Blue = [10]RGB{
	NewHSL(240, 0, 100).ToRGB(),
	NewHSL(240, 0, 90).ToRGB(),
	NewHSL(240, 0, 80).ToRGB(),
	NewHSL(240, 0, 70).ToRGB(),
	NewHSL(240, 0, 60).ToRGB(),
	NewHSL(240, 0, 50).ToRGB(),
	NewHSL(240, 0, 40).ToRGB(),
	NewHSL(240, 0, 30).ToRGB(),
	NewHSL(240, 0, 20).ToRGB(),
	NewHSL(240, 0, 10).ToRGB(),
}

func ColorString(c RGB, s string) string {
	return fmt.Sprintf(
		"\033[38;2;%v;%v;%vm%s\033[0m",
		c.Red, c.Green, c.Blue,
		s,
	)
}
