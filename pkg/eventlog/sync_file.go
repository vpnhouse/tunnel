//go:build darwin || linux
// +build darwin linux

package eventlog

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path"
	"syscall"
	"time"
)

var lockOperationTimeout = time.Millisecond
var lockWaitTimeout = time.Second
var errLocked = errors.New("locked")

type lockState struct {
	File string
	FD   int
}

func (s *lockState) Close() error {
	err := syscall.Close(s.FD)
	if err != nil {
		return err
	}
	return syscall.Unlink(s.File)
}

type eventlogSyncFile struct {
	directory string
}

func NewEventlogSyncFile(directory string) (*eventlogSyncFile, error) {
	err := os.MkdirAll(directory, 0777)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure the direstory %s: %w", directory, err)
	}

	return &eventlogSyncFile{
		directory: directory,
	}, nil
}

func (s *eventlogSyncFile) Acquire(instanceID string, tunnelID string, ttl time.Duration) (bool, error) {
	lock, err := s.tryLock(tunnelID)
	if err != nil {
		if errors.Is(err, errLocked) {
			return false, nil
		}
		return false, err
	}
	defer lock.Close()

	file := s.buildSyncFile(tunnelID)

	data, err := os.ReadFile(file)

	if err == nil {
		if string(data) != instanceID {
			fs, err := os.Stat(file)
			if err != nil {
				return false, err
			}
			if fs.ModTime().UTC().Add(ttl).After(time.Now().UTC()) {
				return false, nil
			}
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return false, err
	}

	err = os.WriteFile(file, []byte(instanceID), 0600)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *eventlogSyncFile) Release(instanceID string, tunnelID string) error {
	lock, err := s.tryLock(tunnelID)
	if err != nil {
		if errors.Is(err, errLocked) {
			return nil
		}
		return err
	}
	defer lock.Close()

	file := s.buildSyncFile(tunnelID)

	data, err := os.ReadFile(file)

	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if string(data) != instanceID {
		return nil
	}

	return os.Remove(file)
}

func (s *eventlogSyncFile) GetPosition(tunnelID string) (Position, error) {
	posFile := s.buildPositionFile(tunnelID)
	stats, err := os.Stat(posFile)
	if errors.Is(err, os.ErrNotExist) {
		return Position{}, ErrPositionNotFound
	}
	if time.Now().Sub(stats.ModTime()) > offsetKeepTimeout {
		_ = os.RemoveAll(posFile)
	}
	data, err := os.ReadFile(posFile)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return Position{}, ErrPositionNotFound
	}

	return positionFromJson(string(data))
}

func (s *eventlogSyncFile) PutPosition(tunnelID string, position Position) error {
	posFile := s.buildPositionFile(tunnelID)
	data := position.ToJson()
	return os.WriteFile(posFile, []byte(data), 0600)
}

func (s *eventlogSyncFile) DeletePosition(tunnelID string) error {
	offsetFile := s.buildPositionFile(tunnelID)
	return os.RemoveAll(offsetFile)
}

func (s *eventlogSyncFile) lock(tunnelID string) (*lockState, error) {
	lockFile := s.buildSyncLockFile(tunnelID)

	// Check and remove lock file in case timout is over
	if fs, err := os.Stat(lockFile); err == nil {
		if fs.ModTime().UTC().Add(lockOperationTimeout).Before(time.Now().UTC()) {
			err := os.RemoveAll(lockFile)
			if err != nil {
				return nil, err
			}
		}
	}

	fd, err := syscall.Open(lockFile, syscall.O_CREAT|syscall.O_RDONLY, 0600)
	if err != nil {
		return nil, err
	}

	err = syscall.Flock(fd, syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		syscall.Close(fd)
	}
	if err == syscall.EWOULDBLOCK {
		return nil, errLocked
	}
	return &lockState{
		File: lockFile,
		FD:   fd,
	}, nil
}

func (s *eventlogSyncFile) tryLock(tunnelID string) (*lockState, error) {
	t := time.NewTimer(lockWaitTimeout)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			return nil, errLocked
		default:
			lock, err := s.lock(tunnelID)
			if err != nil {
				if errors.Is(err, errLocked) {
					time.Sleep(lockOperationTimeout)
					continue
				}
				return nil, err
			}
			return lock, nil
		}
	}
}

func (s *eventlogSyncFile) buildSyncLockFile(value string) string {
	return path.Join(s.directory, fmt.Sprintf("eventlog.lock.%s", base64.RawStdEncoding.EncodeToString([]byte(value))))
}

func (s *eventlogSyncFile) buildSyncFile(value string) string {
	return path.Join(s.directory, fmt.Sprintf("eventlog.sync.%s", base64.RawStdEncoding.EncodeToString([]byte(value))))
}

func (s *eventlogSyncFile) buildPositionFile(value string) string {
	return path.Join(s.directory, fmt.Sprintf("eventlog.position.%s", base64.RawStdEncoding.EncodeToString([]byte(value))))
}
