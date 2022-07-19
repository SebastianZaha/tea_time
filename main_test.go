package main

import "testing"

func TestDrawText(t *testing.T) {
	bs := drawText("a")
	if len(bs) == 0 {
		t.Error("drawText returned 0 bytes .ico")
	}
}
