package conv

const (
	// B represents a byte.
	B = 1
	// KB represents a kilobyte (1000 bytes).
	KB = B * 1000
	// MB represents a megabyte (1000 kB).
	MB = KB * 1000
	// GB represents a gigabyte (1000 MB).
	GB = MB * 1000
	// TB represents a terabyte (1000 GB).
	TB = GB * 1000
	// PB represents a petabyte (1000 TB).
	PB = TB * 1000
	// EB represents an exabyte (1000 PB).
	EB = PB * 1000
)

const byteShift = 10
const (
	_ = 1 << (byteShift * iota)
	// KiB represents a kibibyte (1024 bytes).
	KiB
	// MiB represents a mebibyte (1024 KiB).
	MiB
	// GiB represents a gibibyte (1024 MiB).
	GiB
	// TiB represents a tebibyte (1024 GiB).
	TiB
	// PiB represents a pebibyte (1024 TiB).
	PiB
	// EiB represents an exbibyte (1024 PiB).
	EiB
)
