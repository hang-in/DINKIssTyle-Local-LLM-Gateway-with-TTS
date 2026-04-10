package chatharness

import (
	"os"
	"testing"

	"dinkisstyle-chat/internal/mcp"
)

func TestSessionTrackerKeepsIntermediateProjectionOutOfDBUntilCompletion(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "session_tracker_*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	dbPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(dbPath)

	if err := mcp.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer mcp.CloseDB()

	sessionEntry, err := mcp.UpsertChatSession(mcp.ChatSessionEntry{
		UserID:      "chat_user",
		SessionKey:  "default",
		Status:      "running",
		UIStateJSON: "{}",
	})
	if err != nil {
		t.Fatalf("UpsertChatSession failed: %v", err)
	}

	tracker := NewSessionTracker(
		"chat_user",
		"turn-1",
		sessionEntry,
		true,
		SessionUISnapshot{ToolCards: map[string]SessionToolCardSnapshot{}, Messages: []SessionMessageSnapshot{}},
		"{}",
	)
	state := SessionPersistState{Status: "running", UIStateJSON: "{}"}

	tracker.AppendEvent(state, "user", "message.created", map[string]interface{}{"content": "hello"})

	current, err := mcp.GetCurrentChatSession("chat_user")
	if err != nil {
		t.Fatalf("GetChatSession after user message failed: %v", err)
	}
	snapshotAfterUser := ParseUISnapshot(current.UIStateJSON)
	if len(snapshotAfterUser.Messages) != 1 || snapshotAfterUser.Messages[0].UserContent != "hello" {
		t.Fatalf("expected user message in snapshot, got %+v", snapshotAfterUser.Messages)
	}

	tracker.AppendEvent(state, "assistant", "message.delta", map[string]interface{}{
		"content":      "partial",
		"full_content": "partial",
	})

	current, err = mcp.GetCurrentChatSession("chat_user")
	if err != nil {
		t.Fatalf("GetChatSession after delta failed: %v", err)
	}
	snapshotAfterDelta := ParseUISnapshot(current.UIStateJSON)
	if len(snapshotAfterDelta.Messages) != 1 {
		t.Fatalf("expected 1 snapshot message after delta, got %d", len(snapshotAfterDelta.Messages))
	}
	if len(snapshotAfterDelta.Turns) != 1 {
		t.Fatalf("expected 1 turn projection after delta, got %d", len(snapshotAfterDelta.Turns))
	}
	eventsAfterDelta, err := mcp.ListChatEvents("chat_user", current.ID, 0, 10)
	if err != nil {
		t.Fatalf("ListChatEvents after delta failed: %v", err)
	}
	if len(eventsAfterDelta) != 0 {
		t.Fatalf("expected no persisted chat events before completion, got %d", len(eventsAfterDelta))
	}
	if snapshotAfterDelta.Messages[0].AssistantContent != "" {
		t.Fatalf("expected intermediate assistant content to stay out of persisted snapshot, got %q", snapshotAfterDelta.Messages[0].AssistantContent)
	}

	tracker.AppendEvent(state, "assistant", "request.complete", map[string]interface{}{
		"final_assistant_content": "final answer",
	})

	current, err = mcp.GetCurrentChatSession("chat_user")
	if err != nil {
		t.Fatalf("GetChatSession after completion failed: %v", err)
	}
	snapshotAfterComplete := ParseUISnapshot(current.UIStateJSON)
	if len(snapshotAfterComplete.Messages) != 1 {
		t.Fatalf("expected 1 snapshot message after completion, got %d", len(snapshotAfterComplete.Messages))
	}
	if snapshotAfterComplete.Messages[0].AssistantContent != "final answer" {
		t.Fatalf("expected final assistant content after completion, got %q", snapshotAfterComplete.Messages[0].AssistantContent)
	}
	if len(snapshotAfterComplete.Turns) != 1 || snapshotAfterComplete.Turns[0].Status != "completed" {
		t.Fatalf("expected completed turn projection after completion, got %+v", snapshotAfterComplete.Turns)
	}
	eventsAfterComplete, err := mcp.ListChatEvents("chat_user", current.ID, 0, 10)
	if err != nil {
		t.Fatalf("ListChatEvents after completion failed: %v", err)
	}
	if len(eventsAfterComplete) != 1 {
		t.Fatalf("expected only completion event after completion, got %d", len(eventsAfterComplete))
	}
	if eventsAfterComplete[0].EventType != "request.complete" {
		t.Fatalf("expected request.complete as final persisted event, got %+v", eventsAfterComplete)
	}
}
