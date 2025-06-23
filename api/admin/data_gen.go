package admin

import (
	"math/rand"
	"time"

	"github.com/danmuck/dps_http/configs"
	"github.com/danmuck/dps_http/lib/logs"
)

type DataGenerator struct {
	kill chan any
}

func newDataGenerator() chan any {

	dg := &DataGenerator{kill: make(chan any)}
	dg.start()
	return dg.kill
}

func clamp(x int) int {
	min, max := 25, 669

	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}

func (dg *DataGenerator) start() {
	logs.Dev("DataGenerator started ...")
	go func() {
		for {
			select {
			case <-dg.kill:
				dg.kill = nil
				return
			default:
				time.Sleep(configs.DATAGEN_delay)
				flip := rand.Intn(18)
				n := rand.Intn(750) + 250
				n = clamp(n)

				switch flip {
				case 0:
					go CreateXUsers(n)
				case 1:
					go DeleteXDummies(n)
				case 2:
					go CreateXUsers(n + 750)
				case 3:
					go DeleteXDummies(n + 550)
				case 4:
					go CreateXUsers(n + 2000)
				case 5:
					go DeleteXDummies(n + 1500)
				default:
					go CreateXUsers(max(n-500, 0))
					go DeleteXDummies(max(n-350, 0))
				}
			}
		}
	}()
}
