## Changelog
{{ range .Versions }}
<a name="{{ .Tag.Name }}"></a>
## {{ if .Tag.Previous }}{{ .Tag.Name }}{{ else }}{{ .Tag.Name }}{{ end }} - {{ datetime "2006-01-02" .Tag.Date }}

{{ if .CommitGroups -}}
{{ range .CommitGroups -}}
### {{ .Title }}
{{ range .Commits -}}
- {{.Hash.Short}} {{ if .Scope }}**{{ .Scope }}:** {{ end }}{{ if .Subject }}{{ .Subject }}{{ else }}{{ .Header }}{{ end }}
{{ end }}
{{ end }}
{{ end -}}

### Commits
{{ range .Commits -}}
- {{.Hash.Short}} {{ if .Scope }}**{{ .Scope }}:** {{ end }}{{ if .Subject }}{{ .Subject }}{{ else }}{{ .Header }}{{ end }}
{{ end }}

{{- if .NoteGroups -}}
{{ range .NoteGroups -}}
### {{ .Title }}
{{ range .Notes }}
{{ .Body }}
{{ end }}
{{ end -}}
{{ end -}}

{{ end -}}