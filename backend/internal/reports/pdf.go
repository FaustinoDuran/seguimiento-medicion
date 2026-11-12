package reports

import (
	"bytes"
	"strconv"

	"github.com/go-pdf/fpdf"
)

func RenderPDF(report Report) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()
	translate := pdf.UnicodeTranslatorFromDescriptor("")
	line := func(size float64, style, text string) {
		pdf.SetFont("Helvetica", style, size)
		pdf.MultiCell(0, 6, translate(text), "", "L", false)
	}

	if report.Sprint == nil {
		line(11, "", "Reporte de proyecto")
		line(16, "B", report.Project.Name)
		line(11, "", projectStatusLabel(report.Project.Status)+" · Todo el proyecto")
	} else {
		line(11, "", "Reporte de Sprint "+strconv.FormatInt(report.Sprint.ID, 10))
		line(16, "B", report.Project.Name)
		line(11, "", projectStatusLabel(report.Project.Status)+" · "+report.Sprint.Goal+" · "+sprintStatusLabel(report.Sprint.Status))
	}
	pdf.Ln(3)

	line(13, "B", "Métricas del alcance")
	line(11, "", "Story points planificados "+strconv.FormatInt(report.Metrics.PlannedStoryPoints, 10)+" SP")
	line(11, "", "Story points completados "+strconv.FormatInt(report.Metrics.CompletedStoryPoints, 10)+" SP")
	line(11, "", "Velocidad "+report.Metrics.Velocity.String()+" SP")
	line(11, "", "Horas estimadas "+report.Metrics.EstimatedHours.String()+" h")
	line(11, "", "Horas reales "+report.Metrics.ActualHours.String()+" h")
	line(11, "", "Desviación "+strconv.FormatInt(report.Metrics.Deviation, 10)+" SP")
	line(11, "", "Historias completadas "+report.Metrics.CompletedStoryPercentage.String()+" %")
	line(11, "", "Defectos detectados "+strconv.FormatInt(report.Metrics.DefectsDetected, 10))
	line(11, "", "Defectos resueltos "+strconv.FormatInt(report.Metrics.DefectsResolved, 10))
	pdf.Ln(3)

	line(13, "B", "Historias")
	if len(report.Stories) == 0 {
		line(11, "", "No hay historias del alcance.")
	}
	for _, story := range report.Stories {
		line(11, "", story.Identifier+" "+story.Title+" "+storyStatusLabel(story.Status)+" "+storySprintLabel(story.SprintID)+" "+storyPointsLabel(story.StoryPoints))
	}
	pdf.Ln(3)

	line(13, "B", "Esfuerzo registrado")
	if len(report.Effort) == 0 {
		line(11, "", "No hay registros de esfuerzo.")
	}
	for _, entry := range report.Effort {
		line(11, "", entry.StoryIdentifier+" "+entry.StoryTitle+" "+entry.MemberName+" "+entry.Date+" "+entry.Activity+" "+entry.Hours.String())
	}
	pdf.Ln(3)

	line(13, "B", "Defectos")
	if len(report.Defects) == 0 {
		line(11, "", "No hay defectos del alcance.")
	}
	for _, defect := range report.Defects {
		line(11, "", defect.StoryIdentifier+" "+defect.StoryTitle+" "+defect.Description+" "+severityLabel(defect.Severity)+" "+defectStatusLabel(defect.Status)+" Sprint "+strconv.FormatInt(defect.DetectionSprintID, 10)+" "+resolutionLabel(defect.ResolutionSprintID))
	}

	if err := pdf.Error(); err != nil {
		return nil, err
	}
	var body bytes.Buffer
	if err := pdf.Output(&body); err != nil {
		return nil, err
	}
	return body.Bytes(), nil
}

func projectStatusLabel(status string) string {
	switch status {
	case "planned":
		return "Planificado"
	case "active":
		return "Activo"
	case "closed":
		return "Cerrado"
	default:
		return status
	}
}

func sprintStatusLabel(status string) string {
	if status == "closed" {
		return "Cerrado"
	}
	return "Abierto"
}

func storyStatusLabel(status string) string {
	if status == "done" {
		return "Completada"
	}
	return "Por hacer"
}

func storySprintLabel(sprintID *int64) string {
	if sprintID == nil {
		return "Sin Sprint"
	}
	return "Sprint " + strconv.FormatInt(*sprintID, 10)
}

func storyPointsLabel(points *int64) string {
	if points == nil {
		return "Sin estimar"
	}
	return strconv.FormatInt(*points, 10) + " SP"
}

func severityLabel(severity string) string {
	switch severity {
	case "low":
		return "Baja"
	case "medium":
		return "Media"
	case "high":
		return "Alta"
	case "critical":
		return "Crítica"
	default:
		return severity
	}
}

func defectStatusLabel(status string) string {
	if status == "resolved" {
		return "Resuelto"
	}
	return "Abierto"
}

func resolutionLabel(sprintID *int64) string {
	if sprintID == nil {
		return "Sin resolución"
	}
	return "Sprint " + strconv.FormatInt(*sprintID, 10)
}
