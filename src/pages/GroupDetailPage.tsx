import { useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { AppChrome } from "@/components/layout/AppChrome";
import { Button } from "@/components/ui/Button";
import { Modal } from "@/components/ui/Modal";
import { KnowledgeGraph } from "@/components/graph/KnowledgeGraph";
import { StateBlock } from "@/components/ui/StateBlock";
import { StatusBadge } from "@/components/ui/StatusBadge";
import { useFrameworkGraph, useGenerateFrameworkGraph } from "@/hooks/useGraphs";
import { useGroupDetail } from "@/hooks/useGroups";
import {
  useCreateWebResource,
  useDeleteResource,
  useResources,
  useRetryResource,
  useUploadResource,
} from "@/hooks/useResources";
import { formatDateLabel } from "@/lib/utils";

export function GroupDetailPage() {
  const { groupId = "" } = useParams();
  const [uploadOpen, setUploadOpen] = useState(false);
  const [webOpen, setWebOpen] = useState(false);
  const [selectedFile, setSelectedFile] = useState<File>();
  const [resourceName, setResourceName] = useState("");
  const [webUrl, setWebUrl] = useState("");
  const [webName, setWebName] = useState("");
  const [errorMessage, setErrorMessage] = useState("");

  const groupQuery = useGroupDetail(groupId);
  const resourcesQuery = useResources(groupId);
  const frameworkQuery = useFrameworkGraph(groupId, 2);
  const uploadMutation = useUploadResource(groupId);
  const createWebMutation = useCreateWebResource(groupId);
  const retryMutation = useRetryResource(groupId);
  const deleteMutation = useDeleteResource(groupId);
  const regenerateMutation = useGenerateFrameworkGraph(groupId);

  const group = groupQuery.data;
  const resources = resourcesQuery.data?.items ?? [];

  const sidebarNav = useMemo(
    () => [
      { label: "全部分组", to: "/groups" },
      { label: group?.name ?? "当前分组", active: true },
      { label: "多模态检索设计" },
      { label: "Go + Eino 实践" },
    ],
    [group?.name],
  );

  async function handleUploadSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedFile) {
      setErrorMessage("请选择一个文件");
      return;
    }
    await uploadMutation.mutateAsync({ file: selectedFile, name: resourceName.trim() || undefined });
    setUploadOpen(false);
    setSelectedFile(undefined);
    setResourceName("");
    setErrorMessage("");
  }

  async function handleWebSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    try {
      const parsed = new URL(webUrl);
      if (!/^https?:$/.test(parsed.protocol)) {
        throw new Error("仅支持 http 或 https 链接");
      }
    } catch {
      setErrorMessage("请输入有效的网页地址");
      return;
    }
    await createWebMutation.mutateAsync({ url: webUrl.trim(), name: webName.trim() || undefined });
    setWebOpen(false);
    setWebUrl("");
    setWebName("");
    setErrorMessage("");
  }

  return (
    <>
      <AppChrome
        navItems={sidebarNav}
        main={
          <>
            <header className="topbar split">
              <div>
                <p className="eyebrow">Group Detail</p>
                <h3>{group?.name ?? "加载中..."}</h3>
                <p className="body-copy">{group?.description}</p>
              </div>
              <div className="topbar-actions">
                <Button variant="ghost" onClick={() => setWebOpen(true)}>
                  上传网页
                </Button>
                <Button variant="ghost" onClick={() => setUploadOpen(true)}>
                  上传资源
                </Button>
                <Link to={`/groups/${groupId}/framework`}>
                  <Button>查看框架图谱</Button>
                </Link>
              </div>
            </header>

            <section className="resource-shell">
              <div className="resource-panel">
                <div className="section-head">
                  <h4>资源列表</h4>
                  <span className="muted-text">支持 PDF / DOCX / PPTX / XLSX / TXT / MD / URL</span>
                </div>
                {resourcesQuery.isLoading ? (
                  <div className="resource-table">
                    {Array.from({ length: 4 }).map((_, index) => (
                      <div key={index} className="resource-row skeleton-card" />
                    ))}
                  </div>
                ) : resources.length ? (
                  <div className="resource-table">
                    <div className="resource-row head">
                      <span>资源名称</span>
                      <span>类型</span>
                      <span>更新时间</span>
                      <span>状态</span>
                    </div>
                    {resources.map((resource) => (
                      <div key={resource.id} className="resource-row resource-row-item">
                        <div>
                          <strong>{resource.name}</strong>
                          {resource.error_message ? <p className="danger-text body-copy">{resource.error_message}</p> : null}
                        </div>
                        <span className="pill neutral">{resource.resource_type.toUpperCase()}</span>
                        <span>{formatDateLabel(resource.updated_at)}</span>
                        <div className="resource-status-actions">
                          <StatusBadge status={resource.status} />
                          <div className="table-actions">
                            {resource.status === "completed" ? (
                              <Link className="inline-link" to={`/groups/${groupId}/resources/${resource.id}/graph`}>
                                查看
                              </Link>
                            ) : null}
                            {resource.status === "failed" ? (
                              <button className="inline-link" onClick={() => retryMutation.mutate(resource.id)}>
                                重试
                              </button>
                            ) : null}
                            <button className="inline-link danger-text" onClick={() => deleteMutation.mutate(resource.id)}>
                              删除
                            </button>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                ) : (
                  <StateBlock title="暂无资源" description="上传第一份资料或网页链接后，这里会显示资源处理状态和操作入口。" />
                )}
              </div>

              <div className="preview-panel">
                <div className="section-head">
                  <h4>分组级框架图谱预览</h4>
                  <div className="topbar-actions">
                    <Button variant="ghost" onClick={() => regenerateMutation.mutate()}>
                      重新生成
                    </Button>
                    <Link className="inline-link" to={`/groups/${groupId}/framework`}>
                      进入完整画布
                    </Link>
                  </div>
                </div>
                {frameworkQuery.data ? (
                  <KnowledgeGraph graph={frameworkQuery.data} />
                ) : (
                  <StateBlock title="框架图谱暂不可用" description="当资源达到可用状态后，系统会生成高层框架图谱。" />
                )}
              </div>
            </section>
          </>
        }
        context={
          <>
            <section className="panel">
              <div className="panel-title">
                <h4>处理规则</h4>
              </div>
              <ul className="check-list">
                <li>资源必须先归属于一个分组</li>
                <li>网页会被提取正文并生成 Markdown</li>
                <li>失败状态保留重试入口</li>
              </ul>
            </section>
            <section className="panel">
              <div className="panel-title">
                <h4>本组概览</h4>
              </div>
              <div className="mini-stats">
                <div>
                  <strong>{group?.resource_count ?? 0}</strong>
                  <span>总资源数</span>
                </div>
                <div>
                  <strong>{group?.completed_resource_count ?? 0}</strong>
                  <span>已完成</span>
                </div>
                <div>
                  <strong>{resources.filter((item) => item.status === "failed").length}</strong>
                  <span>失败待重试</span>
                </div>
              </div>
            </section>
          </>
        }
      />

      <Modal open={uploadOpen} title="上传文件资源" description="文件会走异步解析与图谱生成流程。" onClose={() => setUploadOpen(false)}>
        <form className="modal-form" onSubmit={handleUploadSubmit}>
          <label className="upload-dropzone">
            <input
              type="file"
              onChange={(event) => setSelectedFile(event.target.files?.[0])}
              accept=".pdf,.docx,.pptx,.xlsx,.txt,.md"
            />
            <span>{selectedFile ? selectedFile.name : "拖拽或点击选择文件"}</span>
            <small>支持 PDF / DOCX / PPTX / XLSX / TXT / MD</small>
          </label>
          <label className="field">
            <span>自定义资源名称</span>
            <input value={resourceName} onChange={(event) => setResourceName(event.target.value)} placeholder="可选，不填则使用文件名" />
          </label>
          {errorMessage ? <p className="form-error">{errorMessage}</p> : null}
          <div className="modal-actions">
            <Button type="button" variant="ghost" onClick={() => setUploadOpen(false)}>
              取消
            </Button>
            <Button type="submit" disabled={uploadMutation.isPending}>
              {uploadMutation.isPending ? "上传中..." : "开始上传"}
            </Button>
          </div>
        </form>
      </Modal>

      <Modal open={webOpen} title="创建网页资源" description="网页正文会被提取并生成 Markdown 资源。" onClose={() => setWebOpen(false)}>
        <form className="modal-form" onSubmit={handleWebSubmit}>
          <label className="field">
            <span>网页 URL</span>
            <input value={webUrl} onChange={(event) => setWebUrl(event.target.value)} placeholder="https://example.com/article" />
          </label>
          <label className="field">
            <span>资源名称</span>
            <input value={webName} onChange={(event) => setWebName(event.target.value)} placeholder="可选，便于后续辨识" />
          </label>
          {errorMessage ? <p className="form-error">{errorMessage}</p> : null}
          <div className="modal-actions">
            <Button type="button" variant="ghost" onClick={() => setWebOpen(false)}>
              取消
            </Button>
            <Button type="submit" disabled={createWebMutation.isPending}>
              {createWebMutation.isPending ? "提交中..." : "创建网页资源"}
            </Button>
          </div>
        </form>
      </Modal>
    </>
  );
}
