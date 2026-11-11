import type { ReactNode } from 'react'
import { statusLabels } from '../projects/model'
import type { ProjectReport } from './model'

const severity = { low: 'Baja', medium: 'Media', high: 'Alta', critical: 'Crítica' }
const metrics = [
  ['planned_story_points', 'Story points planificados', 'SP'],
  ['completed_story_points', 'Story points completados', 'SP'],
  ['velocity', 'Velocidad', 'SP'],
  ['estimated_hours', 'Horas estimadas', 'h'],
  ['actual_hours', 'Horas reales', 'h'],
  ['deviation', 'Desviación', 'SP'],
  ['completed_story_percentage', 'Historias completadas', '%'],
  ['defects_detected', 'Defectos detectados', ''],
  ['defects_resolved', 'Defectos resueltos', ''],
] as const

function ReportTable({ caption, columns, empty, children }: {
  caption: string; columns: string[]; empty: boolean; children: ReactNode
}) {
  return <div className="report-table-scroll" tabIndex={0} role="region" aria-label={caption}>
    <table className="report-table">
      <caption>{caption}</caption>
      <thead><tr>{columns.map(column => <th key={column} scope="col">{column}</th>)}</tr></thead>
      <tbody>{empty ? <tr><td colSpan={columns.length}>No hay {caption.toLowerCase()}.</td></tr> : children}</tbody>
    </table>
  </div>
}

function StoryReference({ identifier, title }: { identifier: string; title: string }) {
  return <><strong className="report-code">{identifier}</strong><span className="report-story-title">{title}</span></>
}

export function ReportContent({ report }: { report: ProjectReport }) {
  return <article className="report-content">
    <header className="report-heading">
      <p className="eyebrow">Reporte de {report.sprint ? `Sprint ${report.sprint.id}` : 'proyecto'}</p>
      <h1>{report.project.name}</h1>
      <p className="report-context">{statusLabels[report.project.status]} · {report.sprint
        ? `${report.sprint.goal} · ${report.sprint.status === 'closed' ? 'Cerrado' : 'Abierto'}`
        : 'Todo el proyecto'}</p>
    </header>
    <section aria-labelledby="report-metrics-title">
      <h2 id="report-metrics-title" className="report-section-title"><span>01 /</span> Métricas del alcance</h2>
      <dl className="report-metrics">{metrics.map(([key, label, unit]) => <div key={key}>
        <dt>{label}</dt><dd>{String(report.metrics[key])}<small>{unit}</small></dd>
      </div>)}</dl>
    </section>
    <section aria-labelledby="report-stories-title">
      <h2 id="report-stories-title" className="report-section-title"><span>02 /</span> Historias</h2>
      <ReportTable caption="Historias del alcance" columns={['Historia', 'Estado', 'Sprint', 'Estimación']} empty={report.stories.length === 0}>
        {report.stories.map(story => <tr key={story.id}>
          <th scope="row"><StoryReference identifier={story.identifier} title={story.title} /></th>
          <td>{story.status === 'done' ? 'Completada' : 'Por hacer'}</td>
          <td>{story.sprint_id === null ? 'Sin Sprint' : `Sprint ${story.sprint_id}`}</td>
          <td>{story.story_points === null ? 'Sin estimar' : `${story.story_points} SP`}</td>
        </tr>)}
      </ReportTable>
    </section>
    <section aria-labelledby="report-effort-title">
      <h2 id="report-effort-title" className="report-section-title"><span>03 /</span> Esfuerzo registrado</h2>
      <ReportTable caption="Registros de esfuerzo" columns={['Historia', 'Integrante', 'Fecha', 'Actividad', 'Horas']} empty={report.effort.length === 0}>
        {report.effort.map(entry => <tr key={entry.id}>
          <th scope="row"><StoryReference identifier={entry.story_identifier} title={entry.story_title} /></th>
          <td>{entry.member_name}</td><td><time dateTime={entry.date}>{entry.date}</time></td>
          <td>{entry.activity}</td><td>{entry.hours}</td>
        </tr>)}
      </ReportTable>
    </section>
    <section aria-labelledby="report-defects-title">
      <h2 id="report-defects-title" className="report-section-title"><span>04 /</span> Defectos</h2>
      <ReportTable caption="Defectos del alcance" columns={['Historia', 'Descripción', 'Severidad', 'Estado', 'Detección', 'Resolución']} empty={report.defects.length === 0}>
        {report.defects.map(defect => <tr key={defect.id}>
          <th scope="row"><StoryReference identifier={defect.story_identifier} title={defect.story_title} /></th>
          <td>{defect.description}</td><td>{severity[defect.severity]}</td>
          <td>{defect.status === 'resolved' ? 'Resuelto' : 'Abierto'}</td>
          <td>Sprint {defect.detection_sprint_id}</td>
          <td>{defect.resolution_sprint_id === null ? 'Sin resolución' : `Sprint ${defect.resolution_sprint_id}`}</td>
        </tr>)}
      </ReportTable>
    </section>
    <p className="report-footnote">Datos actuales del alcance seleccionado. Una nueva consulta puede reflejar cambios posteriores.</p>
  </article>
}
