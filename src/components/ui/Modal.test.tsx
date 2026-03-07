import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { Modal } from "@/components/ui/Modal";

describe("Modal", () => {
  it("focuses the close button and closes on escape", async () => {
    const user = userEvent.setup();
    const onClose = vi.fn();

    render(
      <Modal open title="测试弹窗" description="用于验证可访问性" onClose={onClose}>
        <div>弹窗内容</div>
      </Modal>,
    );

    const closeButton = screen.getByRole("button", { name: "关闭" });
    expect(closeButton).toHaveFocus();

    await user.keyboard("{Escape}");

    expect(onClose).toHaveBeenCalledTimes(1);
  });
});
