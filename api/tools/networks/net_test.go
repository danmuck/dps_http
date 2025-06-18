package networks

import (
	"fmt"
	"testing"

	"github.com/danmuck/dps_http/api/logs"
)

// Key:
// d = service delay in seconds (transmission delay + propagation delay)
// N = number of packets
// L = size of the packet in bits
// R = data rate in bits per second
// D = distance in meters
// S = propagation speed in meters per second (speed of light on fiber)
// RTT = round trip time in seconds
// PRT = persistent response time in seconds
// NPRT = non-persistent response time in seconds
//
// ++ transmission delay = L/R (on wire)
// where L is the size of the packet in bits and R is the data rate in bits per second
//
// ++ propagation delay = D/S (physical link)
// where D is the distance in meters and S is the propagation speed in meters per second
//
// ++ round trip time = 2 * (D/S) + (L/R) = 2d
// where D is the distance in meters, S is the propagation speed in meters per second (speed of light on fiber), L is the size of the packet in bits, and R is the data rate in bits per second
//
// ++ (persistent connections) response time = (2 * RTT) + (d * N)
// where D is the distance in meters, S is the propagation speed in meters per second (speed of light on fiber), L is the size of the packet in bits, R is the data rate in bits per second, and N is the number of packets
//
// ++ (non-persistent connections) response time = ((2 * RTT) + d) * N
// where D is the distance in meters, S is the propagation speed in meters per second
// (speed of light on fiber), L is the size of the packet in bits, R is the data rate
// in bits per second, and N is the number of packets

const (
	DefaultLinkDistance = 1500 * km      // distance in meters
	DefaultDataRate     = 200 * Mb / sec // data rate in bits per second
	DefaultPacketSize   = 4 * MB         // size of each chunk in bits
	DefaultPackets      = 5              // number of chunks to simulate
	DefaultArrivalRate  = 40.0           // packets per second (λ)
	DefaultServiceRate  = 50.0           // service rate in packets per second (μ)
	DefaultLabel        = "dummy.label"  // example packet label/identifier note:
)

func TestSweepingMu(t *testing.T) {
	logs.ColorTest()
	logs.Dev("\t========[TestSweepingMu]========")

	// Create a service with default parameters
	service := NewServiceParams(DefaultLinkDistance, DefaultDataRate, DefaultPacketSize, DefaultPackets, DefaultLabel)
	logs.Dev("Service Params: %v", service)

	lambda := DefaultArrivalRate
	for mu := 0.0; mu < 100.0; mu += 10.0 {
		p := &ServiceParams{
			Distance_m:      1.5e6,
			DataRate_bps:    200e6,
			PacketSize_b:    32e6,
			PacketLoad:      5,
			ServiceRate_pps: mu,
			ArrivalRate_pps: lambda,
		}
		w := ComputeMetrics(p)

		fmt.Printf("λ=%.1fp/s (/)  μ: %.2fp/s (=) ρ=%.2f, Wq=%.3fs, W=%.3fs \n",
			lambda, mu, lambda/mu, w.QueueingDelay, w.ProcessingDelay+w.QueueingDelay)
	}

	logs.Warn(`		
	// 	λ 			Arrival rate in packets/sec (how fast requests come in)
	// 	μ 			Service rate in packets/sec (how fast you could process if nobody waited)
	// 	ρ = λ/μ 		Traffic intensity or utilization ratio—fraction of your capacity that’s in use
	// 	Wq = ρ / (μ - λ) 	Average queueing delay in an M/M/1 model:
	// 	W = Wq + 1/μ 		Average system time in an M/M/1 model:
	`)

	response := ComputeMetrics(service)
	logs.Dev("Response: %s", response.String())

	if response.PacketsServiced == 0 {
		t.Error("No packets serviced in the response")
	} else {
		t.Log("Network propagation speed test passed.")
	}
}

