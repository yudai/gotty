package webtty

// Option is an option for WebTTY.
type Option func(*WebTTY) error

// WithFixedColumns sets a fixed width to TTY master.
func WithFixedColumns(columns int) Option {
	return func(wt *WebTTY) error {
		wt.columns = columns
		return nil
	}
}

// WithFixedRows sets a fixed height to TTY master.
func WithFixedRows(rows int) Option {
	return func(wt *WebTTY) error {
		wt.rows = rows
		return nil
	}
}

// WithWindowTitle sets the default window title of the session
func WithWindowTitle(windowTitle []byte) Option {
	return func(wt *WebTTY) error {
		wt.windowTitle = windowTitle
		return nil
	}
}

// WithReconnect enables reconnection on the master side.
func WithReconnect(timeInSeconds int) Option {
	return func(wt *WebTTY) error {
		wt.reconnect = timeInSeconds
		return nil
	}
}
