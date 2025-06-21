package networks

import (
	"fmt"
	"strings"
	"time"

	"github.com/danmuck/dps_http/lib/logs"
	"github.com/shirou/gopsutil/net"
)

// Package networks provides utilities for calculating network metrics such as transmission delays, propagation delays, and service rates.
// It includes definitions for various units of measurement, packet structures, and service parameters.
// It also provides methods for creating service parameters and calculating network utilization.
// The API call requires:
// - Link distance in meters
// - Data rate in bits per second
// - Packet size in bits
// - Number of packets to be processed
// - Label for the service parameters

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

	KiB = 1024 * Byte // Kibibyte
	MiB = 1024 * KiB  // Mebibyte
	GiB = 1024 * MiB  // Gibibyte
	TiB = 1024 * GiB  // Tebibyte
	PiB = 1024 * TiB  // Pebibyte

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

type ServiceParams struct {
	Iface           string  `json:"interface"`     // Label for the query
	Distance_m      float64 `json:"distance_m"`    // Physical distance in meters (D)
	DataRate_bps    float64 `json:"data_rate_bps"` // Data rate in bits per second (R) **
	PacketSize_b    float64 `json:"packet_size_b"` // Size of each packet in bits (L) **
	PacketLoad      int     `json:"packets"`       // Number of packets (N) **
	ArrivalRate_pps float64 `json:"lambda"`        // Packets per second (λ)
	ServiceRate_pps float64 `json:"mu"`            // Service rate in packets per second (μ)
}

// Transmission Frame
type Frame struct {
	Source     string    `json:"source"`
	Samples    float64   `json:"sample_size"` // in bits
	Payload    []byte    `json:"payload"`     // actual data -- unused
	Duration_s float64   `json:"duration"`    // duration in seconds
	Timestamp  time.Time `json:"timestamp"`   // timestamp of the frame

	Sent_b   float64 `json:"bits_sent"` // bytes sent
	Recv_b   float64 `json:"bits_recv"` // bytes received
	Sent_pkt uint64  `json:"pkts_sent"` // packets sent (N)
	Recv_pkt uint64  `json:"pkts_recv"` // packets received (N)

	Upload_bps   float64 `json:"upload_bps"`   // upload rate in bits per second (R)
	Download_bps float64 `json:"download_bps"` // download rate in bits per second (R)
	PktsUp_pps   float64 `json:"pkts_up"`      // packet rate in packets per second (mu)
	PktsDown_pps float64 `json:"pkts_down"`    // packets received per second (lamda)
	AvgPktSize   float64 `json:"avg_pkt_size"` // average packet size in bits
}

func (fr *Frame) String() string {
	return fmt.Sprintf(`
	Frame{
		Source: %s,
		Samples: %f,
		Timestamp: %s,
		Duration_s: %f,

		Sent_b: %f,
		Recv_b: %f,
		Sent_pkt: %d,
		Recv_pkt: %d,

		Upload_bps: %f,
		Download_bps: %f,
		PktRate_pps: %f,
		AvgPktSize: %f
	}`,
		fr.Source, fr.Samples, fr.Timestamp, fr.Duration_s,
		fr.Sent_b, fr.Recv_b, fr.Sent_pkt, fr.Recv_pkt,
		fr.Upload_bps, fr.Download_bps, fr.PktsUp_pps, fr.AvgPktSize)
}

// filter networkio counters to specific interfaces
// pernic of true with no ifaces returns all interfaces
func FilterIOCounters(pernic bool, ifaces ...string) []net.IOCountersStat {
	// get counters with given pernic
	stats, err := net.IOCounters(pernic)
	if err != nil {
		logs.Err("unable to read IO counters: %v", err)
		return nil
	}
	// use pernic to decide if we need to filter
	if pernic {
		if ifaces == nil || len(ifaces) == 0 {
			logs.Err("no interfaces specified")
			return stats
		}
		filtered := make([]net.IOCountersStat, len(ifaces))
		for _, stat := range stats {
			for _, filters := range ifaces {
				if strings.Contains(stat.Name, filters) {
					filtered = append(filtered, stat)
				}
			}
		}
		logs.Debug("filtered %d interfaces", len(filtered))
		return filtered
	}
	return stats
}

