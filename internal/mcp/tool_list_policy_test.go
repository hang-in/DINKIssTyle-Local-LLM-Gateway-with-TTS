package mcp

import "testing"

func TestGetToolListForContextFiltersDisabledTools(t *testing.T) {
	tools := GetToolListForContext(true, []string{"search_web", "search_memory"})
	seen := make(map[string]bool, len(tools))
	for _, tool := range tools {
		seen[tool.Name] = true
	}

	if seen["search_web"] {
		t.Fatalf("expected search_web to be filtered from tools/list")
	}
	if seen["search_memory"] {
		t.Fatalf("expected search_memory to be filtered from tools/list")
	}
	if !seen["read_web_page"] {
		t.Fatalf("expected unrelated tool to remain visible")
	}
}

func TestGetToolListForContextFiltersMemoryToolsWhenMemoryDisabled(t *testing.T) {
	tools := GetToolListForContext(false, nil)
	seen := make(map[string]bool, len(tools))
	for _, tool := range tools {
		seen[tool.Name] = true
	}

	for _, memoryTool := range []string{"search_memory", "read_memory", "read_memory_context", "delete_memory", "save_user_fact", "delete_user_fact"} {
		if seen[memoryTool] {
			t.Fatalf("expected %s to be hidden when memory is disabled", memoryTool)
		}
	}
	if !seen["search_web"] {
		t.Fatalf("expected non-memory tool to remain visible")
	}
}
