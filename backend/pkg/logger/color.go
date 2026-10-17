package logger

import (
	"backend/pkg/logger/color"
	"backend/pkg/logger/utils"
	"bytes"
)

const colorStartStr = "\033[38;2;"
const colorEndStr = "\033[0m"

func writeColorStart(b *bytes.Buffer, c color.RGB) {
	b.WriteString(colorStartStr)
	utils.WriteInteger(b, c.Red)
	b.WriteRune(';')
	utils.WriteInteger(b, c.Green)
	b.WriteRune(';')
	utils.WriteInteger(b, c.Blue)
	b.WriteRune('m')
}

func writeColorEnd(b *bytes.Buffer) {
	b.WriteString(colorEndStr)
}
