// Dashboard — redirect to tasks by default.

export function IndexPage() {
  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-semibold">YOLO Factory</h1>
      <p className="text-muted-foreground">
        Task orchestration and management dashboard.
      </p>
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
        {[
          { label: 'Tasks', href: '/factory/tasks', icon: 'check-square' },
          { label: 'Runs', href: '/factory/runs', icon: 'play' },
          { label: 'Repos', href: '/factory/repos', icon: 'git-branch' },
          { label: 'Questions', href: '/factory/questions', icon: 'help-circle' },
        ].map((item) => (
          <a
            key={item.label}
            href={item.href}
            className="rounded-lg border p-4 text-center hover:bg-muted/50 transition-colors"
          >
            <div className="text-lg font-semibold">{item.label}</div>
          </a>
        ))}
      </div>
    </div>
  )
}
