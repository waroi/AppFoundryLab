package incidents

import (
	"testing"
)

func BenchmarkSinkMapLookup(b *testing.B) {
	enabledSinks := map[string]bool{
		"logger":  true,
		"stdout":  true,
		"webhook": true,
	}
	target := "webhook"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = enabledSinks[target]
	}
}
