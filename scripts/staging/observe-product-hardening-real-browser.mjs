import fs from "node:fs/promises";
import { createRequire } from "node:module";
import path from "node:path";
import process from "node:process";

const require = createRequire(new URL("../../frontend/package.json", import.meta.url));
const { chromium } = require("playwright");

const required = (name) => {
  const value = process.env[name];
  if (!value) throw new Error(`${name} is required`);
  return value;
};

const baseUrl = required("CATX_HARDENING_BASE_URL").replace(/\/$/, "");
const username = required("CATX_HARDENING_USERNAME");
const password = required("CATX_HARDENING_PASSWORD");
const clientEmail = required("CATX_HARDENING_CLIENT_EMAIL");
const policyName = required("CATX_HARDENING_POLICY_NAME");
const evidenceDir = required("CATX_HARDENING_EVIDENCE_DIR");
const consoleErrors = [];
let evidencePrefix = "";

await fs.mkdir(evidenceDir, { recursive: true });

const appUrl = (route) => `${baseUrl}/panel${route.startsWith("/") ? route : `/${route}`}`;
const waitForRender = async (page) => {
  await page.waitForLoadState("domcontentloaded");
  await page.waitForTimeout(700);
};
const bodyText = async (page) => (await page.locator("body").innerText()).replace(/\s+/g, " ");
const requireText = async (page, text, label) => {
  const content = await bodyText(page);
  if (!content.includes(text))
    throw new Error(`${label}: missing visible text ${JSON.stringify(text)}`);
};
const requireNoGenericFailure = async (page, label) => {
  const content = (await bodyText(page)).toLowerCase();
  for (const forbidden of ["analytics unavailable", "policy data unavailable", "please enter id"]) {
    if (content.includes(forbidden))
      throw new Error(`${label}: generic failure remained: ${forbidden}`);
  }
};

