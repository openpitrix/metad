// Copyright 2018 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package logger

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"testing"
)

type testing_TBHelper interface {
	Helper()
}

func tAssert(tb testing.TB, condition bool, args ...interface{}) {
	if x, ok := tb.(testing_TBHelper); ok {
		x.Helper() // Go1.9+
	}
	if !condition {
		if msg := fmt.Sprint(args...); msg != "" {
			tb.Fatalf("Assert failed, %s", msg)
		} else {
			tb.Fatalf("Assert failed")
		}
	}
}

func tAssertf(tb testing.TB, condition bool, format string, a ...interface{}) {
	if x, ok := tb.(testing_TBHelper); ok {
		x.Helper() // Go1.9+
	}
	if !condition {
		if msg := fmt.Sprintf(format, a...); msg != "" {
			tb.Fatalf("Assertf failed, %s", msg)
		} else {
			tb.Fatalf("Assertf failed")
		}
	}
}

func tAssertFunc(tb testing.TB, fn func() error) {
	if x, ok := tb.(testing_TBHelper); ok {
		x.Helper() // Go1.9+
	}
	if err := fn(); err != nil {
		tb.Fatalf("AssertFunc failed, %v", err)
	}
}

func tReadBuf(buf *bytes.Buffer) string {
	str := buf.String()
	buf.Reset()
	return str
}

func TestLogger(t *testing.T) {
	buf := new(bytes.Buffer)
	SetOutput(buf)

	Debug("debug log, should ignore by default")
	tAssert(t, tReadBuf(buf) == "")

	Info("info log, should visable")
	tAssertFunc(t, func() error {
		expected := "info log, should visable"
		if msg := tReadBuf(buf); !strings.Contains(msg, expected) {
			return fmt.Errorf("donot contains: %s", expected)
		}
		return nil
	})

	Info("format [%d]", 111)
	tAssertFunc(t, func() error {
		expected := "format [111]"
		if msg := tReadBuf(buf); !strings.Contains(msg, expected) {
			return fmt.Errorf("donot contains: %s", expected)
		}
		return nil
	})

	SetLevelByString("debug")
	Debug("debug log, now it becomes visible")
	tAssertFunc(t, func() error {
		expected := "debug log, now it becomes visible"
		if msg := tReadBuf(buf); !strings.Contains(msg, expected) {
			return fmt.Errorf("donot contains:" + expected)
		}
		return nil
	})

	logger = NewLogger()
	logger.SetPrefix("(prefix)").SetSuffix("(suffix)").SetOutput(buf)

	logger.Warn("log_content")
	log := tReadBuf(buf)
	tAssertFunc(t, func() error {
		re := " -WARNING- \\(prefix\\)log_content \\(testing.go:\\d+\\)\\(suffix\\)"
		matched, err := regexp.MatchString(re, log)
		if err != nil {
			return fmt.Errorf("invalid regexp: %q", err)
		}
		if !matched {
			return fmt.Errorf("regexp failed: %q", re)
		}
		return nil
	})
	t.Log(log)

	logger.HideCallstack()
	logger.Warn("log_content")
	log = tReadBuf(buf)
	tAssertFunc(t, func() error {
		re := " -WARNING- \\(prefix\\)log_content\\(suffix\\)"
		matched, err := regexp.MatchString(re, log)
		if err != nil {
			return fmt.Errorf("invalid regexp: %q", err)
		}
		if !matched {
			return fmt.Errorf("regexp failed: %q", re)
		}
		return nil
	})
	t.Log(log)
}
