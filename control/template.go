package control

import "html/template"

var (
	rootTmpl = template.Must(template.New("__root__").Parse(
		`
{{define "index.html"}}
{{ $defaultTarget := .DefaultStressorTarget }}
<!doctype html>
<html>
	<head>
		<title>Learn Some System Design - LSD - Dashboard</title>
		<link rel="stylesheet" href="/static/styles/main.css">
		<link rel="stylesheet" href="/static/styles/theme.css">
	</head>
	<body>
		<article class="content">
			<h1>Instances</h1>
			<table>
				<thead>
					<tr>
						<th>Name</th>
						<th>Number of requests</th>
					</tr>
				</thead>
				<tbody>
				{{ range $instanceName, $data := .Instances }}
					<tr>
						<td>{{ $data.Name }}</td>
						<td>{{ $data.Metrics.Requests }}</td>
					</tr>
				{{ end }}
				</tbody>
			</table>
		</article>
		<article class="content">
			<h1>Services</h1>
			<table>
				<thead>
					<tr>
						<th>Name</th>
					</tr>
				</thead>
				<tbody>
				{{ range $idx, $data := .Servers }}
					<tr>
						<td><a rel="no-follow" href="{{ $data.Endpoint }}">{{ $data.Service }}</a> ({{ $data.Endpoint }})</td>
					</tr>
				{{ end }}
				</tbody>
			</table>
		</article>
		<article class="content">
			<h1>Stressors</h1>
			<table>
				<thead>
					<tr>
						<th>Name</th>
						<th>Trigger</th>
					</tr>
				</thead>
				<tbody>
				{{ range $idx, $data := .Stressors }}
					<tr>
						<td>
							<a rel="no-follow" href="{{ $data.BaseEndpoint }}/">{{ $data.Name }}</a> ({{ $data.BaseEndpoint }})
						</td>
						<td>
							<form method="POST" action="/actions/trigger-stressor/{{ $data.Name }}">
								<input name="target.endpoint" type="text" value="{{ $defaultTarget }}">
								<button type="submit">Go</button>
							</form>
						</td>
					</tr>
				{{ end }}
				</tbody>
			</table>
		</article>
	</body>
</html>
{{end}}
`))
)
