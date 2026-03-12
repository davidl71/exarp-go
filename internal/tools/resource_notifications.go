package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/davidl71/exarp-go/internal/models"
)

type resourceUpdatesCtxKey struct{}

var resourceUpdatesKey = resourceUpdatesCtxKey{}

// WithResourceUpdateContext seeds the context with the map used for tracking updates.
func WithResourceUpdateContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, resourceUpdatesKey, make(map[string]struct{}))
}

// MarkResourceUpdated records that a specific resource URI has changed.
func MarkResourceUpdated(ctx context.Context, uri string) context.Context {
	if uri == "" || ctx == nil {
		return ctx
	}

	raw := ctx.Value(resourceUpdatesKey)
	updates, _ := raw.(map[string]struct{})
	if updates == nil {
		return ctx
	}

	updates[uri] = struct{}{}
	return ctx
}

// NotifyResources marks multiple URIs as changed.
func NotifyResources(ctx context.Context, uris ...string) context.Context {
	for _, uri := range uris {
		ctx = MarkResourceUpdated(ctx, uri)
	}
	return ctx
}

// ResourceUpdates returns the list of URIs that were marked during this request.
func ResourceUpdates(ctx context.Context) []string {
	if ctx == nil {
		return nil
	}

	raw := ctx.Value(resourceUpdatesKey)
	updates, _ := raw.(map[string]struct{})
	if len(updates) == 0 {
		return nil
	}

	list := make([]string, 0, len(updates))
	for uri := range updates {
		list = append(list, uri)
	}
	return list
}

// MarkTaskResourcesChanged marks the task-related resources as updated.
func MarkTaskResourcesChanged(ctx context.Context, tasks ...*models.Todo2Task) context.Context {
	const (
		tasksURI          = "stdio://tasks"
		tasksSummaryURI   = "stdio://tasks/summary"
		suggestedTasksURI = "stdio://suggested-tasks"
		scorecardURI      = "stdio://scorecard"
	)

	ctx = NotifyResources(ctx, tasksURI, tasksSummaryURI, suggestedTasksURI, scorecardURI)

	for _, task := range tasks {
		if task == nil {
			continue
		}

		ctx = NotifyResources(ctx,
			fmt.Sprintf("stdio://tasks/%s", task.ID),
			fmt.Sprintf("stdio://tasks/status/%s", NormalizeStatusToTitleCase(task.Status)),
		)

		if priority := strings.TrimSpace(task.Priority); priority != "" {
			ctx = NotifyResources(ctx, fmt.Sprintf("stdio://tasks/priority/%s", priority))
		}

		for _, tag := range task.Tags {
			if t := strings.TrimSpace(tag); t != "" {
				ctx = NotifyResources(ctx, fmt.Sprintf("stdio://tasks/tag/%s", t))
			}
		}
	}

	return ctx
}
