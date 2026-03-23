package qr

import (
	"io"

	qrterminal "github.com/mdp/qrterminal/v3"
)

func Render(w io.Writer, url string) {
	qrterminal.GenerateHalfBlock(url, qrterminal.L, w)
}
