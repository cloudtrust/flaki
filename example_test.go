package flaki

import (
	"fmt"
	"log"
)

// This example demonstrates the use of NextID to obtain a
// unique uint64 ID.
func ExampleFlaki_NextID() {
	flaki, err := New()
	if err != nil {
		log.Fatalf("could not create flaki generator: %v", err)
	}

	id, err := flaki.NextID()
	if err != nil {
		log.Fatalf("could not generate ID: %v", err)
	}

	fmt.Printf("Unique ID: %d", id)
}

// This example demonstrates the use of NextIDString to obtain a
// unique string ID.
func ExampleFlaki_NextIDString() {
	flaki, err := New()
	if err != nil {
		log.Fatalf("could not create flaki generator: %v", err)
	}

	id, err := flaki.NextIDString()
	if err != nil {
		log.Fatalf("could not generate ID: %v", err)
	}

	fmt.Printf("Unique ID: %s", id)
}

// This example demonstrates the use of NextID to obtain a
// unique uint64 ID.
func ExampleFlaki_NextValidID() {
	flaki, err := New()
	if err != nil {
		log.Fatalf("could not create flaki generator: %v", err)
	}

	id := flaki.NextValidID()
	if err != nil {
		log.Fatalf("could not generate ID: %v", err)
	}

	fmt.Printf("Unique ID: %d", id)
}

// This example demonstrates the use of NextIDString to obtain a
// unique string ID.
func ExampleFlaki_NextValidIDString() {
	flaki, err := New()
	if err != nil {
		log.Fatalf("could not create flaki generator: %v", err)
	}

	id := flaki.NextValidIDString()
	if err != nil {
		log.Fatalf("could not generate ID: %v", err)
	}

	fmt.Printf("Unique ID: %s", id)
}
