package networks

import (
	"math"

	"github.com/danmuck/dps_http/lib/logs"
)

const (
	NetworkPropagationSpeedActual = 299792458.0 // Speed of light in meters per second
	NetworkPropagationSpeedApprox = 300000000.0 // Approximate speed in meters per second
)

// transmissionDelay = L / R
func transmissionDelay(bits, rate float64) float64 {
	return bits / rate
}

// propagationDelay = D / S
func propagationDelay(distance, speed float64) float64 {
	return distance / speed
}

// RTT = 2×propDelay
func roundTripTime(distance, speed float64) float64 {
	return 2 * propagationDelay(distance, speed)
}

// serviceRate μ = R / L (packets per second)
func serviceRate(rate, bits float64) float64 {
	return rate / bits
}

// averageQueueingDelayMM1 Wq = ρ / (μ − λ), where ρ = λ/μ
// returns +Inf if λ ≥ μ
func averageQueueingDelayMM1(lambda, mu float64) float64 {
	rho := lambda / mu
	den := mu - lambda
	if den <= 0 {
		return math.Inf(1) // truly overloaded
	}
	return rho / den
}

// averageSystemTimeMM1 W = Wq + 1/μ
func averageSystemTimeMM1(lambda, mu float64) float64 {
	wq := averageQueueingDelayMM1(lambda, mu)
	if math.IsInf(wq, 1) {
		return math.Inf(1)
	}
	return wq + 1.0/mu
}

// persistentServiceTime = 2·RTT + N·d_trans
func persistentServiceTime(distance, speed, bits, rate float64, N int) float64 {
	rtt := roundTripTime(distance, speed)
	dTrans := transmissionDelay(bits, rate)
	return 2*rtt + float64(N)*dTrans
}

// nonPersistentServiceTime = (2·RTT + d_trans) · N
func nonPersistentServiceTime(distance, speed, bits, rate float64, N int) float64 {
	rtt := roundTripTime(distance, speed)
	dTrans := transmissionDelay(bits, rate)
	return (2*rtt + dTrans) * float64(N)
}

// ComputeUtilization computes the network utilization based on the provided transmission windows.
// fraction of time link is busy transmitting packets
// Utilization is calculated as the ratio of total transmission time to the total propagation time for persistent connections.
// Non-persistent utilization is calculated similarly but uses non-persistent service time.
// (N * L / R) / (2 * PD + N * L / R) for persistent connection services
// (N * L / R) / ((2 * PD + L / R) * N) for non-persistent connection services
// where PD is the propagation delay, N is the number of packets, L is the packet size, and R is the data rate.
func ComputeUtilization(metrics ...*TransmissionWindow) (util, utilNP float64) {
	logs.Info("[ Running network utilization query ]")
	prop_total := 0.0
	nprop_total := 0.0
	window_trans_total := 0.0
	for _, q := range metrics {
		if q.FramesServiced == 0 {
			logs.Warn("No files in query response, cannot calculate utilization.")
		}
		prop_total += q.PersistentServiceTime
		nprop_total += q.NonPersistentServiceTime
		window_trans_total += q.TotalTransmissionTime
	}
	utilization := window_trans_total / prop_total
	nutilization := window_trans_total / nprop_total
	logs.Debug(`
Utilization --
	Total Query Transmission Time:           %.5f s
	Total Propagation Time (persistent):     %.5f s
	Total Propagation Time (non-persistent): %.5f s
	Utilization (persistent):     %.2f%%
	Utilization (non-persistent): %.2f%%
	`,
		window_trans_total,
		prop_total, nprop_total,
		utilization*100, nutilization*100,
	)

	return utilization, nutilization
}

func ComputeMetrics(frame *ServiceParams) *TransmissionWindow {
	tw := &TransmissionWindow{
		Packets: make([]*Frame, 0, frame.PacketLoad),
	}
	fr := &Frame{Source: frame.Iface, Samples: frame.PacketSize_b}
	for range frame.PacketLoad {
		tw.AddFrame(fr)
	}

	// core delays
	dTrans := transmissionDelay(frame.PacketSize_b, frame.DataRate_bps)
	dProp := propagationDelay(frame.Distance_m, NetworkPropagationSpeedActual)
	rtt := roundTripTime(frame.Distance_m, NetworkPropagationSpeedActual)

	// M/M/1 parameters
	mu := frame.ServiceRate_pps
	lambda := frame.ArrivalRate_pps
	tw.ProcessingDelay = 1.0 / mu
	tw.QueueingDelay = averageQueueingDelayMM1(lambda, mu)
	tw.AverageSystemTimeMM1 = averageSystemTimeMM1(lambda, mu)

	// fill rest
	tw.AvgPacketTransmissionTime = dTrans
	tw.TotalTransmissionTime = float64(frame.PacketLoad) * dTrans
	tw.LinkPropDelay = dProp
	tw.RTT = rtt
	tw.PersistentServiceTime = persistentServiceTime(
		frame.Distance_m, NetworkPropagationSpeedActual, frame.PacketSize_b, frame.DataRate_bps, frame.PacketLoad)
	tw.NonPersistentServiceTime = nonPersistentServiceTime(
		frame.Distance_m, NetworkPropagationSpeedActual, frame.PacketSize_b, frame.DataRate_bps, frame.PacketLoad)
	tw.FramesServiced = frame.PacketLoad
	tw.AvgPacketSize = tw.BitsProcessed / float64(tw.FramesServiced)

	return tw
}
