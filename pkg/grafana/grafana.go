package grafana

import (
	"fmt"
	"reflect"

	"github.com/grafana-tools/sdk"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/logrusorgru/aurora"

	"github.com/richerve/mondiff/pkg/types"
	"github.com/richerve/mondiff/pkg/format"

)

// Diff between 2 dashbaords A and B
type duplicateBoards struct {
	A sdk.Board
	B sdk.Board
}

// Calculate the diff between 2 sdk.Board using cmp package
func (bd duplicateBoards) diff(r types.DiffReporter) string {
	opts := []cmp.Option{
		cmp.Reporter(&r),
		cmp.AllowUnexported(sdk.Board{}),
		// cmpopts.IgnoreUnexported(sdk.Board{}),
		cmpopts.IgnoreFields(sdk.Board{}, "ID", "Version"),
		cmpopts.IgnoreFields(sdk.Panel{}, "ID"),
		cmpopts.IgnoreFields(sdk.CommonPanel{}, "GridPos"),
	}

	cmp.Equal(bd.A, bd.B, opts...)

	return r.String()
}

type indexedFoundBoards map[string]sdk.FoundBoard

func indexFoundBoard(m indexedFoundBoards, fb sdk.FoundBoard, field string) indexedFoundBoards {

	v := reflect.ValueOf(fb)
	f := reflect.Indirect(v).FieldByName(field)

	m[f.String()] = fb
	return m
}

func discoverDashboards(client *sdk.Client) (indexedFoundBoards, error) {

	foundBoards, err := client.SearchDashboards("", false)
	if err != nil {
		return nil, fmt.Errorf("Fail searching for dashboards: %v", err)
	}

	boards := make(indexedFoundBoards)
	for _, foundBoard := range foundBoards {
		boards = indexFoundBoard(boards, foundBoard, "Title")
	}

	return boards, nil
}

// DuplicatedDashboardsWithDiff returns a list of unique and duplicate boards.
func DuplicatedDashboardsWithDiff(clientA, clientB *sdk.Client) (onlyA, onlyB indexedFoundBoards, dupBoards []duplicateBoards, err error) {

	foundBoardsA, err := discoverDashboards(clientA)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Error getting dashboards from client A: %s", err)
	}

	foundBoardsB, err := discoverDashboards(clientB)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Error getting dashboards from client B: %s", err)
	}

	dups := make(indexedFoundBoards)
	for k, v := range foundBoardsA {
		if _, ok := foundBoardsB[k]; ok {
			dups[k] = v
			delete(foundBoardsA, k)
			delete(foundBoardsB, k)
		}
	}

	for title, data := range dups {
		uri := data.URI
		b1, _, err := clientA.GetDashboard(uri)
		if err != nil {
			fmt.Printf("ERROR: Failed getting dashboard %v with title %v from client A\n", uri, title)
		}

		b2, _, err := clientB.GetDashboard(uri)
		if err != nil {
			fmt.Printf("ERROR: Failed gettting dashboard %v with title %v from client B\n", uri, title)
		}

		dupBoards = append(dupBoards, duplicateBoards{A: b1, B: b2})
	}

	return foundBoardsA, foundBoardsB, dupBoards, nil
}

func DashboardsDiffReport(onlyA, onlyB indexedFoundBoards, dups []duplicateBoards) {

	var reporter types.DiffReporter

	fmt.Println()
	format.PrintSectionHeader("Dashboards only in A", "#")
	for k := range onlyA {
		fmt.Printf("%q\n", k)
	}

	fmt.Println()
	format.PrintSectionHeader("Dashboards only in B", "#")
	for k := range onlyB {
		fmt.Printf("%q\n", k)
	}

	fmt.Println()
	format.PrintSectionHeader("Diff between dashboards in both", "#")
	for _, dBoard := range dups {
		if report := dBoard.diff(reporter); report != "" {
			fmt.Printf("# %s:\n", Underline(dBoard.A.Title))
			fmt.Print(report)
		}
	}
}
