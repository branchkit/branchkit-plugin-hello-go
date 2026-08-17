package main

// The settings template. Copy this file into your plugin and edit it — that is
// what it is for. Nothing here is imported from a framework, so you can delete
// half of it, restyle every row, or replace the whole thing with a purpose-built
// editor without asking anyone for a feature.
//
// The split it demonstrates:
//
//	the platform owns the DATA  — which fields exist, their shipped defaults,
//	                              storing the user's changes, composing the two
//	your plugin owns the DRAWING — the markup, the layout, what a click means
//
// That line is deliberate. Sharing the data layer is what removed a whole class
// of bug from this codebase (defaults baked into saved values, so a better
// default never reached anyone who had saved once; a zero value indistinguishable
// from a user choosing zero). Sharing the MARKUP is what nobody should do: every
// plugin's settings want to look different, and a control drawn from a schema
// converges on the blandest thing that fits every schema.

import (
	"encoding/json"
	"fmt"
	"html"

	shared "github.com/branchkit/plugin-sdk-go"
)

// The collection declared in plugin.json under `provides.collections`, with
// `preset: settings`. The platform materializes every declared field at its
// shipped default and applies the user's sparse changes on top, so an untouched
// field always reads the CURRENT default rather than a stale copy saved years
// ago.
const settingsCollection = "plugin.helloworld.config"

// Config mirrors the manifest's field keys. Unknown fields are ignored, so this
// struct may lag the manifest without breaking.
type Config struct {
	Greeting    string `json:"greeting"`
	Punctuation string `json:"punctuation"`
	Shout       bool   `json:"shout"`
}

var settings *shared.SettingsMirror[Config]

// initSettings wires the typed mirror. Must run before plugin.Run() so the
// first fetch lands.
//
// Your manifest must also subscribe to `_platform.collection.updated` — see
// plugin.json. Without it the mirror never hears about changes made anywhere
// else (the Collections tab, another window) and your plugin reads a stale
// value until it restarts. That is a real bug that shipped in this repo once.
func initSettings(p *shared.Plugin) {
	settings = shared.Settings[Config](p, settingsCollection)
}

// config returns the composed settings, with a fallback for the window before
// the first fetch lands. The fallback should match the manifest's defaults;
// plugin.json is the authoritative copy.
func config() Config {
	if settings != nil && settings.Ready() {
		return settings.Get()
	}
	return Config{Greeting: "Hello", Punctuation: "exclamation", Shout: false}
}

// --- Writing: one relay helper, and one handler per control -----------------

// setUserField records ONE user gesture. Settings are `writers: platform_only`
// — a plugin never saves settings on its own initiative — so this says
// `tenant: "_user"`: the choice is the user's and this plugin is the transport.
//
// The platform records the write as `relayed` rather than `user`, because the
// click happened on a surface it did not draw and it will not claim otherwise
// on your behalf. That is visible to the user, who can undo it. Do not use this
// path for anything the user did not just ask for: state your plugin decides on
// its own is domain data, and belongs in its own collection.
func setUserField(key string, value any) error {
	raw, err := json.Marshal(map[string]any{key: value})
	if err != nil {
		return err
	}
	id := settingsCollection
	tenant := "_user"
	_, err = plugin.OverridesApply("patch", settingsCollection, nil, raw, &id, nil, &tenant)
	return err
}

type setGreetingRequest struct {
	Greeting string `json:"greeting"`
}

func handleSetGreeting(req *setGreetingRequest) (any, error) {
	return nil, setUserField("greeting", req.Greeting)
}

type setPunctuationRequest struct {
	Punctuation string `json:"punctuation"`
}

func handleSetPunctuation(req *setPunctuationRequest) (any, error) {
	// Validate before relaying. The platform stores what you send it; a typo
	// here becomes a value your own reader has to defend against later.
	switch req.Punctuation {
	case "period", "exclamation", "none":
	default:
		return nil, fmt.Errorf("unknown punctuation %q", req.Punctuation)
	}
	return nil, setUserField("punctuation", req.Punctuation)
}

type setShoutRequest struct {
	Shout bool `json:"shout"`
}

func handleSetShout(req *setShoutRequest) (any, error) {
	return nil, setUserField("shout", req.Shout)
}

// registerSettingsHandlers wires the three controls above.
func registerSettingsHandlers(p *shared.Plugin) {
	shared.HandleTyped(p, "set_greeting", handleSetGreeting)
	shared.HandleTyped(p, "set_punctuation", handleSetPunctuation)
	shared.HandleTyped(p, "set_shout", handleSetShout)
}

