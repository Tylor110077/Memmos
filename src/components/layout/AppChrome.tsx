import type { ReactNode } from "react";
import { Link } from "react-router-dom";
import { useTheme } from "@/components/theme/ThemeProvider";
import { Button } from "@/components/ui/Button";
import { cn } from "@/lib/utils";

type NavItem = {
  label: string;
  to?: string;
  active?: boolean;
  meta?: string;
};

type AppChromeProps = {
  sidebarTitle?: string;
  navItems: NavItem[];
  recentItems?: NavItem[];
  main: ReactNode;
  context?: ReactNode;
  sidebarFooter?: ReactNode;
  sidebarClassName?: string;
  workspaceClassName?: string;
};

export function AppChrome({
  sidebarTitle,
  navItems,
  recentItems,
  main,
  context,
  sidebarFooter,
  sidebarClassName,
  workspaceClassName,
}: AppChromeProps) {
  const { theme, toggleTheme } = useTheme();

  return (
    <div className="page-shell">
      <div className={cn("workspace", workspaceClassName)}>
        <aside className={cn("sidebar", sidebarClassName)}>
          <div className="brand">
            <div className="brand-mark">KG</div>
            <div>
              <strong>学习知识图谱助手</strong>
              <span>{sidebarTitle ?? "Knowledge Workspace"}</span>
            </div>
          </div>
          <div className="theme-toggle-row">
            <span>{theme === "light" ? "浅色模式" : "深色模式"}</span>
            <Button variant="ghost" className="theme-toggle-btn" onClick={toggleTheme}>
              {theme === "light" ? "切换深色" : "切换浅色"}
            </Button>
          </div>
          <nav className="nav">
            {navItems.map((item) =>
              item.to ? (
                <Link key={item.label} className={cn(item.active && "active")} to={item.to}>
                  {item.label}
                </Link>
              ) : (
                <span key={item.label} className={cn("nav-link", item.active && "active")}>
                  {item.label}
                </span>
              ),
            )}
          </nav>
          {recentItems?.length ? (
            <div className="sidebar-section">
              <p>最近访问</p>
              {recentItems.map((item) => (
                item.to ? (
                  <Link key={item.label} className={cn("recent-item", item.active && "active")} to={item.to}>
                    <strong>{item.label}</strong>
                    <span>{item.meta}</span>
                  </Link>
                ) : (
                  <div key={item.label} className={cn("recent-item", item.active && "active")}>
                    <strong>{item.label}</strong>
                    <span>{item.meta}</span>
                  </div>
                )
              ))}
            </div>
          ) : null}
          {sidebarFooter}
        </aside>
        <main className="main">{main}</main>
        <aside className="context">{context}</aside>
      </div>
    </div>
  );
}
