import { ChangeEvent } from "react";
import type { Attachment } from "../types/issue";

interface Props {
  attachments: Attachment[];
  onUpload: (file: File) => Promise<void>;
  onDownload: (id: number, filename: string) => Promise<void>;
  onDelete: (id: number) => Promise<void>;
}

export function AttachmentPanel({ attachments, onUpload, onDownload, onDelete }: Props) {
  async function upload(event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    if (!file) return;
    await onUpload(file);
    event.target.value = "";
  }

  return (
    <section className="panel">
      <div className="panel-heading">
        <h3>Attachments</h3>
        <label className="upload-button">
          Upload
          <input type="file" onChange={upload} />
        </label>
      </div>
      <div className="attachments">
        {attachments.map((attachment) => (
          <div key={attachment.id} className="attachment-row">
            <div>
              <strong>{attachment.originalFilename}</strong>
              <span>{Math.ceil(attachment.sizeBytes / 1024)} KB · {attachment.contentType}</span>
            </div>
            <div className="button-row">
              <button onClick={() => onDownload(attachment.id, attachment.originalFilename)}>Download</button>
              <button className="danger" onClick={() => onDelete(attachment.id)}>Delete</button>
            </div>
          </div>
        ))}
        {attachments.length === 0 && <div className="empty-list">No attachments yet.</div>}
      </div>
    </section>
  );
}

