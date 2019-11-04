// Package types provides custom types used by other packages
package types

import (
	"strings"
	"fmt"

	"github.com/google/go-cmp/cmp"
	. "github.com/logrusorgru/aurora"
)

// DiffReporter implements cmp.Reporter.
// It is used for all diff reports
type DiffReporter struct {
	path  cmp.Path
	diffs []string
}

func (r *DiffReporter) PushStep(ps cmp.PathStep) {
	r.path = append(r.path, ps)
}

func (r *DiffReporter) Report(rs cmp.Result) {
	if !rs.Equal() {
		vx, vy := r.path.Last().Values()
		r.diffs = append(r.diffs, fmt.Sprintf("\t%#v:\n\t\t-: %+v\n\t\t+: %+v\n", r.path, Green(vx), Yellow(vy)))
	}
}

func (r *DiffReporter) PopStep() {
	r.path = r.path[:len(r.path)-1]
}

func (r *DiffReporter) String() string {
	return strings.Join(r.diffs, "\n")
}
