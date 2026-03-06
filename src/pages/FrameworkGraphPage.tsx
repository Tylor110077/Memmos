import { Link, useParams } from "react-router-dom";
import { AppChrome } from "@/components/layout/AppChrome";
import { KnowledgeGraph } from "@/components/graph/KnowledgeGraph";
import { Button } from "@/components/ui/Button";
import { StateBlock } from "@/components/ui/StateBlock";
import { useFrameworkGraph } from "@/hooks/useGraphs";
import { useGroupEvents } from "@/hooks/useGroupEvents";
import { useGroupDetail } from "@/hooks/useGroups";
import { useGroupSidebar } from "@/hooks/useGroupSidebar";
import { useTrackRecentGroup } from "@/hooks/useRecentGroups";

export function FrameworkGraphPage() {
  const { groupId = "" } = useParams();
  const groupQuery = useGroupDetail(groupId);
  const graphQuery = useFrameworkGraph(groupId, 2);
  useGroupEvents(groupId);
  useTrackRecentGroup(groupId);
  const { navItems, recentItems } = useGroupSidebar(groupId);

  return (
    <AppChrome
      navItems={navItems}
      recentItems={recentItems}
      sidebarClassName="graph-sidebar"
      workspaceClassName="shell-framework"
      sidebarFooter={
        <div className="resource-card">
          <span className="pill blue">Group Overview</span>
          <h4>{groupQuery.data?.name ?? "框架图谱"}</h4>
          <p>高层结构视图，用于建立整体认知而不是查看单节点细节。</p>
        </div>
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
          {graphQuery.data ? (
            <section className="graph-canvas">
              <KnowledgeGraph graph={graphQuery.data} />
            </section>
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
