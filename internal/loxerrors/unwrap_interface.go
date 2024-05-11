package loxerrors

// local interface to be used with errors.Unwrap().
// errors packake does not define separate interface, relies on reflection instead.
type unwrapInterface interface {
	Unwrap() error
}
