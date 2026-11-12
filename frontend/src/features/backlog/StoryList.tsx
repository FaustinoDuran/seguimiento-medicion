import type { Story } from './model'
import { priorityLabels, statusLabels } from './model'

export function StoryList({ stories, onEdit, editingDisabled = false }: {
  stories: Story[]; onEdit?: (story: Story) => void; editingDisabled?: boolean
}) {
  if (stories.length === 0) {
    return <div className="backlog-empty"><h3>Todavía no hay historias</h3>
      <p>Cuando agregues la primera, vas a encontrarla acá.</p></div>
  }
  return (
    <div className="backlog-table-scroll" tabIndex={0} role="region" aria-label="Historias del proyecto">
      <table className="backlog-table">
        <caption className="backlog-sr-only">Todas las historias del proyecto, altas más recientes primero</caption>
        <thead><tr><th scope="col">Identificador</th><th scope="col">Historia</th>
          <th scope="col">Prioridad</th><th scope="col">Estado</th><th scope="col">Estimación</th>
          {onEdit ? <th scope="col">Acciones</th> : null}</tr></thead>
        <tbody>{stories.map(story => (
          <tr key={story.id}>
            <th scope="row" className="backlog-identifier">{story.identifier}</th>
            <td className="backlog-story-title">{story.title}</td>
            <td>{priorityLabels[story.priority] ?? story.priority}</td>
            <td><span className={`backlog-status backlog-status-${story.status}`}>
              {statusLabels[story.status] ?? story.status}</span></td>
            <td className="backlog-points">{story.story_points === null ? 'Sin estimar' : `${story.story_points} SP`}</td>
            {onEdit ? <td><button type="button" className="backlog-quiet-button"
              aria-label={`Editar ${story.identifier}`} disabled={editingDisabled}
              onClick={() => onEdit(story)}>Editar</button></td> : null}
          </tr>
        ))}</tbody>
      </table>
    </div>
  )
}
