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
const unconfiguredClientEmail = required("CATX_HARDENING_UNCONFIGURED_CLIENT_EMAIL");
const policyName = required("CATX_HARDENING_POLICY_NAME");
const evidenceDir = required("CATX_HARDENING_EVIDENCE_DIR");
const consoleErrors = [];
const httpErrors = [];
let evidencePrefix = "";

await fs.mkdir(evidenceDir, { recursive: true });

const appUrl = (route) => `${baseUrl}/panel${route.startsWith("/") ? route : `/${route}`}`;
const waitForRender = async (page) => {
  await page.waitForLoadState("domcontentloaded");
  await page.waitForTimeout(700);
};
const bodyText = async (page) => (await page.locator("body").innerText()).replace(/\s+/g, " ");
const requireText = async (page, text, label) => {
  try {
    await page.waitForFunction(
      (expected) => document.body.innerText.replace(/\s+/g, " ").includes(expected),
      text,
      { timeout: 10_000 },
    );
  } catch {
    throw new Error(`${label}: missing visible text ${JSON.stringify(text)}`);
  }
};
const closeModal = async (modal) => {
  const closeButton = modal.locator(".ant-modal-close");
  if ((await closeButton.count()) > 0) {
    await closeButton.click();
  } else {
    await modal.getByRole("button", { name: "Cancel", exact: true }).click();
  }
  await modal.waitFor({ state: "hidden", timeout: 10_000 });
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

async function openClientAction(page, email, action) {
  const row = page.locator("tr").filter({ hasText: email }).first();
  if ((await row.count()) > 0) {
    await row.getByRole("button", { name: action, exact: true }).click();
    return;
  }
  const card = page.locator(".client-card").filter({ hasText: email }).first();
  await card.waitFor({ state: "visible", timeout: 10_000 });
  if (action === "Edit") {
    await card.getByRole("button", { name: "More", exact: true }).click();
    await page.getByText("Edit", { exact: true }).last().click();
    return;
  }
  await card.getByRole("button", { name: action, exact: true }).click();
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
  const modalLayout = await modal.evaluate((node) => {
    const body = node.querySelector(".ant-modal-body");
    return {
      width: node.getBoundingClientRect().width,
      bodyScrollWidth: body?.scrollWidth ?? 0,
      bodyClientWidth: body?.clientWidth ?? 0,
    };
  });
  if (modalLayout.bodyScrollWidth > modalLayout.bodyClientWidth + 1)
    throw new Error(`policy modal: horizontal overflow ${JSON.stringify(modalLayout)}`);
  if (modalLayout.width > page.viewportSize().width - 16)
    throw new Error(`policy modal: viewport overflow ${JSON.stringify(modalLayout)}`);
  if ((await modal.getByRole("button", { name: "Cancel", exact: true }).count()) !== 1)
    throw new Error("policy modal: Cancel was not localized");
  await modal.getByLabel("Name").fill(currentPolicyName);
  await modal.getByLabel("Description").fill("Structured hardening qualification policy");
  const actionItem = modal.locator(".ant-form-item").filter({ hasText: "Action" }).first();
  const actionSelect = actionItem.getByRole("combobox");
  await actionSelect.click();
  const allowOption = page
    .locator(".ant-select-item-option")
    .filter({ hasText: /^Allow$/ })
    .last();
  await allowOption.waitFor({ state: "attached" });
  await allowOption.evaluate((option) => option.click());
  const selectedAction = actionItem.locator(".ant-select-content");
  await selectedAction.waitFor({ state: "visible" });
  if ((await selectedAction.innerText()).trim() !== "Allow")
    throw new Error("policy create: Allow action was not selected");
  const destinationItem = modal
    .locator(".ant-form-item")
    .filter({ hasText: "Destinations" })
    .first();
  await destinationItem.locator("input").fill("example.com");
  await destinationItem.locator("input").press("Enter");
  await requireText(page, "Advanced policy fields (JSON)", "policy structured editor");
  const policyResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response.url().includes("/panel/api/policies") &&
      response.url().endsWith("/policies"),
  );
  await modal.getByRole("button", { name: "Save", exact: true }).click();
  const savedPolicyResponse = await policyResponse;
  if (!savedPolicyResponse.ok()) {
    const responseBody = (await savedPolicyResponse.text()).slice(0, 500);
    throw new Error(
      `policy create failed: HTTP ${savedPolicyResponse.status()} body=${responseBody}`,
    );
  }
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
  await closeModal(persistedModal);
  const deleteButton = page.getByRole("button", {
    name: new RegExp(`Delete.*${currentPolicyName}`),
  });
  await deleteButton.click();
  await page.getByRole("button", { name: "OK", exact: true }).click();
  await page.getByText(currentPolicyName, { exact: true }).waitFor({ state: "detached" });
  await saveScreenshot(page, "policy-en");
}