func (fr *Frame) PopulateFrame(src string, samples, duration_s float64) {
	// initialize with single sample
	sc := FilterIOCounters(false)
	start := sc[0]

	time.Sleep(time.Second * time.Duration(duration_s))

	ec := FilterIOCounters(false)
	end := ec[0]

	fr.ComputeDeltas(start, end)
	fr.ComputeRates()
	fr.ComputeAvgPktSize()

}

// new frame for overall traffic on all interfaces
func NewFrame(src string, samples, duration_s float64) *Frame {
	fr := &Frame{
		Source:     src,
		Samples:    samples,
		Timestamp:  time.Now(),
		Duration_s: duration_s,
	}
	// initialize with single sample
	sc := FilterIOCounters(false)
	start := sc[0]

	time.Sleep(time.Second * time.Duration(duration_s))

	ec := FilterIOCounters(false)
	end := ec[0]

	fr.ComputeDeltas(start, end)
	fr.ComputeRates()
	fr.ComputeAvgPktSize()

	return fr
}
func NewSample() *Frame {
	return nil
}

func (fr *Frame) ComputeDeltas(start, next net.IOCountersStat) {
	fr.Sent_b = float64(next.BytesSent-start.BytesSent) * 8 // convert to bits
	fr.Recv_b = float64(next.BytesRecv-start.BytesRecv) * 8
	fr.Sent_pkt = next.PacketsSent - start.PacketsSent
	fr.Recv_pkt = next.PacketsRecv - start.PacketsRecv
	logs.Debug("%s, Bytes Sent: %s, Bytes Received: %s, Packets Sent: %d, Packets Received: %d",
		fr.Timestamp.Format("15:04:05"),
		FormatB(fr.Sent_b), FormatB(fr.Recv_b), fr.Sent_pkt, fr.Recv_pkt,
	)
}

func (fr *Frame) ComputeRates() {
	fr.Upload_bps = fr.Sent_b / fr.Duration_s
	fr.Download_bps = fr.Recv_b / fr.Duration_s
	fr.PktsUp_pps = float64(fr.Sent_pkt) / fr.Duration_s
	fr.PktsDown_pps = float64(fr.Recv_pkt) / fr.Duration_s
	logs.Debug("Upload: %s, Download: %s, Packets: %.2f p/s",
		FormatBibi(fr.Upload_bps), FormatBibi(fr.Download_bps), fr.PktsUp_pps)
	return
}

func (fr *Frame) ComputeAvgPktSize() {
	fr.AvgPktSize = (fr.Sent_b + fr.Recv_b) / float64(fr.Sent_pkt+fr.Recv_pkt)
	logs.Debug("Average Packet Size: %s", FormatB(fr.AvgPktSize))
	return
}

type TransmissionWindow struct {
	Packets                   []*Frame // List of packets to query
	BitsProcessed             float64  // Total size of packets in bits
	FramesServiced            int      // Number of packets
	AvgPacketSize             float64  // Average size of packets in bits
	AvgPacketTransmissionTime float64  // Average time to transmit a single packet in seconds
	TotalTransmissionTime     float64  // total transmission time in seconds
	LinkPropDelay             float64  // Link propagation delay in seconds
	ProcessingDelay           float64  // Processing delay in seconds
	QueueingDelay             float64  // Queueing delay in seconds
	RTT                       float64  // Round trip time in seconds
	PersistentServiceTime     float64  // persistent connections in seconds
	NonPersistentServiceTime  float64  // non-persistent connections in seconds
	AverageSystemTimeMM1      float64  // Average system time in M/M/1 queueing model
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
		Iface:           name,
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
	}`, s.Iface, s.Distance_m, s.DataRate_bps, s.PacketSize_b, s.PacketLoad, s.ArrivalRate_pps, s.ServiceRate_pps)
}
