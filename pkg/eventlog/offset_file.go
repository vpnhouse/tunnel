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

type offsetSyncFile struct {
	directory string
}

func NewOffsetSyncFile(directory string) (*offsetSyncFile, error) {
	err := os.MkdirAll(directory, 777)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure the direstory %s: %w", directory, err)
	}

	return &offsetSyncFile{
		directory: directory,
	}, nil
}

func (s *offsetSyncFile) Acquire(instanceID string, tunnelID string, ttl time.Duration) (bool, error) {
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

func (s *offsetSyncFile) Release(instanceID string, tunnelID string) error {
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

func (s *offsetSyncFile) GetOffset(tunnelID string) (Offset, error) {
	offsetFile := s.buildOffsetFile(tunnelID)
	stats, err := os.Stat(offsetFile)
	if errors.Is(err, os.ErrNotExist) {
		return Offset{}, err
	}
	if time.Now().Sub(stats.ModTime()) > offsetKeepTimeout {
		_ = os.RemoveAll(offsetFile)
	}
	data, err := os.ReadFile(offsetFile)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return Offset{TunnelID: tunnelID}, nil
	}

	return offsetFromJson(string(data))
}

func (s *offsetSyncFile) PutOffset(offset Offset) error {
	offsetFile := s.buildOffsetFile(offset.TunnelID)
	data := offset.ToJson()
	return os.WriteFile(offsetFile, []byte(data), 0600)
}

func (s *offsetSyncFile) lock(tunnelID string) (*lockState, error) {
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

func (s *offsetSyncFile) tryLock(tunnelID string) (*lockState, error) {
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

func (s *offsetSyncFile) buildSyncLockFile(value string) string {
	return path.Join(s.directory, fmt.Sprintf("eventlogs.lock.%s", base64.RawStdEncoding.EncodeToString([]byte(value))))
}

func (s *offsetSyncFile) buildSyncFile(value string) string {
	return path.Join(s.directory, fmt.Sprintf("eventlogs.sync.%s", base64.RawStdEncoding.EncodeToString([]byte(value))))
}

func (s *offsetSyncFile) buildOffsetFile(value string) string {
	return path.Join(s.directory, fmt.Sprintf("eventlogs.offset.%s", base64.RawStdEncoding.EncodeToString([]byte(value))))
}
