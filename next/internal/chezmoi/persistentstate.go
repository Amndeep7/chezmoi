package chezmoi

import (
	"log"
	"os"
)

// A PersistentState is a persistent state.
type PersistentState interface {
	Get(bucket, key []byte) ([]byte, error)
	Delete(bucket, key []byte) error
	ForEach(bucket []byte, fn func(k, v []byte) error) error
	OpenOrCreate() error
	Set(bucket, key, value []byte) error
}

// A debugPersistentState wraps a PersistentState and logs to a log.Logger.
type debugPersistentState struct {
	s      PersistentState
	logger *log.Logger
}

// A dryRunPersistentState wraps a PersistentState and drops all writes but
// records that they occurred.
//
// FIXME mock writes (e.g. writes should not affect the underlying
// PersistentState but subsequent reads should return as if the write occurred)
type dryRunPersistentState struct {
	s        PersistentState
	modified bool
}

// A nullPersistentState is an empty PersistentState that returns the zero value
// for all reads and silently consumes all writes.
type nullPersistentState struct{}

// A readOnlyPersistentState wraps a PeristentState but returns an error on any
// write.
type readOnlyPersistentState struct {
	s PersistentState
}

// newDebugPersistentState returns a new debugPersistentState that wraps s and
// logs to logger.
func newDebugPersistentState(s PersistentState, logger *log.Logger) *debugPersistentState {
	return &debugPersistentState{
		s:      s,
		logger: logger,
	}
}

// Delete implements PersistentState.Delete.
func (s *debugPersistentState) Delete(bucket, key []byte) error {
	return s.debugf("Delete(%q, %q)", []interface{}{string(bucket), string(key)}, func() error {
		return s.s.Delete(bucket, key)
	})
}

// ForEach implements PersistentState.ForEach.
func (s *debugPersistentState) ForEach(bucket []byte, fn func(k, v []byte) error) error {
	return s.debugf("ForEach(%q, _)", []interface{}{string(bucket)}, func() error {
		return s.s.ForEach(bucket, fn)
	})
}

// Get implements PersistentState.Get.
func (s *debugPersistentState) Get(bucket, key []byte) ([]byte, error) {
	var value []byte
	err := s.debugf("Get(%q, %q)", []interface{}{string(bucket), string(key)}, func() error {
		var err error
		value, err = s.s.Get(bucket, key)
		return err
	})
	return value, err
}

// OpenOrCreate implements PersistentState.OpenOrCreate.
func (s *debugPersistentState) OpenOrCreate() error {
	return s.debugf("OpenOrCreate", nil, s.s.OpenOrCreate)
}

// Set implements PersistentState.Set.
func (s *debugPersistentState) Set(bucket, key, value []byte) error {
	return s.debugf("Set(%q, %q, %q)", []interface{}{string(bucket), string(key), string(value)}, func() error {
		return s.s.Set(bucket, key, value)
	})
}

// debugf logs the call to f.
func (s *debugPersistentState) debugf(format string, args []interface{}, f func() error) error {
	err := f()
	if err != nil {
		s.logger.Printf(format+" == %v", append(args, err))
	} else {
		s.logger.Printf(format, args...)
	}
	return err
}

// newDryRunPersistentState returns a new dryRunPersistentState that wraps s.
func newDryRunPersistentState(s PersistentState) *dryRunPersistentState {
	return &dryRunPersistentState{
		s: s,
	}
}

// Get implements PersistentState.Get.
func (s *dryRunPersistentState) Get(bucket, key []byte) ([]byte, error) {
	return s.s.Get(bucket, key)
}

// Delete implements PersistentState.Delete.
func (s *dryRunPersistentState) Delete(bucket, key []byte) error {
	s.modified = true
	return nil
}

// ForEach implements PersistentState.ForEach.
func (s *dryRunPersistentState) ForEach(bucket []byte, fn func(k, v []byte) error) error {
	return s.s.ForEach(bucket, fn)
}

// OpenOrCreate implements PersistentState.OpenOrCreate.
func (s *dryRunPersistentState) OpenOrCreate() error {
	s.modified = true // FIXME this will give false negatives if s.s already exists, need to separate create from open
	return s.s.OpenOrCreate()
}

// Set implements PersistentState.Set.
func (s *dryRunPersistentState) Set(bucket, key, value []byte) error {
	s.modified = true
	// FIXME do we need to remember that the value has been set?
	return nil
}

func (nullPersistentState) Get(bucket, key []byte) ([]byte, error)                  { return nil, nil }
func (nullPersistentState) Delete(bucket, key []byte) error                         { return nil }
func (nullPersistentState) ForEach(bucket []byte, fn func(k, v []byte) error) error { return nil }
func (nullPersistentState) OpenOrCreate() error                                     { return nil }
func (nullPersistentState) Set(bucket, key, value []byte) error                     { return nil }

func newReadOnlyPersistentState(s PersistentState) PersistentState {
	return &readOnlyPersistentState{
		s: s,
	}
}

// Get implements PersistentState.Get.
func (s *readOnlyPersistentState) Get(bucket, key []byte) ([]byte, error) {
	return s.s.Get(bucket, key)
}

// Delete implements PersistentState.Delete.
func (s *readOnlyPersistentState) Delete(bucket, key []byte) error {
	return os.ErrPermission
}

// ForEach implements PersistentState.ForEach.
func (s *readOnlyPersistentState) ForEach(bucket []byte, fn func(k, v []byte) error) error {
	return s.s.ForEach(bucket, fn)
}

// OpenOrCreate implements PersistentState.OpenOrCreate.
func (s *readOnlyPersistentState) OpenOrCreate() error {
	return s.s.OpenOrCreate()
}

// Set implements PersistentState.Set.
func (s *readOnlyPersistentState) Set(bucket, key, value []byte) error {
	return os.ErrPermission
}
