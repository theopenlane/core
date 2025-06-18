# Events

## Domain Scan

```go
func handleDomainScan(evt soiree.Event) error {
	payload, ok := evt.Payload().(DomainScanEvent)
	if !ok {
		log.Error().Msg("invalid domain scan payload")
		return nil
	}

	sc := evt.Client().(*ScanClient)
	client := sc.Client
	ctx := evt.Context()

	report, err := requestDomainScan(ctx, sc, payload.Domain)
	status := enums.ScanStatusCompleted
	if err != nil {
		log.Error().Err(err).Msg("domain scan failed")
		status = enums.ScanStatusFailed
	}

	techIDs := []string{}
	for name, info := range report.Technologies {
		description := info.Description
		if description == "" && client.Summarizer != nil {
			sum, serr := client.Summarizer.Summarize(ctx, "Asset info: "+name)
			if serr == nil {
				description = sum
			}
		}
		tech, terr := client.Asset.Query().Where(asset.Name(name)).Only(ctx)
		if terr != nil {
			tech, terr = client.Asset.Create().
				SetAssetType(enums.AssetTypeTechnology).
				SetName(name).
				SetDescription(description).
				SetWebsite(info.Website).
				SetCpe(info.CPE).
				SetCategories(info.Categories).
				Save(ctx)
		}
		if terr != nil {
			log.Error().Err(terr).Msg("persisting asset")
			continue
		}
		techIDs = append(techIDs, tech.ID)
	}

	meta := map[string]any{}
	if b, mErr := json.Marshal(report); mErr == nil {
		_ = json.Unmarshal(b, &meta)
	}

	update := client.Scan.UpdateOneID(payload.ScanID).
		SetStatus(status).
		SetScanType(enums.ScanTypeDomain).
		SetMetadata(meta).
		ClearAssets().
		AddAssetIDs(techIDs...)
	if _, err := update.Save(ctx); err != nil {
		log.Error().Err(err).Msg("updating domain scan")
	}

	return nil
}
```

## Vulnerability scan

```go
func handleVulnerabilityScan(evt soiree.Event) error {
	payload, ok := evt.Payload().(VulnerabilityScanEvent)
	if !ok {
		log.Error().Msg("invalid vulnerability scan payload")
		return nil
	}

	sc := evt.Client().(*ScanClient)
	client := sc.Client
	ctx := evt.Context()

	report, err := requestVulnerabilityScan(ctx, sc, payload.Domain)
	status := enums.ScanStatusCompleted
	if err != nil {
		log.Error().Err(err).Msg("vulnerability scan failed")
		status = enums.ScanStatusFailed
	}

	vulnIDs := []string{}
	for _, v := range report.Vulnerabilities {
		rec, verr := client.Risk.Create().
			SetName(v.CVEID).
			SetRiskType("vulnerability").
			SetDetails(v.Description).
			SetCvssScore(v.Score).
			SetCvss(map[string]any{"vector": v.CVSSVector, "version": v.CVSSVersion}).
			SetReference(v.Reference).
			Save(ctx)
		if verr != nil {
			log.Error().Err(verr).Msg("persisting risk")
			continue
		}

		vulnIDs = append(vulnIDs, rec.ID)

		if v.Score >= defaultTaskThreshold {
			title := fmt.Sprintf("Remediate %s", v.CVEID)
			_, terr := client.Task.Create().
				SetTitle(title).
				SetDetails(v.Description).
				SetCategory("vulnerability remediation").
				SetOwnerID(payload.OrgID).
				SetRiskID(rec.ID).
				Save(ctx)
			if terr != nil {
				log.Error().Err(terr).Msg("creating remediation task")
			}
		}
	}

	meta := map[string]any{}
	if b, mErr := json.Marshal(report); mErr == nil {
		_ = json.Unmarshal(b, &meta)
	}

	update := client.Scan.UpdateOneID(payload.ScanID).
		SetStatus(status).
		SetScanType(enums.ScanTypeVulnerability).
		SetMetadata(meta).
		ClearRisks().
		AddRiskIDs(vulnIDs...)
	if _, err := update.Save(ctx); err != nil {
		log.Error().Err(err).Msg("updating vulnerability scan")
	}

	return nil
}
```