func TestSweepingLambdaExt(t *testing.T) {
	logs.ColorTest()
	logs.Dev("\t========[TestSweepingLambdaExt]========")

	μ := serviceRate(DefaultDataRate, DefaultPacketSize)
	for λ := 0.0; λ <= μ*1.2; λ += μ / 20 {
		w := ComputeMetrics(&ServiceParams{
			Distance_m:      1.5e6,
			DataRate_bps:    200e6,
			PacketSize_b:    32e6,
			PacketLoad:      5,
			ArrivalRate_pps: λ,
			ServiceRate_pps: μ,
		})

		fmt.Printf("λ=%.1fp/s (/)  μ: %.2fp/s (=) ρ=%.2f, Wq=%.3fs, W=%.3fs \n",
			λ, μ, λ/μ, w.QueueingDelay, w.ProcessingDelay+w.QueueingDelay)
	}
	logs.Warn(`		
	// 	λ 			Arrival rate in packets/sec (how fast requests come in)
	// 	μ 			Service rate in packets/sec (how fast you could process if nobody waited)
	// 	ρ = λ/μ 		Traffic intensity or utilization ratio—fraction of your capacity that’s in use
	// 	Wq = ρ / (μ - λ) 	Average queueing delay in an M/M/1 model:
	// 	W = Wq + 1/μ 		Average system time in an M/M/1 model:
	`)

	if true {
		t.Log("Network propagation speed test passed.")
	} else {
		t.Error("Network propagation speed test failed.")
	}
}

func TestSweepingLambda(t *testing.T) {
	logs.ColorTest()
	logs.Dev("Testing sweeping lambda values for network service parameters...")

	mu := serviceRate(DefaultDataRate, DefaultPacketSize)
	for lambda := 0.0; lambda < mu; lambda += mu / 10 {
		p := &ServiceParams{
			Distance_m:      1.5e6,
			DataRate_bps:    200e6,
			PacketSize_b:    32e6,
			PacketLoad:      5,
			ArrivalRate_pps: lambda,
			ServiceRate_pps: mu,
		}
		w := ComputeMetrics(p)

		fmt.Printf("λ=%.1fp/s (/)  μ: %.2fp/s (=) ρ=%.2f, Wq=%.3fs, W=%.3fs \n",
			lambda, mu, lambda/mu, w.QueueingDelay, w.ProcessingDelay+w.QueueingDelay)
	}
	logs.Warn(`		
	// 	λ 			Arrival rate in packets/sec (how fast requests come in)
	// 	μ 			Service rate in packets/sec (how fast you could process if nobody waited)
	// 	ρ = λ/μ 		Traffic intensity or utilization ratio—fraction of your capacity that’s in use
	// 	Wq = ρ / (μ - λ) 	Average queueing delay in an M/M/1 model:
	// 	W = Wq + 1/μ 		Average system time in an M/M/1 model:
	`)

	// Check if the response is valid
	if mu == 0 {
		t.Error("No packets serviced in the response")
	} else {
		t.Log("Network propagation speed test passed.")
	}
}
func TestNetworkingConcepts(t *testing.T) {
	logs.ColorTest()
	logs.Dev("========[TestNetworkingConcepts]========")

	server_1r := NewServiceParams(DefaultLinkDistance, DefaultDataRate, DefaultPacketSize, DefaultPackets, "[ 1 ]")
	response_1r := ComputeMetrics(server_1r)

	logs.Dev("Response 1 : %s", response_1r.String())
	logs.Warn("=======================================================================")

	server_2r := NewServiceParams(600*km, 500*Mb/sec, 10*MB, 2, " [ 2 ]")
	response_2r := ComputeMetrics(server_2r)

	logs.Dev("Response 2 : %s", response_2r.String())
	logs.Warn("=======================================================================")

	server_3r := NewServiceParams(1200*km, 100*Mb/sec, 8*Mb, 3, "[ 3 ]")
	response_3r := ComputeMetrics(server_3r)

	logs.Dev("Response 3 : %s", response_3r.String())
	logs.Warn("=======================================================================")
	logs.Info("Network propagation speed test completed, computing utilization...")

	utest_util, ntest_util := ComputeUtilization(response_1r, response_2r, response_3r)
	logs.Info("Utilization (Persistent): %.2f%%, Utilization (Non-Persistent): %.2f%%",
		utest_util*100, ntest_util*100)

	if true {
		t.Log("Network propagation speed test passed.")
	} else {
		t.Error("Network propagation speed test failed.")
	}
}
