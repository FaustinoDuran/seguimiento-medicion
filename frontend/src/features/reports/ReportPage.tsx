import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { getReport } from './api'
import type { ProjectReport } from './model'
import { ReportContent } from './ReportContent'
import './reports.css'

function ReportLoader({ projectId }: { projectId: string }) {
  const [report, setReport] = useState<ProjectReport | null>(null)
  useEffect(() => {
    const controller = new AbortController()
    getReport(projectId, controller.signal).then(value => {
      if (!controller.signal.aborted) setReport(value)
    }).catch(() => {})
    return () => controller.abort()
  }, [projectId])
  return <>
    <Link className="text-link" to={`/projects/${projectId}`}>← Volver al proyecto</Link>
    {report ? <ReportContent report={report} /> : <p role="status">Cargando reporte…</p>}
  </>
}
export function ReportPage() {
  const { projectId = '' } = useParams()
  return <section className="report-page"><ReportLoader key={projectId} projectId={projectId} /></section>
}
