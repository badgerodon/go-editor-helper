import json, urllib.request

import sublime, sublime_plugin

def rpc(method, params):
	req = urllib.request.Request(
		"http://127.0.0.1:9999/jsonrpc",
		data=bytearray(json.dumps({
			"method": "Service." + method,
			"params": params,
			"id": 1
		}), "utf8"),
		headers={"Content-Type": "application/json"},
		method="POST"
	)
	f = urllib.request.urlopen(req)
	res = str(f.read(), "utf8")
	f.close()
	res = json.loads(res)
	return (res['result'], res['error'])

def sel(view, i=0):
	try:
		s = view.sel()
		if s is not None and i < len(s):
			return s[i]
	except Exception:
		pass

	return sublime.Region(0, 0)

def is_go_source_view(view=None):
	if view is None:
		return False

	return view.score_selector(sel(view).begin(), "source.go") > 0

class GoFormatCommand(sublime_plugin.TextCommand):
	def is_enabled(self):
		return is_go_source_view(self.view)

	def run(self, edit):
		variables = self.view.window().extract_variables()
		region = sublime.Region(0, self.view.size())

		res, err = rpc("Format", [{
			"file": variables["file"],
			"text": self.view.substr(region)
		}])
		if err:
			sublime.error_message("format error: " + err)
		else:
			self.view.replace(edit, region, res["text"])

class GoGoToDefinitionCommand(sublime_plugin.TextCommand):
	def is_enabled(self):
		return is_go_source_view(self.view)

	def run(self, edit):
		variables = self.view.window().extract_variables()
		s = self.view.sel()
		if s is not None and len(s) > 0:
			res, err = rpc("GoToDefinition", [{
				"file": variables["file"],
				"text": self.view.substr(sublime.Region(0, self.view.size())),
				"offset": s[0].a
			}])
			if err:
				sublime.error_message("go_to_definition error: " + err)
			else:
				view = sublime.active_window().open_file(
					"{}:{}:{}".format(res["file"], res["line"], res["column"]),
					sublime.ENCODED_POSITION
				)

class GoLintCommand(sublime_plugin.TextCommand):
	def is_enabled(self):
		return is_go_source_view(self.view)

	def run(self, edit):
		variables = self.view.window().extract_variables()
		res, err = rpc("Lint", [{
			"file": variables["file"]
		}])
		if err:
			sublime.error_message("lint error: " + err)
		else:
			print("RESPONSE", res)


class Events(sublime_plugin.EventListener):
	def on_pre_save(self, view):
		if is_go_source_view(view):
			view.run_command("go_format")

	def on_post_save(self, view):
		if is_go_source_view(view):
			view.run_command("go_lint")

def plugin_loaded():
	sublime.log_commands(True)
