package nanoidutil

import (
	crand "crypto/rand"
	"fmt"
	"io"
)

var (
	simplexAlpha = []byte("bcdfghjknpqrtvxyz")
	simplexMixed = append(simplexAlpha, []byte("0123456789")...)
)

func Simplex() string {
	return DeterministicSimplex(crand.Reader)
}

func DeterministicSimplex(r io.Reader) string {
	return string(append(simplexGenerate(r, simplexAlpha, 1), simplexGenerate(r, simplexMixed, 11)...))
}

func simplexGenerate(r io.Reader, corpus []byte, count int) []byte {
	res := make([]byte, count)
	corpusLen := uint(len(corpus))

	readLen, err := r.Read(res)
	if err != nil {
		panic(fmt.Errorf("nanoidutil: read bytes: %v", err))
	} else if readLen != count {
		panic(fmt.Errorf("nanoidutil: mismatch read length: expected %d, got %d", count, readLen))
	}

	for i := 0; i < count; i++ {
		res[i] = corpus[uint(res[i])%corpusLen]
	}

	return res
}
