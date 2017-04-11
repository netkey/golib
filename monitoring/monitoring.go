// Tideland Go Library - Monitoring
//
// Copyright (C) 2009-2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package monitoring

//--------------------
// IMPORTS
//--------------------

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

//--------------------
// CONSTANTS
//--------------------

const (
	etmTLine  = "+----------------------------------------------------+------------+--------------------+--------------------+--------------------+\n"
	etmHeader = "| Measuring Point Name                               | Count      | Min Dur            | Max Dur            | Avg Dur            |\n"
	etmFormat = "| %-50s | %10d | %18s | %18s | %18s |\n"

	ssvTLine  = "+----------------------------------------------------+-----------+---------------+---------------+---------------+---------------+\n"
	ssvHeader = "| Stay-Set Variable Name                             | Count     | Act Value     | Min Value     | Max Value     | Avg Value     |\n"
	ssvFormat = "| %-50s | %9d | %13d | %13d | %13d | %13d |\n"

	dsrTLine  = "+----------------------------------------------------+---------------------------------------------------------------------------+\n"
	dsrHeader = "| Dynamic Status                                     | Value                                                                     |\n"
	dsrFormat = "| %-50s | %-73s |\n"
)

//--------------------
// INTERFACES
//--------------------

// IDFilter allows to add filter for execution time measurings,
// stay-set values, and dynamic status retriever. If set only
// monitorings with the filter returning true will be done.
type IDFilter func(id string) bool

// Measuring defines one execution time measuring containing the ID and
// the starting time of the measuring and able to pass this data after
// the end of the measuring to its backend.
type Measuring interface {
	// EndMeasuring ends the measuring and passes its
	// data to the backend.
	EndMeasuring() time.Duration
}

// MeasuringPoint defines the collected information for one execution
// time measuring point.
type MeasuringPoint interface {
	fmt.Stringer

	// ID returns the identifier of the measuring point.
	ID() string

	// Count returns how often this point has been measured.
	Count() int64

	// MinDuration returns the shortest execution time.
	MinDuration() time.Duration

	// MaxDuration returns the longest execution time.
	MaxDuration() time.Duration

	// AvgDuration returns the average execution time.
	AvgDuration() time.Duration
}

// MeasuringPoints is a set of measuring points.
type MeasuringPoints []MeasuringPoint

// Implement the sort interface.

func (m MeasuringPoints) Len() int           { return len(m) }
func (m MeasuringPoints) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m MeasuringPoints) Less(i, j int) bool { return m[i].ID() < m[j].ID() }

// StaySetVariable contains the cumulated values
// for one stay-set variable.
type StaySetVariable interface {
	fmt.Stringer

	// ID returns the identifier of the stay-set variable.
	ID() string

	// Count returns how often the value has been changed.
	Count() int64

	// ActValue returns the current value of the variable.
	ActValue() int64

	// MinValue returns the minimum value of the variable.
	MinValue() int64

	// MaxValue returns the maximum value of the variable.
	MaxValue() int64

	// MinValue returns the average value of the variable.
	AvgValue() int64
}

// StaySetVariables is a set of stay-set variables.
type StaySetVariables []StaySetVariable

// Implement the sort interface.

func (s StaySetVariables) Len() int           { return len(s) }
func (s StaySetVariables) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s StaySetVariables) Less(i, j int) bool { return s[i].ID() < s[j].ID() }

// DynamicStatusRetriever is called by the server and
// returns a current status as string.
type DynamicStatusRetriever func() (string, error)

// DynamicStatusValue contains one retrieved value.
type DynamicStatusValue interface {
	fmt.Stringer

	// ID returns the identifier of the status value.
	ID() string

	// Value returns the retrieved value as string.
	Value() string
}

// DynamicStatusValues is a set of dynamic status values.
type DynamicStatusValues []DynamicStatusValue

// Implement the sort interface.

func (d DynamicStatusValues) Len() int           { return len(d) }
func (d DynamicStatusValues) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
func (d DynamicStatusValues) Less(i, j int) bool { return d[i].ID() < d[j].ID() }

