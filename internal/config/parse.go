package config

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

var (
	errNotBoolean    = errors.New("must be either true or false")
	errEmpty         = errors.New("must not be empty")
	errNotOneOf      = errors.New("is not an allowed value")
	errNotURL        = errors.New("must be a valid URL")
	errNotAllowedURI = errors.New("has a scheme that is not allowed")
)

// lookupFunc has the same shape as os.LookupEnv.
type lookupFunc func(key string) (string, bool)

// loader reads typed values and collects every problem, so one startup
// failure reports all misconfigured variables at once.
type loader struct {
	lookup lookupFunc
	errs   []error
}

func (l *loader) fail(name string, err error) {
	l.errs = append(l.errs, fmt.Errorf("%s %w", name, err))
}

// str returns the variable or fallback when it is unset.
func (l *loader) str(name, fallback string) string {
	if v, ok := l.lookup(name); ok {
		return v
	}

	return fallback
}

// oneOf returns the variable (or fallback when unset) if it is in allowed.
func (l *loader) oneOf(name, fallback string, allowed ...string) string {
	v := l.str(name, fallback)
	for _, a := range allowed {
		if v == a {
			return v
		}
	}
	l.fail(name, fmt.Errorf("%w: %s", errNotOneOf, strings.Join(allowed, ", ")))

	return fallback
}

// boolean accepts only "true" or "false".
func (l *loader) boolean(name string, fallback bool) bool {
	raw, ok := l.lookup(name)
	if !ok {
		return fallback
	}

	switch raw {
	case "true":
		return true
	case "false":
		return false
	default:
		l.fail(name, errNotBoolean)

		return fallback
	}
}

// requiredURL returns the variable, which must be set and use one of schemes.
func (l *loader) requiredURL(name string, schemes ...string) string {
	raw, ok := l.lookup(name)
	if !ok || strings.TrimSpace(raw) == "" {
		l.fail(name, errEmpty)

		return ""
	}

	return l.checkURL(name, raw, schemes)
}

// optionalURL is like requiredURL but an unset or empty variable yields "".
func (l *loader) optionalURL(name string, schemes ...string) string {
	raw, ok := l.lookup(name)
	if !ok || strings.TrimSpace(raw) == "" {
		return ""
	}

	return l.checkURL(name, raw, schemes)
}

func (l *loader) checkURL(name, raw string, schemes []string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		l.fail(name, errNotURL)

		return ""
	}
	for _, s := range schemes {
		if u.Scheme == s {
			return raw
		}
	}
	l.fail(
		name,
		fmt.Errorf("%w: %s", errNotAllowedURI, strings.Join(schemes, ", ")),
	)

	return ""
}
