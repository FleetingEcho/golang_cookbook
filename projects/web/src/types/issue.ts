export type IssueStatus = "open" | "in_progress" | "closed";
export type IssuePriority = "low" | "medium" | "high";
export type IssueType = "bug" | "feature" | "task" | "question";

export interface Issue {
  id: number;
  title: string;
  description: string;
  status: IssueStatus;
  priority: IssuePriority;
  issueType: IssueType;
  assignee: string | null;
  createdBy: string;
  createdAt: string;
  updatedAt: string;
}

export interface Label {
  id: number;
  name: string;
  color: string;
}

export interface Comment {
  id: number;
  issueId: number;
  author: string;
  body: string;
  createdAt: string;
}

export interface Attachment {
  id: number;
  issueId: number;
  originalFilename: string;
  storedFilename: string;
  contentType: string;
  sizeBytes: number;
  createdAt: string;
}

export interface IssueDetail extends Issue {
  labels: Label[];
  comments: Comment[];
  attachments: Attachment[];
}

export interface IssueFilters {
  status?: string;
  priority?: string;
  issueType?: string;
  labelId?: string;
  search?: string;
  limit?: number;
  offset?: number;
}

export interface CreateIssueInput {
  title: string;
  description: string;
  priority: IssuePriority;
  issueType: IssueType;
  assignee?: string | null;
  createdBy: string;
  labelIds: number[];
}

export interface UpdateIssueInput {
  title?: string;
  description?: string;
  status?: IssueStatus;
  priority?: IssuePriority;
  issueType?: IssueType;
  assignee?: string | null;
  labelIds?: number[];
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  limit: number;
  offset: number;
}
