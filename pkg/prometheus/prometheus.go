package prometheus

import (
	"fmt"
	"context"
	"time"

	"github.com/prometheus/client_golang/api/prometheus/v1"
	promapi "github.com/prometheus/client_golang/api"
	"github.com/google/go-cmp/cmp"

	"github.com/richerve/mondiff/pkg/types"
	"github.com/richerve/mondiff/pkg/format"
)

type duplicateRules struct {
	A v1.Rules
	B v1.Rules
}

// Calculate the diff between 2 sdk.Board using cmp package
func (dr duplicateRules) diff(r types.Reporter) string {
	opts := []cmp.Option{
		cmp.Reporter(&r),
		// cmpopts.IgnoreUnexported(sdk.Board{}),
		// cmpopts.IgnoreFields(sdk.Board{}, "ID", "Version"),
		// cmpopts.IgnoreFields(sdk.Panel{}, "ID"),
		// cmpopts.IgnoreFields(sdk.CommonPanel{}, "GridPos"),
	}

	cmp.Equal(dr.A, dr.B, opts...)

	return r.String()
}

func DiscoverRuleGroups(client promapi.Client) (*types.Set, error) {

	api := v1.NewAPI(client)

	ruleGroups := types.NewSet()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rulesResult, err := api.Rules(ctx)
	if err != nil {
		return ruleGroups, fmt.Errorf("Error getting rules: %v", err)
	}

	for _, ruleGroup := range rulesResult.Groups {
		ruleGroups.Add(ruleGroup.Name, ruleGroup.Rules)
	}

	return ruleGroups, nil
}

func DuplicateRuleGroupsByName(rgA, rgB *types.Set) *types.Set {

	return rgA.Intersection(rgB)
}

func DuplicatedRuleGroupsWithDiff(rgA, rgB *types.Set) (onlyA, onlyB *types.Set, dupRuleGroups []duplicateRules) {

	dups := DuplicateRuleGroupsByName(rgA, rgB)
	onlyA = rgA.Difference(dups)
	onlyB = rgB.Difference(dups)

	for name := range dups.Items {
		rulesA := rgA.Items[name].(v1.Rules)
		rulesB := rgB.Items[name].(v1.Rules)
		dupRuleGroups = append(dupRuleGroups, duplicateRules{A: rulesA, B: rulesB})
	}
	return
}

func RulesDiffReport(onlyA, onlyB *types.Set, dups []duplicateRules) {

	var reporter types.Reporter

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
