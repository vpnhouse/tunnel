package eventlog

type subscribeOptions struct {
	ActiveLog           bool
	Position            EventlogPosition
	SkipEventAtPosition bool // Skip event at position (i.e. not publish it)
}

type SubscribeOption func(opts *subscribeOptions) error

// Started from Active log disregarding the Position
func WithActiveLog() SubscribeOption {
	return func(opts *subscribeOptions) error {
		opts.ActiveLog = true
		return nil
	}
}

// Used only if ActiveLog is not defined
func WithPosition(position EventlogPosition) SubscribeOption {
	return func(opts *subscribeOptions) error {
		err := position.validate()
		if err != nil {
			return err
		}
		opts.Position = position
		return nil
	}
}

func WithSkipEventAtPosition(skip bool) SubscribeOption {
	return func(opts *subscribeOptions) error {
		opts.SkipEventAtPosition = skip
		return nil
	}
}
