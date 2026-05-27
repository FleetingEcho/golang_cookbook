import { FormEvent, useState } from "react";
import type { CreateIssueInput, IssuePriority, IssueType, Label } from "../types/issue";
import { AvatarCircle } from "./AvatarCircle";

interface Props {
  labels: Label[];
  onSubmit: (input: CreateIssueInput) => Promise<void>;
}

export function IssueForm({ labels, onSubmit }: Props) {
  const [submitting, setSubmitting] = useState(false);
  const [labelIds, setLabelIds] = useState<number[]>([]);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    setSubmitting(true);
    try {
      await onSubmit({
        title: String(form.get("title") ?? ""),
        description: String(form.get("description") ?? ""),
        priority: String(form.get("priority")) as IssuePriority,
        issueType: String(form.get("issue_type")) as IssueType,
        assignee: String(form.get("assignee") || "") || null,
        createdBy: String(form.get("created_by") || "dev"),
        labelIds: labelIds
      });
      event.currentTarget.reset();
      setLabelIds([]);
    } finally {
      setSubmitting(false);
    }
  }

  function toggleLabel(id: number) {
    setLabelIds((current) => (current.includes(id) ? current.filter((item) => item !== id) : [...current, id]));
  }

  return (
    <form className="create-form" onSubmit={submit}>
      <label>
        Title
        <input name="title" required placeholder="Short, searchable summary" />
      </label>
      <label>
        Description
        <textarea name="description" required rows={5} placeholder="Steps, context, expected behavior" />
      </label>
      <div className="two-col">
        <label>
          Priority
          <select name="priority" defaultValue="medium">
            <option value="high">High</option>
            <option value="medium">Medium</option>
            <option value="low">Low</option>
          </select>
        </label>
        <label>
          Type
          <select name="issue_type" defaultValue="bug">
            <option value="bug">Bug</option>
            <option value="feature">Feature</option>
            <option value="task">Task</option>
            <option value="question">Question</option>
          </select>
        </label>
      </div>
      <div className="two-col">
        <label>
          Assignee
          <input name="assignee" placeholder="Optional" />
        </label>
        <label>
          Created by
          <span className="avatar-input-wrap">
            <AvatarCircle name="Teng" size={18} />
            <input name="created_by" defaultValue="Teng" />
          </span>
        </label>
      </div>
      <div className="label-picker">
        {labels.map((label) => (
          <button
            key={label.id}
            type="button"
            className={labelIds.includes(label.id) ? "picked" : ""}
            onClick={() => toggleLabel(label.id)}
            style={{ ["--swatch" as string]: label.color }}
          >
            {label.name}
          </button>
        ))}
      </div>
      <button className="primary" disabled={submitting}>
        {submitting ? "Creating..." : "Create issue"}
      </button>
    </form>
  );
}

