import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { beforeEach, describe, expect, it } from "vitest";
import { ToastProvider } from "@/components/feedback/ToastProvider";
import { clearRecentGroups, readRecentGroupIds } from "@/hooks/useRecentGroups";
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
  beforeEach(() => {
    clearRecentGroups();
  });

  it("renders resources and can create a web resource", async () => {
    const user = userEvent.setup();
    const resourceName = "new-web-resource";
    renderPage();

    expect(await screen.findByText("OpenAI Agents Cookbook.pdf")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "多模态检索设计" })).toHaveAttribute("href", "/groups/grp_mm");
    expect(readRecentGroupIds()).toContain("grp_agent");

    await user.click(screen.getByRole("button", { name: "上传网页" }));
    fireEvent.change(screen.getByPlaceholderText("https://example.com/article"), {
      target: { value: "https://example.com/new-article" },
    });
    fireEvent.change(screen.getByPlaceholderText("可选，便于后续辨识"), {
      target: { value: resourceName },
    });
    await user.click(screen.getByRole("button", { name: "创建网页资源" }));

    await waitFor(() => expect(screen.getByText(resourceName)).toBeInTheDocument());
  });
});
