import { FormEvent } from "react";
import type { Comment } from "../types/issue";
import { AvatarCircle } from "./AvatarCircle";

interface Props {
  comments: Comment[];
  onSubmit: (author: string, body: string) => Promise<void>;
}

export function CommentList({ comments, onSubmit }: Props) {
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    await onSubmit(String(form.get("author") || "Teng"), String(form.get("body") || ""));
    event.currentTarget.reset();
  }

  return (
    <section className="panel">
      <h3>Comments</h3>
      <div className="comments">
        {comments.map((comment) => (
          <article key={comment.id} className="comment">
            <div className="comment-author">
              <AvatarCircle name={comment.author} size={20} />
              <strong>{comment.author}</strong>
            </div>
            <p>{comment.body}</p>
            <time>{comment.createdAt}</time>
          </article>
        ))}
      </div>
      <form className="comment-form" onSubmit={submit}>
        <input name="author" defaultValue="Teng" placeholder="Author" />
        <textarea name="body" required rows={3} placeholder="Add an update, question, or resolution note" />
        <div className="button-row" style={{ justifyContent: "flex-end" }}>
          <button className="primary">Add comment</button>
        </div>
      </form>
    </section>
  );
}

