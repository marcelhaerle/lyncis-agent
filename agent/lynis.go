package agent

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/marcelhaerle/lyncis-agent/api"
)

func RunLynis() (api.ScanPayload, error) {
	payload := api.ScanPayload{
		RawData:  make(map[string]interface{}),
		Findings: []api.ScanFinding{},
	}

	// 1. Execute Lynis
	// Note: Run() will return an error if the exit code is non-zero, which Lynis might do if it finds warnings.
	// We ignore the error and prioritize parsing the report file instead.
	cmd := exec.Command("lynis", "audit", "system", "--cronjob")
	_ = cmd.Run()

	// 2. Parse report file
	reportPath := "/var/log/lynis-report.dat"
	file, err := os.Open(reportPath)
	if err != nil {
		return payload, fmt.Errorf("failed to open lynis report: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		payload.RawData[key] = val

		// Extract hardening index
		if key == "hardening_index" {
			if idx, err := strconv.Atoi(val); err == nil {
				payload.HardeningIndex = idx
			}
		}

		// Extract warnings and suggestions
		// Format is typically TEST-ID|Description related to finding|
		if strings.HasPrefix(key, "warning[]") {
			fParts := strings.Split(val, "|")
			if len(fParts) >= 2 {
				payload.Findings = append(payload.Findings, api.ScanFinding{
					Severity:    "warning",
					TestID:      fParts[0],
					Description: fParts[1],
				})
			}
		} else if strings.HasPrefix(key, "suggestion[]") {
			fParts := strings.Split(val, "|")
			if len(fParts) >= 2 {
				payload.Findings = append(payload.Findings, api.ScanFinding{
					Severity:    "suggestion",
					TestID:      fParts[0],
					Description: fParts[1],
				})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return payload, fmt.Errorf("error reading lynis report: %w", err)
	}

	return payload, nil
}
