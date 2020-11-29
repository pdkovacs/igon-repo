package pingpong

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPingpongPing(t *testing.T) {
	in := "ping"
	expected := "pong"
	actual := Foo(in)
	assert.Equal(t, expected, actual)
}

func TestPingpongBlurp(t *testing.T) {
	in := "bing"
	expected := "blurp"
	actual := Foo(in)
	assert.Equal(t, expected, actual)
}
