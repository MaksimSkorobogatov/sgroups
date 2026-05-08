package patterns

import (
	"github.com/H-BF/corlib/pkg/patterns/observer"
)

type (
	// Observer is an alias
	Observer = observer.Observer

	// EventType is an alias
	EventType = observer.EventType

	// EventReceiver is an alias
	EventReceiver = observer.EventReceiver

	// Subject is an alias
	Subject = observer.Subject

	// NoNotifySubject is an alias
	NoNotifySubject = observer.NotNotifiableSubject

	// NullSubject is an alias
	NullSubject = observer.NullSubject

	// NullObserver it does nothing
	NullObserver = observer.NullObserver
)

var (
	// NewSubject creates an Subject instance
	NewSubject = observer.NewSubject

	// NewObserver -
	NewObserver = observer.NewObserver
)
