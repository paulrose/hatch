import { useState } from "react";
import { Header } from "@/components/header";
import { ProjectList } from "@/components/project-list";
import { RouteMap } from "@/components/route-map";
import { AddProjectDialog } from "@/components/add-project-dialog";
import { EditProjectDialog } from "@/components/edit-project-dialog";
import { useProjects } from "@/hooks/use-projects";
import { useHealth } from "@/hooks/use-health";
import { useStatus } from "@/hooks/use-status";

function App() {
  const { projects, add, update, remove, toggle, loading, error } =
    useProjects();
  const { lookup } = useHealth();
  const { status } = useStatus();

  const [addOpen, setAddOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<string | null>(null);

  function handleDelete(name: string) {
    if (confirm(`Delete project "${name}"?`)) {
      remove(name);
    }
  }

  const editProject = editTarget ? projects[editTarget] : null;

  return (
    <div className="min-h-screen bg-linen">
      <Header status={status} onAddProject={() => setAddOpen(true)} />
      <main className="mx-auto max-w-5xl p-4">
        {loading && (
          <p className="py-16 text-center text-text-muted">Loading…</p>
        )}
        {error && (
          <p className="py-16 text-center text-destructive">{error}</p>
        )}
        {!loading && !error && (
          <>
            <ProjectList
              projects={projects}
              healthLookup={lookup}
              onToggle={toggle}
              onEdit={setEditTarget}
              onDelete={handleDelete}
              onAdd={() => setAddOpen(true)}
            />
            <RouteMap projects={projects} healthLookup={lookup} />
          </>
        )}
      </main>
      <AddProjectDialog open={addOpen} onOpenChange={setAddOpen} onAdd={add} />
      {editTarget && editProject && (
        <EditProjectDialog
          open={!!editTarget}
          onOpenChange={(open) => {
            if (!open) setEditTarget(null);
          }}
          name={editTarget}
          project={editProject}
          onSave={update}
        />
      )}
    </div>
  );
}

export default App;
