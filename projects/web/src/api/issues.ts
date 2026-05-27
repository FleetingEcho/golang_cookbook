import { apiKey, apiRequest, downloadUrl } from "./client";
import type {
  Attachment,
  Comment,
  CreateIssueInput,
  Issue,
  IssueDetail,
  IssueFilters,
  Label,
  PaginatedResponse,
  UpdateIssueInput
} from "../types/issue";

function queryString(filters: IssueFilters): string {
  const params = new URLSearchParams();
  Object.entries(filters).forEach(([key, value]) => {
    if (value !== undefined && value !== null) params.set(key, String(value));
  });
  const query = params.toString();
  return query ? `?${query}` : "";
}

export async function listIssues(filters: IssueFilters): Promise<{ items: Issue[]; total: number }> {
  const res = await apiRequest<PaginatedResponse<Issue>>(`/issues${queryString(filters)}`);
  return { items: res.items, total: res.total };
}

export function getIssue(id: number): Promise<IssueDetail> {
  return apiRequest<IssueDetail>(`/issues/${id}`);
}

export function createIssue(input: CreateIssueInput): Promise<IssueDetail> {
  return apiRequest<IssueDetail>("/issues", {
    method: "POST",
    body: JSON.stringify(input)
  });
}

export function updateIssue(id: number, input: UpdateIssueInput): Promise<IssueDetail> {
  return apiRequest<IssueDetail>(`/issues/${id}`, {
    method: "PATCH",
    body: JSON.stringify(input)
  });
}

export function deleteIssue(id: number): Promise<{ deleted: boolean }> {
  return apiRequest<{ deleted: boolean }>(`/issues/${id}`, { method: "DELETE" });
}

export function listLabels(): Promise<Label[]> {
  return apiRequest<Label[]>("/labels");
}

export function createComment(issueId: number, author: string, body: string): Promise<Comment> {
  return apiRequest<Comment>(`/issues/${issueId}/comments`, {
    method: "POST",
    body: JSON.stringify({ author, body })
  });
}

export function uploadAttachment(issueId: number, file: File): Promise<Attachment> {
  const body = new FormData();
  body.append("file", file);
  return apiRequest<Attachment>(`/issues/${issueId}/attachments`, {
    method: "POST",
    body
  });
}

export function deleteAttachment(id: number): Promise<{ deleted: boolean }> {
  return apiRequest<{ deleted: boolean }>(`/attachments/${id}`, { method: "DELETE" });
}

export async function downloadAttachment(id: number, filename: string): Promise<void> {
  const response = await fetch(downloadUrl(`/attachments/${id}/download`), {
    headers: { "x-api-key": apiKey() }
  });
  if (!response.ok) throw new Error("Download failed");

  const blob = await response.blob();
  const url = window.URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  link.click();
  window.URL.revokeObjectURL(url);
}
