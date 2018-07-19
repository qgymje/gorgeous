package gorgeous

import (
	"log"
	"testing"
	"time"
)

func Test_supervisor(t *testing.T) {
	var f = func() {
		log.Println("I'm paniced")
		panic("i'm panicking")
	}
	supervisor(worker(f), 10)
	time.Sleep(1 * time.Second)
}
