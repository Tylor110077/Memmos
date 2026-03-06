import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it } from "vitest";
import { ToastProvider } from "@/components/feedback/ToastProvider";
import { clearRecentGroups, seedRecentGroups } from "@/hooks/useRecentGroups";
import { GroupsPage } from "@/pages/GroupsPage";

function renderPage(initialEntry = "/groups") {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });

  return render(
    <QueryClientProvider client={queryClient}>
      <ToastProvider>
        <MemoryRouter initialEntries={[initialEntry]}>
          <GroupsPage />
        </MemoryRouter>
      </ToastProvider>
    </QueryClientProvider>,
  );
}

describe("GroupsPage", () => {
  beforeEach(() => {
    clearRecentGroups();
  });

  it("renders the list and can create a group", async () => {
    const user = userEvent.setup();
    const groupName = "agent-test-group";
    renderPage();

    expect((await screen.findAllByText("LLM Agent 体系")).length).toBeGreaterThan(0);

    await user.click(screen.getAllByRole("button", { name: "创建分组" })[0]);
    fireEvent.change(screen.getByPlaceholderText("例如：LLM Agent 体系"), {
      target: { value: groupName },
    });
    await user.click(screen.getAllByRole("button", { name: "创建分组" })[1]);

    await waitFor(() =>
      expect(screen.getAllByRole("link", { name: "进入分组" })[0].closest(".group-card")).toHaveTextContent(groupName),
    );
  });

  it("opens create modal from the empty add card", async () => {
    const user = userEvent.setup();
    renderPage();

    await screen.findByRole("button", { name: /新建一个学习分组/ });
    await user.click(screen.getByRole("button", { name: /新建一个学习分组/ }));

    expect(screen.getByRole("dialog", { name: "创建分组" })).toBeInTheDocument();
  });

  it("supports recent view and exposes navigation links", async () => {
    seedRecentGroups(["grp_mm", "grp_go"]);
    renderPage("/groups?view=recent");

    expect(await screen.findByRole("link", { name: "最近访问" })).toHaveAttribute("href", "/groups?view=recent");
    expect(await screen.findByRole("heading", { name: "多模态检索设计", level: 4 })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Go + Eino 实践", level: 4 })).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "LLM Agent 体系", level: 4 })).not.toBeInTheDocument();
  });
});
