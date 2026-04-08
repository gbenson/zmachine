package logfollower

type Receiver interface {
	Receive(*Record)
}

// The ReceiverFunc type allows the use of ordinary functions as
// [Receiver]s. If f is a function with the appropriate signature,
// ReceiverFunc(f) is a [Receiver] that calls f.
type ReceiverFunc func(*Record)

// [Receive] implements [Receiver].
func (f ReceiverFunc) Receive(r *Record) {
	f(r)
}
