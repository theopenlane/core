//go:build examples

package main

import (
	"fmt"

	"github.com/theopenlane/shared/soiree"
)

func CustomPanicHandler(recoveredPanic any) {
	fmt.Printf("Recovered from panic: %v", recoveredPanic)
}

func main() {
	pool := soiree.NewEventPool(soiree.WithPanicHandler(CustomPanicHandler))
	topic := "user.created"

	listener := func(ctx *soiree.EventContext) error {
		panic(fmt.Sprintf("George Costanza when there's a fire: %s", ctx.Event().Topic()))
	}

	if _, err := pool.On(topic, listener); err != nil {
		panic(err)
	}

	pool.Emit(topic, soiree.NewBaseEvent(topic, "sfunk"))

	fmt.Println("Application continues running despite the panic - isn't programming great")
}