async function exerciseClientTraffic(page, { exerciseUnconfigured = true } = {}) {
  await openRoute(
    page,
    `/clients?search=${encodeURIComponent(clientEmail)}`,
    "clients",
    clientEmail,
  );
  await openClientAction(page, clientEmail, "Edit");
  const editModal = page.locator(".ant-modal").last();
  await editModal.getByRole("tab", { name: "Traffic control", exact: true }).click();
  await requireText(
    page,
    "Rate/speed shaping is unsupported for generic Xray users.",
    "client edit traffic enforcement state",
  );
  const trafficPane = editModal.locator('[role="tabpanel"][aria-hidden="false"]');
  const trafficNumbers = trafficPane.getByRole("spinbutton");
  if ((await trafficNumbers.count()) < 2) throw new Error("client edit traffic: controls missing");
  await trafficNumbers.nth(1).fill("1048576");
  const trafficSaveResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "PUT" &&
      response.url().includes("/panel/api/traffic-control/clients/") &&
      response.url().endsWith("/policy") &&
      response.ok(),
  );
  await trafficPane.getByRole("button", { name: "Save", exact: true }).click();
  const savedTraffic = await trafficSaveResponse;
  const savedTrafficBody = await savedTraffic.json();
  if (savedTrafficBody.obj?.clientEmail !== clientEmail)
    throw new Error(`client edit traffic: save targeted ${savedTrafficBody.obj?.clientEmail}`);
  await closeModal(editModal);
  await openRoute(
    page,
    `/clients?search=${encodeURIComponent(clientEmail)}`,
    "clients reopen",
    clientEmail,
  );
  await openClientAction(page, clientEmail, "Edit");
  const reopenedModal = page.locator(".ant-modal").last();
  await reopenedModal.getByRole("tab", { name: "Traffic control", exact: true }).click();
  const reopenedNumbers = reopenedModal
    .locator('[role="tabpanel"][aria-hidden="false"]')
    .getByRole("spinbutton");
  if ((await reopenedNumbers.nth(1).inputValue()) !== "1048576")
    throw new Error("client edit traffic: quota did not persist after reopen");
  await closeModal(reopenedModal);
  await openClientAction(page, clientEmail, "Client Information");
  const infoModal = page.locator(".ant-modal").last();
  await requireText(page, "Traffic control", "client information traffic state");
  if ((await infoModal.getByRole("button", { name: "Save", exact: true }).count()) > 0) {
    throw new Error("client information: mutable Save control leaked into read-only traffic view");
  }
  await saveScreenshot(page, "client-traffic-en");
  await closeModal(infoModal);

  if (!exerciseUnconfigured) return;

  await openRoute(
    page,
    `/clients?search=${encodeURIComponent(unconfiguredClientEmail)}`,
    "unconfigured client",
    unconfiguredClientEmail,
  );
  const initialUnconfigured = await page.evaluate(async (url) => {
    const response = await fetch(url);
    return response.json();
  }, appUrl(`/api/traffic-control/clients/${encodeURIComponent(unconfiguredClientEmail)}/policy`));
  if (initialUnconfigured.obj?.state !== "unconfigured")
    throw new Error(`unconfigured client precondition changed: ${JSON.stringify(initialUnconfigured)}`);
  const unconfiguredPolicyResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "GET" &&
      response.url().includes(
        `/panel/api/traffic-control/clients/${encodeURIComponent(unconfiguredClientEmail)}/policy`,
      ),
  );
  await openClientAction(page, unconfiguredClientEmail, "Edit");
  const unconfiguredModal = page.locator(".ant-modal").last();
  await unconfiguredModal.getByRole("tab", { name: "Traffic control", exact: true }).click();
  const unconfiguredPolicyBody = await (await unconfiguredPolicyResponse).text();
  try {
    await requireText(page, "Not configured", "unconfigured traffic state");
  } catch (error) {
    await fs.writeFile(
      path.join(evidenceDir, `${evidencePrefix}-unconfigured-client.html`),
      await unconfiguredModal.evaluate((node) => node.outerHTML),
      "utf8",
    );
    throw new Error(`${error.message}; traffic API=${unconfiguredPolicyBody}`);
  }
  await unconfiguredModal
    .getByRole("button", { name: "Configure traffic control", exact: true })
    .click();
  await requireText(page, "Quota and window", "traffic configure controls");
  const unconfiguredSaveResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "PUT" &&
      response.url().includes("/panel/api/traffic-control/clients/") &&
      response.url().endsWith("/policy") &&
      response.ok(),
  );
  await unconfiguredModal
    .locator('[role="tabpanel"][aria-hidden="false"]')
    .getByRole("button", { name: "Save", exact: true })
    .click();
  await unconfiguredSaveResponse;
  await closeModal(unconfiguredModal);
  await openRoute(
    page,
    `/clients?search=${encodeURIComponent(unconfiguredClientEmail)}`,
    "configured client reopen",
    unconfiguredClientEmail,
  );
  await openClientAction(page, unconfiguredClientEmail, "Edit");
  const configuredModal = page.locator(".ant-modal").last();
  await configuredModal.getByRole("tab", { name: "Traffic control", exact: true }).click();
  await requireText(page, "Quota and window", "configured traffic after UI save");
  if (
    (await configuredModal
      .locator('[role="tabpanel"][aria-hidden="false"]')
      .getByRole("spinbutton")
      .count()) < 2
  )
    throw new Error("configured traffic after UI save: controls did not persist");
  await closeModal(configuredModal);
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
  await requireText(page, "Состояние среды выполнения", "Russian runtime-state label");
  await openRoute(page, "/policies", "policy-ru", "Движок политик");
  await openRoute(page, "/fleet-updates", "fleet-updates-ru", "Обновление узлов");
  const rawMachineWords = /(^|\s)(ready|aborted|stable|local|unsupported)(\s|$)/i;
  const fleetText = await bodyText(page);
  if (rawMachineWords.test(fleetText))
    throw new Error(`fleet updates RU: untranslated machine value visible: ${fleetText}`);
  await saveScreenshot(page, "policy-ru");
  await page.context().addCookies([{ name: "lang", value: "en-US", url: baseUrl }]);
}

