// Package types provides custom types used by other packages
package types

import (
	"strings"
	"sort"
	"fmt"

	"github.com/google/go-cmp/cmp"
	. "github.com/logrusorgru/aurora"
)

// Set is a set-like type that is indexed by a string.
// But each item can have any arbitrary value
type Set struct {
	Items map[string]interface{}
}

func NewSet() *Set {
	items := make(map[string]interface{})

	return &Set{Items: items}
}

func (s *Set) Add(item string, value interface{}) {
	if !s.Has(item) {
		s.Items[item] = value
	}
}

func (s *Set) Has(item string) bool {
	_, ok := s.Items[item]
	return ok
}

func (s *Set) Delete(item string) bool {
	if s.Has(item) {
		delete(s.Items, item)
		return true
	}
	return false
}

func (s *Set) Len() int {
	return len(s.Items)
}

func (s *Set) Keys() (result []string) {

	for k := range s.Items {
		result = append(result, k)
	}

	sort.Strings(result)
	return
}

func (s *Set) String() string {

	return strings.Join(s.Keys(), "\n")
}

func(s *Set) Intersection(s2 *Set) *Set {
	resultSet := NewSet()

	for k, v := range s2.Items {
		if _, ok := s.Items[k]; ok {
			resultSet.Items[k] = v
		}
	}
	return resultSet
}

func(s *Set) Difference(s2 *Set) *Set {
	resultSet := NewSet()

	for k, v := range s.Items {
		if _, ok := s2.Items[k]; !ok {
			resultSet.Items[k] = v
		}
	}
	return resultSet
}

// Reporter implements cmp.Reporter
// it is used for all diff reports
type Reporter struct {
	path  cmp.Path
	diffs []string
}

func (r *Reporter) PushStep(ps cmp.PathStep) {
	r.path = append(r.path, ps)
}

func (r *Reporter) Report(rs cmp.Result) {
	if !rs.Equal() {
		vx, vy := r.path.Last().Values()
		r.diffs = append(r.diffs, fmt.Sprintf("\t%#v:\n\t\t-: %+v\n\t\t+: %+v\n", r.path, Green(vx), Yellow(vy)))
	}
}

func (r *Reporter) PopStep() {
	r.path = r.path[:len(r.path)-1]
}

func (r *Reporter) String() string {
	return strings.Join(r.diffs, "\n")
}
