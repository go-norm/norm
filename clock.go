// Copyright 2021 Joe Chen. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package norm

import (
	"time"
)

// NowFunc is function to return the current time.
type NowFunc func() time.Time

// Now returns the current UTC time with time.Microsecond truncated because
// PostgreSQL 9.6 does not support saving microsecond. This is particularly
// useful when trying to compare time values between Go and what we get back
// from the PostgreSQL.
func Now() time.Time {
	return time.Now().UTC().Truncate(time.Microsecond)
}