async function exerciseFleetAndAudit(page) {
  await openRoute(page, "/audit", "audit", "Audit");
  const auditFilter = page.getByPlaceholder("Event type");
  await auditFilter.fill("review.no-such-event");
  await page.getByRole("button", { name: "Filter", exact: true }).click();
  await requireText(page, "No audit events", "audit zero-event state");
  await auditFilter.fill("auth.login");
  await page.getByRole("button", { name: "Filter", exact: true }).click();
  await requireText(page, "auth.login", "synthetic admin audit event");
  await openRoute(page, "/fleet", "fleet", "Fleet");
  await requireText(page, "No managed nodes yet", "fleet empty state");
  await requireText(page, "Manage nodes", "fleet upstream action");
  await openRoute(page, "/fleet-updates", "fleet updates", "Fleet updates");
  await requireText(page, "Campaigns", "fleet updates workspace");
}

async function exerciseCanonicalSponsors(page) {
  await openRoute(page, "/catx/sponsors", "sponsors", "Sponsors");
  const manage = page.getByRole("button", { name: /Manage sponsors/ });
  if ((await manage.count()) !== 1) {
    await fs.writeFile(
      path.join(evidenceDir, `${evidencePrefix}-sponsors.html`),
      await page.locator("body").evaluate((node) => node.outerHTML),
      "utf8",
    );
    throw new Error(`sponsors: management action missing; body=${await bodyText(page)}`);
  }
  await manage.click();
  await page.waitForURL(/\/panel\/catx\/sponsors\/manage/);
  await requireText(page, "Sponsors management", "canonical sponsor management action");
}

const browserLaunchOptions = {
  headless: true,
  args: process.getuid?.() === 0 ? ["--no-sandbox", "--disable-setuid-sandbox"] : [],
};
if (process.env.CATX_HARDENING_CHROMIUM_EXECUTABLE_PATH) {
  browserLaunchOptions.executablePath = process.env.CATX_HARDENING_CHROMIUM_EXECUTABLE_PATH;
}
const browser = await chromium.launch(browserLaunchOptions);

try {
  for (const [name, viewport] of [
    ["desktop", { width: 1920, height: 1080 }],
    ["compact", { width: 1366, height: 768 }],
  ]) {
    evidencePrefix = name;
    const context = await browser.newContext({ viewport });
    const page = await context.newPage();
    page.on("console", (message) => {
      const location = message.location().url || "unknown";
      const expectedFavicon404 =
        message.type() === "error" &&
        /favicon\.ico(?:$|[?#])/.test(location) &&
        /404/.test(message.text());
      if (message.type() === "error" && !expectedFavicon404)
        consoleErrors.push(
          `${name} console.error: ${message.text()} @ ${location}`,
        );
    });
    page.on("pageerror", (error) => consoleErrors.push(`${name} pageerror: ${error.message}`));
    page.on("response", (response) => {
      if (response.status() >= 400)
        httpErrors.push(`${name} HTTP ${response.status()} ${response.request().method()} ${response.url()}`);
    });
    page.on("requestfailed", (request) =>
      httpErrors.push(`${name} request failed ${request.method()} ${request.url()} ${request.failure()?.errorText || "unknown"}`),
    );
    await login(page);
    await openRoute(page, "/", "dashboard", "Overview");
    await toggleSettingsAndRestore(page);
    await openRoute(page, "/activity", "analytics", "Client activity");
    await saveScreenshot(page, "activity-en");
    await exercisePolicy(page);
    await exerciseClientTraffic(page, { exerciseUnconfigured: name === "desktop" });
    await exercisePortal(page);
    await exerciseFleetAndAudit(page);
    await exerciseCanonicalSponsors(page);
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

if (consoleErrors.length > 0 || httpErrors.length > 0) {
  throw new Error(
    `browser qualification diagnostics failed:\n${[...consoleErrors, ...httpErrors].join("\n")}`,
  );
}
