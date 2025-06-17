package networks

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Packet struct {
	Source        string  `json:"source"`
	Destination   string  `json:"destination"`
	PayloadSize_b float64 `json:"payload_size"` // in bits
	Payload       []byte  `json:"payload"`      // actual data
}

type ServiceParams struct {
	Label           string  `json:"label"`         // Label for the query
	Distance_m      float64 `json:"distance_m"`    // Distance in meters (D)
	DataRate_bps    float64 `json:"data_rate_bps"` // Data rate in bits per second (R)
	PacketSize_b    float64 `json:"packet_size_b"` // Size of each packet in bits (L)
	PacketLoad      int     `json:"packets"`       // Number of packets (N)
	ArrivalRate_pps float64 `json:"arrival_rate"`  // Packets per second (λ)
	ServiceRate_pps float64 `json:"mu"`            // Service rate in packets per second (μ)

}
type TransmissionWindow struct {
	Packets                   []*Packet // List of packets to query
	BitsProcessed             float64   // Total size of packets in bits
	PacketsServiced           int       // Number of packets
	AvgPacketSize             float64   // Average size of packets in bits
	AvgPacketTransmissionTime float64   // Average time to transmit a single packet in seconds
	TotalTransmissionTime     float64   // total transmission time in seconds
	LinkPropDelay             float64   // Link propagation delay in seconds
	ProcessingDelay           float64   // Processing delay in seconds
	QueueingDelay             float64   // Queueing delay in seconds
	RTT                       float64   // Round trip time in seconds
	PersistentServiceTime     float64   // persistent connections in seconds
	NonPersistentServiceTime  float64   // non-persistent connections in seconds
	AverageSystemTimeMM1      float64   // Average system time in M/M/1 queueing model
	// PacketTransmissionTime    float64   // Time to transmit a single packet in seconds

}

type NetworkMetrics struct {
	TransmissionLog  []*TransmissionWindow `json:"transmission_log"`
	NetworkLatency   float64               `json:"network_latency"`   // in milliseconds
	NetworkSpeed     float64               `json:"network_speed"`     // in Mbps
	NetworkBandwidth float64               `json:"network_bandwidth"` // in Mbps
	NetworkJitter    float64               `json:"network_jitter"`    // in milliseconds
}

func NewServiceParams(link_distance, data_rate, size float64, packets int, name string) *ServiceParams {
	return &ServiceParams{
		Label:           name,
		Distance_m:      link_distance,
		DataRate_bps:    data_rate,
		PacketSize_b:    size,
		PacketLoad:      packets,
		ArrivalRate_pps: 40.0,             // (λ) how fast packets arrive (1 pkt/sec) debug:
		ServiceRate_pps: data_rate / size, // (μ) how fast you could serve them if no queueing debug:
	}
}
func (s *ServiceParams) String() string {
	return fmt.Sprintf(`
	ServiceParams {
		Label: %s,
		(D) Distance (m): %.2f,
		(R) Data Rate (bps): %.2f,
		(L) Packet Size (b): %.2f,
		(N) Packet Load: %d,
		(λ) Arrival Rate (pps): %.2f,
		(μ) Service Rate (pps): %.2f
	}`, s.Label, s.Distance_m, s.DataRate_bps, s.PacketSize_b, s.PacketLoad, s.ArrivalRate_pps, s.ServiceRate_pps)
}

// default units
func FormatB(value float64) string {
	return FormatBits(value, 3, 2)
}

// FormatSize converts “value” into the largest human-readable unit so that
// 1 ≤ integer-part ≤ maxIntDigits, rounds to decDigits, trims trailing zeros,
// and appends the unit suffix.
func FormatBits(value float64, maxIntDigits, decDigits int) string {
	type unit struct {
		name string
		size float64
	}
	units := []unit{
		{"PB", PB}, {"Tb", Tb}, {"GB", GB}, {"Mb", Mb},
		{"MB", MB}, {"KB", KB}, {"Kb", Kb}, {"B", Byte}, {"b", Bit},
	}

	// pick the first unit where v = value/size ≥ 1 and integer-digits(v) ≤ maxIntDigits
	for _, u := range units {
		if value >= u.size {
			v := value / u.size
			if integerDigits(v) <= maxIntDigits {
				return formatFloat(v, decDigits) + " " + u.name
			}
		}
	}

	// fallback: just bits
	v := value / Bit
	return formatFloat(v, decDigits) + "b"
}

// integerDigits returns the count of digits left of the decimal in |v|.
// (e.g. v=0.5→1, v=12.3→2, v=1234→4)
func integerDigits(v float64) int {
	v = math.Abs(v)
	if v < 1 {
		return 1
	}
	return int(math.Floor(math.Log10(v))) + 1
}

// formatFloat produces a string with exactly decDigits places, then
// trims any trailing “0”s and a trailing “.” if present.
func formatFloat(v float64, decDigits int) string {
	s := strconv.FormatFloat(v, 'f', decDigits, 64)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}
	return s
}
