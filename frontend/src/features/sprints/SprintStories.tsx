import type { AssignedStory } from './model'
import { storyStatusLabels } from './model'

export function SprintStories({ stories, onComplete, completingId }: {
  stories: AssignedStory[]
  onComplete?: (story: AssignedStory) => void
  completingId?: number | null
}) {
  if (stories.length === 0) {
    return (
      <div className="sprints-empty">
        <h3>Todavía no hay historias asignadas</h3>
        <p>Cuando asignes la primera, vas a encontrarla acá.</p>
      </div>
    )
  }

  return (
    <div className="sprints-table-scroll" tabIndex={0} role="region" aria-label="Historias del sprint">
      <table className="sprints-table">
        <caption className="sprints-sr-only">Historias asignadas a este Sprint</caption>
        <thead>
          <tr>
            <th scope="col">Identificador</th>
            <th scope="col">Historia</th>
            <th scope="col">Estado</th>
            {onComplete ? <th scope="col">Acción</th> : null}
          </tr>
        </thead>
        <tbody>
          {stories.map(story => (
            <tr key={story.story_id}>
              <th scope="row" className="sprints-identifier">{story.identifier}</th>
              <td className="sprints-goal-cell">{story.title}</td>
              <td>
                <span className={`sprints-status sprints-status-${story.status}`}>
                  {storyStatusLabels[story.status] ?? story.status}
                </span>
              </td>
              {onComplete
                ? (
                  <td>
                    {story.status === 'todo'
                      ? (
                        <button className="sprints-quiet-button" type="button"
                          disabled={completingId === story.story_id}
                          onClick={() => onComplete(story)}>
                          {completingId === story.story_id ? 'Completando…' : 'Completar'}
                        </button>
                      )
                      : <span className="sprints-muted">Sin acciones</span>}
                  </td>
                )
                : null}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
