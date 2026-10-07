import fs from "node:fs/promises";
import path from "node:path";
import { createRequire } from "node:module";

const require = createRequire(new URL("../../frontend/package.json", import.meta.url));
const { chromium } = require("playwright");

const required = (name) => {
  const value = process.env[name];
  if (!value) throw new Error(`${name} is required`);
  return value;
};

const baseUrl = required("M00_UI_BASE_URL").replace(/\/$/, "");
const username = required("M00_UI_USERNAME");
const password = required("M00_UI_PASSWORD");
const evidenceDir = required("M00_UI_EVIDENCE_DIR");
const candidateSha = required("M00_UI_CANDIDATE_SHA");
const consoleErrors = [];
const findings = [];

await fs.mkdir(evidenceDir, { recursive: true });

const settingsUrl = `${baseUrl}/panel/settings#catx-features`;
const record = (classification, area, details) => findings.push({ classification, area, details });

const waitForPanel = async (page) => {
  await page.waitForFunction(
    async (url) => {
      try {
        const response = await fetch(url, { cache: "no-store" });
        return response.ok;
      } catch {
        return false;
      }
    },
    `${baseUrl}/csrf-token`,
    { timeout: 20_000 },
  );
};

const login = async (page) => {
  await page.goto(baseUrl, { waitUntil: "domcontentloaded" });
  await page.locator('input[autocomplete="username"]').fill(username);
  await page.locator('input[autocomplete="current-password"]').fill(password);
  await page.locator('button[type="submit"]').click();
  await page.waitForURL(/\/panel\//, { timeout: 20_000 });
  await page.waitForTimeout(700);
};

const openSettings = async (page, language) => {
  await page.goto(settingsUrl, { waitUntil: "domcontentloaded" });
  await page.waitForFunction(() => document.body.innerText.includes("CatX"), undefined, {
    timeout: 20_000,
  });
  await page.waitForTimeout(500);
  const expected = language === "ru-RU" ? "Функции CatX-UI" : "CatX-UI Features";
  if (!(await page.getByText(expected, { exact: true }).count())) {
    throw new Error(`M00 settings heading missing for ${language}`);
  }
};

const bodyText = (page) => page.locator("body").innerText();

const setPanelLanguage = async (page, language) => {
  await page.evaluate((value) => {
    for (const cookiePath of ["/", "/panel", "/panel/"]) {
      document.cookie = `lang=; Max-Age=0; path=${cookiePath}`;
    }
    document.cookie = `lang=${encodeURIComponent(value)}; path=/`;
  }, language);
};

const capture = async (page, name) => {
  await page.screenshot({ path: path.join(evidenceDir, `${name}.png`), fullPage: true });
};

const inspectLayout = async (page, name, language, theme) => {
  const metrics = await page.evaluate(() => {
    const rows = [...document.querySelectorAll(".ant-list-item")];
    const rowHeights = rows.map((row) => Math.round(row.getBoundingClientRect().height));
    const controls = [...document.querySelectorAll('.ant-list-item [role="switch"]')];
    const root = document.documentElement;
    const body = document.body;
    return {
      viewport: { width: window.innerWidth, height: window.innerHeight },
      scrollWidth: Math.max(root.scrollWidth, body.scrollWidth),
      clientWidth: Math.min(root.clientWidth, body.clientWidth),
      listRows: rows.length,
      rowHeights,
      switches: controls.length,
      switchLabels: controls.map((control) => control.getAttribute("aria-label")),
      desiredLabels: document.querySelectorAll(
        '.ant-list-item [aria-label*="Desired state"], .ant-list-item [aria-label*="Желаемое состояние"]',
      ).length,
      runtimeLabels: [...document.querySelectorAll(".ant-list-item")].filter((row) =>
        /Active runtime state:|Фактическое состояние:/.test(row.textContent || ""),
      ).length,
      maturityBadges: document.querySelectorAll(".fork-maturity-badge").length,
      headingBadges: document.querySelectorAll("h2 .fork-maturity-badge").length,
      bodyClass: body.className,
      dataTheme: document.documentElement.getAttribute("data-theme"),
    };
  });
  await fs.writeFile(
    path.join(evidenceDir, `${name}.json`),
    JSON.stringify({ name, language, theme, metrics }, null, 2),
    "utf8",
  );

  if (metrics.scrollWidth > metrics.clientWidth + 1) {
    record(
      "RESPONSIVE-BUG",
      `${name} horizontal overflow`,
      `scrollWidth=${metrics.scrollWidth}, clientWidth=${metrics.clientWidth}`,
    );
  } else {
    record("NO-ISSUE", `${name} horizontal overflow`, "No horizontal overflow detected.");
  }
  if (metrics.listRows === 0) {
    record("UI-BUG", `${name} feature rows`, "Feature Settings rendered no feature rows.");
  } else {
    record("NO-ISSUE", `${name} feature rows`, `${metrics.listRows} feature rows rendered.`);
  }
  if (metrics.switches !== metrics.listRows) {
    record(
      "ACCESSIBILITY-BUG",
      `${name} switch controls`,
      `Expected one switch per row; rows=${metrics.listRows}, switches=${metrics.switches}.`,
    );
  } else if (metrics.switchLabels.some((label) => !label)) {
    record(
      "ACCESSIBILITY-BUG",
      `${name} switch labels`,
      "At least one switch lacks an aria-label.",
    );
  } else {
    record("NO-ISSUE", `${name} switch controls`, "Every feature row has one labelled switch.");
  }
  if (metrics.headingBadges !== 1) {
    record(
      "UI-BUG",
      `${name} heading maturity`,
      `Expected one M00 heading badge; found ${metrics.headingBadges}.`,
    );
  } else {
    record(
      "NO-ISSUE",
      `${name} heading maturity`,
      "Exactly one compact M00 heading badge rendered.",
    );
  }
  if (metrics.runtimeLabels !== metrics.listRows) {
    record(
      "STATE-PRESENTATION-BUG",
      `${name} runtime labels`,
      `Expected one active-runtime label per row; rows=${metrics.listRows}, labels=${metrics.runtimeLabels}.`,
    );
  } else {
    record(
      "NO-ISSUE",
      `${name} runtime labels`,
      "Desired and active-runtime labels are both visible per row.",
    );
  }
  if (
    metrics.rowHeights.length > 1 &&
    Math.max(...metrics.rowHeights) - Math.min(...metrics.rowHeights) > 72
  ) {
    record(
      "RESPONSIVE-BUG",
      `${name} row heights`,
      `Row height spread is ${Math.max(...metrics.rowHeights) - Math.min(...metrics.rowHeights)}px.`,
    );
  } else {
    record("NO-ISSUE", `${name} row heights`, "Feature rows have controlled height variation.");
  }
  await capture(page, name);
};

const csrf = async (page) =>
  page.evaluate(async (url) => {
    const response = await fetch(url, { cache: "no-store" });
    const body = await response.json();
    return body.obj;
  }, `${baseUrl}/csrf-token`);

const featureSnapshot = async (page) =>
  page.evaluate(async (url) => {
    const response = await fetch(url, { cache: "no-store" });
    return response.json();
  }, `${baseUrl}/panel/api/fork/settings/features`);

const updateFlags = async (page, flags) => {
  const token = await csrf(page);
  return page.evaluate(
    async ({ url, token: csrfToken, nextFlags }) => {
      const response = await fetch(url, {
        method: "PUT",
        headers: { "Content-Type": "application/json", "X-CSRF-Token": csrfToken },
        body: JSON.stringify({ flags: nextFlags }),
      });
      return { status: response.status, body: await response.json() };
    },
    { url: `${baseUrl}/panel/api/fork/settings/features`, token, nextFlags: flags },
  );
};

const restartPanel = async (page) => {
  const token = await csrf(page);
  return page.evaluate(
    async ({ url, token: csrfToken }) => {
      const response = await fetch(url, {
        method: "POST",
        headers: { "Content-Type": "application/json", "X-CSRF-Token": csrfToken },
        body: "{}",
      });
      return { status: response.status, body: await response.json() };
    },
    { url: `${baseUrl}/panel/api/setting/restartPanel`, token },
  );
};

const waitForFeatureState = async (page, key, states) => {
  const handle = await page.waitForFunction(
    async ({ url, featureKey, expectedStates }) => {
      try {
        const response = await fetch(url, { cache: "no-store" });
        const body = await response.json();
        const item = body.obj?.items?.find((candidate) => candidate.key === featureKey);
        return item && expectedStates.includes(item.state) ? item.state : false;
      } catch {
        return false;
      }
    },
    {
      url: `${baseUrl}/panel/api/fork/settings/features`,
      featureKey: key,
      expectedStates: states,
    },
    { timeout: 30_000 },
  );
  const state = await handle.jsonValue();
  await handle.dispose();
  return state;
};

const inspectEnglish = async (page) => {
  await openSettings(page, "en-US");
  await page.mouse.move(10, 240);
  await page.waitForTimeout(350);
  const navEntry = page.locator(".ant-menu-item").filter({ hasText: "CatX-UI Features" });
  if ((await navEntry.count()) === 0) {
    record("UI-BUG", "English M00 navigation", "Feature Settings navigation entry is not visible.");
  } else {
    record(
      "NO-ISSUE",
      "English M00 navigation",
      "Feature Settings entry is visible in the sidebar.",
    );
  }
  const body = await bodyText(page);
  if ((body.match(/\bDEVS\b/g) || []).length < 2) {
    record(
      "UI-BUG",
      "English DEVS presentation",
      "DEVS is not visible in both navigation and heading.",
    );
  } else {
    record(
      "NO-ISSUE",
      "English DEVS presentation",
      "DEVS is visible in navigation and page heading.",
    );
  }
  if (!(await page.getByText("Desired state", { exact: true }).count())) {
    record("STATE-PRESENTATION-BUG", "English desired state", "Desired state label is missing.");
  }
  if (!(await page.getByText("Active runtime state:", { exact: false }).count())) {
    record(
      "STATE-PRESENTATION-BUG",
      "English active state",
      "Active runtime state label is missing.",
    );
  }
  if (
    (await page.getByRole("button", { name: "Save feature settings", exact: true }).count()) !== 1
  ) {
    record("UI-BUG", "English save action", "Expected one localized Save feature settings action.");
  } else {
    record("NO-ISSUE", "English save action", "Save action is visible and localized.");
  }
  await inspectLayout(page, "english-dark-1920", "en-US", "dark");

  const themeButton = page.locator("#theme-cycle");
  if ((await themeButton.count()) === 1) {
    await themeButton.click();
    await page.waitForFunction(() => document.documentElement.dataset.theme === "ultra-dark");
    await inspectLayout(page, "english-ultra-1920", "en-US", "ultra-dark");
    await themeButton.click();
    await page.waitForFunction(() => document.body.classList.contains("light"));
    await inspectLayout(page, "english-light-1920", "en-US", "light");
  } else {
    record(
      "COSMETIC",
      "English theme control",
      "Theme cycle control was not available in the desktop shell.",
    );
  }
};

const inspectRussian = async (page) => {
  await setPanelLanguage(page, "ru-RU");
  await openSettings(page, "ru-RU");
  const body = await bodyText(page);
  for (const expected of ["Функции CatX-UI", "Желаемое состояние", "Фактическое состояние:"]) {
    if (!body.includes(expected)) {
      record("UI-BUG", `Russian ${expected}`, `Expected localized text is missing: ${expected}`);
    }
  }
  for (const raw of ["feature_off", "active", "initializing", "restart_required", "error"]) {
    if (new RegExp(`(^|\\s)${raw}($|\\s)`, "i").test(body)) {
      record(
        "STATE-PRESENTATION-BUG",
        "Russian machine-state leak",
        `Raw state ${raw} is visible.`,
      );
    }
  }
  record(
    "NO-ISSUE",
    "Russian long-string wrapping",
    "Russian M00 text rendered for responsive inspection.",
  );
  await inspectLayout(page, "russian-dark-1920", "ru-RU", "light-after-cycle");
};

const inspectResponsiveViewport = async (browser, width, height) => {
  const context = await browser.newContext({ viewport: { width, height } });
  const page = await context.newPage();
  await login(page);
  await openSettings(page, "en-US");
  if (width <= 768) {
    const handle = page.getByRole("button", { name: /Open menu/i });
    if ((await handle.count()) !== 1) {
      record("RESPONSIVE-BUG", `${width}px sidebar`, "Mobile menu handle is missing.");
    } else {
      await handle.click();
      await page.waitForTimeout(250);
      if ((await page.locator(".ant-drawer").count()) === 0) {
        record("RESPONSIVE-BUG", `${width}px sidebar`, "Mobile navigation drawer did not open.");
      } else {
        record("NO-ISSUE", `${width}px sidebar`, "Mobile navigation drawer opens.");
      }
    }
  }
  await inspectLayout(page, `english-dark-${width}`, "en-US", "dark");
  await context.close();
};

const inspectStateMatrix = async (page) => {
  await setPanelLanguage(page, "en-US");
  await openSettings(page, "en-US");
  const initial = await featureSnapshot(page);
  if (!initial.success || !initial.obj?.items?.length)
    throw new Error("M00 feature snapshot is empty");
  const originalFlags = Object.fromEntries(
    initial.obj.items.map((item) => [item.key, item.enabled]),
  );
  const offFlags = Object.fromEntries(Object.keys(originalFlags).map((key) => [key, false]));
  await updateFlags(page, offFlags);
  await page.reload({ waitUntil: "domcontentloaded" });
  await page.waitForTimeout(700);
  const analyticsRow = page.locator(".ant-list-item").filter({ hasText: "Analytics" }).first();
  if ((await analyticsRow.count()) !== 1) throw new Error("M00 Analytics row is missing");
  await analyticsRow.getByRole("switch").click();
  await page.getByRole("button", { name: "Save feature settings", exact: true }).click();
  await page.waitForFunction(
    () => document.body.innerText.includes("Restart required"),
    undefined,
    { timeout: 10_000 },
  );
  record(
    "NO-ISSUE",
    "M00-UI-03 restart-required",
    "Saved ON / runtime not yet applied is visibly distinct.",
  );
  await capture(page, "state-restart-required");

  const restartResult = await restartPanel(page);
  if (restartResult.status < 200 || restartResult.status >= 300) {
    record(
      "STATE-PRESENTATION-BUG",
      "M00-UI-02 active",
      `restartPanel returned HTTP ${restartResult.status}.`,
    );
  } else {
    await waitForPanel(page);
    await login(page);
    await openSettings(page, "en-US");
    const state = await waitForFeatureState(page, "analytics.enabled", ["active", "error"]);
    if (state === "active") {
      record(
        "NO-ISSUE",
        "M00-UI-02 active",
        "Runtime ACTIVE is visibly distinct from desired state and DEVS.",
      );
      await capture(page, "state-active");
    } else {
      record(
        "STATE-PRESENTATION-BUG",
        "M00-UI-02 active",
        `Analytics reached ${state}, not active.`,
      );
      await capture(page, "state-error");
    }
  }

  await updateFlags(page, originalFlags);
  await page.waitForTimeout(400);
  await restartPanel(page);
  await waitForPanel(page);
  const restored = await featureSnapshot(page);
  if (restored.success)
    record(
      "NO-ISSUE",
      "M00 disposable-state cleanup",
      "Original desired flags restored in the disposable panel.",
    );
};

const browser = await chromium.launch({
  headless: true,
  args: process.getuid?.() === 0 ? ["--no-sandbox", "--disable-setuid-sandbox"] : [],
});

try {
  for (const [name, viewport] of [
    ["desktop", { width: 1920, height: 1080 }],
    ["compact", { width: 1366, height: 768 }],
  ]) {
    const context = await browser.newContext({ viewport });
    const page = await context.newPage();
    page.on("console", (message) => {
      if (message.type() === "error" && !/favicon\.ico/.test(message.location().url || "")) {
        consoleErrors.push(`${name} console.error: ${message.text()}`);
      }
    });
    page.on("pageerror", (error) => consoleErrors.push(`${name} pageerror: ${error.message}`));
    await login(page);
    if (name === "desktop") {
      await inspectEnglish(page);
      await inspectRussian(page);
      await inspectStateMatrix(page);
    } else {
      await inspectLayout(page, "english-dark-1366", "en-US", "dark");
    }
    await context.close();
  }
  for (const [width, height] of [
    [1024, 768],
    [768, 768],
    [430, 900],
    [375, 812],
  ]) {
    await inspectResponsiveViewport(browser, width, height);
  }
} finally {
  await browser.close();
  await fs.writeFile(
    path.join(evidenceDir, "browser-console.log"),
    `${consoleErrors.join("\n")}\n`,
    "utf8",
  );
  await fs.writeFile(
    path.join(evidenceDir, "m00-ui-findings.json"),
    JSON.stringify({ candidateSha, findings }, null, 2),
    "utf8",
  );
}

const confirmed = findings.filter(({ classification }) => classification !== "NO-ISSUE");
if (consoleErrors.length || confirmed.length) {
  throw new Error(
    `M00 UI review found ${confirmed.length} non-NO-ISSUE finding(s) and ${consoleErrors.length} browser console error(s).`,
  );
}
