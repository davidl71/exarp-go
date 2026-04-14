// tui3270_dispatch.go — TRANS-style command dispatch table for 3270 ISPF command line.
package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/davidl71/exarp-go/internal/database"
	"github.com/racingmars/go3270"
)

// t3270DispatchFn handles a command verb with already uppercased argv tail.
type t3270DispatchFn func(state *tui3270State, currentTx go3270.Tx, args []string) (go3270.Tx, any, error)

// t3270VerbDispatch maps the first command token (uppercase) to a handler (AGENT is handled in handleCommand).
var t3270VerbDispatch map[string]t3270DispatchFn

func init() {
	reg := func(keys []string, fn t3270DispatchFn) {
		for _, k := range keys {
			t3270VerbDispatch[k] = fn
		}
	}
	t3270VerbDispatch = make(map[string]t3270DispatchFn)

	reg([]string{"1"}, func(s *tui3270State, _ go3270.Tx, _ []string) (go3270.Tx, any, error) {
		s.command = ""
		return s.taskListTransaction, s, nil
	})
	reg([]string{"2"}, func(s *tui3270State, _ go3270.Tx, _ []string) (go3270.Tx, any, error) {
		s.command = ""
		return s.configTransaction, s, nil
	})
	reg([]string{"3"}, func(s *tui3270State, _ go3270.Tx, _ []string) (go3270.Tx, any, error) {
		s.command = ""
		return s.scorecardTransaction, s, nil
	})
	reg([]string{"4"}, func(s *tui3270State, _ go3270.Tx, _ []string) (go3270.Tx, any, error) {
		s.command = ""
		return s.handoffTransaction, s, nil
	})
	reg([]string{"5"}, func(s *tui3270State, _ go3270.Tx, _ []string) (go3270.Tx, any, error) {
		s.command = ""
		return nil, nil, nil
	})
	reg([]string{"7"}, func(s *tui3270State, _ go3270.Tx, _ []string) (go3270.Tx, any, error) {
		s.command = ""
		return s.healthTransaction, s, nil
	})

	reg([]string{"SC", "SCORECARD"}, func(s *tui3270State, _ go3270.Tx, _ []string) (go3270.Tx, any, error) {
		s.pushSession("Tasks", s.taskListTransaction)
		s.command = ""
		return s.scorecardTransaction, s, nil
	})
	reg([]string{"HANDOFFS", "HO"}, func(s *tui3270State, _ go3270.Tx, _ []string) (go3270.Tx, any, error) {
		s.pushSession("Tasks", s.taskListTransaction)
		s.command = ""
		return s.handoffTransaction, s, nil
	})
	reg([]string{"MENU", "M", "MAIN"}, func(s *tui3270State, _ go3270.Tx, _ []string) (go3270.Tx, any, error) {
		s.command = ""
		return s.mainMenuTransaction, s, nil
	})
	reg([]string{"TASKS", "T"}, func(s *tui3270State, _ go3270.Tx, _ []string) (go3270.Tx, any, error) {
		s.command = ""
		return s.taskListTransaction, s, nil
	})
	reg([]string{"CONFIG"}, func(s *tui3270State, _ go3270.Tx, _ []string) (go3270.Tx, any, error) {
		s.command = ""
		return s.configTransaction, s, nil
	})
	reg([]string{"HELP", "H"}, func(s *tui3270State, _ go3270.Tx, _ []string) (go3270.Tx, any, error) {
		s.command = ""
		return s.helpTransaction, s, nil
	})
	reg([]string{"HEALTH", "SDSF"}, func(s *tui3270State, _ go3270.Tx, _ []string) (go3270.Tx, any, error) {
		s.pushSession("Tasks", s.taskListTransaction)
		s.command = ""
		return s.healthTransaction, s, nil
	})
	reg([]string{"GIT", "GITLOG"}, func(s *tui3270State, _ go3270.Tx, _ []string) (go3270.Tx, any, error) {
		s.pushSession("Tasks", s.taskListTransaction)
		s.command = ""
		return s.gitDashboardTransaction, s, nil
	})
	reg([]string{"SPRINT", "BOARD"}, func(s *tui3270State, _ go3270.Tx, _ []string) (go3270.Tx, any, error) {
		s.pushSession("Tasks", s.taskListTransaction)
		s.command = ""
		return s.sprintBoardTransaction, s, nil
	})
	reg([]string{"SWAP"}, func(s *tui3270State, _ go3270.Tx, _ []string) (go3270.Tx, any, error) {
		s.command = ""
		sess := s.popSession()
		if sess != nil {
			return sess.tx, s, nil
		}
		return s.mainMenuTransaction, s, nil
	})

	reg([]string{"FIND", "F"}, func(s *tui3270State, _ go3270.Tx, args []string) (go3270.Tx, any, error) {
		if len(args) > 0 {
			s.filter = strings.Join(args, " ")
			s.cursor = 0
			s.listOffset = 0
			ctx := context.Background()

			var err error

			s.tasks, err = s.loadTasksForStatus(ctx, s.status)
			if err == nil {
				filtered := []*database.Todo2Task{}
				searchTerm := strings.ToLower(s.filter)

				for _, task := range s.tasks {
					content := strings.ToLower(task.Content + " " + task.LongDescription)
					if strings.Contains(content, searchTerm) {
						filtered = append(filtered, task)
					}
				}

				s.tasks = filtered
			}
		} else {
			s.filter = ""
			ctx := context.Background()

			var err error

			s.tasks, err = s.loadTasksForStatus(ctx, s.status)
			if err != nil {
				logError(context.Background(), "Error reloading tasks", "error", err, "operation", "reloadTasks")
			}
		}

		s.command = ""

		return s.taskListTransaction, s, nil
	})

	reg([]string{"RESET", "RES"}, func(s *tui3270State, _ go3270.Tx, _ []string) (go3270.Tx, any, error) {
		s.filter = ""
		s.command = ""
		ctx := context.Background()

		var err error

		s.tasks, err = s.loadTasksForStatus(ctx, s.status)
		if err != nil {
			logError(context.Background(), "Error reloading tasks", "error", err, "operation", "reloadTasks")
		}

		return s.taskListTransaction, s, nil
	})

	reg([]string{"EDIT", "E"}, func(s *tui3270State, currentTx go3270.Tx, args []string) (go3270.Tx, any, error) {
		if len(args) > 0 {
			taskID := args[0]
			if strings.HasPrefix(taskID, "T-") {
				ctx := context.Background()

				task, err := getTaskViaMCP(ctx, s.server, taskID)
				if err == nil {
					s.selectedTask = task
					s.command = ""

					return s.taskEditorTransaction, s, nil
				}
			} else {
				var lineNum int
				if _, err := fmt.Sscanf(taskID, "%d", &lineNum); err == nil {
					if lineNum > 0 && lineNum <= len(s.tasks) {
						s.selectedTask = s.tasks[lineNum-1]
						s.cursor = lineNum - 1
						s.command = ""

						return s.taskEditorTransaction, s, nil
					}
				}
			}
		} else if s.cursor < len(s.tasks) {
			s.selectedTask = s.tasks[s.cursor]
			s.command = ""

			return s.taskEditorTransaction, s, nil
		}

		s.command = ""

		return currentTx, s, nil
	})

	reg([]string{"VIEW", "V"}, func(s *tui3270State, currentTx go3270.Tx, args []string) (go3270.Tx, any, error) {
		if len(args) > 0 {
			taskID := args[0]
			if strings.HasPrefix(taskID, "T-") {
				ctx := context.Background()

				task, err := getTaskViaMCP(ctx, s.server, taskID)
				if err == nil {
					s.selectedTask = task
					s.command = ""

					return s.taskDetailTransaction, s, nil
				}
			}
		} else if s.cursor < len(s.tasks) {
			s.selectedTask = s.tasks[s.cursor]
			s.command = ""

			return s.taskDetailTransaction, s, nil
		}

		s.command = ""

		return currentTx, s, nil
	})

	reg([]string{"TOP"}, func(s *tui3270State, _ go3270.Tx, _ []string) (go3270.Tx, any, error) {
		s.cursor = 0
		s.listOffset = 0
		s.command = ""

		return s.taskListTransaction, s, nil
	})

	reg([]string{"BOTTOM", "BOT"}, func(s *tui3270State, _ go3270.Tx, _ []string) (go3270.Tx, any, error) {
		if len(s.tasks) > 0 {
			s.cursor = len(s.tasks) - 1
			mv := 18
			if s.devInfo != nil {
				mv = t3270MaxVisible(s.devInfo)
			}
			if s.cursor >= mv {
				s.listOffset = s.cursor - mv + 1
			}
		}

		s.command = ""

		return s.taskListTransaction, s, nil
	})

	reg([]string{"RUN"}, func(s *tui3270State, _ go3270.Tx, args []string) (go3270.Tx, any, error) {
		s.command = ""

		sub := ""
		if len(args) > 0 {
			sub = args[0]
		}

		switch strings.ToUpper(sub) {
		case "TASK":
			if s.cursor < len(s.tasks) {
				task := s.tasks[s.cursor]
				prompt := PromptForTask(task.ID, task.Content)
				r := RunChildAgent(s.projectRoot, prompt)

				return s.showChildAgentResultTransaction(r.Message, s.taskListTransaction), s, nil
			}

			return s.showChildAgentResultTransaction("No task selected", s.taskListTransaction), s, nil
		case "PLAN":
			prompt := PromptForPlan(s.projectRoot)
			r := RunChildAgent(s.projectRoot, prompt)

			return s.showChildAgentResultTransaction(r.Message, s.taskListTransaction), s, nil
		case "WAVE":
			if len(s.tasks) == 0 {
				return s.showChildAgentResultTransaction("No tasks", s.taskListTransaction), s, nil
			}

			level, ids, err := firstWaveTaskIDs(s.projectRoot, s.tasks)
			if err != nil {
				return s.showChildAgentResultTransaction("No waves", s.taskListTransaction), s, nil
			}

			prompt := PromptForWave(level, ids)
			r := RunChildAgent(s.projectRoot, prompt)

			return s.showChildAgentResultTransaction(r.Message, s.taskListTransaction), s, nil
		case "HANDOFF":
			ctx := context.Background()

			entries, err := fetchHandoffs(ctx, s.server, 5)
			if err != nil || len(entries) == 0 {
				return s.showChildAgentResultTransaction("No handoffs", s.taskListTransaction), s, nil
			}

			h := entries[0]
			steps := make([]interface{}, len(h.NextSteps))
			for i, st := range h.NextSteps {
				steps[i] = st
			}

			prompt := PromptForHandoff(h.Summary, steps)
			r := RunChildAgent(s.projectRoot, prompt)

			return s.showChildAgentResultTransaction(r.Message, s.taskListTransaction), s, nil
		default:
			return s.showChildAgentResultTransaction("RUN TASK|PLAN|WAVE|HANDOFF", s.taskListTransaction), s, nil
		}
	})
}
