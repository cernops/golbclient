package network

import "time"

// NetScanner : Network interface contract
type NetScanner interface {
	Run(time.Duration)
}
