// Tideland Go Library - Monitoring
//
// Copyright (C) 2009-2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

// Package monitoring of the Tideland Go Library supports three kinds of
// system monitoring. They are helpful to understand what's happening
// inside a system during runtime. So execution times can be measured
// and analyzed, stay-set variables integrated and dynamic control
// value retrieval provided. The backend is exchangeable. So the
// StandardBackend workes like described above, the NullBackend does
// nothing, and own implementations can integrate external systems.
// Additionally filters can be added to reduce the monitoring to
// the points of interest.
package monitoring

// EOF
