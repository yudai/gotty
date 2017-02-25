package webtty

import (
	"encoding/json"

	"github.com/pkg/errors"
)

// Option is an option for WebTTY.
type Option func(*WebTTY) error

// WithPermitWrite sets a WebTTY to accept input from slaves.
func WithPermitWrite() Option {
	return func(wt *WebTTY) error {
		wt.permitWrite = true
		return nil
	}
}

// WithFixedSize sets a fixed size to TTY master.
func WithFixedSize(width int, height int) Option {
	return func(wt *WebTTY) error {
		wt.width = width
		wt.height = height
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

// WithMasterPreferences sets an optional configuration of master.
func WithMasterPreferences(preferences interface{}) Option {
	return func(wt *WebTTY) error {
		prefs, err := json.Marshal(preferences)
		if err != nil {
			return errors.Wrapf(err, "failed to marshal preferences as JSON")
		}
		wt.masterPrefs = prefs
		return nil
	}
}
