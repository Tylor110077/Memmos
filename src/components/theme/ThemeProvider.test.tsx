import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { ThemeProvider, useTheme } from "@/components/theme/ThemeProvider";

function ThemeProbe() {
  const { theme, toggleTheme } = useTheme();

  return (
    <button type="button" onClick={toggleTheme}>
      当前主题 {theme}
    </button>
  );
}

describe("ThemeProvider", () => {
  const storage = new Map<string, string>();

  beforeEach(() => {
    Object.defineProperty(window, "localStorage", {
      configurable: true,
      value: {
        getItem: (key: string) => storage.get(key) ?? null,
        setItem: (key: string, value: string) => {
          storage.set(key, value);
        },
        clear: () => {
          storage.clear();
        },
      },
    });
  });

  afterEach(() => {
    storage.clear();
    delete document.documentElement.dataset.theme;
    document.documentElement.style.colorScheme = "";
  });

  it("hydrates from storage and toggles theme", async () => {
    window.localStorage.setItem("goaipj-theme", "dark");
    const user = userEvent.setup();

    render(
      <ThemeProvider>
        <ThemeProbe />
      </ThemeProvider>,
    );

    expect(screen.getByRole("button", { name: "当前主题 dark" })).toBeInTheDocument();
    expect(document.documentElement.dataset.theme).toBe("dark");
    expect(document.documentElement.style.colorScheme).toBe("dark");

    await user.click(screen.getByRole("button", { name: "当前主题 dark" }));

    expect(screen.getByRole("button", { name: "当前主题 light" })).toBeInTheDocument();
    expect(window.localStorage.getItem("goaipj-theme")).toBe("light");
    expect(document.documentElement.dataset.theme).toBe("light");
  });
});
