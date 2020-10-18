package config

// MiscConfiguration configuration:
// configuration proxy etc
// TimeOutSeed: auto delete torrents from client after X hours
type MiscConfiguration struct {
	Proxy  string // proxy:port
	Active bool
}
