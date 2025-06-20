package networks

import (
	"fmt"

	"slices"

	"github.com/danmuck/dps_http/api/logs"
)

func (l *TransmissionWindow) AddFrame(pkt *Frame) {
	l.Packets = append(l.Packets, pkt)
	l.BitsProcessed += pkt.Sent_b + pkt.Recv_b
	l.FramesServiced++
	// logs.Debug("Added packet: %s, total size: %s, total count: %d",
	// 	pkt.Source, FormatBits(l.BitsProcessed, 2, 2), l.PacketsServiced)
}

func (l *TransmissionWindow) RemoveFrame(label string) {
	for i, pkt := range l.Packets {
		if pkt.Source == label {
			l.BitsProcessed -= (pkt.Sent_b + pkt.Recv_b)
			l.FramesServiced--
			l.Packets = slices.Delete(l.Packets, i, i+1)
			logs.Debug("Removed packet: %s, total size: %.2f bits, total count: %d", label, l.BitsProcessed, l.FramesServiced)
			return
		}
	}
	logs.Warn("packet not found: %s", label)
}

func (l *TransmissionWindow) String() string {
	return fmt.Sprintf(`
	TransmissionWindow {
		Packets Serviced (p): %d,				// total number of packets serviced
		Bits Processed: %s,					// total size of packets
		Avg Packet Size: %s,					// average size of packets

		Avg Packet Transmission Time: %.5fs, 		// time to transmit single packet to wire
		Total Transmission Time: %.5fs,			// time to transmit all packets back to back

		Queueing Delay: %.5fs,					// (ρ/(μ-λ)) average queueing delay in M/M/1 queueing model
		Processing Delay: %.5fs,				// (1/μ) processing delay in M/M/1 queueing model
		Link Prop Delay: %.5fs,				// pd = (D/S) one way physical link propagation delay in seconds
		RTT: %.5fs,						// (2pd) round trip propagation time

		Average System Time MM1: %.5fs,				// (Wq + 1/μ) average system time in M/M/1 queueing model
		Persistent Service Time: %.5fs,			// persistent connections
		Non Persistent Service Time: %.5fs,			// non-persistent connections
		Packets: %d,						// number of packets in the transmission window
	}`, l.FramesServiced, FormatB(l.BitsProcessed), FormatB(l.AvgPacketSize),
		l.AvgPacketTransmissionTime, l.TotalTransmissionTime,

		l.QueueingDelay, l.ProcessingDelay,
		l.LinkPropDelay, l.RTT,

		l.AverageSystemTimeMM1,
		l.PersistentServiceTime,
		l.NonPersistentServiceTime, len(l.Packets),
	)
}
