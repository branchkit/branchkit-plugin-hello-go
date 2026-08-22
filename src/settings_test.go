package main

import (
	"strings"
	"testing"
)

// Copy this alongside settings.go. A settings tab is ordinary string building,
// so it is ordinary to test — which is worth having, because the failure mode
// of a hand-written control is silent: a button posts to a method name that
// does not exist, and the click just does nothing.

func TestRenderSettingsTabDrawsEveryControl(t *testing.T) {
	html := renderSettingsTab()

	for _, want := range []string{
		"Greeting",     // text row label
		"Ending",       // enum row label
		"Shout it",     // bool row label
		"Hello, Drew!", // the live preview, built from the current settings
	} {
		if !strings.Contains(html, want) {
			t.Errorf("settings tab is missing %q", want)
		}
	}

	// Every control must post to a method this plugin actually registers.
	// `branchkit.MethodPost` builds the URL, so what this catches is a method
	// name that was renamed on one side only.
	for _, method := range []string{"set_greeting", "set_punctuation", "set_shout"} {
		if !strings.Contains(html, "/methods/"+method) {
			t.Errorf("no control posts to %q", method)
		}
	}
}

// The descriptions come from the same place the platform's own view reads them,
// so a field described in plugin.json is described identically wherever it is
// drawn. Keep them in the manifest, not inline in the markup.
func TestRowsCarryTheirDescription(t *testing.T) {
	html := renderSettingsTab()
	if !strings.Contains(html, "The word typed before the name.") {
		t.Error("the greeting row lost its description")
	}
}

func TestGreetingReflectsSettings(t *testing.T) {
	// Before the mirror's first fetch, config() falls back to the manifest
	// defaults — so a plugin started cold still greets correctly.
	if got := greetingFor("Drew"); got != "Hello, Drew!" {
		t.Errorf("greetingFor(Drew) = %q, want %q", got, "Hello, Drew!")
	}
}
