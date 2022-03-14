// Copyright 2021 The Uranium Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package xerror

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"

	openapi "github.com/comradevpn/api/go/server/common"
	"github.com/comradevpn/tunnel/pkg/version"
	"github.com/getsentry/sentry-go"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var errorByCodeCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: "errors",
	Name:      "by_code",
	Help:      "number of errors partitioned by code, label, and version info",
}, []string{"code_name", "label", "tag", "commit", "feature", "caller"})

func init() {
	prometheus.MustRegister(errorByCodeCounter)
}

type ErrorType struct {
	httpCode int
	codeName openapi.ErrorResult
}

var (
	EInternalErrorType         = &ErrorType{http.StatusInternalServerError, openapi.ErrorResultINTERNALERROR}
	EInvalidArgumentType       = &ErrorType{http.StatusBadRequest, openapi.ErrorResultINVALIDARGUMENT}
	EEntryNotFoundType         = &ErrorType{http.StatusNotFound, openapi.ErrorResultNOTFOUND}
	EExistsType                = &ErrorType{http.StatusConflict, openapi.ErrorResultENTRYEXISTS}
	EStorageErrorType          = &ErrorType{http.StatusInternalServerError, openapi.ErrorResultSTORAGEERROR}
	ETunnelErrorType           = &ErrorType{http.StatusInternalServerError, openapi.ErrorResultTUNNELERROR}
	EUnauthorizedType          = &ErrorType{http.StatusUnauthorized, openapi.ErrorResultUNAUTHORIZED}
	EAuthenticationFailedType  = &ErrorType{http.StatusUnauthorized, openapi.ErrorResultAUTHFAILED}
	ENotEnoughSpaceType        = &ErrorType{http.StatusInsufficientStorage, openapi.ErrorResultINSUFFICIENTSTORAGE}
	EUnavailableType           = &ErrorType{http.StatusServiceUnavailable, openapi.ErrorResultSERVICEUNAVAILABLE}
	EConfigurationRequiredType = &ErrorType{http.StatusConflict, openapi.ErrorResultCONFIGURATIONREQUIRED}
	EForbiddenType             = &ErrorType{http.StatusForbidden, openapi.ErrorResultFORBIDDEN}
	EInvalidConfigurationType  = &ErrorType{http.StatusInternalServerError, openapi.ErrorResultINVALIDCONFIGURATION}
)

func EInternalError(description string, err error, fields ...zap.Field) *Error {
	return newError(EInternalErrorType, description, secretiveSerializer, err, nil, fields...)
}

func WInternalError(label, description string, err error, fields ...zap.Field) *Error {
	return newWarning(EInternalErrorType, description, secretiveSerializer, err, nil, label, fields...)
}

func EInvalidArgument(description string, err error, fields ...zap.Field) *Error {
	return newError(EInvalidArgumentType, description, defaultSerializer, err, nil, fields...)
}

func WInvalidArgument(label, description string, err error, fields ...zap.Field) *Error {
	return newWarning(EInvalidArgumentType, description, defaultSerializer, err, nil, label, fields...)
}

func EInvalidField(description string, failedField string, err error, fields ...zap.Field) *Error {
	return newError(EInvalidArgumentType, description, defaultSerializer, err, &failedField, fields...)
}

func WInvalidField(label, description string, failedField string, err error, fields ...zap.Field) *Error {
	return newWarning(EInvalidArgumentType, description, defaultSerializer, err, &failedField, label, fields...)
}

func EEntryNotFound(description string, err error, fields ...zap.Field) *Error {
	return newError(EEntryNotFoundType, description, defaultSerializer, err, nil, fields...)
}

func WEntryNotFound(label, description string, err error, fields ...zap.Field) *Error {
	return newWarning(EEntryNotFoundType, description, defaultSerializer, err, nil, label, fields...)
}

func EExists(description string, err error, fields ...zap.Field) *Error {
	return newError(EExistsType, description, defaultSerializer, err, nil, fields...)
}

func WExists(label, description string, err error, fields ...zap.Field) *Error {
	return newWarning(EExistsType, description, defaultSerializer, err, nil, label, fields...)
}

func EStorageError(description string, err error, fields ...zap.Field) *Error {
	return newError(EStorageErrorType, description, defaultSerializer, err, nil, fields...)
}

func WStorageError(label, description string, err error, fields ...zap.Field) *Error {
	return newWarning(EStorageErrorType, description, defaultSerializer, err, nil, label, fields...)
}

func ETunnelError(description string, err error, fields ...zap.Field) *Error {
	return newError(ETunnelErrorType, description, defaultSerializer, err, nil, fields...)
}

func WTunnelError(label, description string, err error, fields ...zap.Field) *Error {
	return newWarning(ETunnelErrorType, description, defaultSerializer, err, nil, label, fields...)
}

func EUnauthorized(description string, err error, fields ...zap.Field) *Error {
	return newError(EUnauthorizedType, description, secretiveSerializer, err, nil, fields...)
}

func WUnauthorized(label, description string, err error, fields ...zap.Field) *Error {
	return newWarning(EUnauthorizedType, description, secretiveSerializer, err, nil, label, fields...)
}

func EAuthenticationFailed(description string, err error, fields ...zap.Field) *Error {
	return newError(EAuthenticationFailedType, description, secretiveSerializer, err, nil, fields...)
}

func WAuthenticationFailed(label, description string, err error, fields ...zap.Field) *Error {
	return newWarning(EAuthenticationFailedType, description, secretiveSerializer, err, nil, label, fields...)
}

func ENotEnoughSpace(description string, err error, fields ...zap.Field) *Error {
	return newError(ENotEnoughSpaceType, description, defaultSerializer, err, nil, fields...)
}

