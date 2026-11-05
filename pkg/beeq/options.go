package beeq

import (
	"time"

	"github.com/panjf2000/gnet/v2"
)

const (
	MQTT31  = 3
	MQTT311 = 4
	MQTT50  = 5
)

type Options struct {
	MaxPacketSize        int
	MaxConnPending       int
	KeepAliveTimeout     time.Duration
	ReusePort            bool
	TCPNoDelay           bool
	TCPKeepAlive         bool
	TCPKeepAlivePeriod   time.Duration
	Multicore            bool
	EnableTicker         bool
	SupportedProtocols   []byte
	BlacklistMaxAttempts int
	BlacklistBlockTime   time.Duration
	CleanSession         bool
}

func DefaultOptions() Options {
	return Options{
		MaxPacketSize:        64 * 1024,
		MaxConnPending:       1024 * 1024,
		KeepAliveTimeout:     30 * time.Second,
		ReusePort:            false,
		TCPNoDelay:           true,
		TCPKeepAlive:         true,
		TCPKeepAlivePeriod:   60 * time.Second,
		Multicore:            true,
		EnableTicker:         true,
		SupportedProtocols:   []byte{MQTT31, MQTT311, MQTT50},
		BlacklistMaxAttempts: 10,
		BlacklistBlockTime:   60 * time.Second,
		CleanSession:         false,
	}
}

func (opts Options) toGnetOptions() []gnet.Option {
	gnetOpts := []gnet.Option{}

	if opts.Multicore {
		gnetOpts = append(gnetOpts, gnet.WithMulticore(true))
	}
	if opts.EnableTicker {
		gnetOpts = append(gnetOpts, gnet.WithTicker(true))
	}
	if opts.ReusePort {
		gnetOpts = append(gnetOpts, gnet.WithReusePort(true))
	}
	if opts.TCPNoDelay {
		gnetOpts = append(gnetOpts, gnet.WithTCPNoDelay(gnet.TCPNoDelay))
	}
	if opts.TCPKeepAlive {
		gnetOpts = append(gnetOpts, gnet.WithTCPKeepAlive(opts.TCPKeepAlivePeriod))
	}

	return gnetOpts
}

func (opts Options) isProtocolSupported(version byte) bool {
	for _, v := range opts.SupportedProtocols {
		if v == version {
			return true
		}
	}
	return false
}
