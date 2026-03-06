import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { useToast } from "@/components/feedback/ToastProvider";
import { AppChrome } from "@/components/layout/AppChrome";
import { Breadcrumbs } from "@/components/layout/Breadcrumbs";
import { KnowledgeGraph } from "@/components/graph/KnowledgeGraph";
import { Button } from "@/components/ui/Button";
import { StateBlock } from "@/components/ui/StateBlock";
import { useConversation, useCreateConversation, useStreamMessage } from "@/hooks/useConversation";
import { useExpandNode, useNodeDetail, useResourceGraph } from "@/hooks/useGraphs";
import { useGroupEvents } from "@/hooks/useGroupEvents";
import { useGroupSidebar } from "@/hooks/useGroupSidebar";
import { useTrackRecentGroup } from "@/hooks/useRecentGroups";
import { useResourceDetail } from "@/hooks/useResources";
import { useGraphWorkbenchStore } from "@/store/graphWorkbench";
import type { ConversationMessage } from "@/api/types";

export function ResourceGraphPage() {
  const { groupId = "", resourceId = "" } = useParams();
  const [draft, setDraft] = useState("");
  const [conversationId, setConversationId] = useState<string>(import.meta.env.VITE_API_MODE === "live" ? "" : "conv_loop");
  const [pendingUserMessage, setPendingUserMessage] = useState<ConversationMessage | null>(null);
  const [streamingMessage, setStreamingMessage] = useState<ConversationMessage | null>(null);
  const { pushToast } = useToast();
  const resourceQuery = useResourceDetail(resourceId, groupId);
  useGroupEvents(groupId);
  useTrackRecentGroup(groupId);
  const { navItems: sidebarNav, recentItems } = useGroupSidebar(groupId);
  const { selectedNodeId, includeExpansion, maxLevel, setSelectedNodeId, setIncludeExpansion, setMaxLevel, reset } =
    useGraphWorkbenchStore();
  const graphQuery = useResourceGraph(resourceId, { includeExpansion, maxLevel });
  const graphId = graphQuery.data?.graph.id ?? "";
  const nodeId = selectedNodeId ?? graphQuery.data?.graph.root_node_id ?? undefined;
  const nodeDetailQuery = useNodeDetail(graphId, nodeId);
  const expandMutation = useExpandNode(graphId, resourceId, groupId);
  const createConversation = useCreateConversation();
  const conversationQuery = useConversation(conversationId);
  const streamMessage = useStreamMessage(conversationId);

  useEffect(() => {
    return () => reset();
  }, [reset]);

  useEffect(() => {
    if (!selectedNodeId && graphQuery.data?.graph.root_node_id) {
      setSelectedNodeId(graphQuery.data.graph.root_node_id);
    }
  }, [graphQuery.data?.graph.root_node_id, selectedNodeId, setSelectedNodeId]);

  const messages = [
    ...(conversationQuery.data?.messages ?? []),
    ...(pendingUserMessage ? [pendingUserMessage] : []),
    ...(streamingMessage ? [streamingMessage] : []),
  ];
  const resource = resourceQuery.data;

  async function ensureConversation() {
    if (conversationId) {
      return conversationId;
    }
    const conversation = await createConversation.mutateAsync({
      group_id: groupId,
      graph_id: graphId,
      current_node_id: nodeId ?? "",
      title: nodeDetailQuery.data?.name,
    });
    setConversationId(conversation.id);
    return conversation.id;
  }

  async function handleSendMessage(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!draft.trim()) return;
    const content = draft.trim();
    const existingConversationId = await ensureConversation();
    if (!conversationId) {
      setConversationId(existingConversationId);
    }
    const now = new Date().toISOString();
    setPendingUserMessage({
      id: `msg_pending_user_${now}`,
      conversation_id: existingConversationId,
      current_node_id: nodeId ?? "",
      role: "user",
      content,
      citations: { chunk_ids: [], node_ids: [] },
      created_at: now,
    });
    setStreamingMessage({
      id: "msg_streaming_assistant",
      conversation_id: existingConversationId,
      current_node_id: nodeId ?? "",
      role: "assistant",
      content: "",
      citations: { chunk_ids: [], node_ids: [] },
      created_at: now,
    });
    try {
      await streamMessage.mutateAsync({
        conversationId: existingConversationId,
        content,
        handlers: {
          onStart: (messageId) =>
            setStreamingMessage((current) => (current ? { ...current, id: messageId } : current)),
          onDelta: (delta) =>
            setStreamingMessage((current) => (current ? { ...current, content: `${current.content}${delta}` } : current)),
          onDone: (message) => {
            setPendingUserMessage(null);
            setStreamingMessage(message);
          },
        },
      });
      setPendingUserMessage(null);
      setStreamingMessage(null);
      setDraft("");
      pushToast({ title: "问题已发送", description: "回答已围绕当前节点刷新。", tone: "success" });
    } catch (error) {
      setPendingUserMessage(null);
      setStreamingMessage(null);
      pushToast({
        title: "发送失败",
        description: error instanceof Error ? error.message : "当前问题发送失败，请稍后重试。",
        tone: "error",
      });
    }
  }

  if (resource?.status === "failed") {
    return (
      <AppChrome
        navItems={sidebarNav}
        recentItems={recentItems}
        sidebarClassName="graph-sidebar"
        workspaceClassName="shell-graph"
        main={
          <StateBlock
            title="资源处理失败"
            description={resource.error_message ?? "当前资源无法生成图谱，请先重试处理。"}
          />
        }
        context={<div />}
      />
    );
  }

  if (resource && resource.status !== "completed") {
    return (
      <AppChrome
        navItems={sidebarNav}
        recentItems={recentItems}
        sidebarClassName="graph-sidebar"
        workspaceClassName="shell-graph"
        main={<StateBlock title="图谱生成中" description="资源尚未完成处理，图谱会在后端任务完成后自动可见。" />}
        context={<div />}
      />
    );
  }

  return (
    <AppChrome
      navItems={sidebarNav}
      recentItems={recentItems}
      sidebarClassName="graph-sidebar"
      workspaceClassName="shell-graph"
      sidebarFooter={
        <>
          <div className="resource-card">
            <span className="pill neutral">{resource?.resource_type.toUpperCase()} 资源图谱</span>
            <h4>{resource?.name}</h4>
            <p>当前聚焦：{nodeDetailQuery.data?.name ?? "主干主题"}</p>
          </div>
          <div className="control-card">
            <h5>层级过滤</h5>
            <div className="segmented">
              <button className={maxLevel === 2 ? "active" : ""} onClick={() => setMaxLevel(2)}>
                1-2
              </button>
              <button className={maxLevel === 3 ? "active" : ""} onClick={() => setMaxLevel(3)}>
                1-3
              </button>
              <button className={maxLevel === undefined ? "active" : ""} onClick={() => setMaxLevel(undefined)}>
                全部
              </button>
            </div>
          </div>
          <div className="control-card">
            <h5>显示选项</h5>
            <label className="switch-row">
              <span>显示扩展节点</span>
              <input type="checkbox" checked={includeExpansion} onChange={(event) => setIncludeExpansion(event.target.checked)} />
            </label>
          </div>
        </>
      }
      main={
        <>
          <header className="topbar">
            <div>
              <Breadcrumbs
                items={[
                  { label: "分组列表", to: "/groups" },
                  { label: "分组详情", to: `/groups/${groupId}` },
                  { label: resource?.name ?? "资源图谱" },
                ]}
              />
              <p className="eyebrow">Resource Graph</p>
              <h3>资源级知识图谱</h3>
              <p className="body-copy">当前资源：{resource?.name}</p>
            </div>
            <div className="topbar-actions">
              <Link to={`/groups/${groupId}`}>
                <Button variant="ghost">返回分组</Button>
              </Link>
            </div>
          </header>
          {graphQuery.data ? (
            <section className="graph-canvas">
              <KnowledgeGraph
                graph={graphQuery.data}
                selectedNodeId={nodeId}
                onSelectNode={setSelectedNodeId}
                exportName={`${resource?.name ?? "resource"}-knowledge-graph`}
              />
            </section>
          ) : (
            <StateBlock title="暂无图谱数据" description="当前资源缺少可渲染的图谱结构。" />
          )}
        </>
      }
      context={
        <div className="chat-context">
          <section className="panel">
            <div className="panel-title">
              <h4>节点详情</h4>
              <span className="pill blue">当前节点</span>
            </div>
            {nodeDetailQuery.data ? (
              <>
                <div className="detail-block">
                  <h5>{nodeDetailQuery.data.name}</h5>
                  <p>{nodeDetailQuery.data.description}</p>
                </div>
                <div className="detail-block">
                  <h6>在当前主题中的意义</h6>
                  <p>{nodeDetailQuery.data.meaning}</p>
                </div>
                <div className="detail-block">
                  <h6>相关关系</h6>
                  <div className="relation-list">
                    {nodeDetailQuery.data.neighbors.map((neighbor) => (
                      <div key={neighbor.node_id} className="relation-item">
                        <strong>{neighbor.node_name}</strong>
                        <span>
                          {neighbor.direction === "out" ? "指向" : "来自"} · {neighbor.relation_type}
                        </span>
                      </div>
                    ))}
                  </div>
                </div>
                <div className="detail-block">
                  <h6>相关例子</h6>
                  {nodeDetailQuery.data.examples.map((example) => (
                    <p key={example.id}>
                      {example.example_text}
                      <span className="example-source">{example.source_type === "extracted" ? "原始提取示例" : "系统生成示例"}</span>
                    </p>
                  ))}
                </div>
                <Button
                  block
                  onClick={() => nodeId && expandMutation.mutate({ nodeId, reason: "我想更理解这个概念的直接支撑知识" })}
                  disabled={expandMutation.isPending}
                >
                  {expandMutation.isPending ? "扩展处理中..." : "生成直接相关扩展节点"}
                </Button>
              </>
            ) : (
              <StateBlock title="请选择节点" description="点击图谱节点后，这里会展示解释、意义和相邻关系。" />
            )}
          </section>

          <section className="panel chat-panel">
            <div className="panel-title">
              <h4>侧边栏对话</h4>
            </div>
            <div className="chat-list" role="log" aria-live="polite" aria-busy={streamMessage.isPending}>
              {messages.map((message) => (
                <div key={message.id} className={`chat-msg ${message.role === "user" ? "user" : "ai"}`}>
                  <span className="role">{message.role === "assistant" ? "助理" : message.role === "system" ? "系统" : "你"}</span>
                  <p>{message.content}</p>
                  {message.role === "assistant" && (message.citations.node_ids.length > 0 || message.citations.chunk_ids.length > 0) ? (
                    <div className="citation-row">
                      {message.citations.node_ids.length > 0 ? <span>引用节点 {message.citations.node_ids.length}</span> : null}
                      {message.citations.chunk_ids.length > 0 ? <span>引用片段 {message.citations.chunk_ids.length}</span> : null}
                    </div>
                  ) : null}
                </div>
              ))}
            </div>
            <form className="chat-form" onSubmit={handleSendMessage}>
              <label className="sr-only" htmlFor="resource-graph-question">
                输入问题
              </label>
              <textarea
                id="resource-graph-question"
                className="chat-input"
                rows={3}
                placeholder="继续追问当前节点的意义、例子或相邻关系"
                value={draft}
                onChange={(event) => setDraft(event.target.value)}
              />
              <Button type="submit" disabled={streamMessage.isPending}>
                {streamMessage.isPending ? "发送中..." : "发送问题"}
              </Button>
            </form>
          </section>
        </div>
      }
    />
  );
}
