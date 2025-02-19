package main

import (
	"context"

	"github.com/theopenlane/core/pkg/sleuth/dnsx"
)

func main() {
	client, err := dnsx.NewDNSX(dnsx.WithOutputCDN(true))
	if err != nil {
		panic(err)
	}

	report, err := client.GetDomainDNSRecords(context.TODO(), "theopenlane.io")
	if err != nil {
		panic(err)
	}

	dnsx.PrintDNSRecordsReportTable(report)
}
