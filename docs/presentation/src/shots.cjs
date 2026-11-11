// Снимки экранов интерфейса со стенда http://localhost для слайда «Разработанные интерфейсы».
// Окно 1440×810 (на меньшей ширине вёрстка сжимается) с плотностью 1,5 — картинка 2160×1215.
// Номера проектов и узла — из демонстрационных данных стенда (task stack:seed).
// node shots.cjs <каталог> [имя ...]
const { chromium } = require('/home/michael/Projects/game-developer-hub/src/frontend/node_modules/playwright-core');
const out = process.argv[2], only = process.argv.slice(3);
const BASE = 'http://localhost';
const USERS = { dev: 'dev@welwise.com', mod: 'moderator@welwise.com', adm: 'admin@welwise.com' };
// [файл, роль, адрес, действия перед снимком]
const SHOTS = [
  ['ui-draft', 'dev', '/projects/5/draft'],
  ['ui-stats', 'dev', '/projects/12/stats'],
  ['ui-servers', 'dev', '/projects/14/servers'],
  ['ui-node', 'dev', '/nodes/1'],
  ['ui-queue', 'mod', '/moderator/queue'],
  ['ui-review', 'mod', '/moderator/projects/14', async (p) => {
    await p.getByText('Песочница', { exact: true }).first().click().catch(() => {});
    await p.waitForTimeout(2500);
  }],
  ['ui-catalog', 'mod', '/moderator/catalog'],
  ['ui-admin', 'adm', '/admin/developers'],
];
(async () => {
  const b = await chromium.launch();
  const pages = {};
  const login = async (role) => {
    if (pages[role]) return pages[role];
    const ctx = await b.newContext({ viewport: { width: 1440, height: 810 }, deviceScaleFactor: 1.5, colorScheme: 'light', locale: 'ru-RU' });
    const p = await ctx.newPage();
    await p.goto(BASE + '/login', { waitUntil: 'networkidle' });
    await p.evaluate(() => localStorage.setItem('theme', 'light'));
    await p.fill('input[type=email]', USERS[role]);
    await p.fill('input[type=password]', 'password123');
    await p.getByRole('button', { name: 'Войти' }).click();
    await p.waitForURL((u) => !u.pathname.startsWith('/login'), { timeout: 15000 });
    return (pages[role] = p);
  };
  for (const [name, role, url, act] of SHOTS) {
    if (only.length && !only.includes(name)) continue;
    const p = await login(role);
    await p.goto(BASE + url, { waitUntil: 'networkidle' });
    await p.waitForTimeout(1500);
    if (act) await act(p);
    await p.screenshot({ path: `${out}/${name}.jpg`, type: 'jpeg', quality: 88 });
    console.log(name, p.url());
  }
  await b.close();
})().catch((e) => { console.error(e); process.exit(1); });
