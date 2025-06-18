package networks

import "fmt"

const (
	Bit  = 1.0
	Byte = 8 * Bit
	Kb   = 1000 * Bit  // Kilobit
	KB   = 1000 * Byte // Kilobyte
	Mb   = 1000 * Kb   // Megabit
	MB   = 1000 * KB   // Megabyte
	Gb   = 1000 * Mb   // Gigabit
	GB   = 1000 * MB   // Gigabyte
	Tb   = 1000 * Gb   // Terabit
	TB   = 1000 * GB   // Terabyte
	Pb   = 1000 * Tb   // Petabit
	PB   = 1000 * TB   // Petabyte

	m  = 1.0
	km = 1000 * m // Kilometer

	s     = 1.0
	sec   = 1 * s     // Second
	min   = 60 * s    // Minute
	hour  = 60 * min  // Hour
	day   = 24 * hour // Day
	week  = 7 * day   // Week
	month = 30 * day  // Month
	year  = 365 * day // Year
)

const (
	Delay = iota
	Propagation
	Transmission
	Processing
	Queueing
)

// File Query methods
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
