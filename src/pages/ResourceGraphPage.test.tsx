import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { describe, expect, it } from "vitest";
import { ToastProvider } from "@/components/feedback/ToastProvider";
import { ResourceGraphPage } from "@/pages/ResourceGraphPage";

function renderPage() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });

  return render(
    <QueryClientProvider client={queryClient}>
      <ToastProvider>
        <MemoryRouter initialEntries={["/groups/grp_agent/resources/res_cookbook/graph"]}>
          <Routes>
            <Route path="/groups/:groupId/resources/:resourceId/graph" element={<ResourceGraphPage />} />
          </Routes>
        </MemoryRouter>
      </ToastProvider>
    </QueryClientProvider>,
  );
}

describe("ResourceGraphPage", () => {
  it("renders chat history and streams a follow-up question", async () => {
    const user = userEvent.setup();
    renderPage();

    expect(await screen.findByText("资源级知识图谱")).toBeInTheDocument();
    expect(await screen.findByText("Agent 执行循环", { selector: "h5" })).toBeInTheDocument();
    expect(screen.getByText("它和工具调用节点是什么关系？")).toBeInTheDocument();

    await user.type(screen.getByPlaceholderText("继续追问当前节点的意义、例子或相邻关系"), "这个节点的核心价值是什么？");
    await user.click(screen.getByRole("button", { name: "发送问题" }));

    expect(await screen.findByRole("button", { name: "发送中..." })).toBeDisabled();
    expect(within(screen.getByRole("log")).getByText("这个节点的核心价值是什么？")).toBeInTheDocument();
    expect(await screen.findByText("这个节点在当前知识结构中承担主干解释作用，并把相关概念串成可追问的学习链路。")).toBeInTheDocument();
  });
});