// Backend defines the interface for a type managing all
// the information provided or needed by the public functions
// of the monitoring package.
type Backend interface {
	// BeginMeasuring starts a new measuring with a given id.
	BeginMeasuring(id string) Measuring

	// ReadMeasuringPoint returns the measuring point for an id.
	ReadMeasuringPoint(id string) (MeasuringPoint, error)

	// MeasuringPointsDo performs the function f for
	// all measuring points.
	MeasuringPointsDo(f func(MeasuringPoint)) error

	// SetVariable sets a value of a stay-set variable.
	SetVariable(id string, v int64)

	// IncrVariable increases a variable.
	IncrVariable(id string)

	// DecrVariable decreases a variable.
	DecrVariable(id string)

	// ReadVariable returns the stay-set variable for an id.
	ReadVariable(id string) (StaySetVariable, error)

	// StaySetVariablesDo performs the function f for all
	// variables.
	StaySetVariablesDo(f func(StaySetVariable)) error

	// Register registers a new dynamic status retriever function.
	Register(id string, rf DynamicStatusRetriever)

	// ReadStatus returns the dynamic status for an id.
	ReadStatus(id string) (string, error)

	// DynamicStatusValuesDo performs the function f for all
	// status values.
	DynamicStatusValuesDo(f func(DynamicStatusValue)) error

	// SetMeasuringFilter sets the new filter for measurings
	// and returns the current one.
	SetMeasuringsFilter(f IDFilter) IDFilter

	// SetMeasuringFilter sets the new filter for variables
	// and returns the current one.
	SetVariablesFilter(f IDFilter) IDFilter

	// SetRetrieversFilter sets the new filter for status retrievers
	// and returns the current one.
	SetRetrieversFilter(f IDFilter) IDFilter

	// Reset clears all monitored values.
	Reset() error

	// Stop tells the backend that a new one has been set.
	Stop()
}

//--------------------
// MONITORING API
//--------------------

// monitoring manages the global monitor.
type monitoring struct {
	sync.RWMutex
	b Backend
}

// backend ensures, that a backend is set. By default
// it's the standard one.
func (m *monitoring) backend() Backend {
	if m.b == nil {
		m.b = NewStandardBackend()
	}
	return m.b
}

// setBackend sets the current backend. If one
// is already set it will be stopped.
func (m *monitoring) setBackend(mb Backend) {
	if m.b != nil {
		m.b.Stop()
	}
	m.b = mb
}

// monitor is the global monitor.
var monitor = &monitoring{}

// SetBackend allows to switch the monitoring backend.
func SetBackend(mb Backend) {
	monitor.Lock()
	defer monitor.Unlock()
	monitor.setBackend(mb)
}

// BeginMeasuring starts a new measuring with a given id.
// All measurings with the same id will be aggregated.
func BeginMeasuring(id string) Measuring {
	monitor.RLock()
	defer monitor.RUnlock()
	return monitor.backend().BeginMeasuring(id)
}

// Measure the execution of a function.
func Measure(id string, f func()) time.Duration {
	m := BeginMeasuring(id)
	f()
	return m.EndMeasuring()
}

// ReadMeasuringPoint returns the measuring point for an id.
func ReadMeasuringPoint(id string) (MeasuringPoint, error) {
	monitor.RLock()
	defer monitor.RUnlock()
	return monitor.backend().ReadMeasuringPoint(id)
}

// MeasuringPointsDo performs the function f for
// all measuring points.
func MeasuringPointsDo(f func(MeasuringPoint)) error {
	monitor.RLock()
	defer monitor.RUnlock()
	return monitor.backend().MeasuringPointsDo(f)
}

// MeasuringPointsWrite prints the measuring points for which
// the passed function returns true to the passed writer.
func MeasuringPointsWrite(w io.Writer, ff func(MeasuringPoint) bool) error {
	fmt.Fprint(w, etmTLine)
	fmt.Fprint(w, etmHeader)
	fmt.Fprint(w, etmTLine)
	if err := MeasuringPointsDo(func(mp MeasuringPoint) {
		if ff(mp) {
			fmt.Fprintf(w, etmFormat, mp.ID(), mp.Count(), mp.MinDuration(), mp.MaxDuration(), mp.AvgDuration())
		}
	}); err != nil {
		return err
	}
	fmt.Fprint(w, etmTLine)
	return nil
}

// MeasuringPointsPrintAll prints all measuring points
// to STDOUT.
func MeasuringPointsPrintAll() error {
	return MeasuringPointsWrite(os.Stdout, func(mp MeasuringPoint) bool { return true })
}

// SetVariable sets a value of a stay-set variable.
func SetVariable(id string, v int64) {
	monitor.RLock()
	defer monitor.RUnlock()
	monitor.backend().SetVariable(id, v)
}

