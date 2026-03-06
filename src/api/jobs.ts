import { apiRequest } from "@/api/client";
import type { Job } from "@/api/types";

export function getJob(jobId: string) {
  return apiRequest<Job>(`/jobs/${jobId}`);
}
