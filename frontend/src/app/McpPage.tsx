const MCP_URL = 'https://seguimiento-medicion.aedestec.com/mcp'

export function McpPage() {
  return (
    <section className="page-intro">
      <h1>MCP</h1>
      <p className="lead">
        Conectá un cliente de IA para consultar proyectos, historias, sprints, métricas e informes.
        El servidor es de solo lectura y no pide autenticación.
      </p>
      <div className="mcp-guide">
        <h2>Dirección del servidor</h2>
        <a className="mcp-url" href={MCP_URL}>
          {MCP_URL}
        </a>
        <h2>Conectarlo en ChatGPT</h2>
        <ol>
          <li>
            En ChatGPT web, abrí Plugins, después + y Add custom MCP server. Si la cuenta muestra la
            interfaz anterior, usá Configuración, Apps y Create.
          </li>
          <li>Nombre: Seguimiento y medición.</li>
          <li>URL: {MCP_URL}</li>
          <li>Autenticación: ninguna.</li>
          <li>Revisá las herramientas descubiertas, creá el plugin e instalalo.</li>
          <li>En un chat nuevo, seleccioná el plugin con @ y pedile los proyectos.</li>
        </ol>
        <p>
          La disponibilidad depende del plan y de los permisos del espacio de trabajo.{' '}
          <a
            className="text-link"
            href="https://help.openai.com/en/articles/12584461-connecting-custom-mcp-servers-in-chatgpt"
          >
            Guía oficial de ChatGPT
          </a>
        </p>
      </div>
    </section>
  )
}
