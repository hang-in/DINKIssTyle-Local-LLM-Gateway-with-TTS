package core

import (
	"strings"
	"testing"
)

func TestCompactRecentTurnContentPreservesSignalLines(t *testing.T) {
	long := strings.Repeat("intro text ", 40) +
		"\nerror: failed to open /tmp/demo.txt: permission denied" +
		"\n$ go test ./..." +
		"\n" + strings.Repeat("tail text ", 30)

	got := compactRecentTurnContent(long, 220)

	if !strings.Contains(got, "error: failed to open /tmp/demo.txt") {
		t.Fatalf("expected error line to survive compaction, got %q", got)
	}
	if !strings.Contains(got, "$ go test ./...") {
		t.Fatalf("expected command line to survive compaction, got %q", got)
	}
	if len([]rune(got)) > 260 {
		t.Fatalf("expected compacted text to stay reasonably close to budget, got %d chars", len([]rune(got)))
	}
}

func TestBuildRecentContextFromSnapshotUsesLatestTurnPriority(t *testing.T) {
	snapshot := chatSessionUISnapshot{
		Messages: []chatSessionMessageSnapshot{
			{
				TurnID:           "turn-1",
				UserContent:      "Earlier user question that should be compacted because it is older.",
				AssistantContent: "Earlier assistant answer that should also be compacted.",
			},
			{
				TurnID:           "turn-2",
				UserContent:      strings.Repeat("latest user context ", 60) + "\n/path/to/file.go:42",
				AssistantContent: strings.Repeat("latest assistant context ", 50) + "\nerror: panic recovered",
			},
		},
	}

	got, turns := buildRecentContextFromSnapshot(snapshot, 4)

	if turns != 2 {
		t.Fatalf("expected 2 turns, got %d", turns)
	}
	if !strings.Contains(got, "Turn -2") || !strings.Contains(got, "Turn -1") {
		t.Fatalf("expected relative turn labels in context, got %q", got)
	}
	if !strings.Contains(got, "/path/to/file.go:42") {
		t.Fatalf("expected latest turn file path to survive, got %q", got)
	}
	if !strings.Contains(got, "error: panic recovered") {
		t.Fatalf("expected latest assistant signal to survive, got %q", got)
	}
}