func WNotEnoughSpace(label, description string, err error, fields ...zap.Field) *Error {
	return newWarning(ENotEnoughSpaceType, description, defaultSerializer, err, nil, label, fields...)
}

func EUnavailable(description string, err error, fields ...zap.Field) *Error {
	return newError(EUnavailableType, description, defaultSerializer, err, nil, fields...)
}

func WUnavailable(label, description string, err error, fields ...zap.Field) *Error {
	return newWarning(EUnavailableType, description, defaultSerializer, err, nil, label, fields...)
}

func EConfigurationRequired(msg string) *Error {
	return newError(EConfigurationRequiredType, msg, defaultSerializer, nil, nil)
}

func EForbidden(msg string) *Error {
	return newError(EForbiddenType, msg, defaultSerializer, nil, nil)
}

func EInvalidConfiguration(msg string, field string) *Error {
	return newError(EInvalidConfigurationType, msg, defaultSerializer, nil, &field)
}

type errorSerializerFunc func(error *Error) (int, []byte)

func marshalError(oError *openapi.Error) []byte {
	j, err := json.MarshalIndent(oError, "", "  ")
	if err != nil {
		zap.L().Fatal("can't marshal error", zap.Any("oError", oError), zap.Error(err))
	}

	return j
}

func defaultSerializer(err *Error) (int, []byte) {
	oError := &openapi.Error{
		Result: err.errorType.codeName,
		Error:  &err.description,
		Field:  err.failedField,
	}

	if err.nestedError != nil {
		details := err.nestedError.Error()
		oError.Details = &details
	}

	return err.errorType.httpCode, marshalError(oError)
}

func secretiveSerializer(err *Error) (int, []byte) {
	oError := &openapi.Error{
		Result: err.errorType.codeName,
		Error:  &err.description,
	}

	return err.errorType.httpCode, marshalError(oError)
}

type Error struct {
	errorType    *ErrorType
	description  string
	nestedError  error
	failedField  *string
	warningLabel string

	serializer          errorSerializerFunc
	externalLoggerLevel string
}

func (e *Error) Is(target error) bool {
	if err2, ok := target.(*Error); ok {
		return e.errorType == err2.errorType
	}

	return false
}

func (e *Error) Unwrap() error {
	return e.nestedError
}

func (e *Error) Error() string {
	text := e.description
	if e.nestedError != nil {
		text = text + ": " + e.nestedError.Error()
	}
	return text
}

func newError(errorType *ErrorType, description string, serializer errorSerializerFunc, err error, failedField *string, fields ...zap.Field) *Error {
	e := &Error{
		errorType:           errorType,
		description:         description,
		nestedError:         err,
		serializer:          serializer,
		failedField:         failedField,
		externalLoggerLevel: string(sentry.LevelError),
	}

	sendToExternalServices(e, fields...)
	zap.L().Error(e.Error(), fields...)
	return e
}

func newWarning(errorType *ErrorType, msg string, serializer errorSerializerFunc, err error, failedField *string, label string, fields ...zap.Field) *Error {
	if len(label) == 0 {
		label = "unset"
	}
	w := &Error{
		errorType:           errorType,
		description:         msg,
		nestedError:         err,
		serializer:          serializer,
		failedField:         failedField,
		warningLabel:        label,
		externalLoggerLevel: string(sentry.LevelWarning),
	}

	sendToExternalServices(w, fields...)
	zap.L().Warn(w.Error(), fields...)
	return w
}

// ErrorToHttpResponse returns http status code and body bytes.
func ErrorToHttpResponse(err error) (int, []byte) {
	switch t := err.(type) {
	case *Error:
		return t.serializer(t)
	}

	msg := err.Error()
	oError := &openapi.Error{
		Result: "UNKNOWN_ERROR",
		Error:  &msg,
	}
	return http.StatusInternalServerError, marshalError(oError)
}

func sendToExternalServices(e *Error, fields ...zap.Field) {
	// prometheus counter
	errorByCodeCounter.WithLabelValues(
		string(e.errorType.codeName),
		e.warningLabel,
		version.GetTag(),
		version.GetCommit(),
		version.GetFeature(),
		getCaller(),
	).Inc()

	// sentry
	// fill the scope with error-related fields and push an error
	// within that scope.
	sentry.CurrentHub().WithScope(func(scope *sentry.Scope) {
		scope.SetTag("err_type", string(e.errorType.codeName))

		if e.failedField != nil {
			scope.SetExtra("failed_field", *e.failedField)
		}
		if len(e.warningLabel) == 0 {
			scope.SetExtra("warn_label", e.warningLabel)
		}

		if len(fields) > 0 {
			encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{})
			buf, err := encoder.EncodeEntry(zapcore.Entry{}, fields)
			if err == nil {
				scope.SetExtra("zap_fields", buf.String())
			}
		}

		if e.nestedError == nil {
			scope.SetLevel(sentry.Level(e.externalLoggerLevel))
			sentry.CaptureMessage(e.description)
		} else {
			scope.SetExtra("message", e.description)
			sentry.CaptureException(e.nestedError)
		}
	})
}

func getCaller() string {
	// skip callers in this file, so (srcFile, line) points
	// to the one who invoked common.EInternal(...)
	_, srcFile, line, ok := runtime.Caller(4)
	if !ok {
		return "unknown"
	}

	srcFile = cutCallerFilePath(srcFile)
	return fmt.Sprintf("%s:%d", srcFile, line)
}

// /home/user/src/project/package/foo.go -> package/foo.go
func cutCallerFilePath(file string) string {
	oneSlash := false
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == os.PathSeparator {
			if oneSlash {
				file = file[i+1:]
				break
			}
			oneSlash = true
		}
	}
	return file
}
