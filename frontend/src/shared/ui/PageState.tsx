import type { ReactNode } from 'react'

type PageStateProps = {
  className: string
  title?: string
  children: ReactNode
}

export function LoadingState({ className, children }: Pick<PageStateProps, 'className' | 'children'>) {
  return <p className={className} role="status">{children}</p>
}

export function EmptyState({ className, title, children }: PageStateProps) {
  return (
    <div className={className}>
      {title ? <h3>{title}</h3> : null}
      <p>{children}</p>
    </div>
  )
}

export function ErrorState({
  className,
  title,
  children,
  action,
}: PageStateProps & { action?: ReactNode }) {
  return (
    <div className={className} role="alert">
      {title ? <h2>{title}</h2> : null}
      <p>{children}</p>
      {action}
    </div>
  )
}
