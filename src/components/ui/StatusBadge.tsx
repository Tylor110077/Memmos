import type { ResourceStatus } from "@/api/types";
import { cn } from "@/lib/utils";

const labelMap: Record<ResourceStatus, string> = {
  uploaded: "已上传",
  parsing: "解析中",
  normalizing: "标准化中",
  graph_generating: "图谱生成中",
  completed: "已完成",
  failed: "失败",
};

export function StatusBadge({ status }: { status: ResourceStatus }) {
  return <span className={cn("status", status === "completed" && "success", status === "failed" && "danger", status !== "completed" && status !== "failed" && "warning")}>{labelMap[status]}</span>;
}