async function login(page) {
  await page.goto(baseUrl, { waitUntil: "domcontentloaded" });
  await page.locator('input[autocomplete="username"]').fill(username);
  await page.locator('input[autocomplete="current-password"]').fill(password);
  await page.locator('button[type="submit"]').click();
  await page.waitForURL(/\/panel\//, { timeout: 20_000 });
  await waitForRender(page);
}

async function openRoute(page, route, label, visibleText) {
  await page.goto(appUrl(route), { waitUntil: "domcontentloaded" });
  await waitForRender(page);
  if (!page.url().includes("/panel/")) throw new Error(`${label}: redirected out of panel`);
  await requireNoGenericFailure(page, label);
  if (visibleText) await requireText(page, visibleText, label);
}

async function saveScreenshot(page, name) {
  await page.screenshot({
    path: path.join(evidenceDir, `${evidencePrefix}-${name}.png`),
    fullPage: true,
  });
}

async function toggleSettingsAndRestore(page) {
  await openRoute(page, "/settings#catx-features", "settings", "CatX-UI Features");
  const sponsorItem = page.locator(".ant-list-item").filter({ hasText: "Sponsors" }).first();
  const switchControl = sponsorItem.getByRole("switch");
  if ((await switchControl.count()) !== 1)
    throw new Error("settings: Sponsors switch was not rendered");
  const saveFeatureSettings = async () => {
    let saveButton = page.getByRole("button", {
      name: "Save feature settings",
      exact: true,
    });
    try {
      await saveButton.waitFor({ state: "visible" });
    } catch (error) {
      saveButton = page.locator("button").filter({ hasText: "Save feature settings" }).last();
      await saveButton.waitFor({ state: "attached" });
      await saveButton.scrollIntoViewIfNeeded();
    }
    const saved = page.waitForResponse(
      (response) =>
        response.request().method() === "PUT" &&
        response.url().includes("/panel/api/fork/settings/features") &&
        response.ok(),
    );
    await saveButton.click({ force: true });
    await saved;
    await page.waitForFunction(
      () => !document.querySelector("button.ant-btn-loading"),
      undefined,
      { timeout: 10_000 },
    );
  };
  await switchControl.click();
  await saveFeatureSettings();
  await requireText(page, "Restart required", "settings saved pending restart");
  await switchControl.click();
  await saveFeatureSettings();
  await page.waitForTimeout(500);
  await requireText(page, "Active", "settings active runtime state");
  await saveScreenshot(page, "settings-en");
}

async function exercisePolicy(page) {
  const currentPolicyName = `${policyName} ${evidencePrefix}`;
  await openRoute(page, "/policies", "policy", "Policy engine");
  await page.getByRole("button", { name: "New policy" }).click();
  const modal = page.locator(".ant-modal").last();
  await modal.getByLabel("Name").fill(currentPolicyName);
  await modal.getByLabel("Description").fill("Structured hardening qualification policy");
  const actionItem = modal.locator(".ant-form-item").filter({ hasText: "Action" }).first();
  const actionSelect = actionItem.getByRole("combobox");
  const allowOption = page.locator('[role="option"][aria-label="Allow"]');
  const allowSelected = await allowOption.evaluateAll((options) =>
    options.some((option) => option.getAttribute("aria-selected") === "true"),
  );
  if (!allowSelected) {
    await actionSelect.click();
    await page.locator('[role="option"][aria-label="Allow"]:visible').last().click();
  }
  const destinationItem = modal
    .locator(".ant-form-item")
    .filter({ hasText: "Destinations" })
    .first();
  await destinationItem.locator("input").fill("example.com");
  await destinationItem.locator("input").press("Enter");
  await requireText(page, "Advanced policy fields (JSON)", "policy structured editor");
  await modal.getByRole("button", { name: "Save", exact: true }).click();
  await page.getByText(currentPolicyName, { exact: true }).first().waitFor({ timeout: 10_000 });
  const editButton = page.getByRole("button", { name: new RegExp(`Edit.*${currentPolicyName}`) });
  await editButton.click();
  const editModal = page.locator(".ant-modal").last();
  await editModal.getByLabel("Description").fill("Updated structured qualification policy");
  await editModal.getByRole("button", { name: "Save", exact: true }).click();
  await editModal.waitFor({ state: "hidden" });
  await editButton.click();
  const persistedModal = page.locator(".ant-modal").last();
  const persistedDescription = persistedModal.getByLabel("Description", { exact: true });
  if ((await persistedDescription.inputValue()) !== "Updated structured qualification policy")
    throw new Error("policy edit result: updated description was not persisted");
  await page.keyboard.press("Escape");
  await saveScreenshot(page, "policy-en");
}

async function exerciseClientTraffic(page) {
  await openRoute(
    page,
    `/clients?search=${encodeURIComponent(clientEmail)}`,
    "clients",
    clientEmail,
  );
  const row = page.locator("tr").filter({ hasText: clientEmail }).first();
  if ((await row.count()) > 0) {
    await row.getByRole("button", { name: "Edit", exact: true }).click();
  } else {
    const card = page.locator(".client-card").filter({ hasText: clientEmail }).first();
    await card.getByRole("button", { name: "More", exact: true }).click();
    await page.getByText("Edit", { exact: true }).last().click();
  }
  const editModal = page.locator(".ant-modal").last();
  await editModal.getByRole("tab", { name: "Traffic control", exact: true }).click();
  await requireText(page, "Upload B/s", "client edit traffic panel");
  await requireText(page, "Download B/s", "client edit traffic panel");
  await page.keyboard.press("Escape");
  await page.waitForTimeout(300);
  const infoRow = page.locator("tr").filter({ hasText: clientEmail }).first();
  if ((await infoRow.count()) > 0) {
    await infoRow.getByRole("button", { name: "Client information", exact: true }).click();
  } else {
    const card = page.locator(".client-card").filter({ hasText: clientEmail }).first();
    await card.getByRole("button", { name: "Client information", exact: true }).click();
  }
  const infoModal = page.locator(".ant-modal").last();
  await requireText(page, "Traffic control", "client information traffic state");
  if ((await infoModal.getByRole("button", { name: "Save", exact: true }).count()) > 0) {
    throw new Error("client information: mutable Save control leaked into read-only traffic view");
  }
  await saveScreenshot(page, "client-traffic-en");
  await page.keyboard.press("Escape");
}

async function exercisePortal(page) {
  await openRoute(page, "/portal-access", "portal", "Portal access");
  const clientItem = page.locator(".ant-form-item").filter({ hasText: "Client" }).first();
  await clientItem.getByRole("combobox").click();
  await page.getByText(clientEmail, { exact: true }).last().click();
  await page.getByRole("button", { name: "Issue token", exact: true }).click();
  await page
    .getByText("Copy this token now. It will not be shown again.", { exact: true })
    .waitFor({ timeout: 10_000 });
}

async function exerciseLocale(page) {
  await page.context().addCookies([{ name: "lang", value: "ru-RU", url: baseUrl }]);
  await openRoute(page, "/settings#catx-features", "settings-ru", "Функции CatX-UI");
  await requireText(page, "Состояние выполнения", "Russian runtime-state label");
  await openRoute(page, "/policies", "policy-ru", "Движок политик");
  await saveScreenshot(page, "policy-ru");
  await page.context().addCookies([{ name: "lang", value: "en-US", url: baseUrl }]);
}

const browser = await chromium.launch({
  headless: true,
  args: process.getuid?.() === 0 ? ["--no-sandbox", "--disable-setuid-sandbox"] : [],
});

try {
  for (const [name, viewport] of [
    ["desktop", { width: 1920, height: 1080 }],
    ["compact", { width: 1366, height: 768 }],
  ]) {
    evidencePrefix = name;
    const context = await browser.newContext({ viewport });
    const page = await context.newPage();
    page.on("console", (message) => {
      if (message.type() === "error")
        consoleErrors.push(`${name} console.error: ${message.text()}`);
    });
    page.on("pageerror", (error) => consoleErrors.push(`${name} pageerror: ${error.message}`));
    await login(page);
    await openRoute(page, "/", "dashboard", "Overview");
    await toggleSettingsAndRestore(page);
    await openRoute(page, "/activity", "analytics", "Client activity");
    await saveScreenshot(page, "activity-en");
    await exercisePolicy(page);
    await exerciseClientTraffic(page);
    await exercisePortal(page);
    await openRoute(page, "/catx/sponsors", "sponsors", "Sponsors");
    await exerciseLocale(page);
    await context.close();
  }
} finally {
  await fs.writeFile(
    path.join(evidenceDir, "browser-console.log"),
    `${consoleErrors.join("\n")}\n`,
    "utf8",
  );
  await browser.close();
}

if (consoleErrors.length > 0) {
  throw new Error(`browser console qualification failed:\n${consoleErrors.join("\n")}`);
}
