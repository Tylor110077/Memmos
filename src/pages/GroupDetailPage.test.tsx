import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { describe, expect, it } from "vitest";
import { ToastProvider } from "@/components/feedback/ToastProvider";
import { GroupDetailPage } from "@/pages/GroupDetailPage";

function renderPage() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });

  return render(
    <QueryClientProvider client={queryClient}>
      <ToastProvider>
        <MemoryRouter initialEntries={["/groups/grp_agent"]}>
          <Routes>
            <Route path="/groups/:groupId" element={<GroupDetailPage />} />
          </Routes>
        </MemoryRouter>
      </ToastProvider>
    </QueryClientProvider>,
  );
}

describe("GroupDetailPage", () => {
  it("renders resources and can create a web resource", async () => {
    const user = userEvent.setup();
    renderPage();

    expect(await screen.findByText("OpenAI Agents Cookbook.pdf")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "上传网页" }));
    await user.type(screen.getByPlaceholderText("https://example.com/article"), "https://example.com/new-article");
    await user.type(screen.getByPlaceholderText("可选，便于后续辨识"), "新的网页资源");
    await user.click(screen.getByRole("button", { name: "创建网页资源" }));

    await waitFor(() => expect(screen.getByText("新的网页资源")).toBeInTheDocument());
  });
});
