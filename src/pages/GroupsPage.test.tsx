import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";
import { GroupsPage } from "@/pages/GroupsPage";

function renderPage() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });

  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter>
        <GroupsPage />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe("GroupsPage", () => {
  it("renders the list and can create a group", async () => {
    const user = userEvent.setup();
    renderPage();

    expect((await screen.findAllByText("LLM Agent 体系")).length).toBeGreaterThan(0);

    await user.click(screen.getAllByRole("button", { name: "创建分组" })[0]);
    await user.type(screen.getByPlaceholderText("例如：LLM Agent 体系"), "测试分组");
    await user.click(screen.getAllByRole("button", { name: "创建分组" })[1]);

    await waitFor(() => expect(screen.getAllByText("测试分组").length).toBeGreaterThan(0));
  });
});
