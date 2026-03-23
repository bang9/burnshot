package qr

import (
	"io"

	qrterminal "github.com/mdp/qrterminal/v3"
)

func Render(w io.Writer, url string) {
	qrterminal.GenerateWithConfig(url, qrterminal.Config{
		Level:     qrterminal.L,
		Writer:    w,
		BlackChar: qrterminal.BLACK,
		WhiteChar: qrterminal.WHITE,
		QuietZone: 1,
	})
}
