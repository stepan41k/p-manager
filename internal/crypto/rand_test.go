package crypto

import (
	"testing"
)


func TestGeneratePassword(t *testing.T) {
	tests := []struct{
		name string
		n int
	}{
		{
			name: "length 10",
			n: 10,
		},
		{
			name: "length 20",
			n: 20,
		},
		{
			name: "length 30",
			n: 30,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GeneratePassword(tt.n)
			t.Log(got)
		})
	}
}

func TestGeneratePassword_Uniqueness(t *testing.T) {
	iterations := 1000000
	seen := make(map[string]struct{})
	
	for i := 0; i < iterations; i++ {
		s := GeneratePassword(16)
		if _, ok := seen[s]; ok {
			t.Errorf("Collision detected at iteration %d: %s", i, s)
		} 
		seen[s] = struct{}{}
	}
}