package networks

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/danmuck/dps_http/api/logs"
	"github.com/shirou/gopsutil/net"
)

// default formatting constants helper
func FormatB(value float64) string {
	return FormatBits(value, 3, 2)
}
func FormatBibi(value float64) string {
	return FormatBitsIbi(value, 2, 2)
}

// FormatSize converts “value” bits into the largest human-readable unit so that
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
	return formatFloat(v, decDigits) + " b"
}

// FormatBitsIbi converts “value” bits into the largest human-readable binary unit so that
// 1 ≤ integer-part ≤ maxIntDigits, rounds to decDigits, trims trailing zeros,
// and appends the unit suffix.
func FormatBitsIbi(value float64, maxIntDigits, decDigits int) string {
	type unit struct {
		name string
		size float64
	}
	units := []unit{
		{"PiB", PiB}, {"TiB", TiB}, {"GiB", GiB}, {"MiB", MiB},
		{"KiB", KiB}, {"B", Byte}, {"b", Bit},
	}

	for _, u := range units {
		if value >= u.size {
			v := value / u.size
			if integerDigits(v) <= maxIntDigits {
				return formatFloat(v, decDigits) + " " + u.name
			}
		}
	}
	// fallback to bits
	v := value / Bit
	return formatFloat(v, decDigits) + " b"
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

// note: rmm post testing
// refactored out
func GenerateFrame(distance_m, duration_s float64) *Frame {
	ival := 1 * time.Second
	// 1. Get the initial cumulative counters (aggregate over all interfaces)
	bootstrap := FilterIOCounters(false)

	fr := &Frame{
		Source:     "interface or all",
		Duration_s: duration_s,
		Payload:    make([]byte, int(256/8)), // convert bits to bytes note: 256 is filler!!
		Timestamp:  time.Now(),
		Samples:    0,

		Sent_b:   0,
		Recv_b:   0,
		Sent_pkt: 0,
		Recv_pkt: 0,

		Upload_bps:   0,
		Download_bps: 0,
		PktsUp_pps:   0,
	}

	// use index 0 because pernic=false returns a single aggregate counter
	prev := bootstrap[0]
	ticker := time.NewTicker(ival)
	defer ticker.Stop()
	for now := range ticker.C {

		// 2. Read them again
		counters, err := net.IOCounters(false)
		if err != nil {
			logs.Warn("error reading IO counters: %v", err)
			continue
		}
		curr := counters[0]

		// 3. Calculate deltas
		frSent := float64(curr.BytesSent - prev.BytesSent)
		frRecv := float64(curr.BytesRecv - prev.BytesRecv)
		frSentPkts := float64(curr.PacketsSent - prev.PacketsSent)
		frRecvPkts := float64(curr.PacketsRecv - prev.PacketsRecv)

		logs.Dev("%s, Bytes Sent: %.2f, Bytes Received: %.2f, Packets Sent: %.2f, Packets Received: %.2f",
			now.Format("15:04:05"), frSent, frRecv,
			frSentPkts, frRecvPkts)
		// 4. Compute rates (bytes per second)
		//    delta bytes / ticker seconds == bytes/sec
		upload_Bps := frSent / ival.Seconds()
		download_Bps := frRecv / ival.Seconds()
		packets_ps := frSentPkts + frRecvPkts/ival.Seconds()
		avgPacketSize_B := (frSent + frRecv) / (frSentPkts + frRecvPkts)

		logs.Warn("Upload: %.2f B/s, Download: %.2f B/s, Packets/s: %.2f, Avg Packet Size: %.2f B",
			upload_Bps, download_Bps, packets_ps, avgPacketSize_B)

		// conver to bits per second
		upload_bps := upload_Bps * 8 // bits per second
		download_bps := download_Bps * 8
		logs.Info("Upload: %.2f bps, Download: %.2f bps", upload_bps, download_bps)

		// 5. (Optional) convert to more human units, e.g. KiB/s
		uploadKiB := upload_Bps / 1024
		downloadKiB := download_Bps / 1024
		logs.Info("Upload: %s/s, Download: %s/s",
			FormatBibi(upload_bps), FormatBibi(download_bps))
		// 6. Print

		logs.Log("%s → Upload: %.2f KiB/s, Download: %.2f KiB/s\n",
			now.Format("15:04:05"), uploadKiB, downloadKiB)

		// 7. Prepare for next iteration
		prev = curr
	}
	return fr
}
