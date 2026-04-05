// task_wave.go — wave remaining helpers (parallel-execution-waves.json ∩ open tasks).
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/models"
	"github.com/davidl71/exarp-go/internal/tools"
	mcpcli "github.com/davidl71/mcp-go-core/pkg/mcp/cli"
)

type waveFile struct {
	Waves map[string][]string `json:"waves"`
}

// handleTaskWaveCommand implements `task wave remaining <0|1|2> [--batch N]` and `task wave ids <0|1|2>`.
func handleTaskWaveCommand(parsed *mcpcli.Args) error {
	pos := parsed.Positional
	if len(pos) < 2 {
		return fmt.Errorf(`usage: task wave remaining <0|1|2> [--batch 15]
       task wave ids <0|1|2>`)
	}
	sub := pos[0]
	waveIdx, err := strconv.Atoi(pos[1])
	if err != nil || waveIdx < 0 || waveIdx > 2 {
		return fmt.Errorf("wave index must be 0, 1, or 2")
	}

	root, err := tools.FindProjectRoot()
	if err != nil {
		return err
	}
	path := filepath.Join(root, ".cursor", "plans", "parallel-execution-waves.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read waves file %s: %w (set PROJECT_ROOT or add .cursor/plans/parallel-execution-waves.json)", path, err)
	}
	var wf waveFile
	if err := json.Unmarshal(raw, &wf); err != nil {
		return fmt.Errorf("parse waves json: %w", err)
	}
	key := strconv.Itoa(waveIdx)
	ids := wf.Waves[key]
	if len(ids) == 0 {
		return fmt.Errorf("no task ids for wave %d in %s", waveIdx, path)
	}

	switch sub {
	case "ids":
		for _, id := range ids {
			fmt.Println(id)
		}
		return nil
	case "remaining":
		open, err := database.ListTasks(context.Background(), &database.TaskFilters{Statuses: models.OpenStatuses()})
		if err != nil {
			return fmt.Errorf("list open tasks: %w", err)
		}
		openSet := make(map[string]bool, len(open))
		for _, t := range open {
			if t != nil {
				openSet[t.ID] = true
			}
		}
		var rem []string
		for _, id := range ids {
			if openSet[id] {
				rem = append(rem, id)
			}
		}
		batch := 0
		if b, err := strconv.Atoi(parsed.GetFlag("batch", "")); err == nil && b > 0 {
			batch = b
		}
		if CLIOutputOpts.JSON {
			out := map[string]interface{}{
				"wave":            waveIdx,
				"remaining_count": len(rem),
				"remaining_ids":   rem,
			}
			if batch > 0 && len(rem) > batch {
				out["next_batch_ids"] = rem[:batch]
			}
			raw, _ := json.Marshal(out)
			fmt.Println(string(raw))
			return nil
		}
		fmt.Printf("# Wave %d remaining: %d (open statuses ∩ wave ids)\n", waveIdx, len(rem))
		for _, id := range rem {
			fmt.Println(id)
		}
		if batch > 0 && len(rem) > batch {
			fmt.Printf("\n# Next batch (first %d):\n", batch)
			fmt.Println(strings.Join(rem[:batch], "\n"))
		}
		return nil
	default:
		return fmt.Errorf("unknown wave subcommand %q (use: remaining, ids)", sub)
	}
}
