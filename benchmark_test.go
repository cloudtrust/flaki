package flaki

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkNextID(b *testing.B) {
	var flaki, err = New()
	assert.Nil(b, err)

	for n := 0; n < b.N; n++ {
		flaki.NextID()
	}
}

func BenchmarkNextIDString(b *testing.B) {
	var flaki, err = New()
	assert.Nil(b, err)

	for n := 0; n < b.N; n++ {
		flaki.NextIDString()
	}
}

func BenchmarkNextValidID(b *testing.B) {
	var flaki, err = New()
	assert.Nil(b, err)

	for n := 0; n < b.N; n++ {
		flaki.NextValidID()
	}
}

func BenchmarkNextValidIDString(b *testing.B) {
	var flaki, err = New()
	assert.Nil(b, err)

	for n := 0; n < b.N; n++ {
		flaki.NextValidIDString()
	}
}
