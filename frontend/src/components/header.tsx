import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import type { DaemonStatus } from "@/types";
import { Plus } from "lucide-react";

interface HeaderProps {
  status: DaemonStatus | null;
  onAddProject: () => void;
}

export function Header({ status, onAddProject }: HeaderProps) {
  return (
    <header className="border-b border-border bg-card/50 backdrop-blur-sm">
      <div className="mx-auto flex max-w-5xl items-center justify-between px-4 py-3">
        <div className="flex items-center gap-3">
          <h1 className="font-heading text-2xl text-muted-teal uppercase tracking-tight">
            Hatch
          </h1>
          {status && (
            <>
              <Badge variant="secondary">{status.version}</Badge>
              <span className="text-xs text-text-muted">
                up {status.uptime}
              </span>
            </>
          )}
        </div>
        <Button size="sm" onClick={onAddProject}>
          <Plus />
          Add Project
        </Button>
      </div>
    </header>
  );
}