// IncrVariable increases a stay-set variable.
func IncrVariable(id string) {
	monitor.RLock()
	defer monitor.RUnlock()
	monitor.backend().IncrVariable(id)
}

// DecrVariable decreases a stay-set variable.
func DecrVariable(id string) {
	monitor.RLock()
	defer monitor.RUnlock()
	monitor.backend().DecrVariable(id)
}

// ReadVariable returns the stay-set variable for an id.
func ReadVariable(id string) (StaySetVariable, error) {
	monitor.RLock()
	defer monitor.RUnlock()
	return monitor.backend().ReadVariable(id)
}

// StaySetVariablesDo performs the function f for all
// variables.
func StaySetVariablesDo(f func(StaySetVariable)) error {
	monitor.RLock()
	defer monitor.RUnlock()
	return monitor.backend().StaySetVariablesDo(f)
}

// StaySetVariablesWrite prints the stay-set variables for which
// the passed function returns true to the passed writer.
func StaySetVariablesWrite(w io.Writer, ff func(StaySetVariable) bool) error {
	fmt.Fprint(w, ssvTLine)
	fmt.Fprint(w, ssvHeader)
	fmt.Fprint(w, ssvTLine)
	if err := StaySetVariablesDo(func(ssv StaySetVariable) {
		if ff(ssv) {
			fmt.Fprintf(w, ssvFormat, ssv.ID(), ssv.Count(), ssv.ActValue(), ssv.MinValue(), ssv.MaxValue(), ssv.AvgValue())
		}
	}); err != nil {
		return err
	}
	fmt.Fprint(w, ssvTLine)
	return nil
}

// StaySetVariablesPrintAll prints all stay-set variables
// to STDOUT.
func StaySetVariablesPrintAll() error {
	return StaySetVariablesWrite(os.Stdout, func(ssv StaySetVariable) bool { return true })
}

// Register registers a new dynamic status retriever function.
func Register(id string, rf DynamicStatusRetriever) {
	monitor.RLock()
	defer monitor.RUnlock()
	monitor.backend().Register(id, rf)
}

// ReadStatus returns the dynamic status for an id.
func ReadStatus(id string) (string, error) {
	monitor.RLock()
	defer monitor.RUnlock()
	return monitor.backend().ReadStatus(id)
}

// DynamicStatusValuesDo performs the function f for all
// status values.
func DynamicStatusValuesDo(f func(DynamicStatusValue)) error {
	monitor.RLock()
	defer monitor.RUnlock()
	return monitor.backend().DynamicStatusValuesDo(f)
}

// DynamicStatusValuesWrite prints the status values for which
// the passed function returns true to the passed writer.
func DynamicStatusValuesWrite(w io.Writer, ff func(DynamicStatusValue) bool) error {
	fmt.Fprint(w, dsrTLine)
	fmt.Fprint(w, dsrHeader)
	fmt.Fprint(w, dsrTLine)
	if err := DynamicStatusValuesDo(func(dsv DynamicStatusValue) {
		if ff(dsv) {
			fmt.Fprintf(w, dsrFormat, dsv.ID(), dsv.Value())
		}
	}); err != nil {
		return err
	}
	fmt.Fprint(w, dsrTLine)
	return nil
}

// DynamicStatusValuesPrintAll prints all status values to STDOUT.
func DynamicStatusValuesPrintAll() error {
	return DynamicStatusValuesWrite(os.Stdout, func(dsv DynamicStatusValue) bool { return true })
}

// SetMeasuringsFilter sets the new filter for measurings
// and returns the current one.
func SetMeasuringsFilter(f IDFilter) IDFilter {
	monitor.RLock()
	defer monitor.RUnlock()
	return monitor.backend().SetMeasuringsFilter(f)
}

// SetVariablesFilter sets the new filter for variables
// and returns the current one.
func SetVariablesFilter(f IDFilter) IDFilter {
	monitor.RLock()
	defer monitor.RUnlock()
	return monitor.backend().SetVariablesFilter(f)
}

// SetRetrieversFilter sets the new filter for status retrievers
// and returns the current one.
func SetRetrieversFilter(f IDFilter) IDFilter {
	monitor.RLock()
	defer monitor.RUnlock()
	return monitor.backend().SetRetrieversFilter(f)
}

// Reset clears all monitored values.
func Reset() error {
	monitor.RLock()
	defer monitor.RUnlock()
	return monitor.backend().Reset()
}

// EOF
