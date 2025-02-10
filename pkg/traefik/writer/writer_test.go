package writer_test

import (
	"117503445/traefik-provider-frp/pkg/traefik/writer"
	_ "embed"
	"testing"
)

//go:embed test.txt
var text string

func TestXxx(t *testing.T) {
	c := writer.GetHttpContent(text)
	println(c)
}
