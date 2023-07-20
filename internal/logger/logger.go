// Copyright (c) Pedersen authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package logger

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type UTCFormatter struct {
	logrus.Formatter
}

func (u UTCFormatter) Format(e *logrus.Entry) ([]byte, error) {
	e.Time = e.Time.UTC()
	return u.Formatter.Format(e)
}

func writerForOutput(output string) (io.Writer, error) {
	if len(output) < 1 {
		return io.Discard, nil
	} else if strings.ToLower(output) == "stdout" {
		return os.Stdout, nil
	}

	logFile, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}

	logrus.RegisterExitHandler(func() {
		if logFile == nil {
			return
		}

		logFile.Close() //nolint: errcheck
	})

	return logFile, nil
}

func Init(logLevel string, output string) error {
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		return err
	}

	logrus.SetLevel(level)

	logrus.SetFormatter(UTCFormatter{&logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
		},
		TimestampFormat: time.RFC3339Nano,
	}})

	logWriter, err := writerForOutput(output)
	if err != nil {
		return err
	}

	logrus.SetOutput(logWriter)

	return nil
}
