package flock

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

// Flock provides advisory file locking using flock(2).
type Flock struct {
	path string
	fd   int
}

// New creates a new Flock for the given file path.
func New(path string) *Flock {
	return &Flock{path: path, fd: -1}
}

// TryLock attempts to acquire an exclusive lock without blocking.
// Returns (true, nil) if the lock was acquired, (false, nil) if another
// process holds the lock, or (false, err) on unexpected errors.
func (f *Flock) TryLock() (bool, error) {
	fd, err := unix.Open(f.path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return false, err
	}

	err = unix.Flock(fd, unix.LOCK_EX|unix.LOCK_NB)
	if err != nil {
		unix.Close(fd)
		if errors.Is(err, unix.EWOULDBLOCK) {
			return false, nil
		}
		return false, err
	}

	f.fd = fd
	return true, nil
}

// Unlock releases the lock and closes the file descriptor.
func (f *Flock) Unlock() error {
	if f.fd < 0 {
		return nil
	}
	err := unix.Close(f.fd)
	f.fd = -1
	return err
}
