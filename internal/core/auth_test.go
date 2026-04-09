package core

import (
	"reflect"
	"testing"
)

func TestExpandDisabledToolAliases(t *testing.T) {
	got := expandDisabledToolAliases([]string{"personal_memory", "search_web", "search_memory"})
	want := []string{"search_memory", "read_memory", "read_memory_context", "delete_memory", "search_web"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expandDisabledToolAliases() = %#v, want %#v", got, want)
	}
}

func TestCollapseDisabledToolsForUI(t *testing.T) {
	got := collapseDisabledToolsForUI([]string{"search_memory", "read_memory_context", "search_web"})
	want := []string{"personal_memory", "search_web"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("collapseDisabledToolsForUI() = %#v, want %#v", got, want)
	}
}
