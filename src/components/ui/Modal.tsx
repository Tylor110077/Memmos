import { type PropsWithChildren, useEffect } from "react";
import { createPortal } from "react-dom";
import { Button } from "@/components/ui/Button";

type ModalProps = PropsWithChildren<{
  open: boolean;
  title: string;
  description?: string;
  onClose: () => void;
}>;

export function Modal({ open, title, description, onClose, children }: ModalProps) {
  useEffect(() => {
    if (!open) return;
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        onClose();
      }
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [open, onClose]);

  if (!open) return null;

  return createPortal(
    <div className="modal-backdrop" role="presentation" onClick={onClose}>
      <div className="modal-card" role="dialog" aria-modal="true" aria-label={title} onClick={(event) => event.stopPropagation()}>
        <div className="modal-head">
          <div>
            <p className="eyebrow">Dialog</p>
            <h4>{title}</h4>
            {description ? <p className="body-copy">{description}</p> : null}
          </div>
          <Button variant="ghost" onClick={onClose}>
            关闭
          </Button>
        </div>
        {children}
      </div>
    </div>,
    document.body,
  );
}
