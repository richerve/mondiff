package prometheus

import (
	"fmt"
	"context"
	"time"
	"reflect"

	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/client_golang/api"
	"github.com/google/go-cmp/cmp"

	"github.com/richerve/mondiff/pkg/types"
	"github.com/richerve/mondiff/pkg/format"
)

type duplicateRules struct {
	A v1.Rules
	B v1.Rules
}

// Calculate the diff between 2 sdk.Board using cmp package
func (dr duplicateRules) diff(r types.DiffReporter) string {
	cmp.Equal(dr.A, dr.B, cmp.Reporter(&r))

	return r.String()
}

type indexedRuleGroups map[string]v1.RuleGroup

func indexRuleGroup(m indexedRuleGroups, rg v1.RuleGroup, field string) indexedRuleGroups {

	v := reflect.ValueOf(rg)
	f := reflect.Indirect(v).FieldByName(field)

	m[f.String()] = rg
	return m
}

func discoverRuleGroups(client api.Client) (indexedRuleGroups, error) {

	promapi := v1.NewAPI(client)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rulesResult, err := promapi.Rules(ctx)
	if err != nil {
		return nil, err
	}

	ruleGroups := make(indexedRuleGroups)
	for _, ruleGroup := range rulesResult.Groups {
		ruleGroups = indexRuleGroup(ruleGroups, ruleGroup, "Name")
	}

	return ruleGroups, nil
}

func DuplicatedRuleGroupsWithDiff(clientA, clientB api.Client) (onlyA, onlyB indexedRuleGroups, dupRuleGroups []duplicateRules, err error) {

	rgA, err := discoverRuleGroups(clientA)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Error discovering rules on client A: %v", err)
	}

	rgB, err := discoverRuleGroups(clientB)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Error discovering rules on client B: %v", err)
	}

	dups := make(indexedRuleGroups)
	for k, v := range rgA {
		if _, ok := rgB[k]; ok {
			dups[k] = v
			delete(rgA, k)
			delete(rgB, k)
		}
	}

	for name := range dups {
		rulesA := rgA[name].Rules
		rulesB := rgB[name].Rules
		dupRuleGroups = append(dupRuleGroups, duplicateRules{A: rulesA, B: rulesB})
	}

	return rgA, rgB, dupRuleGroups, nil
}

func RulesDiffReport(onlyA, onlyB indexedRuleGroups, dups []duplicateRules) {

	var reporter types.DiffReporter

	format.PrintSectionHeader("Rules only in A", "#")
	fmt.Println(onlyA)

	format.PrintSectionHeader("Rules only in B", "#")
	fmt.Println(onlyB)

	format.PrintSectionHeader("Diff between rules in both", "#")
	for _, dRule := range dups {
		if report := dRule.diff(reporter); report != "" {
			fmt.Print(report)
		}
	}
}
