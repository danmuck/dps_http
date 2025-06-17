package networks

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
