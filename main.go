package main

import (
	"fmt"
	"os"
	"net/http"
	"crypto/tls"
	"strings"

	"github.com/grafana-tools/sdk"
	"github.com/spf13/viper"
	"github.com/prometheus/client_golang/api"

	"github.com/richerve/mondiff/pkg/grafana"
	"github.com/richerve/mondiff/pkg/prometheus"
)

func getProfileInfo(arg string) (url, auth string, err error) {
	// config := viper.GetViper()

	if viper.InConfig(arg) {
		urlPattern := fmt.Sprintf("%s.url", arg)
		authPattern := fmt.Sprintf("%s.auth", arg)

		url = viper.GetString(urlPattern)
		auth = viper.GetString(authPattern)
	} else {
		return "", "", fmt.Errorf("Argument %s is not a profile in config", arg)
	}

	return url, auth, nil
}

func main() {

	// Config
	viper.SetConfigType("toml")
	viper.SetConfigName("config")
	viper.AddConfigPath("$HOME/.mondiff")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			panic(fmt.Errorf("Config file not found: %s \n", err))
		} else {
			panic(fmt.Errorf("Error reading config file: %s \n", err))
		}
	}

	if len(os.Args[1:]) != 2 {
		fmt.Fprint(os.Stderr, "ERROR: Exactly 2 args are needed, either urls or config profiles")
		os.Exit(1)
	}

	urlA, authA, err := getProfileInfo(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v", err)
		os.Exit(1)
	}
	urlB, authB, err := getProfileInfo(os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v", err)
		os.Exit(1)
	}

	if strings.Contains(urlA, "grafana") {
		// When testing against a local machine, most likely the SAN name will not be valid
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		insecureHTTPClient := &http.Client{Transport: tr}

		grafanaClientA := sdk.NewClient(urlA, authA, insecureHTTPClient)
		grafanaClientB := sdk.NewClient(urlB, authB, insecureHTTPClient)

		onlyA, onlyB, dups, err := grafana.DuplicatedDashboardsWithDiff(grafanaClientA, grafanaClientB)
		if err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}

		grafana.DashboardsDiffReport(onlyA, onlyB, dups)
	}

	if strings.Contains(urlA, "prometheus") {

		configA := api.Config{
			Address: urlA,
		}

		configB := api.Config{
			Address: urlB,
		}

		promClientA, err := api.NewClient(configA)
		if err != nil {
			panic(err)
		}
		promClientB, err := api.NewClient(configB)
		if err != nil {
			panic(err)
		}

		rgA, err := prometheus.DiscoverRuleGroups(promClientA)
		rgB, err := prometheus.DiscoverRuleGroups(promClientB)

		onlyA, onlyB, dups := prometheus.DuplicatedRuleGroupsWithDiff(rgA, rgB)

		prometheus.RulesDiffReport(onlyA, onlyB, dups)
	}
}
