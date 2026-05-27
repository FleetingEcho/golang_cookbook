import { useCallback, useEffect, useMemo, useState } from "react";
import {
  createComment,
  createIssue,
  deleteAttachment,
  deleteIssue,
  downloadAttachment,
  getIssue,
  listIssues,
  listLabels,
  updateIssue,
  uploadAttachment
} from "./api/issues";
import { AttachmentPanel } from "./components/AttachmentPanel";
import { CommentList } from "./components/CommentList";
import { FilterBar } from "./components/FilterBar";
import { IssueDetailPanel } from "./components/IssueDetail";
import { IssueForm } from "./components/IssueForm";
import { IssueList } from "./components/IssueList";
import { ThemeToggle } from "./components/ThemeToggle";
import type { CreateIssueInput, Issue, IssueDetail, IssueFilters, Label } from "./types/issue";

export default function App() {
  const [issues, setIssues] = useState<Issue[]>([]);
  const [labels, setLabels] = useState<Label[]>([]);
  // 从 URL 中读取初始 selectedId，支持直接访问 /issues/5
  const [selectedId, setSelectedId] = useState<number | null>(() => {
    const m = window.location.pathname.match(/^\/issues\/(\d+)/);
    return m ? parseInt(m[1], 10) : null;
  });
  const [selected, setSelected] = useState<IssueDetail | null>(null);
  const [filters, setFilters] = useState<IssueFilters>({});
  const [totalCount, setTotalCount] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [backend, setBackend] = useState<"checking" | "ok" | "error">("checking");

  const checkHealth = useCallback(async () => {
    try {
      const res = await fetch("http://127.0.0.1:3001/health");
      setBackend(res.ok ? "ok" : "error");
    } catch {
      setBackend("error");
    }
  }, []);

  const selectedLabelIds = useMemo(() => selected?.labels.map((label) => label.id) ?? [], [selected]);

  async function refreshList(nextFilters = filters) {
    setLoading(true);
    setError(null);
    try {
      const [issueData, labelData] = await Promise.all([listIssues(nextFilters), listLabels()]);
      setIssues(issueData.items);
      setTotalCount(issueData.total);
      setLabels(labelData);
      if (!selectedId && issueData.items[0]) setSelectedId(issueData.items[0].id);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load issues");
    } finally {
      setLoading(false);
    }
  }

  async function handleLoadMore() {
    const next = { ...filters, offset: issues.length };
    setLoading(true);
    try {
      const more = await listIssues(next);
      setIssues((prev) => [...prev, ...more.items]);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load more");
    } finally {
      setLoading(false);
    }
  }

  async function refreshDetail(id: number | null = selectedId) {
    if (!id) {
      setSelected(null);
      return;
    }
    setError(null);
    try {
      setSelected(await getIssue(id));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load issue");
    }
  }

  // URL 与 selectedId 同步
  useEffect(() => {
    if (selectedId) {
      history.replaceState(null, "", `/issues/${selectedId}`);
    } else {
      history.replaceState(null, "", "/");
    }
  }, [selectedId]);

  // 监听浏览器前进/后退
  useEffect(() => {
    const onPop = () => {
      const m = window.location.pathname.match(/^\/issues\/(\d+)/);
      setSelectedId(m ? parseInt(m[1], 10) : null);
    };
    window.addEventListener("popstate", onPop);
    return () => window.removeEventListener("popstate", onPop);
  }, []);

  useEffect(() => {
    void refreshList();
    checkHealth();
  }, []);

  useEffect(() => {
    void refreshDetail(selectedId);
  }, [selectedId]);

  useEffect(() => {
    const interval = setInterval(checkHealth, 15_000);
    return () => clearInterval(interval);
  }, [checkHealth]);

  async function handleFilterChange(next: IssueFilters) {
    setFilters(next);
    await refreshList(next);
  }

  async function handleCreate(input: CreateIssueInput) {
    const created = await createIssue(input);
    await refreshList();
    setSelectedId(created.id);
  }

  async function handleStatus(status: IssueDetail["status"]) {
    if (!selected) return;
    const updated = await updateIssue(selected.id, { status });
    setSelected(updated);
    await refreshList();
  }

  async function handleSaveIssue(input: Partial<IssueDetail>) {
    if (!selected) return;
    const updated = await updateIssue(selected.id, {
      title: input.title,
      description: input.description,
      priority: input.priority,
      issueType: input.issueType,
      assignee: input.assignee,
      labelIds: selectedLabelIds
    });
    setSelected(updated);
    await refreshList();
  }

  async function handleDeleteIssue() {
    if (!selected) return;
    await deleteIssue(selected.id);
    setSelectedId(null);
    setSelected(null);
    await refreshList();
  }

  async function handleComment(author: string, body: string) {
    if (!selected) return;
    await createComment(selected.id, author, body);
    await refreshDetail(selected.id);
  }

  async function handleUpload(file: File) {
    if (!selected) return;
    await uploadAttachment(selected.id, file);
    await refreshDetail(selected.id);
  }

  async function handleDeleteAttachment(id: number) {
    await deleteAttachment(id);
    await refreshDetail();
  }

  return (
    <main className="app-shell">
      <section className="topbar">
        <div className="topbar-left">
          <div className="brand-mark" aria-hidden="true">IT</div>
          <div>
            <p className="eyebrow">Axum · SQLite · React</p>
            <h1>Issue Tracker</h1>
          </div>
        </div>
        <div className="topbar-right">
          <div className={`status-pill ${backend}`}>
            {backend === "checking" ? "⋯" : backend === "ok" ? "API ok" : "API error"}
          </div>
          <ThemeToggle />
        </div>
      </section>

      {error && <div className="error-banner">{error}</div>}

      <section className="workspace">
        <aside className="sidebar">
          <FilterBar filters={filters} labels={labels} onChange={handleFilterChange} />
          <div className="list-header">
            <span>Issues</span>
            <span className="list-count">{totalCount}</span>
          </div>
          <IssueList issues={issues} selectedId={selectedId} onSelect={setSelectedId} />
          {issues.length < totalCount && (
            <button className="load-more" onClick={handleLoadMore} disabled={loading}>
              {loading ? "Loading…" : `Load more (${issues.length}/${totalCount})`}
            </button>
          )}
        </aside>

        <section className="detail-pane">
          {selected ? (
            <>
              <IssueDetailPanel
                issue={selected}
                labels={labels}
                selectedLabelIds={selectedLabelIds}
                onStatusChange={handleStatus}
                onSave={handleSaveIssue}
                onDelete={handleDeleteIssue}
              />
              <CommentList comments={selected.comments} onSubmit={handleComment} />
              <AttachmentPanel
                attachments={selected.attachments}
                onUpload={handleUpload}
                onDownload={downloadAttachment}
                onDelete={handleDeleteAttachment}
              />
            </>
          ) : (
            <div className="empty-state">Select an issue from the list, or create a new one.</div>
          )}
        </section>

        <aside className="create-pane">
          <div className="create-form-header">
            <h2>New issue</h2>
          </div>
          <IssueForm labels={labels} onSubmit={handleCreate} />
        </aside>
      </section>
    </main>
  );
}

