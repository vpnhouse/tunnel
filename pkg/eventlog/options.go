package eventlog

import (
	"time"
)

type options struct {
	// self signed client's tls setup
	// client negotiate against tunnel API to gather CA (self signed)
	SelfSigned bool
	// tunnel port (GRPC)
	TunnelPort string
	// authSecret to sign requests to the server both grpc and http
	AuthSecret string
	// tunnel key to extra check
	// extra check is ommited if empty
	TunnelKey string
	// indicate to use http request to tunnel api (used for debug purposes)
	NoSSL bool
	// Stop idle timeout
	// > 0 the client stops listen events and close connection
	// in case silence (idle) from the server during timeout
	StopIdleTimeout time.Duration
	// Optional tunnel id (used in tracking offset)
	TunnelID string
	// Enforce to start reading from active log disregarding to stored offset
	ForceStartFromActiveLog bool

	// lock ttl timeout (note all client instances mut use same lock ttl)
	// default: time.Minute
	LockTtl time.Duration
	// Extra timeout to prolongate lock ttl
	// default: 30 * time.Second
	LockProlongateTimeout time.Duration
	// report position interval
	// default 5 * time.Second
	ReportPositionInterval time.Duration
	// Wait timeout to output the collected event
	// default 5 * time.Second
	WaitOutputWriteTimeout time.Duration
}

type Option func(opts *options) error

func WithSelfSignedTLS() Option {
	return func(opts *options) error {
		opts.SelfSigned = true
		return nil
	}
}

func WithTunnelPort(port string) Option {
	return func(opts *options) error {
		opts.TunnelPort = port
		return nil
	}
}

func WithNoSSL() Option {
	return func(opts *options) error {
		opts.NoSSL = true
		return nil
	}
}

func WithAuthSecret(authSecret string) Option {
	return func(opts *options) error {
		opts.AuthSecret = authSecret
		return nil
	}
}

func WithStopIdleTimeout(stopIdleTimeout time.Duration) Option {
	return func(opts *options) error {
		opts.StopIdleTimeout = stopIdleTimeout
		return nil
	}
}

func WithStartFromActiveLog() Option {
	return func(opts *options) error {
		opts.ForceStartFromActiveLog = true
		return nil
	}
}

func WithLockTtl(lockTtl time.Duration) Option {
	return func(opts *options) error {
		opts.LockTtl = lockTtl
		return nil
	}
}

func WithLockProlongateTimeout(lockProlongateTimeout time.Duration) Option {
	return func(opts *options) error {
		opts.LockProlongateTimeout = lockProlongateTimeout
		return nil
	}
}

func WithReportPositionInterval(reportPositionInterval time.Duration) Option {
	return func(opts *options) error {
		opts.ReportPositionInterval = reportPositionInterval
		return nil
	}
}

func WithWaitOutputWriteTimeout(waitOutputWriteTimeout time.Duration) Option {
	return func(opts *options) error {
		opts.WaitOutputWriteTimeout = waitOutputWriteTimeout
		return nil
	}
}
