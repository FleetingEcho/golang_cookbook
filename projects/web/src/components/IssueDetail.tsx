import { FormEvent, useState } from "react";
import type { IssueDetail, IssuePriority, IssueStatus, IssueType, Label } from "../types/issue";
import { AvatarCircle } from "./AvatarCircle";

interface Props {
  issue: IssueDetail;
  labels: Label[];
  selectedLabelIds: number[];
  onStatusChange: (status: IssueStatus) => Promise<void>;
  onSave: (input: Partial<IssueDetail>) => Promise<void>;
  onDelete: () => Promise<void>;
}

const STATUS_LABEL: Record<IssueStatus, string> = {
  open: "Open",
  in_progress: "In progress",
  closed: "Closed"
};

export function IssueDetailPanel({ issue, onStatusChange, onSave, onDelete }: Props) {
  const [editing, setEditing] = useState(false);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    await onSave({
      title: String(form.get("title")),
      description: String(form.get("description")),
      priority: String(form.get("priority")) as IssuePriority,
      issueType: String(form.get("issue_type")) as IssueType,
      assignee: String(form.get("assignee") || "") || null
    });
    setEditing(false);
  }

  return (
    <article className="issue-detail">
      <header className="detail-header">
        <div>
          <p className="eyebrow">Issue #{issue.id}</p>
          <h2>{issue.title}</h2>
        </div>
        <div className="status-tabs" role="tablist" aria-label="Status">
          {(["open", "in_progress", "closed"] as IssueStatus[]).map((status) => (
            <button
              key={status}
              role="tab"
              aria-selected={issue.status === status}
              className={issue.status === status ? "active" : ""}
              onClick={() => onStatusChange(status)}
            >
              {STATUS_LABEL[status]}
            </button>
          ))}
        </div>
      </header>

      <div className="badge-row">
        <span className={`priority ${issue.priority}`}>{issue.priority}</span>
        <span className={`type-chip ${issue.issueType}`}>{issue.issueType}</span>
        <span className="assignee">
          {issue.assignee ? `@ ${issue.assignee}` : "Unassigned"}
        </span>
        {issue.labels.map((label) => (
          <span
            key={label.id}
            className="label-chip"
            style={{ ["--swatch" as string]: label.color }}
          >
            {label.name}
          </span>
        ))}
      </div>

      {editing ? (
        <form className="edit-form" onSubmit={submit}>
          <input name="title" defaultValue={issue.title} />
          <textarea name="description" defaultValue={issue.description} rows={6} />
          <div className="two-col">
            <select name="priority" defaultValue={issue.priority}>
              <option value="high">High</option>
              <option value="medium">Medium</option>
              <option value="low">Low</option>
            </select>
            <select name="issue_type" defaultValue={issue.issueType}>
              <option value="bug">Bug</option>
              <option value="feature">Feature</option>
              <option value="task">Task</option>
              <option value="question">Question</option>
            </select>
          </div>
          <input name="assignee" defaultValue={issue.assignee ?? ""} placeholder="Assignee" />
          <div className="button-row">
            <button className="primary">Save</button>
            <button type="button" onClick={() => setEditing(false)}>Cancel</button>
          </div>
        </form>
      ) : (
        <p className="description">{issue.description}</p>
      )}

      <footer className="detail-footer">
        <div className="button-row">
          {!editing && <button onClick={() => setEditing(true)}>Edit</button>}
          <button className="danger" onClick={onDelete}>Delete</button>
        </div>
        <span className="detail-footer-meta">
          <AvatarCircle name={issue.createdBy} size={18} /> Created {new Date(issue.createdAt).toLocaleString()} · by {issue.createdBy}
        </span>
      </footer>
    </article>
  );
}
