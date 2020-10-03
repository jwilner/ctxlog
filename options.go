package ctxlog

// OptDebug filters the output to the debug level or above.
func OptDebug(s *settings) {
	s.filter = levelDebug
}

// OptInfo filters the output to the info level or above.
func OptInfo(s *settings) {
	s.filter = levelInfo
}

// OptWarn filters the output to the warn level or above.
func OptWarn(s *settings) {
	s.filter = levelWarn
}

// OptError filters the output to the error level.
func OptError(s *settings) {
	s.filter = levelError
}

// OptTime includes the time in the output, formatted according to time.RFC3339Nano
func OptTime(s *settings) {
	s.flags |= flagTime
}

// OptCaller includes the caller in the output.
func OptCaller(long bool) Option {
	set, unset := flagCallerShort, flagCallerLong
	if long {
		set, unset = flagCallerLong, flagCallerShort
	}
	return func(s *settings) {
		s.flags |= set
		s.flags &^= unset
	}
}

// Option controls the output of a logger
type Option func(s *settings)
