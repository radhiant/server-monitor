package collector

import (
	"sync"
	"time"

	netutil "github.com/shirou/gopsutil/v4/net"

	"server-monitor/backend/internal/model"
)

type ifacePrevState struct {
	rxBytes   uint64
	txBytes   uint64
	rxPackets uint64
	txPackets uint64
}

// NetworkCollector tracks per-interface delta transfer rates.
type NetworkCollector struct {
	mu          sync.Mutex
	lastTime    time.Time
	prevStates  map[string]ifacePrevState
	initialized bool
}

// NewNetworkCollector creates an initialized NetworkCollector.
func NewNetworkCollector() *NetworkCollector {
	return &NetworkCollector{
		prevStates: make(map[string]ifacePrevState),
	}
}

// Collect gathers interface statistics and computes delta rates per second.
func (nc *NetworkCollector) Collect() []model.NetworkInterface {
	nc.mu.Lock()
	defer nc.mu.Unlock()

	now := time.Now()
	counters, err := netutil.IOCounters(true)
	if err != nil {
		return []model.NetworkInterface{}
	}

	deltaSec := now.Sub(nc.lastTime).Seconds()
	if !nc.initialized || deltaSec <= 0 {
		deltaSec = 1.0
	}

	result := make([]model.NetworkInterface, 0, len(counters))

	for _, c := range counters {
		// Filter out empty or loopback with zero activity if necessary, but keep primary interfaces
		var rxRate, txRate, rxPktRate, txPktRate float64

		if nc.initialized {
			if prev, exists := nc.prevStates[c.Name]; exists {
				if c.BytesRecv >= prev.rxBytes {
					rxRate = float64(c.BytesRecv-prev.rxBytes) / deltaSec
				}
				if c.BytesSent >= prev.txBytes {
					txRate = float64(c.BytesSent-prev.txBytes) / deltaSec
				}
				if c.PacketsRecv >= prev.rxPackets {
					rxPktRate = float64(c.PacketsRecv-prev.rxPackets) / deltaSec
				}
				if c.PacketsSent >= prev.txPackets {
					txPktRate = float64(c.PacketsSent-prev.txPackets) / deltaSec
				}
			}
		}

		nc.prevStates[c.Name] = ifacePrevState{
			rxBytes:   c.BytesRecv,
			txBytes:   c.BytesSent,
			rxPackets: c.PacketsRecv,
			txPackets: c.PacketsSent,
		}

		result = append(result, model.NetworkInterface{
			Name:         c.Name,
			RxBytesSec:   rxRate,
			TxBytesSec:   txRate,
			RxPacketsSec: rxPktRate,
			TxPacketsSec: txPktRate,
			TotalRxBytes: c.BytesRecv,
			TotalTxBytes: c.BytesSent,
		})
	}

	nc.lastTime = now
	nc.initialized = true

	return result
}
