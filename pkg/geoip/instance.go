package geoip

import (
	"net"

	"github.com/oschwald/maxminddb-golang"
	"github.com/vpnhouse/tunnel/pkg/xerror"
	"go.uber.org/zap"
)

type Instance struct {
	dbCountry *maxminddb.Reader
	running   bool
}

func NewGeoip(path string) (*Instance, error) {
	db, err := maxminddb.Open(path)
	if err != nil || db == nil {
		return nil, xerror.EInternalError("can't open maxminddb country database", err, zap.String("path", path))
	}

	return &Instance{
		dbCountry: db,
		running:   true,
	}, nil
}

func (instance *Instance) GetCountry(ip net.IP) (string, error) {
	var record struct {
		Country struct {
			ISOCode string `maxminddb:"iso_code"`
		} `maxminddb:"country"`
	}

	err := instance.dbCountry.Lookup(ip, &record)
	if err != nil {
		return "", xerror.EInternalError("can't lookup country", err)
	}

	return record.Country.ISOCode, nil
}

func (instance *Instance) Shutdown() error {
	err := instance.dbCountry.Close()
	if err != nil {
		return xerror.EInternalError("can't close maxminddb country database", err)
	}
	instance.running = false
	return nil
}

func (instance *Instance) Running() bool {
	return instance.running
}
