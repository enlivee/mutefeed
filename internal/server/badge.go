package server

import (
	"bytes"
	"fmt"
	"html/template"
)

var badgeTmpl = template.Must(template.New("badge").Parse(`<svg xmlns="http://www.w3.org/2000/svg" width="{{.TotalWidth}}" height="20" role="img" aria-label="{{.LeftLabel}}: {{.RightLabel}}">
  <title>{{.LeftLabel}}: {{.RightLabel}}</title>
  <linearGradient id="s" x2="0" y2="100%">
    <stop offset="0" stop-color="#bbb" stop-opacity=".1"/>
    <stop offset="1" stop-opacity=".1"/>
  </linearGradient>
  <clipPath id="r">
    <rect width="{{.TotalWidth}}" height="20" rx="3" fill="#fff"/>
  </clipPath>
  <g clip-path="url(#r)">
    <rect width="{{.LeftWidth}}" height="20" fill="{{.LeftColor}}"/>
    <rect x="{{.LeftWidth}}" width="{{.RightWidth}}" height="20" fill="{{.RightColor}}"/>
    <rect width="{{.TotalWidth}}" height="20" fill="url(#s)"/>
  </g>
  <g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" text-rendering="geometricPrecision" font-size="110">
    <text aria-hidden="true" x="{{.LeftX}}" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="{{.LeftTextLen}}">{{.LeftLabel}}</text>
    <text x="{{.LeftX}}" y="140" transform="scale(.1)" fill="#fff" textLength="{{.LeftTextLen}}">{{.LeftLabel}}</text>
    <text aria-hidden="true" x="{{.RightX}}" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="{{.RightTextLen}}">{{.RightLabel}}</text>
    <text x="{{.RightX}}" y="140" transform="scale(.1)" fill="#fff" textLength="{{.RightTextLen}}">{{.RightLabel}}</text>
  </g>
</svg>`))

type badgeData struct {
	TotalWidth   int
	LeftWidth    int
	RightWidth   int
	LeftX        int
	RightX       int
	LeftTextLen  int
	RightTextLen int
	LeftLabel    string
	RightLabel   string
	LeftColor    string
	RightColor   string
}

// measureTextWidth approximates character widths for 11px Verdana/sans-serif font.
func measureTextWidth(s string) int {
	w := 0
	for _, r := range s {
		switch {
		case r == 'm' || r == 'M' || r == 'W' || r == 'w':
			w += 10
		case r == '%':
			w += 9
		case r == 'i' || r == 'l' || r == '!' || r == '|' || r == ':' || r == ';' || r == '\'':
			w += 3
		case r == 'f' || r == 't' || r == 'j' || r == 'r' || r == ' ' || r == '1' || r == '-' || r == '.' || r == '(' || r == ')':
			w += 5
		case r >= 'A' && r <= 'Z':
			w += 8
		default: // a-z, digits 0, 2-9
			w += 7
		}
	}
	if w < 10 {
		w = 10
	}
	return w
}

// BadgeColor determines the right-side badge color based on clean percentage.
// >= 90%: emerald (#2ed573)
// >= 70%: amber (#ffa502)
// < 70%:  ruby (#ff4757)
func BadgeColor(percentage int) string {
	if percentage >= 90 {
		return "#2ed573"
	}
	if percentage >= 70 {
		return "#ffa502"
	}
	return "#ff4757"
}

// GenerateBadge builds a Shields.io-style SVG badge with the provided labels and colors.
func GenerateBadge(left, right, rightColor string) (string, error) {
	leftTextW := measureTextWidth(left)
	rightTextW := measureTextWidth(right)
	leftWidth := leftTextW + 14
	rightWidth := rightTextW + 14
	totalWidth := leftWidth + rightWidth

	leftX := (leftWidth * 10) / 2
	rightX := (leftWidth * 10) + ((rightWidth * 10) / 2)
	leftTextLen := leftTextW * 10
	rightTextLen := rightTextW * 10

	data := badgeData{
		TotalWidth:   totalWidth,
		LeftWidth:    leftWidth,
		RightWidth:   rightWidth,
		LeftX:        leftX,
		RightX:       rightX,
		LeftTextLen:  leftTextLen,
		RightTextLen: rightTextLen,
		LeftLabel:    left,
		RightLabel:   right,
		LeftColor:    "#2f3542", // slate background
		RightColor:   rightColor,
	}

	var buf bytes.Buffer
	if err := badgeTmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenerateCleanBadge generates an SVG badge for a given clean percentage.
func GenerateCleanBadge(percentage int) (string, error) {
	rightLabel := fmt.Sprintf("%d%% clean", percentage)
	color := BadgeColor(percentage)
	return GenerateBadge("mutefeed", rightLabel, color)
}

// GenerateErrorBadge generates a fallback error SVG badge [mutefeed | error].
func GenerateErrorBadge() string {
	svg, err := GenerateBadge("mutefeed", "error", "#ff4757")
	if err != nil {
		return `<svg xmlns="http://www.w3.org/2000/svg" width="112" height="20"><rect width="69" height="20" fill="#2f3542"/><rect x="69" width="43" height="20" fill="#ff4757"/></svg>`
	}
	return svg
}
