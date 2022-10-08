package utils_test

import (
	"interview-fm-backend/internal/utils"
	"testing"
)

func TestGenerateKey(t *testing.T) {
	table := []struct {
		src string
		res string
	}{
		{"", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{"abc", "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"},
		{"hello", "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"},
	}
	for _, tc := range table {
		res := utils.GenerateKey(tc.src)
		if res != tc.res {
			t.Errorf("expected %s, got %s", tc.res, res)
		}
	}
}

func BenchmarkGenerateKey_A(b *testing.B) {
	sample := "Hello world"
	for i := 0; i < b.N; i++ {
		utils.GenerateKey_A(sample)
	}
}

func BenchmarkGenerateKey_B(b *testing.B) {
	sample := "Hello world"
	for i := 0; i < b.N; i++ {
		utils.GenerateKey_B(sample)
	}
}

func BenchmarkGenerateKey_C(b *testing.B) {
	sample := "Hello world"
	for i := 0; i < b.N; i++ {
		utils.GenerateKey(sample)
	}
}
