package main

import "testing"

func TestEnsureDirectDL(t *testing.T) {
	v := MySqlBackend{}
	err := v.EnsureDirectDL(2)
	if err != nil {
		t.Error(err)
	}
}
