import { Button } from "@/components/ui/Button";

type StateBlockProps = {
  title: string;
  description: string;
  actionLabel?: string;
  onAction?: () => void;
};

export function StateBlock({ title, description, actionLabel, onAction }: StateBlockProps) {
  return (
    <div className="state-block">
      <div className="empty-mark">{actionLabel ? "!" : "·"}</div>
      <h4>{title}</h4>
      <p>{description}</p>
      {actionLabel && onAction ? (
        <Button variant="ghost" onClick={onAction}>
          {actionLabel}
        </Button>
      ) : null}
    </div>
  );
}
