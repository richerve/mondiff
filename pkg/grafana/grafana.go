package grafana

import (
	"fmt"

	"github.com/grafana-tools/sdk"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/logrusorgru/aurora"

	"github.com/richerve/mondiff/pkg/types"
	"github.com/richerve/mondiff/pkg/format"

)

type DiscoveredDashboard struct {
	URI string
}

func newDiscoveredDashboard(uri string) DiscoveredDashboard {
	return DiscoveredDashboard{
		URI: uri,
	}
}

// Di ff between 2 dashbaords A and B
type duplicateBoards struct {
	A sdk.Board
	B sdk.Board
}

// Calculate the diff between 2 sdk.Board using cmp package
func (bd duplicateBoards) diff(r types.Reporter) string {
	opts := []cmp.Option{
		cmp.Reporter(&r),
		cmpopts.IgnoreUnexported(sdk.Board{}),
		cmpopts.IgnoreFields(sdk.Board{}, "ID", "Version"),
		cmpopts.IgnoreFields(sdk.Panel{}, "ID"),
		cmpopts.IgnoreFields(sdk.CommonPanel{}, "GridPos"),
	}

	cmp.Equal(bd.A, bd.B, opts...)

	return r.String()
}

func DiscoverDashboards(client *sdk.Client) (*types.Set, error) {

	// boards will have a discoveredDashboard indexed by FoundBoard.Title
	boards := types.NewSet()

	foundBoards, err := client.SearchDashboards("", false)
	if err != nil {
		return boards, fmt.Errorf("Fail searching for dashboards: %s", err)
	}

	for _, foundBoard := range foundBoards {
		boards.Add(foundBoard.Title, newDiscoveredDashboard(foundBoard.URI))
	}

	return boards, nil
}

func DuplicatedDashboardsByURI(boardsA, boardsB *types.Set) *types.Set {

	return boardsA.Intersection(boardsB)
}

// DuplicatedDashboardsWithDiff returns a list of unique and duplicate boards.
func DuplicatedDashboardsWithDiff(clientA, clientB *sdk.Client) (onlyA *types.Set, onlyB *types.Set, dupBoards []duplicateBoards, err error) {

	boardsA, err := DiscoverDashboards(clientA)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Error getting dashbaords from client A: %s", err)
	}

	boardsB, err := DiscoverDashboards(clientB)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Error getting dashbaords from client B: %s", err)
	}

	dups := DuplicatedDashboardsByURI(boardsA, boardsB)
	onlyA = boardsA.Difference(dups)
	onlyB = boardsB.Difference(dups)

	for title, data := range dups.Items {
		dboard, ok := data.(DiscoveredDashboard)
		if !ok {
			fmt.Printf("ERROR: Can't read discovered dashboard %v", dboard)
		}

		b1, _, err := clientA.GetDashboard(dboard.URI)
		if err != nil {
			fmt.Printf("ERROR: Failed getting dashboard %q from client A\n", title)
		}
		b2, _, err := clientB.GetDashboard(dboard.URI)
		if err != nil {
			fmt.Printf("ERROR: Failed gettting dashboard %q from client B\n", title)
		}

		dupBoards = append(dupBoards, duplicateBoards{A: b1, B: b2})
	}

	return onlyA, onlyB, dupBoards, nil
}

func DashboardsDiffReport(onlyA, onlyB *types.Set, dups []duplicateBoards) {

	var reporter types.Reporter

	format.PrintSectionHeader("Dashboards only in A", "#")
	fmt.Println(onlyA)

	format.PrintSectionHeader("Dashboards only in B", "#")
	fmt.Println(onlyB)

	format.PrintSectionHeader("Diff between dashboards in both", "#")
	for _, dBoard := range dups {
		if report := dBoard.diff(reporter); report != "" {
			fmt.Printf("# %s:\n", Underline(dBoard.A.Title))
			fmt.Print(report)
		}
	}
}