// --- Drawing: your markup, start to finish ----------------------------------

// renderSettingsTab draws the Settings tab. Plain string building, because a
// template language is a choice each plugin makes for itself — several plugins
// in this repo use templ instead, and the platform does not care.
//
// The controls post to this plugin's own methods with `shared.MethodPost`,
// which builds the `/v1/plugins/<id>/methods/<name>` URL so the id is never
// spelled by hand. Datastar posts it; the tab re-renders from the server.
// No client-side state, no fetch, no JSON assembled in the browser.
func renderSettingsTab() string {
	c := config()

	rows := textRow(
		"Greeting", "The word typed before the name.",
		c.Greeting,
		// $el.value is the input's current text at click time.
		shared.MethodPost("set_greeting", "{greeting: $el.previousElementSibling.value}"),
	) + enumRow(
		"Ending", "How the greeting finishes.",
		c.Punctuation,
		[][2]string{
			{"period", "Hello, Drew."},
			{"exclamation", "Hello, Drew!"},
			{"none", "Hello, Drew"},
		},
	) + boolRow(
		"Shout it", "Type the greeting in capitals.",
		c.Shout,
		shared.MethodPost("set_shout", fmt.Sprintf("{shout: %t}", !c.Shout)),
	)

	return `<div id="settings-table-container" style="padding: 16px; font-family: system-ui;">` +
		`<h2 style="margin: 0 0 4px 0; font-size: 16px;">Greeting</h2>` +
		`<p style="color: #888; margin: 0 0 16px 0; font-size: 13px;">Shapes what the ` +
		`<em>hello branchkit</em> command types at your cursor.</p>` +
		`<div class="settings-table">` + rows + `</div>` +
		`<p style="color: #666; margin: 16px 0 0 0; font-size: 12px;">Preview: ` +
		html.EscapeString(greetingFor("Drew")) + `</p>` +
		`</div>`
}

// row is the shared shell: a label, the line saying what the setting does, and
// whatever control you hand it. The description belongs next to the control —
// a label names a setting, the description says why you would touch it.
func row(label, description, control string) string {
	return `<div class="settings-row" style="display: flex; align-items: center; gap: 16px; padding: 12px 0; border-bottom: 1px solid #2a2a2a;">` +
		`<div style="flex: 1;">` +
		`<div style="font-size: 14px; font-weight: 600;">` + html.EscapeString(label) + `</div>` +
		`<div style="font-size: 12px; color: #888; margin-top: 2px;">` + html.EscapeString(description) + `</div>` +
		`</div><div>` + control + `</div></div>`
}

func textRow(label, description, value, onSave string) string {
	return row(label, description,
		`<input type="text" value="`+html.EscapeString(value)+`" `+
			`style="background: #1c1c1e; border: 1px solid #3a3a3c; color: inherit; border-radius: 4px; padding: 4px 8px; font-size: 13px;">`+
			`<button style="margin-left: 8px; font-size: 12px;" data-on:click="`+html.EscapeString(onSave)+`">Save</button>`)
}

func boolRow(label, description string, on bool, onToggle string) string {
	state := "Off"
	if on {
		state = "On"
	}
	return row(label, description,
		`<button style="min-width: 56px; font-size: 12px;" data-on:click="`+html.EscapeString(onToggle)+`">`+state+`</button>`)
}

// enumRow shows one button per value. Note the option LABELS — "Hello, Drew!"
// reads better than the stored value `exclamation`, and only your plugin knows
// that. This is the clearest case for owning your own markup.
func enumRow(label, description, current string, options [][2]string) string {
	buttons := ""
	for _, opt := range options {
		value, text := opt[0], opt[1]
		style := "font-size: 12px; margin-left: 6px;"
		if value == current {
			style += " font-weight: 700; border-color: #4a9eff;"
		}
		buttons += `<button style="` + style + `" data-on:click="` +
			html.EscapeString(shared.MethodPost("set_punctuation", `{punctuation: '`+value+`'}`)) +
			`">` + html.EscapeString(text) + `</button>`
	}
	return row(label, description, buttons)
}

// greetingFor applies the settings — the reason any of this exists.
func greetingFor(name string) string {
	c := config()
	text := c.Greeting + ", " + name
	switch c.Punctuation {
	case "period":
		text += "."
	case "exclamation":
		text += "!"
	}
	if c.Shout {
		text = shoutText(text)
	}
	return text
}
