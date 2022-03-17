// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package eventlog

import (
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/vpnhouse/tunnel/pkg/xerror"
	"github.com/google/uuid"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

const (
	dirFileNameTemplate = "%d_%d_%s"
)

// namedFile is how the fsStorage stores its files.
// Actual file name on a disk must contain sequential log number,
// creation timestamp, and logID for a caller ($number_$timestamp_$uuid).
type namedFile struct {
	seq       int
	timestamp int64
	uuid      string
}

func (m namedFile) String() string {
	return fmt.Sprintf(dirFileNameTemplate, m.seq, m.timestamp, m.uuid)
}

func newDirFile(seq int) namedFile {
	return namedFile{
		seq:       seq,
		timestamp: time.Now().Unix(),
		uuid:      uuid.New().String(),
	}
}

type StorageConfig struct {
	// path to a log dir
	Dir string `json:"dir"`
	// number of files to maintain
	MaxFiles int `json:"max_files"`
	// for how long we want to write to a single logfile
	Period time.Duration `json:"period"`
	// how many bytes we want to write to a single logfile
	Size int64 `json:"size"`
}

// fsStorage implements logs storage on fs.
type fsStorage struct {
	// lock guards the struct fields
	lock sync.Mutex

	// currentLog is the active log file
	currentLog namedFile
	// current log fd for *writing* only,
	// each reader must maintain its own read-only fd
	// of the particular piece of logs.
	currentFD      WriteSyncCloser
	currentWritten int64
	currentBtime   time.Time

	// rotated logs lives here,
	// ordered from older to newer ones.
	rotated []namedFile

	config StorageConfig

	// keep the fs instance per storage to be able
	// to run tests in parallel
	_fs afero.Fs
}

func newFsStorage(cfg StorageConfig, fss ...afero.Fs) (*fsStorage, error) {
	var files []namedFile

	var filesys afero.Fs
	if len(fss) > 0 {
		filesys = fss[0]
	} else {
		filesys = afero.NewOsFs()
	}

	err := afero.Walk(filesys, cfg.Dir, func(path string, d fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		f, err := parseLogFileName(d.Name())
		if err != nil {
			return err
		}

		files = append(files, f)
		return nil
	})

	if err != nil {
		return nil, err
	}

	// since Walk returns files in lexical order,
	// we have to explicitly sort them by the number.
	sort.Slice(files, func(i, j int) bool {
		return files[i].seq < files[j].seq
	})

	if err := validateFilesSequence(files); err != nil {
		return nil, err
	}

	if len(files) == 0 {
		files = append(files, newDirFile(1))
	}

	currentLog := files[len(files)-1]
	fd, size, err := openWriteOnly(filesys, filepath.Join(cfg.Dir, currentLog.String()))
	if err != nil {
		return nil, err
	}

	return &fsStorage{
		currentLog:     currentLog,
		currentFD:      fd,
		currentWritten: size,
		currentBtime:   time.Unix(currentLog.timestamp, 0),
		rotated:        alterFileIndexSize(files[:len(files)-1], cfg.MaxFiles),
		config:         cfg,
		_fs:            filesys,
	}, nil
}

func (ds *fsStorage) OpenLog(logID string, offset int64) (io.ReadCloser, error) {
	ds.lock.Lock()
	defer ds.lock.Unlock()

	file, err := ds.lookup(logID)
	if err != nil {
		return nil, err
	}

	path := filepath.Join(ds.config.Dir, file.String())
	return openReadOnly(ds._fs, path, offset)
}

func (ds *fsStorage) lookup(logID string) (namedFile, error) {
	if ds.currentLog.uuid == logID {
		return ds.currentLog, nil
	}

	for _, file := range ds.rotated {
		if file.uuid == logID {
			return file, nil
		}
	}
	return namedFile{}, fmt.Errorf("unknown file")
}

// write given buf to the file.
// It will optionally rotate the file if limits exceeded.
func (ds *fsStorage) Write(buf []byte) error {
	n, err := ds.currentFD.Write(buf)
	if err != nil {
		return xerror.EStorageError("failed to write to the underlying file",
			err, zap.Stringer("log_file", ds.currentLog),
			zap.Int64("at_offset", ds.currentWritten))
	}

	ds.currentWritten += int64(n)

	if ds.mustRotate() {
		ds.rotateCurrentLog()
	}

	return nil
}

// mustRotate returns true if we have to rotate currentFD.
func (ds *fsStorage) mustRotate() bool {
	lifetime := time.Now().UTC().Sub(ds.currentBtime)

	if ds.config.Size > 0 && ds.currentWritten >= ds.config.Size {
		zap.L().Info("must rotate current log (size)",
			zap.Int64("current_write", ds.currentWritten),
			zap.Int64("max_size", ds.config.Size))
		return true
	}

	if ds.config.Period > 0 && lifetime > ds.config.Period {
		zap.L().Info("must rotate current log (ttl)", zap.Duration("lifetime", lifetime))
		return true
	}

	return false
}

func (ds *fsStorage) generateNextLogName() namedFile {
	return namedFile{
		seq:       ds.currentLog.seq + 1,
		timestamp: time.Now().Unix(),
		uuid:      uuid.New().String(),
	}
}

func (ds *fsStorage) rotateCurrentLog() {
	// try open file first, in case of error
	// we can still write to the existing file
	// and retry on a next write() call.

	nextLog := ds.generateNextLogName()
	nextlogPath := filepath.Join(ds.config.Dir, nextLog.String())
	// drop the size here since it is always zero for new files.
	fd, _, err := openWriteOnly(ds._fs, nextlogPath)
	if err != nil {
		zap.L().Error("failed to open next log", zap.Error(err), zap.String("path", nextlogPath))
		return
	}

	// this lock is for log subscribers,
	// the Push() method will continue collecting
	// events in the buffered chan while the
	// rotation is in progress
	ds.lock.Lock()
	defer ds.lock.Unlock()

	_ = ds.currentFD.Sync()
	_ = ds.currentFD.Close()

	// we have some slot allocated, move current log to a rotated list.
	// len(ds.rotated) == 0 means that we got the MaxFiles=0 option,
	// so no older log stored and managed.
	if len(ds.rotated) > 0 {
		ds.rotated = append(ds.rotated[1:], ds.currentLog)
	}

	ds.currentLog = nextLog
	ds.currentFD = fd
	ds.currentWritten = 0
	ds.currentBtime = time.Now().UTC()

	zap.L().Debug("next log allocated", zap.String("log_id", nextLog.uuid))
}

// HasLog returns true if log we manage a file with a given ID.
func (ds *fsStorage) HasLog(logID string) bool {
	ds.lock.Lock()
	defer ds.lock.Unlock()

	if _, err := ds.lookup(logID); err != nil {
		return false
	}
	return true
}

// FirstLog returns the name of the earliest known logs,
// if nothing was rotated so far, it returns the active file.
func (ds *fsStorage) FirstLog() string {
	ds.lock.Lock()
	defer ds.lock.Unlock()

	for _, file := range ds.rotated {
		if len(file.uuid) > 0 {
			return file.uuid
		}
	}

	return ds.currentLog.uuid
}

// NextLog returns next log to read from.
// Returns an error if no next log for a given log can be found.
// This may happen for very slow clients, who reads the very first log
// longer than we write and rotate MaxFiles number of logs.
func (ds *fsStorage) NextLog(current string) (string, error) {
	ds.lock.Lock()
	defer ds.lock.Unlock()

	if current == ds.currentLog.uuid {
		return current, nil
	}

	for i, file := range ds.rotated {
		if file.uuid == current {
			if i+1 == len(ds.rotated) {
				return ds.currentLog.uuid, nil
			}
			return ds.rotated[i+1].uuid, nil
		}
	}

	return "", fmt.Errorf("unable to find next log: given logID is too old")
}

func (ds *fsStorage) Close() {
	_ = ds.currentFD.Sync()
	_ = ds.currentFD.Close()
}

func parseLogFileName(s string) (namedFile, error) {
	var f namedFile
	n, err := fmt.Sscanf(s, dirFileNameTemplate, &f.seq, &f.timestamp, &f.uuid)
	if err != nil {
		return namedFile{}, fmt.Errorf("invalid log name `%s`: expected name template is `$num_$timestamp_$uuid`", s)
	}

	if n != 3 {
		return namedFile{}, fmt.Errorf("invalid log name `%s`: expected name template is `$num_$timestamp_$uuid`", s)
	}

	if _, err := uuid.Parse(f.uuid); err != nil {
		return namedFile{}, fmt.Errorf("invalid log name: failed to parse UUID from `%s`", s)
	}

	return f, nil
}

func validateFilesSequence(files []namedFile) error {
	if len(files) == 0 {
		return nil
	}

	seq := files[0].seq
	for _, file := range files {
		if file.seq != seq {
			return fmt.Errorf("invalid sequence at `%s`, expecting %d", file, seq)
		}
		seq++
	}

	return nil
}

// alterFileIndexSize shrinks or expands given slice
// according to the maxFiles value from a config.
func alterFileIndexSize(given []namedFile, maxFiles int) []namedFile {
	if len(given) > maxFiles {
		k := len(given) - maxFiles
		return given[k:]
	}
	if maxFiles > len(given) {
		k := maxFiles - len(given)
		return append(make([]namedFile, k), given...)
	}
	return given
}
