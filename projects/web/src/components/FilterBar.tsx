import type { IssueFilters, Label } from "../types/issue";

interface Props {
  filters: IssueFilters;
  labels: Label[];
  onChange: (filters: IssueFilters) => void;
}

export function FilterBar({ filters, labels, onChange }: Props) {
  function update(key: keyof IssueFilters, value: string) {
    onChange({ ...filters, [key]: value || undefined });
  }

  return (
    <div className="filter-bar">
      <div className="search-wrap">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
          <circle cx="11" cy="11" r="7" />
          <path d="m21 21-4.3-4.3" />
        </svg>
        <input
          value={filters.search ?? ""}
          onChange={(event) => update("search", event.target.value)}
          placeholder="Search title or description"
        />
      </div>
      <div className="filter-grid">
        <select value={filters.status ?? ""} onChange={(event) => update("status", event.target.value)}>
          <option value="">All status</option>
          <option value="open">Open</option>
          <option value="in_progress">In progress</option>
          <option value="closed">Closed</option>
        </select>
        <select value={filters.priority ?? ""} onChange={(event) => update("priority", event.target.value)}>
          <option value="">All priority</option>
          <option value="high">High</option>
          <option value="medium">Medium</option>
          <option value="low">Low</option>
        </select>
        <select value={filters.issueType ?? ""} onChange={(event) => update("issueType", event.target.value)}>
          <option value="">All type</option>
          <option value="bug">Bug</option>
          <option value="feature">Feature</option>
          <option value="task">Task</option>
          <option value="question">Question</option>
        </select>
        <select value={filters.labelId ?? ""} onChange={(event) => update("labelId", event.target.value)}>
          <option value="">All labels</option>
          {labels.map((label) => (
            <option key={label.id} value={label.id}>
              {label.name}
            </option>
          ))}
        </select>
      </div>
    </div>
  );
}
