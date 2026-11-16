import { Component } from 'react'
import type { ErrorInfo, ReactNode } from 'react'

type AppErrorBoundaryProps = {
  children: ReactNode
}

type AppErrorBoundaryState = {
  failed: boolean
}

export class AppErrorBoundary extends Component<AppErrorBoundaryProps, AppErrorBoundaryState> {
  state: AppErrorBoundaryState = { failed: false }

  static getDerivedStateFromError(): AppErrorBoundaryState {
    return { failed: true }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo): void {
    console.error('Unexpected render error', error, errorInfo)
  }

  render(): ReactNode {
    if (this.state.failed) {
      return (
        <main className="fatal-error" id="main-content">
          <p className="eyebrow">Error inesperado</p>
          <h1>No pudimos mostrar esta pantalla</h1>
          <p className="lead">
            Recargá la aplicación. Si el problema continúa, comprobá la conexión con el backend.
          </p>
          <button className="button" type="button" onClick={() => window.location.reload()}>
            Recargar aplicación
          </button>
        </main>
      )
    }

    return this.props.children
  }
}
