package main

import (
	"strings"

	shared "github.com/branchkit/plugin-sdk-go"
)

type RenderSettingsRequest struct {
	TabKey string `json:"tab_key"`
}

// plugin is the process-wide handle, so the settings relay in settings.go can
// reach it without threading it through every call.
var plugin *shared.Plugin

// shoutText is here rather than in settings.go to keep the template file free
// of anything specific to this example.
func shoutText(s string) string { return strings.ToUpper(s) }

func main() {
	plugin = shared.NewPlugin()

	// Settings: declared in plugin.json, stored by the platform, read here
	// through a typed mirror. See settings.go — that file is the template to
	// copy into your own plugin.
	initSettings(plugin)
	registerSettingsHandlers(plugin)

	// HandleGreet and GreetParams are both generated from plugin.json into
	// actions_gen.go by `just gen-plugins`, so the action string is never
	// spelled here and the params arrive typed. Prefer this over the untyped
	// plugin.HandleAction — this plugin is the example others copy.
	HandleGreet(plugin, func(p GreetParams, _ *shared.OnActionRequest) (any, error) {
		name := "BranchKit"
		if p.Name != nil {
			name = *p.Name
		}

		// Generated wrapper, not plugin.Call("input.type_text", map[string]any{…}):
		// the method name and the argument shape are checked at compile time.
		// The greeting is built from the user's settings, which is the whole
		// point of declaring them: read the composed value, never a copy this
		// plugin saved for itself.
		return nil, plugin.InputTypeText(greetingFor(name))
	})

	shared.HandleTyped(plugin, "render_settings", func(req *RenderSettingsRequest) (any, error) {
		if req.TabKey == "settings" {
			return shared.RenderSettingsResponse{HTML: renderSettingsTab()}, nil
		}
		return shared.RenderSettingsResponse{
			HTML: `<div style="padding: 16px; font-family: system-ui;">
	<h2 style="margin: 0 0 12px 0;">Hello World Plugin</h2>
	<p style="color: #888; margin: 0 0 16px 0;">A minimal BranchKit plugin that types a greeting at the cursor.</p>

	<h3 style="margin: 0 0 8px 0;">Keybind</h3>
	<p>Press <kbd style="background: #333; padding: 2px 6px; border-radius: 3px;">Alt+Shift+H</kbd> to type "Hello, BranchKit!" at your cursor.</p>

	<h3 style="margin: 16px 0 8px 0;">Voice Commands</h3>
	<p style="color: #888; margin: 0 0 8px 0;">Activate command mode (see Voice plugin settings for your keybind), then say:</p>
	<table style="border-collapse: collapse; width: 100%;">
		<tr>
			<td style="padding: 6px 12px; border-bottom: 1px solid #333;"><em>"hello branchkit"</em></td>
			<td style="padding: 6px 12px; border-bottom: 1px solid #333; color: #888;">Types "Hello, BranchKit!"</td>
		</tr>
		<tr>
			<td style="padding: 6px 12px;"><em>"hello &lt;name&gt;"</em></td>
			<td style="padding: 6px 12px; color: #888;">Types "Hello, &lt;name&gt;!" with any spoken word</td>
		</tr>
	</table>

	<p style="color: #666; margin: 16px 0 0 0; font-size: 13px;">
		Open a text editor first — the greeting is typed at your cursor position.
	</p>
</div>`,
		}, nil
	})

	plugin.OnReady(func() {
		shared.Log("helloworld", "all plugins ready")
	})

	plugin.Run()
}
