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

// ── 分页配置 ───────────────────────────────────────────────────────────────
const PAGE_SIZES = [10, 25, 50] as const;

export default function App() {
  const [issues, setIssues] = useState<Issue[]>([]);
  const [labels, setLabels] = useState<Label[]>([]);
  // 从 URL 中读取初始 selectedId
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

  // ── 分页状态 ──────────────────────────────────────────────────────────
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState<number>(25);

  const totalPages = Math.max(1, Math.ceil(totalCount / pageSize));
  const showingFrom = (page - 1) * pageSize + (issues.length > 0 ? 1 : 0);
  const showingTo = Math.min((page - 1) * pageSize + issues.length, totalCount);

  const checkHealth = useCallback(async () => {
    try {
      const res = await fetch("http://127.0.0.1:3001/health");
      setBackend(res.ok ? "ok" : "error");
    } catch {
      setBackend("error");
    }
  }, []);

  const selectedLabelIds = useMemo(() => selected?.labels.map((label) => label.id) ?? [], [selected]);

  async function refreshList(nextFilters = filters, nextPage = page, nextPageSize = pageSize) {
    setLoading(true);
    setError(null);
    try {
      const query: IssueFilters = {
        ...nextFilters,
        limit: nextPageSize,
        offset: (nextPage - 1) * nextPageSize,
      };
      const [issueData, labelData] = await Promise.all([listIssues(query), listLabels()]);
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
    history.replaceState(null, "", selectedId ? `/issues/${selectedId}` : "/");
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

  // ── 分页导航 ──────────────────────────────────────────────────────────
  function goToPage(p: number) {
    if (p < 1 || p > totalPages || p === page || loading) return;
    setPage(p);
    refreshList(filters, p, pageSize);
  }

  // 生成页码列表，例如 [1, '...', 4, 5, 6, '...', 20]
  function getPageNumbers(): (number | "ellipsis")[] {
    const pages: (number | "ellipsis")[] = [];
    const total = totalPages;
    const current = page;
    const delta = 2; // 当前页两侧显示几个页码

    // 固定显示第 1 页
    pages.push(1);

    const rangeStart = Math.max(2, current - delta);
    const rangeEnd = Math.min(total - 1, current + delta);

    if (rangeStart > 2) pages.push("ellipsis");

    for (let i = rangeStart; i <= rangeEnd; i++) {
      pages.push(i);
    }

    if (rangeEnd < total - 1) pages.push("ellipsis");

    if (total > 1) pages.push(total);

    return pages;
  }

  function handlePageSizeChange(newSize: number) {
    setPageSize(newSize);
    setPage(1);
    refreshList(filters, 1, newSize);
  }

  async function handleFilterChange(next: IssueFilters) {
    setFilters(next);
    setPage(1);
    await refreshList(next, 1, pageSize);
  }

  async function handleCreate(input: CreateIssueInput) {
    const created = await createIssue(input);
    setPage(1);
    await refreshList(filters, 1, pageSize);
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
          {/* ── List header: title + count + page size ─────────────── */}
          <div className="list-header">
            <span>Issues</span>
            <span className="list-count">{totalCount}</span>
            <div className="list-header-spacer" />
            <select
              value={pageSize}
              onChange={(e) => handlePageSizeChange(Number(e.target.value))}
              className="pagination-select"
            >
              {PAGE_SIZES.map((s) => (
                <option key={s} value={s}>
                  {s} / page
                </option>
              ))}
            </select>
          </div>
          <IssueList issues={issues} selectedId={selectedId} onSelect={setSelectedId} />

          {/* ── 页码 ────────────────────────────────────────────────── */}
          <div className="pagination">
            {totalCount > 0 && (
              <div className="pagination-range">
                {showingFrom}–{showingTo} of {totalCount}
              </div>
            )}
            <div className="pagination-controls">
              <button
                className="pagination-btn"
                onClick={() => goToPage(page - 1)}
                disabled={page <= 1}
              >
                ‹
              </button>

              {getPageNumbers().map((p, i) =>
                p === "ellipsis" ? (
                  <span key={`e-${i}`} className="pagination-ellipsis">…</span>
                ) : (
                  <button
                    key={p}
                    className={`pagination-btn ${p === page ? "active" : ""}`}
                    onClick={() => goToPage(p)}
                  >
                    {p}
                  </button>
                )
              )}

              <button
                className="pagination-btn"
                onClick={() => goToPage(page + 1)}
                disabled={page >= totalPages}
              >
                ›
              </button>
            </div>
          </div>
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
