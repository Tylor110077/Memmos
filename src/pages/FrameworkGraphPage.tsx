import { useEffect, useMemo } from "react";
import { Link, useParams } from "react-router-dom";
import { filterGraphView } from "@/components/graph/filterGraph";
import { AppChrome } from "@/components/layout/AppChrome";
import { KnowledgeGraph } from "@/components/graph/KnowledgeGraph";
import { Button } from "@/components/ui/Button";
import { StateBlock } from "@/components/ui/StateBlock";
import { useFrameworkGraph } from "@/hooks/useGraphs";
import { useGroupEvents } from "@/hooks/useGroupEvents";
import { useGroupDetail } from "@/hooks/useGroups";
import { useGroupSidebar } from "@/hooks/useGroupSidebar";
import { useTrackRecentGroup } from "@/hooks/useRecentGroups";
import { useGraphWorkbenchStore } from "@/store/graphWorkbench";
import type { NodeType } from "@/api/types";

const nodeTypeLabels: Record<NodeType, string> = {
  topic: "主题",
  subtopic: "子主题",
  concept: "概念",
  method: "方法",
  rule: "规则",
  conclusion: "结论",
  example: "示例",
};

export function FrameworkGraphPage() {
  const { groupId = "" } = useParams();
  const groupQuery = useGroupDetail(groupId);
  const { maxLevel, searchTerm, nodeTypeFilter, setMaxLevel, setSearchTerm, toggleNodeType, clearNodeTypeFilter, reset } =
    useGraphWorkbenchStore();
  const graphQuery = useFrameworkGraph(groupId, maxLevel);
  useGroupEvents(groupId);
  useTrackRecentGroup(groupId);
  const { navItems, recentItems } = useGroupSidebar(groupId);
  const filteredGraph = useMemo(
    () =>
      graphQuery.data
        ? filterGraphView(graphQuery.data, {
            searchTerm,
            nodeTypeFilter,
          })
        : undefined,
    [graphQuery.data, nodeTypeFilter, searchTerm],
  );
  const availableNodeTypes = useMemo(
    () => Array.from(new Set((graphQuery.data?.nodes ?? []).map((node) => node.node_type))),
    [graphQuery.data?.nodes],
  );

  useEffect(() => {
    return () => reset();
  }, [reset]);

  return (
    <AppChrome
      navItems={navItems}
      recentItems={recentItems}
      sidebarClassName="graph-sidebar"
      workspaceClassName="shell-framework"
      sidebarFooter={
        <>
          <div className="resource-card">
            <span className="pill blue">Group Overview</span>
            <h4>{groupQuery.data?.name ?? "框架图谱"}</h4>
            <p>高层结构视图，用于建立整体认知而不是查看单节点细节。</p>
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
            <h5>关键词筛选</h5>
            <input
              className="searchbox compact"
              value={searchTerm}
              onChange={(event) => setSearchTerm(event.target.value)}
              placeholder="搜索框架节点"
            />
          </div>
          <div className="control-card">
            <div className="filter-head">
              <h5>节点类型</h5>
              <button className="inline-link" onClick={clearNodeTypeFilter} type="button">
                清空
              </button>
            </div>
            <div className="filter-chip-grid">
              {availableNodeTypes.map((nodeType) => (
                <button
                  key={nodeType}
                  type="button"
                  className={`filter-chip ${nodeTypeFilter.includes(nodeType) ? "active" : ""}`}
                  onClick={() => toggleNodeType(nodeType)}
                >
                  {nodeTypeLabels[nodeType]}
                </button>
              ))}
            </div>
          </div>
        </>
      }
      main={
        <>
          <header className="topbar">
            <div>
              <p className="eyebrow">Framework Graph</p>
              <h3>分组级框架图谱</h3>
            </div>
            <div className="topbar-actions">
              <Link to={`/groups/${groupId}`}>
                <Button variant="ghost">返回分组</Button>
              </Link>
            </div>
          </header>
          {filteredGraph && filteredGraph.nodes.length > 0 ? (
            <section className="graph-canvas">
              <KnowledgeGraph graph={filteredGraph} exportName={`${groupQuery.data?.name ?? "framework"}-framework-graph`} />
            </section>
          ) : graphQuery.data ? (
            <StateBlock title="当前筛选下暂无结果" description="可以清空关键词或节点类型筛选，恢复完整图谱视图。" />
          ) : (
            <StateBlock title="框架图谱尚未生成" description="等至少一部分资源处理完成后，这里会显示分组级骨架视图。" />
          )}
        </>
      }
      context={
        <>
          <section className="panel">
            <div className="panel-title">
              <h4>高层节点说明</h4>
            </div>
            <div className="detail-block">
              <h5>执行与反馈</h5>
              <p>连接规划、工具调用与观察结果，是多个资源中反复出现的共性主线。</p>
            </div>
          </section>
          <section className="panel">
            <div className="panel-title">
              <h4>使用目标</h4>
            </div>
            <ul className="check-list">
              <li>帮助用户先建立整体骨架</li>
              <li>发现资源之间的交叉主题</li>
              <li>作为资源级图谱的导航入口</li>
            </ul>
          </section>
        </>
      }
    />
  );
}
