package tools

import (
	"context"
	"testing"

	"github.com/davidl71/exarp-go/internal/models"
)

func TestResourceNotificationsTrackUpdates(t *testing.T) {
	ctx := context.Background()
	ctx = WithResourceUpdateContext(ctx)
	ctx = MarkResourceUpdated(ctx, "stdio://tasks")
	ctx = MarkResourceUpdated(ctx, "stdio://scorecard")

	got := ResourceUpdates(ctx)
	if len(got) != 2 {
		t.Fatalf("unexpected resource updates: %v", got)
	}
}

func TestMarkTaskResourcesChangedIncludesTags(t *testing.T) {
	ctx := context.Background()
	ctx = WithResourceUpdateContext(ctx)
	task := &models.Todo2Task{
		ID:       "T-123",
		Status:   "todo",
		Priority: "high",
		Tags:     []string{"dev", "urgent"},
	}

	ctx = MarkTaskResourcesChanged(ctx, task)

	wantURI := "stdio://tasks/tag/urgent"
	got := ResourceUpdates(ctx)
	found := false
	for _, uri := range got {
		if uri == wantURI {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("expected %s in updates %v", wantURI, got)
	}
}
