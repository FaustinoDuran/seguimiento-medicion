import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { getReport } from './api'
import type { ProjectReport, ReportSprint } from './model'
import { ReportContent } from './ReportContent'
import './reports.css'

function ReportLoader({ projectId }: { projectId: string }) {
  const [selected, setSelected] = useState('')
  const [options, setOptions] = useState<ReportSprint[]>([])
  const [report, setReport] = useState<ProjectReport | null>(null)
  useEffect(() => {
    const controller = new AbortController()
    getReport(projectId, selected, controller.signal).then(value => {
      if (!controller.signal.aborted) { setReport(value); setOptions(value.available_sprints) }
    }).catch(() => {})
    return () => controller.abort()
  }, [projectId, selected])
  return <>
    <Link className="text-link" to={`/projects/${projectId}`}>← Volver al proyecto</Link>
    <div className="report-selector">
      <label htmlFor="report-scope">Alcance del reporte</label>
      <select id="report-scope" value={selected} onChange={event => {
        setReport(null)
        setSelected(event.target.value)
      }}>
        <option value="">Todo el proyecto</option>
        {options.map(sprint => <option key={sprint.id} value={sprint.id}>
          Sprint {sprint.id} · {sprint.goal} · {sprint.status === 'closed' ? 'Cerrado' : 'Abierto'}
        </option>)}
      </select>
    </div>
    {report ? <ReportContent report={report} /> : <p role="status">Cargando reporte…</p>}
  </>
}
export function ReportPage() {
  const { projectId = '' } = useParams()
  return <section className="report-page"><ReportLoader key={projectId} projectId={projectId} /></section>
}
