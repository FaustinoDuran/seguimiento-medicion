package reports_test

import (
	"bytes"
	"compress/zlib"
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/metrics"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/reports"
)

func TestProjectReportPDFContainsTheReport(t *testing.T) {
	report := reports.Report{
		Project:          reports.Project{ID: 7, Name: "Seguimiento", Status: "active"},
		AvailableSprints: []reports.Sprint{{ID: 9, Goal: "Selector interno", Status: "open"}},
		Metrics: reports.Metrics{
			ProjectID: 7, PlannedStoryPoints: 8, ActualHours: metrics.Amount(150),
		},
		Stories: []reports.Story{{
			ID: 1, Identifier: "HU-A", Title: "Exportar informe", Status: "todo",
		}},
		Effort: []reports.Effort{{
			ID: 1, StoryIdentifier: "HU-A", StoryTitle: "Exportar informe",
			MemberName: "Ana", Date: "2026-11-10", Activity: "Revisar", Hours: metrics.Amount(150),
		}},
		Defects: []reports.Defect{{
			ID: 1, StoryIdentifier: "HU-A", StoryTitle: "Exportar informe",
			Description: "Total mal calculado", Severity: "high", Status: "open",
			DetectionSprintID: 2,
		}},
	}

	file, err := reports.RenderPDF(report)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(file, []byte("%PDF")) {
		t.Fatalf("expected a PDF file, got %q", file)
	}
	text := pdfPlainText(t, file)
	for _, want := range []string{
		"Seguimiento", "Activo", "Todo el proyecto",
		"Story points planificados", "8", "Horas reales", "1.5",
		"HU-A", "Exportar informe", "Por hacer", "Sin Sprint", "Sin estimar",
		"Ana", "2026-11-10", "Revisar",
		"Total mal calculado", "Alta", "Abierto", "Sprint 2", "Sin resolución",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("PDF text missing %q\n%s", want, text)
		}
	}
	if strings.Contains(text, "Selector interno") {
		t.Errorf("PDF must not include the sprint selector\n%s", text)
	}
}

func TestEmptyReportPDFKeepsSectionMessages(t *testing.T) {
	file, err := reports.RenderPDF(reports.Report{
		Project: reports.Project{ID: 3, Name: "Vacio", Status: "planned"},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertPDFText(t, file, "Vacio", "Planificado", "Todo el proyecto",
		"No hay historias del alcance.", "No hay registros de esfuerzo.", "No hay defectos del alcance.")
}

func pdfPlainText(t *testing.T, file []byte) string {
	t.Helper()
	var text strings.Builder
	rest := file
	for {
		start := bytes.Index(rest, []byte("stream"))
		if start < 0 {
			break
		}
		rest = rest[start+len("stream"):]
		if len(rest) > 0 && rest[0] == '\r' {
			rest = rest[1:]
		}
		if len(rest) > 0 && rest[0] == '\n' {
			rest = rest[1:]
		}
		end := bytes.Index(rest, []byte("endstream"))
		if end < 0 {
			t.Fatal("pdf stream without endstream")
		}
		raw := bytes.TrimRight(rest[:end], "\r\n")
		rest = rest[end+len("endstream"):]
		decoded := raw
		reader, err := zlib.NewReader(bytes.NewReader(raw))
		if err == nil {
			inflated, readErr := io.ReadAll(reader)
			_ = reader.Close()
			if readErr == nil {
				decoded = inflated
			}
		}
		text.WriteString(pdfContentText(decoded))
		text.WriteByte('\n')
	}
	if text.Len() == 0 {
		t.Fatalf("pdf has no readable text: %q", file)
	}
	return text.String()
}

func pdfContentText(content []byte) string {
	var text strings.Builder
	for i := 0; i < len(content); i++ {
		switch content[i] {
		case '(':
			literal, next := pdfLiteral(content, i)
			text.WriteString(winAnsi(literal))
			i = next - 1
		case '<':
			if i+1 < len(content) && content[i+1] == '<' {
				i++
				continue
			}
			hex, next := pdfHex(content, i)
			text.WriteString(hex)
			i = next - 1
		}
	}
	return text.String()
}

func pdfLiteral(content []byte, start int) ([]byte, int) {
	var raw []byte
	depth := 1
	for i := start + 1; i < len(content); i++ {
		switch content[i] {
		case '\\':
			if i+1 >= len(content) {
				return raw, len(content)
			}
			next := content[i+1]
			switch next {
			case 'n':
				raw = append(raw, '\n')
			case 'r':
				raw = append(raw, '\r')
			case 't':
				raw = append(raw, '\t')
			case 'b':
				raw = append(raw, '\b')
			case 'f':
				raw = append(raw, '\f')
			case '(', ')', '\\':
				raw = append(raw, next)
			default:
				if next >= '0' && next <= '7' {
					octal := []byte{next}
					j := i + 2
					for len(octal) < 3 && j < len(content) && content[j] >= '0' && content[j] <= '7' {
						octal = append(octal, content[j])
						j++
					}
					value, _ := strconv.ParseInt(string(octal), 8, 16)
					raw = append(raw, byte(value))
					i = j - 1
					continue
				}
				raw = append(raw, next)
			}
			i++
		case '(':
			depth++
			raw = append(raw, '(')
		case ')':
			depth--
			if depth == 0 {
				return raw, i + 1
			}
			raw = append(raw, ')')
		default:
			raw = append(raw, content[i])
		}
	}
	return raw, len(content)
}

func pdfHex(content []byte, start int) (string, int) {
	end := bytes.IndexByte(content[start+1:], '>')
	if end < 0 {
		return "", len(content)
	}
	digits := content[start+1 : start+1+end]
	var packed []byte
	for _, digit := range digits {
		if digit != ' ' && digit != '\n' && digit != '\r' && digit != '\t' {
			packed = append(packed, digit)
		}
	}
	if len(packed)%2 == 1 {
		packed = append(packed, '0')
	}
	decoded := make([]byte, len(packed)/2)
	for i := 0; i < len(decoded); i++ {
		value, err := strconv.ParseUint(string(packed[i*2:i*2+2]), 16, 8)
		if err != nil {
			return "", start + 2 + end
		}
		decoded[i] = byte(value)
	}
	if bytes.HasPrefix(decoded, []byte{0xFE, 0xFF}) {
		var text strings.Builder
		for i := 2; i+1 < len(decoded); i += 2 {
			text.WriteRune(rune(decoded[i])<<8 | rune(decoded[i+1]))
		}
		return text.String(), start + 2 + end
	}
	return winAnsi(decoded), start + 2 + end
}

func winAnsi(raw []byte) string {
	var text strings.Builder
	for _, value := range raw {
		text.WriteRune(rune(value))
	}
	return text.String()
}
