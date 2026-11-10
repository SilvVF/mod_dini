package main

import (
	"log"
	"testing"
)

func TestHuiHiu(t *testing.T) {

	h := &HuiHuiProvider{}

	page, err := h.Page(t.Context(), "8277")

	if err != nil {
		t.Error(err)
	}

	log.Println(page)

	t.Error(nil)
}
