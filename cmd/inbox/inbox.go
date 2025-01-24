package inbox

import (
	"context"
	"fmt"
	"log"
	"sync"

	"memphis/cmd/common"
	"memphis/cmd/contractor"
	"time"
)

func Run(ctx context.Context, wg *sync.WaitGroup, jobs chan contractor.Job) {
	defer wg.Done()
	log.Println("Inbox() started")

	i := 0
	for {
		select {
		case <-ctx.Done():
			log.Println("context.Done() in Inbox()")
			return

		default:
			jobs <- contractor.Job{
				Descriptor: contractor.JobDescriptor{
					ID:    contractor.JobID(fmt.Sprintf("%v", i)),
					JType: "anyType",
				},
				ExecFn: common.DefFunc,
				Args:   i,
			}
			i++
			time.Sleep(1 * time.Second)

		}
	}
}
