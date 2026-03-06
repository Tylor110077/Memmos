import { useMemo, useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { useToast } from "@/components/feedback/ToastProvider";
import { AppChrome } from "@/components/layout/AppChrome";
import { Breadcrumbs } from "@/components/layout/Breadcrumbs";
import { Button } from "@/components/ui/Button";
import { Modal } from "@/components/ui/Modal";
import { StateBlock } from "@/components/ui/StateBlock";
import { useGroupResourceIndex } from "@/hooks/useGroupResourceIndex";
import { useCreateGroup, useDeleteGroup, useGroups, useUpdateGroup } from "@/hooks/useGroups";
import { useRecentGroupIds } from "@/hooks/useRecentGroups";
import { formatDateLabel } from "@/lib/utils";

type GroupFormState = {
  name: string;
  description: string;
};

const emptyForm = { name: "", description: "" };

export function GroupsPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const [keyword, setKeyword] = useState("");
  const [createOpen, setCreateOpen] = useState(false);
  const [renameGroupId, setRenameGroupId] = useState<string>();
  const [deleteGroupId, setDeleteGroupId] = useState<string>();
  const [form, setForm] = useState<GroupFormState>(emptyForm);
  const [errorMessage, setErrorMessage] = useState("");
  const { pushToast } = useToast();

  const groupsQuery = useGroups(keyword);
  const recentGroupIds = useRecentGroupIds();
  const createGroup = useCreateGroup();
  const deleteGroup = useDeleteGroup();
  const updateGroup = useUpdateGroup(renameGroupId ?? "");
  const groups = groupsQuery.data?.items ?? [];
  const { resourcesByGroupId, resources, isLoading: isResourceIndexLoading } = useGroupResourceIndex(groups);
  const requestedView = searchParams.get("view") ?? "all";
  const currentView = ["all", "recent", "processing", "framework"].includes(requestedView) ? requestedView : "all";

  const renameTarget = useMemo(
    () => groupsQuery.data?.items.find((group) => group.id === renameGroupId),
    [groupsQuery.data?.items, renameGroupId],
  );

  const navItems = [
    { label: "全部分组", to: "/groups", active: currentView === "all" },
    { label: "最近访问", to: "/groups?view=recent", active: currentView === "recent" },
    { label: "处理中资源", to: "/groups?view=processing", active: currentView === "processing" },
    { label: "框架图谱", to: "/groups?view=framework", active: currentView === "framework" },
  ];

  const recentItems = recentGroupIds
    .map((groupId) => groups.find((group) => group.id === groupId))
    .filter((group): group is (typeof groups)[number] => Boolean(group))
    .map((group) => ({
      label: group.name,
      to: `/groups/${group.id}`,
      meta: `${group.resource_count} 个资源`,
    }));

  const filteredGroups = useMemo(() => {
    if (currentView === "recent") {
      const recentLookup = new Map(recentGroupIds.map((groupId, index) => [groupId, index]));
      return groups
        .filter((group) => recentLookup.has(group.id))
        .sort((left, right) => (recentLookup.get(left.id) ?? 0) - (recentLookup.get(right.id) ?? 0));
    }

    if (currentView === "processing") {
      return groups.filter((group) =>
        (resourcesByGroupId[group.id] ?? []).some((resource) =>
          ["uploaded", "parsing", "normalizing", "graph_generating"].includes(resource.status),
        ),
      );
    }

    if (currentView === "framework") {
      return groups.filter((group) => (resourcesByGroupId[group.id] ?? []).some((resource) => resource.status === "completed"));
    }

    return groups;
  }, [currentView, groups, recentGroupIds, resourcesByGroupId]);

  const statusStats = useMemo(
    () => ({
      completed: resources.filter((resource) => resource.status === "completed").length,
      generating: resources.filter((resource) =>
        ["uploaded", "parsing", "normalizing", "graph_generating"].includes(resource.status),
      ).length,
      failed: resources.filter((resource) => resource.status === "failed").length,
    }),
    [resources],
  );

  const emptyStateCopy = {
    all: "创建第一个学习分组后，这里会展示你的知识空间。",
    recent: "访问过分组后，这里会展示最近打开的学习空间。",
    processing: "当某个分组中存在处理中资源时，会出现在这里。",
    framework: "当分组具备高层框架图谱后，会出现在这里。",
  } as const;

  async function handleCreateSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setErrorMessage("");
    if (!form.name.trim()) {
      setErrorMessage("分组名称不能为空");
      return;
    }
    if (form.name.length > 30) {
      setErrorMessage("分组名称不能超过 30 个字符");
      return;
    }
    try {
      await createGroup.mutateAsync({ name: form.name.trim(), description: form.description.trim() || undefined });
      setCreateOpen(false);
      setForm(emptyForm);
      pushToast({ title: "分组已创建", description: "新的学习分组已经加入列表。", tone: "success" });
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : "创建失败");
    }
  }

  async function handleRenameSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!renameGroupId) return;
    try {
      await updateGroup.mutateAsync({ name: form.name.trim(), description: form.description.trim() || undefined });
      setRenameGroupId(undefined);
      setForm(emptyForm);
      pushToast({ title: "分组已更新", description: "名称和描述已经同步刷新。", tone: "success" });
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : "更新失败");
    }
  }

  async function handleDeleteConfirm() {
    if (!deleteGroupId) return;
    await deleteGroup.mutateAsync(deleteGroupId);
    setDeleteGroupId(undefined);
    pushToast({ title: "分组已删除", description: "列表已经移除对应项目。", tone: "info" });
  }

  return (
    <>
      <AppChrome
        navItems={navItems}
        recentItems={recentItems}
        main={
          <>
            <header className="topbar">
              <div>
                <Breadcrumbs items={[{ label: "学习知识图谱助手" }, { label: "分组列表" }]} />
                <p className="eyebrow">Project Spaces</p>
                <h3>分组列表</h3>
              </div>
              <div className="topbar-actions">
                <input
                  className="searchbox"
                  placeholder="搜索分组、资源或主题"
                  value={keyword}
                  onChange={(event) => setKeyword(event.target.value)}
                />
                <Button onClick={() => setCreateOpen(true)}>创建分组</Button>
              </div>
            </header>

            {groupsQuery.isLoading ? (
              <div className="card-grid">
                {Array.from({ length: 4 }).map((_, index) => (
                  <div key={index} className="group-card skeleton-card" />
                ))}
              </div>
            ) : filteredGroups.length ? (
              <section className="card-grid">
                {filteredGroups.map((group, index) => (
                  <article key={group.id} className="group-card">
                    <div className="group-card-head">
                      <span className={`group-icon ${index === 1 ? "green" : index === 2 ? "purple" : ""}`}>
                        {group.name.slice(0, 2).toUpperCase()}
                      </span>
                      <span className="pill blue">{group.resource_count} 个资源</span>
                    </div>
                    <h4>{group.name}</h4>
                    <p>{group.description || "暂无描述"}</p>
                    <div className="meta-row">
                      <span>创建于 {formatDateLabel(group.created_at)}</span>
                      <span>知识空间已创建</span>
                    </div>
                    <div className="card-actions">
                      <Link className="inline-link" to={`/groups/${group.id}`}>
                        进入分组
                      </Link>
                      <button
                        className="inline-link"
                        onClick={() => {
                          setRenameGroupId(group.id);
                          setForm({
                            name: group.name,
                            description: group.description ?? "",
                          });
                        }}
                      >
                        重命名
                      </button>
                      <button className="inline-link danger-text" onClick={() => setDeleteGroupId(group.id)}>
                        删除
                      </button>
                    </div>
                  </article>
                ))}
                <button
                  type="button"
                  className="group-card muted group-card-action"
                  onClick={() => setCreateOpen(true)}
                >
                  <div className="empty-state">
                    <div className="empty-mark">+</div>
                    <h4>新建一个学习分组</h4>
                    <p>先创建项目空间，再上传资料与网页链接。</p>
                  </div>
                </button>
              </section>
            ) : (
              <StateBlock
                title="暂无分组"
                description={emptyStateCopy[currentView as keyof typeof emptyStateCopy] ?? emptyStateCopy.all}
                actionLabel={currentView === "all" ? "创建分组" : "查看全部分组"}
                onAction={() => {
                  if (currentView === "all") {
                    setCreateOpen(true);
                    return;
                  }
                  navigate("/groups");
                }}
              />
            )}
          </>
        }
        context={
          <>
            <section className="panel">
              <div className="panel-title">
                <h4>设计意图</h4>
              </div>
              <p className="body-copy">
                首页强调“项目空间”这一一级对象，让用户先理解系统按分组组织资料，而不是先进入单个文件。
              </p>
            </section>
            <section className="panel">
              <div className="panel-title">
                <h4>状态分布</h4>
              </div>
              <div className="mini-stats">
                <div>
                  <strong>{isResourceIndexLoading ? "..." : statusStats.completed}</strong>
                  <span>已完成资源</span>
                </div>
                <div>
                  <strong>{isResourceIndexLoading ? "..." : statusStats.generating}</strong>
                  <span>图谱生成中</span>
                </div>
                <div>
                  <strong>{isResourceIndexLoading ? "..." : statusStats.failed}</strong>
                  <span>失败待重试</span>
                </div>
              </div>
            </section>
          </>
        }
      />

      <Modal
        open={createOpen}
        title="创建分组"
        description="先建立知识空间，再上传资源和网页链接。"
        onClose={() => {
          setCreateOpen(false);
          setForm(emptyForm);
          setErrorMessage("");
        }}
      >
        <form className="modal-form" onSubmit={handleCreateSubmit}>
          <label className="field">
            <span>分组名称</span>
            <input
              value={form.name}
              onChange={(event) => setForm((current) => ({ ...current, name: event.target.value }))}
              placeholder="例如：LLM Agent 体系"
            />
          </label>
          <label className="field">
            <span>分组描述</span>
            <textarea
              value={form.description}
              onChange={(event) => setForm((current) => ({ ...current, description: event.target.value }))}
              placeholder="说明该分组的学习目标与资料范围"
              rows={4}
            />
          </label>
          {errorMessage ? <p className="form-error">{errorMessage}</p> : null}
          <div className="modal-actions">
            <Button type="button" variant="ghost" onClick={() => setCreateOpen(false)}>
              取消
            </Button>
            <Button type="submit" disabled={createGroup.isPending}>
              {createGroup.isPending ? "创建中..." : "创建分组"}
            </Button>
          </div>
        </form>
      </Modal>

      <Modal
        open={Boolean(renameTarget)}
        title="重命名分组"
        description="更新分组名称和描述，列表会立即刷新。"
        onClose={() => {
          setRenameGroupId(undefined);
          setForm(emptyForm);
          setErrorMessage("");
        }}
      >
        <form className="modal-form" onSubmit={handleRenameSubmit}>
          <label className="field">
            <span>分组名称</span>
            <input value={form.name} onChange={(event) => setForm((current) => ({ ...current, name: event.target.value }))} />
          </label>
          <label className="field">
            <span>分组描述</span>
            <textarea
              rows={4}
              value={form.description}
              onChange={(event) => setForm((current) => ({ ...current, description: event.target.value }))}
            />
          </label>
          {errorMessage ? <p className="form-error">{errorMessage}</p> : null}
          <div className="modal-actions">
            <Button type="button" variant="ghost" onClick={() => setRenameGroupId(undefined)}>
              取消
            </Button>
            <Button type="submit" disabled={updateGroup.isPending}>
              保存更新
            </Button>
          </div>
        </form>
      </Modal>

      <Modal open={Boolean(deleteGroupId)} title="删除分组" description="删除后该分组下的资源入口将从列表中移除。" onClose={() => setDeleteGroupId(undefined)}>
        <div className="confirm-block">
          <p>确认删除当前分组吗？这个操作会立即刷新列表。</p>
          <div className="modal-actions">
            <Button type="button" variant="ghost" onClick={() => setDeleteGroupId(undefined)}>
              取消
            </Button>
            <Button type="button" variant="danger" onClick={handleDeleteConfirm} disabled={deleteGroup.isPending}>
              删除分组
            </Button>
          </div>
        </div>
      </Modal>
    </>
  );
}
