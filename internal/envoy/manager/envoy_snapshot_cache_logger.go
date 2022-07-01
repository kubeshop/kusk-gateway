// MIT License
//
// Copyright (c) 2022 Kubeshop
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// A note on logging:
// `l.logger.V(1).Info(...)` is effectively debug level.

package manager

import (
	"fmt"

	envoy_log "github.com/envoyproxy/go-control-plane/pkg/log"
	"github.com/go-logr/logr"
)

// EnvoySnapshotCacheLogger wraps a `logr.Logger` logger.
type EnvoySnapshotCacheLogger struct {
	logger logr.Logger
}

func NewEnvoySnapshotCacheLogger(logger logr.Logger) envoy_log.Logger {
	return &EnvoySnapshotCacheLogger{
		logger: logger.WithName("SnapshotCache"),
	}
}

// Debugf logs a formatted debugging message.
func (l EnvoySnapshotCacheLogger) Debugf(format string, args ...interface{}) {
	l.logger.V(1).Info(fmt.Sprintf(format, args...))
}

// Infof logs a formatted informational message.
func (l EnvoySnapshotCacheLogger) Infof(format string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, args...))
}

// Warnf logs a formatted warning message.
func (l EnvoySnapshotCacheLogger) Warnf(format string, args ...interface{}) {
	l.logger.Error(nil, fmt.Sprintf(format, args...))
}

// Errorf logs a formatted error message.
func (l EnvoySnapshotCacheLogger) Errorf(format string, args ...interface{}) {
	l.logger.Error(nil, fmt.Sprintf(format, args...))
}
