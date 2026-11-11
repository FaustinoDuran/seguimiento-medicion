import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { ApiError } from '../../shared/api/client'
import { getReport } from './api'
import type { ProjectReport, ReportSprint } from './model'
import { ReportContent } from './ReportContent'
import './reports.css'

type PageState =
  | { kind: 'loading' }
  | { kind: 'ready'; report: ProjectReport }
  | { kind: 'missing' }
  | { kind: 'error' }

function MissingReport() {
  return <div className="report-alert">
    <h1>Proyecto o Sprint no encontrado</h1>
    <p>No encontramos el alcance solicitado en este proyecto.</p>
    <Link className="text-link" to="/projects">Volver al listado</Link>
  </div>
}

function ReportLoader({ projectId }: { projectId: string }) {
  const [selected, setSelected] = useState('')
  const [options, setOptions] = useState<ReportSprint[]>([])
  const [state, setState] = useState<PageState>({ kind: 'loading' })
  const [attempt, setAttempt] = useState(0)
  useEffect(() => {
    const controller = new AbortController()
    getReport(projectId, selected, controller.signal).then(report => {
      if (controller.signal.aborted) return
      setState({ kind: 'ready', report })
      setOptions(report.available_sprints)
    }).catch(error => {
      if (controller.signal.aborted) return
      setState({ kind: error instanceof ApiError && error.status === 404 ? 'missing' : 'error' })
    })
    return () => controller.abort()
  }, [projectId, selected, attempt])
  return <>
    <Link className="text-link" to={`/projects/${projectId}`}>← Volver al proyecto</Link>
    <div className="report-selector">
      <label htmlFor="report-scope">Alcance del reporte</label>
      <select id="report-scope" value={selected} onChange={event => {
        setState({ kind: 'loading' })
        setSelected(event.target.value)
      }}>
        <option value="">Todo el proyecto</option>
        {options.map(sprint => <option key={sprint.id} value={sprint.id}>
          Sprint {sprint.id} · {sprint.goal} · {sprint.status === 'closed' ? 'Cerrado' : 'Abierto'}
        </option>)}
      </select>
    </div>
    {state.kind === 'ready' ? <ReportContent report={state.report} /> : null}
    {state.kind === 'loading' ? <p className="report-loading" role="status">Cargando reporte…</p> : null}
    {state.kind === 'missing' ? <MissingReport /> : null}
    {state.kind === 'error' ? <div className="report-alert" role="alert">
      <h1>No pudimos cargar el reporte</h1>
      <p>Comprobá la conexión y volvé a intentarlo.</p>
      <button type="button" className="button" onClick={() => {
        setState({ kind: 'loading' })
        setAttempt(previous => previous + 1)
      }}>Reintentar carga</button>
    </div> : null}
  </>
}
export function ReportPage() {
  const { projectId = '' } = useParams()
  return <section className="report-page">{/^[1-9]\d*$/.test(projectId)
    ? <ReportLoader key={projectId} projectId={projectId} />
    : <MissingReport />}</section>
}
