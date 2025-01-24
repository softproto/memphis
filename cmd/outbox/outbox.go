package outbox

import (
	"context"
	"log"
	"sync"

	"memphis/cmd/contractor"
)

func Run(ctx context.Context, wg *sync.WaitGroup, jobs chan contractor.Job) {
	defer wg.Done()
	log.Println("Outbox() started")

	for {
		select {
		case <-ctx.Done():
			log.Println("context.Done() in Outbox()")
			return

		default:

		}
	}
}
