import { Navigate, createBrowserRouter } from "react-router-dom";
import { GroupDetailPage } from "@/pages/GroupDetailPage";
import { GroupsPage } from "@/pages/GroupsPage";
import { FrameworkGraphPage } from "@/pages/FrameworkGraphPage";
import { ResourceGraphPage } from "@/pages/ResourceGraphPage";

export const router = createBrowserRouter([
  {
    path: "/",
    element: <Navigate to="/groups" replace />,
  },
  {
    path: "/groups",
    element: <GroupsPage />,
  },
  {
    path: "/groups/:groupId",
    element: <GroupDetailPage />,
  },
  {
    path: "/groups/:groupId/framework",
    element: <FrameworkGraphPage />,
  },
  {
    path: "/groups/:groupId/resources/:resourceId/graph",
    element: <ResourceGraphPage />,
  },
]);
