// Tideland Go Library - Logger - Unit Tests
//
// Copyright (C) 2012-2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package logger_test

//--------------------
// IMPORTS
//--------------------

import (
	"log"
	"os"
	"testing"

	"github.com/tideland/golib/audit"
	"github.com/tideland/golib/logger"
)

//--------------------
// TESTS
//--------------------

// TestGetSetLevel tests the setting of the logging level.
func TestGetSetLevel(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	level := logger.Level()
	defer logger.SetLevel(level)

	tl, fetch := logger.NewTestLogger()
	logger.SetLogger(tl)

	logger.SetLevel(logger.LevelDebug)
	logger.Debugf("Debug.")
	logger.Infof("Info.")
	logger.Warningf("Warning.")
	logger.Errorf("Error.")
	logger.Criticalf("Critical.")

	assert.Length(fetch(), 5)

	logger.SetLevel(logger.LevelError)
	logger.Debugf("Debug.")
	logger.Infof("Info.")
	logger.Warningf("Warning.")
	logger.Errorf("Error.")
	logger.Criticalf("Critical.")

	assert.Length(fetch(), 2)
}

// TestGetSetLevelString tests the setting of the
// logging level with a string.
func TestGetSetLevelString(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	level := logger.Level()
	defer logger.SetLevel(level)

	tl, fetch := logger.NewTestLogger()
	logger.SetLogger(tl)

	logger.SetLevelString("dEbUg")
	logger.Debugf("Debug.")
	logger.Infof("Info.")
	logger.Warningf("Warning.")
	logger.Errorf("Error.")
	logger.Criticalf("Critical.")

	assert.Length(fetch(), 5)

	logger.SetLevelString("error")
	logger.Debugf("Debug.")
	logger.Infof("Info.")
	logger.Warningf("Warning.")
	logger.Errorf("Error.")
	logger.Criticalf("Critical.")

	assert.Length(fetch(), 2)

	logger.SetLevelString("dont-know-what-you-mean")
	logger.Debugf("Debug.")
	logger.Infof("Info.")
	logger.Warningf("Warning.")
	logger.Errorf("Error.")
	logger.Criticalf("Critical.")

	assert.Length(fetch(), 2)
}

// TestFiltering tests the filtering of the logging.
func TestFiltering(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	level := logger.Level()
	defer logger.SetLevel(level)

	tl, fetch := logger.NewTestLogger()
	logger.SetLogger(tl)

	logger.SetLevel(logger.LevelDebug)
	logger.SetFilter(func(level logger.LogLevel, info, msg string) bool {
		return level >= logger.LevelWarning && level <= logger.LevelError
	})

	logger.Debugf("Debug.")
	logger.Infof("Info.")
	logger.Warningf("Warning.")
	logger.Errorf("Error.")
	logger.Criticalf("Critical.")

	assert.Length(fetch(), 3)

	logger.UnsetFilter()

	logger.Debugf("Debug.")
	logger.Infof("Info.")
	logger.Warningf("Warning.")
	logger.Errorf("Error.")
	logger.Criticalf("Critical.")

	assert.Length(fetch(), 5)
}

// TestGoLogger tests logging with the go logger.
func TestGoLogger(t *testing.T) {
	level := logger.Level()
	defer logger.SetLevel(level)

	log.SetOutput(os.Stdout)

	logger.SetLevel(logger.LevelDebug)
	logger.SetLogger(logger.NewGoLogger())

	logger.Debugf("Debug.")
	logger.Infof("Info.")
	logger.Warningf("Warning.")
	logger.Errorf("Error.")
	logger.Criticalf("Critical.")
}

// TestSysLogger tests logging with the syslogger.
func TestSysLogger(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	level := logger.Level()
	defer logger.SetLevel(level)

	logger.SetLevel(logger.LevelDebug)

	sl, err := logger.NewSysLogger("GOAS")
	assert.Nil(err)
	logger.SetLogger(sl)

	logger.Debugf("Debug.")
	logger.Infof("Info.")
	logger.Warningf("Warning.")
	logger.Errorf("Error.")
	logger.Criticalf("Critical.")
}

// TestFatalExit tests the call of the fatal exiter after a
// fatal error log.
func TestFatalExit(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	level := logger.Level()
	defer logger.SetLevel(level)

	tl, fetch := logger.NewTestLogger()
	logger.SetLogger(tl)

	exited := false
	fatalExiter := func() {
		exited = true
	}

	logger.SetFatalExiter(fatalExiter)

	logger.Fatalf("fatal")
	assert.Length(fetch(), 1)
	assert.True(exited)
}

// EOF
