import { chromium } from 'playwright';

const baseURL = process.argv[2];
if (!baseURL) {
  throw new Error('usage: node observe-public-locales.mjs BASE_URL');
}

const cases = [
  { language: 'ru-RU', direction: 'ltr', login: 'Войти', username: 'Имя пользователя' },
  { language: 'fa-IR', direction: 'rtl', login: 'ورود', username: 'نام‌کاربری' },
];

const browser = await chromium.launch({ headless: true, args: ['--no-sandbox'] });
try {
  for (const testCase of cases) {
    const context = await browser.newContext();
    await context.addCookies([
      {
        name: 'lang',
        value: testCase.language,
        url: `${new URL(baseURL).origin}/`,
      },
    ]);
    const page = await context.newPage();
    await page.goto(`${baseURL.replace(/\/$/, '')}/`, { waitUntil: 'networkidle' });
    await page.getByRole('button', { name: testCase.login, exact: true }).waitFor();
    await page.getByLabel(testCase.username, { exact: true }).waitFor();
    const direction = await page.evaluate(() => ({
      html: document.documentElement.dir,
      body: document.body.dir,
    }));
    if (direction.html !== testCase.direction || direction.body !== testCase.direction) {
      throw new Error(
        `${testCase.language} direction html=${direction.html} body=${direction.body}; expected ${testCase.direction}`,
      );
    }
    console.log(
      `public locale probe: PASS language=${testCase.language} direction=${testCase.direction}`,
    );
    await context.close();
  }
} finally {
  await browser.close();
}
