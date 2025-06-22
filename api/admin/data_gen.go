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
	min, max := 5, 669

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
				flip := rand.Intn(69)
				nd := rand.Intn(75) + 25
				na := rand.Intn(75) + 25
				d := clamp(nd)
				c := clamp(na)
				n := clamp(nd + na)
				switch flip {
				case 0:
					go CreateXUsers(n)
				case 1:
					go DeleteXDummies(n)
				case 2:
					go CreateXUsers(c + 75)
					go DeleteXDummies(d + 75)
				case 3:
					go CreateXUsers(c + 75)
					go CreateXUsers(c + 75)
				case 4:
					go DeleteXDummies(d + 75)
					go DeleteXDummies(d + 75)
				case 5:
					go CreateXUsers(c + 500)
					go DeleteXDummies(d + 500)
				default:
					go CreateXUsers(c + 5)
					go DeleteXDummies(d + 5)
				}
			}
		}
	}()
}
