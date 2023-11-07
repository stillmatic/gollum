package gollum_test

import (
	"fmt"
	"testing"

	math "github.com/chewxy/math32"
	"github.com/stillmatic/gollum/internal/testutil"
	"github.com/viterin/vek/vek32"
)

func weaviateCosineSimilarity(a []float32, b []float32) (float32, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("vectors have different dimensions")
	}

	var (
		sumProduct float32
		sumASquare float32
		sumBSquare float32
	)

	for i := range a {
		sumProduct += (a[i] * b[i])
		sumASquare += (a[i] * a[i])
		sumBSquare += (b[i] * b[i])
	}

	return sumProduct / (math.Sqrt(sumASquare) * math.Sqrt(sumBSquare)), nil
}

func BenchmarkCosSim(b *testing.B) {
	ns := []int{256, 512, 768, 1024}
	b.Run("weaviate", func(b *testing.B) {
		for _, n := range ns {
			A := testutil.GetRandomEmbedding(n)
			B := testutil.GetRandomEmbedding(n)
			b.ResetTimer()
			b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					f, err := weaviateCosineSimilarity(A, B)
					if err != nil {
						panic(err)
					}
					_ = f
				}
			})
		}
	})
	b.Run("vek", func(b *testing.B) {
		for _, n := range ns {
			A := testutil.GetRandomEmbedding(n)
			B := testutil.GetRandomEmbedding(n)
			b.ResetTimer()
			b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					f := vek32.CosineSimilarity(A, B)
					_ = f
				}
			})
		}
	})
}

// func BenchmarkGonum32(b *testing.B) {
// 	vs := []int{256, 512, 768, 1024}
// 	for _, n := range vs {
// 		A := getRandomEmbedding(n)
// 		B := getRandomEmbedding(n)
// 		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
// 			for i := 0; i < b.N; i++ {
// 				f, err := gallant.GonumSim(A, B)
// 				if err != nil {
// 					panic(err)
// 				}
// 				_ = f
// 			}
// 		})
// 	}
// }

// func BenchmarkGonum64(b *testing.B) {
// 	vs := []int{256, 512, 768, 1024}
// 	for _, n := range vs {
// 		A := getRandomEmbedding(n)
// 		B := getRandomEmbedding(n)
// 		a_ := make([]float64, len(A))
// 		b_ := make([]float64, len(B))
// 		for i := range A {
// 			a_[i] = float64(A[i])
// 			b_[i] = float64(B[i])
// 		}
// 		b.Run("Naive", func(b *testing.B) {
// 			for i := 0; i < b.N; i++ {
// 				f, err := gallant.GonumSim64(a_, b_)
// 				if err != nil {
// 					panic(err)
// 				}
// 				_ = f
// 			}
// 		})
// 		am := mat.NewDense(1, len(a_), a_)
// 		bm := mat.NewDense(1, len(b_), b_)
// 		b.Run("prealloc", func(b *testing.B) {
// 			for i := 0; i < b.N; i++ {
// 				var dot mat.Dense
// 				dot.Mul(am, bm.T())
// 				_ = dot.At(0, 0)
// 			}
// 		})
// 	}
// }

// goos: linux
// goarch: amd64
// cpu: AMD Ryzen 9 7950X 16-Core Processor
// BenchmarkWeaviateCosSim/256-32         	 8028375	       154.8 ns/op	       0 B/op	       0 allocs/op
// BenchmarkWeaviateCosSim/512-32         	 3958342	       300.1 ns/op	       0 B/op	       0 allocs/op
// BenchmarkWeaviateCosSim/768-32         	 2677993	       456.1 ns/op	       0 B/op	       0 allocs/op
// BenchmarkWeaviateCosSim/1024-32        	 2002258	       601.6 ns/op	       0 B/op	       0 allocs/op
// BenchmarkVek32CosSim/256-32            	81166414	        15.43 ns/op	       0 B/op	       0 allocs/op
// BenchmarkVek32CosSim/512-32            	46376474	        26.72 ns/op	       0 B/op	       0 allocs/op
// BenchmarkVek32CosSim/768-32            	30476739	        39.65 ns/op	       0 B/op	       0 allocs/op
// BenchmarkVek32CosSim/1024-32           	22698370	        51.63 ns/op	       0 B/op	       0 allocs/op
// BenchmarkGonum32/256-32                	 1664224	       738.9 ns/op	    4312 B/op	       7 allocs/op
// BenchmarkGonum32/512-32                	  914896	      1212 ns/op	    8408 B/op	       7 allocs/op
// BenchmarkGonum32/768-32                	  718142	      1634 ns/op	   12504 B/op	       7 allocs/op
// BenchmarkGonum32/1024-32               	  573609	      2885 ns/op	   16600 B/op	       7 allocs/op
// BenchmarkGonum64/Naive-32              	 6329708	       189.8 ns/op	     216 B/op	       5 allocs/op
// BenchmarkGonum64/prealloc-32           	 9126764	       144.8 ns/op	      88 B/op	       3 allocs/op
// BenchmarkGonum64/Naive#01-32           	 5176160	       222.0 ns/op	     216 B/op	       5 allocs/op
// BenchmarkGonum64/prealloc#01-32        	 6484794	       185.2 ns/op	      88 B/op	       3 allocs/op
// BenchmarkGonum64/Naive#02-32           	 4569102	       266.8 ns/op	     216 B/op	       5 allocs/op
// BenchmarkGonum64/prealloc#02-32        	 5566400	       225.0 ns/op	      88 B/op	       3 allocs/op
// BenchmarkGonum64/Naive#03-32           	 3910498	       300.0 ns/op	     216 B/op	       5 allocs/op
// BenchmarkGonum64/prealloc#03-32        	 4585336	       265.9 ns/op	      88 B/op	       3 allocs/op

// goos: darwin
// goarch: arm64
// pkg: github.com/stillmatic/gollum
// BenchmarkCosSim/weaviate/256-10     	 4330524	       274.5 ns/op	       0 B/op	       0 allocs/op
// BenchmarkCosSim/weaviate/512-10     	 1995426	       605.6 ns/op	       0 B/op	       0 allocs/op
// BenchmarkCosSim/weaviate/768-10     	 1312820	       917.6 ns/op	       0 B/op	       0 allocs/op
// BenchmarkCosSim/weaviate/1024-10    	  973432	      1232 ns/op	       0 B/op	       0 allocs/op
// BenchmarkCosSim/vek/256-10          	 4335747	       272.0 ns/op	       0 B/op	       0 allocs/op
// BenchmarkCosSim/vek/512-10          	 2027366	       596.2 ns/op	       0 B/op	       0 allocs/op
// BenchmarkCosSim/vek/768-10          	 1310983	       925.2 ns/op	       0 B/op	       0 allocs/op
// BenchmarkCosSim/vek/1024-10         	  969460	      1233 ns/op	       0 B/op	       0 allocs/op
// PASS

// General high level takeaways:
// - vek32 is best option if SIMD is available
// - gonum64 is fast and scales better than weaviate but requires allocs and using f64
// - gonum64 probably has SIMD intrinsics?
// - weaviate implementation is 'fine', appears to have linear scaling with vector size
// - for something in the hotpath I think it's worth it to use vek32 when possible
// - on mac, vek32 and weaviate are more or less identical
