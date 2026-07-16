package ziwei

import "testing"

// BenchmarkGenerate 排盘全流程(安星+四化+大限)。
func BenchmarkGenerate(b *testing.B) {
	in := BirthInfo{Year: 1990, Month: 6, Day: 15, Hour: 5, Gender: Male}
	opt := Options{ReferenceYear: 2026}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := Generate(in, opt); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGenerateWithPatterns 排盘 + 格局判定。
func BenchmarkGenerateWithPatterns(b *testing.B) {
	in := BirthInfo{Year: 1990, Month: 6, Day: 15, Hour: 5, Gender: Male}
	opt := Options{ReferenceYear: 2026}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		chart, err := Generate(in, opt)
		if err != nil {
			b.Fatal(err)
		}
		DetectPatterns(chart)
	}
}
