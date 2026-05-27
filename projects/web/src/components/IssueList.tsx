import type { Issue } from "../types/issue";

interface Props {
  issues: Issue[];
  selectedId: number | null;
  onSelect: (id: number) => void;
}

const STATUS_LABEL: Record<Issue["status"], string> = {
  open: "Open",
  in_progress: "In progress",
  closed: "Closed"
};

export function IssueList({ issues, selectedId, onSelect }: Props) {
  return (
    <div className="issue-list">
      {issues.map((issue) => (
        <button
          key={issue.id}
          className={`issue-row ${selectedId === issue.id ? "selected" : ""}`}
          onClick={() => onSelect(issue.id)}
        >
          <span className={`status-icon ${issue.status}`} aria-label={STATUS_LABEL[issue.status]} />
          <span className="issue-title">{issue.title}</span>
          <span className={`priority ${issue.priority}`}>{issue.priority}</span>
          <span className="issue-meta">
            <span>#{issue.id}</span>
            <span className="sep">·</span>
            <span className={`type-chip ${issue.issueType}`}>{issue.issueType}</span>
            {issue.assignee && (
              <>
                <span className="sep">·</span>
                <span>{issue.assignee}</span>
              </>
            )}
          </span>
        </button>
      ))}
      {issues.length === 0 && <div className="empty-list">No issues match the current filters.</div>}
    </div>
  );
}
