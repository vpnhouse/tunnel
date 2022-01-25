package authorizer

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/Codename-Uranium/tunnel/pkg/xerror"
	"github.com/dgrijalva/jwt-go"
	"github.com/marcw/cachecontrol"
	"go.uber.org/zap"
)

// TODO: Use firebase SDK to verify tokens

type FirebaseKeyKeeper struct {
	lock    sync.RWMutex
	keys    map[string]*rsa.PublicKey
	cancel  context.CancelFunc
	bgSync  sync.WaitGroup
	running bool
}

func NewFirebaseKeyKeeper() (*FirebaseKeyKeeper, error) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	keeper := &FirebaseKeyKeeper{
		cancel:  cancel,
		running: true,
	}

	// warn: misleading comment
	// Do one round on fetch. If failed - repeat it in background.
	refreshInterval, err := keeper.fetch()
	if err != nil {
		return nil, xerror.EInternalError("Can't initialize firebase keys", err)
	}

	keeper.bgSync.Add(1)

	// Pass refresh interval to fetcher routine
	go keeper.fetcher(ctx, refreshInterval)

	return keeper, nil
}

func (keeper *FirebaseKeyKeeper) Shutdown() error {
	keeper.cancel()
	keeper.bgSync.Wait()
	keeper.running = false
	return nil
}

func (keeper *FirebaseKeyKeeper) Running() bool {
	return keeper.running
}

func (keeper *FirebaseKeyKeeper) KeyHelper(token *jwt.Token) (interface{}, error) {
	keeper.lock.RLock()
	defer keeper.lock.RUnlock()

	idValue, ok := token.Header["kid"]
	if !ok {
		return nil, xerror.EAuthenticationFailed("invalid token", fmt.Errorf("firebase key id is not set"))
	}

	var id string
	switch v := idValue.(type) {
	case string:
		id = v
	default:
		return nil, xerror.EAuthenticationFailed("invalid firebase token", fmt.Errorf("unsupported firebase key id type"))
	}
	key, keyFound := keeper.keys[id]
	if !keyFound {
		return nil, xerror.EAuthenticationFailed("unknown firebase key", nil)
	}

	return key, nil
}

func parseKey(keyString string) (*rsa.PublicKey, error) {
	pemData, _ := pem.Decode([]byte(keyString))
	if pemData == nil {
		return nil, xerror.EInternalError("Can't decode firebase PEM data", nil)
	}

	cert, err := x509.ParseCertificate(pemData.Bytes)
	if err != nil {
		return nil, xerror.EInternalError("Can't parse firebase public key", err)
	}

	publicKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, xerror.EInternalError("Firebase certificate is not an RSA key", nil)
	}
	return publicKey, nil
}

func (keeper *FirebaseKeyKeeper) parseKeysFromResponse(response *http.Response) (map[string]*rsa.PublicKey, error) {
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	keys := make(map[string]string)
	err = json.Unmarshal(body, &keys)

	newKeys := make(map[string]*rsa.PublicKey)
	for id, keyString := range keys {
		zap.L().Debug("Received firebase key", zap.String("id", id), zap.String("key", keyString))
		key, err := parseKey(keyString)
		if err != nil {
			return nil, err
		}

		newKeys[id] = key
	}

	return newKeys, nil
}

func (keeper *FirebaseKeyKeeper) parseRefreshIntervalFromResponse(response *http.Response) time.Duration {
	ccString := response.Header.Get("Cache-Control")
	if ccString == "" {
		zap.L().Warn("no cache-control header from firebase")
		return defaultFirebaseRefreshInterval
	}

	cc := cachecontrol.Parse(ccString)
	maxAge := cc.MaxAge()
	if maxAge <= time.Duration(0) {
		zap.L().Warn("failed to parse cache-control header", zap.String("raw", ccString))
		return defaultFirebaseRefreshInterval
	}

	// Common practice is to refresh at lifetime/2
	return maxAge / 2
}

func (keeper *FirebaseKeyKeeper) fetcher(ctx context.Context, refreshInterval time.Duration) {
	defer keeper.bgSync.Done()

	t := time.NewTimer(refreshInterval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			nextInterval, err := keeper.fetch()
			if err != nil {
				nextInterval = defaultFirebaseRefreshInterval
			}

			t.Reset(nextInterval)
		}
	}
}

func (keeper *FirebaseKeyKeeper) fetch() (time.Duration, error) {
	response, err := http.Get("https://www.googleapis.com/robot/v1/metadata/x509/securetoken@system.gserviceaccount.com")
	if err != nil {
		return 0, xerror.EInternalError("failed to fetch keys from firebase", err)
	}

	newKeys, err := keeper.parseKeysFromResponse(response)
	if err != nil {
		return 0, xerror.EInternalError("failed to parse firebase keys", err)
	}

	refreshInterval := keeper.parseRefreshIntervalFromResponse(response)

	keeper.lock.Lock()
	defer keeper.lock.Unlock()
	keeper.keys = newKeys

	return refreshInterval, nil
}
