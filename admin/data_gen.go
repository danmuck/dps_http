package admin

// import (
// 	"math/rand"
// 	"time"

// 	"github.com/danmuck/dps_lib/logs"
// )

// var DATAGEN_delay = 2 * time.Second // delay between data generation events

// type DataGenerator struct {
// 	kill chan any
// }

// func newDataGenerator() chan any {

// 	dg := &DataGenerator{kill: make(chan any)}
// 	dg.start()
// 	return dg.kill
// }

// func clamp(x int) int {
// 	min, max := 25, 669

// 	if x < min {
// 		return min
// 	}
// 	if x > max {
// 		return max
// 	}
// 	return x
// }

// func (dg *DataGenerator) start() {
// 	logs.Dev("DataGenerator started ...")
// 	go func() {
// 		for {
// 			select {
// 			case <-dg.kill:
// 				dg.kill = nil
// 				return
// 			default:
// 				time.Sleep(DATAGEN_delay)
// 				flip := rand.Intn(18)
// 				n := rand.Intn(75) + 25
// 				n = clamp(n)

// 				switch flip {
// 				case 0:
// 					go CreateXUsers(n)
// 				case 1:
// 					go DeleteXDummies(n)
// 				case 2:
// 					go CreateXUsers(n + 75)
// 				case 3:
// 					go DeleteXDummies(n + 55)
// 				case 4:
// 					go CreateXUsers(n + 200)
// 				case 5:
// 					go DeleteXDummies(n + 15)
// 				default:
// 					go CreateXUsers(max(n-50, 0))
// 					go DeleteXDummies(max(n-35, 0))
// 				}
// 			}
// 		}
// 	}()
// }
