import { test, expect, type Page } from '@playwright/test';
import type { User, Topic, Post, Water } from '../src/api/types';

const topicId = '11111111-1111-4111-8111-111111111111';
const categoryId = '22222222-2222-4222-8222-222222222222';
const user: User = {
  id: '33333333-3333-4333-8333-333333333333', username: 'РыболовПодмосковья', display_name: 'РыболовПодмосковья',
  avatar_url: null, region: null, bio: null, joined_at: '2026-10-20T10:00:00Z',
};
const topic: Topic = {
  id: topicId, category_id: categoryId, author_id: user.id, title: 'Обсуждаем осеннюю рыбалку',
  slug: 'autumn', body: 'Первая строка.\nВторая строка.', type: 'discussion', pinned: false, locked: false,
  created_at: '2026-10-20T10:00:00Z', last_activity_at: '2026-10-20T11:00:00Z',
  replies_count: 3, views_count: 10, author_name: user.display_name, category_name: 'Общие вопросы',
};
const post = (id: number, body: string): Post => ({ id: String(id), topic_id: topicId, author_id: user.id,
  body, created_at: '2026-10-20T11:00:00Z', edited_at: null, parent_post_id: null, author_name: user.display_name });
const water = (id: number): Water => ({ id: String(id), name: 'Озеро ' + id, slug: 'water-' + id, kind: 'lake',
  region_id: null, description: 'Иллюстративная карточка', cover_url: null });

async function mockApi(page: Page) {
  const state = { authenticated: false, registered: false, failFeed: false, waterLimits: [] as string[], replyBody: '', profile: { ...user } as User };
  const avatarPNG = Buffer.from('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/mfoAAAAASUVORK5CYII=', 'base64');
  await page.route('https://fonts.googleapis.com/**', (route) => route.abort());
  await page.route('https://fonts.gstatic.com/**', (route) => route.abort());
  await page.route('**/api/v1/**', async (route) => {
    const request = route.request();
    const url = new URL(request.url());
    const path = url.pathname.replace('/api/v1', '');
    if (path === '/auth/me') return route.fulfill(state.authenticated ? { json: state.profile } :
      { status: 401, json: { error: { code: 'UNAUTHORIZED', message: 'unauthorized', request_id: 'test' } } });
    if (path === '/auth/register') { state.registered = true; return route.fulfill({ status: 202, body: '' }); }
    if (path === '/auth/register/verify') { state.authenticated = true; return route.fulfill({ status: 204 }); }
    if (path === '/auth/login') { state.authenticated = true; return route.fulfill({ status: 204 }); }
    if (path === '/auth/logout') { state.authenticated = false; return route.fulfill({ status: 204 }); }
    if (path === '/auth/avatar' && request.method() === 'PUT') {
      state.profile = { ...state.profile, avatar_url: `/api/v1/users/${user.id}/avatar?v=1` };
      return route.fulfill({ json: state.profile });
    }
    if (path === `/users/${user.id}/avatar`) return route.fulfill({ contentType: 'image/png', body: avatarPNG });
    if (path === '/forum/categories') return route.fulfill({ json: [{ id: categoryId, name: 'Общие вопросы', slug: 'general', parent_id: null }] });
    if (path === '/feed') return route.fulfill(state.failFeed ?
      { status: 503, json: { error: { code: 'INTERNAL', message: 'unavailable', request_id: 'test' } } } :
      { json: { items: [{ kind: 'topic', topic, report: null }], next_cursor: null } });
    if (path === '/forum/topics') return route.fulfill({ json: { items: [topic], next_cursor: null } });
    if (path === '/forum/topics/' + topicId) return route.fulfill({ json: { ...topic, replies_count: state.replyBody ? 4 : 3 } });
    if (path === '/forum/topics/' + topicId + '/posts') {
      if (request.method() === 'POST') {
        state.replyBody = request.postDataJSON().body;
        return route.fulfill({ status: 201, json: post(4, state.replyBody) });
      }
      return route.fulfill({ json: url.searchParams.has('cursor') ?
        { items: [post(3, 'Третий ответ'), ...(state.replyBody ? [post(4, state.replyBody)] : [])], next_cursor: null } :
        { items: [post(1, 'Первый ответ'), post(2, 'Второй ответ')], next_cursor: 'posts-page-2' } });
    }
    if (path === '/waters') {
      const limit = url.searchParams.get('limit') ?? '20'; state.waterLimits.push(limit);
      return route.fulfill({ json: limit === '20' ? url.searchParams.has('cursor') ?
        { items: [water(22)], next_cursor: null } : { items: [water(20), water(21)], next_cursor: 'waters-page-2' } :
        { items: [water(Number(limit))], next_cursor: null } });
    }
    if (path.startsWith('/waters/')) return route.fulfill({ json: water(Number(path.split('/').at(-1))) });
    if (path === '/reports') return route.fulfill({ json: { items: [], next_cursor: null } });
    if (path === '/search') return route.fulfill({ json: { topics: [], reports: [], waters: [water(20)], users: [] } });
    if (path === '/notifications') return route.fulfill({ json: { items: [], next_cursor: null } });
    if (path === '/follows') return route.fulfill({ status: 204 });
    return route.fulfill({ status: 404, json: { error: { code: 'NOT_FOUND', message: 'not found', request_id: 'test' } } });
  });
  return state;
}

test('profile avatar upload updates the card and mobile create button stays centered', async ({ page }, testInfo) => {
  const state = await mockApi(page); state.authenticated = true;
  await page.goto('/profile');
  await expect(page.getByRole('heading', { name: user.display_name, exact: true })).toBeVisible();
  const createButton = page.getByRole('button', { name: 'Создать', exact: true });
  if (testInfo.project.name === 'mobile') {
    const box = await createButton.boundingBox();
    const icon = await createButton.locator('svg').boundingBox();
    expect(box && icon && Math.abs(box.x + box.width / 2 - icon.x - icon.width / 2)).toBeLessThan(1);
    expect(box && icon && Math.abs(box.y + box.height / 2 - icon.y - icon.height / 2)).toBeLessThan(1);
    expect(await createButton.evaluate((element) => getComputedStyle(element).borderRadius)).toBe('50%');
  }
  const fileChooserPromise = page.waitForEvent('filechooser');
  await page.getByRole('button', { name: 'Загрузить аватар' }).click();
  const fileChooser = await fileChooserPromise;
  await fileChooser.setFiles({ name: 'avatar.png', mimeType: 'image/png',
    buffer: Buffer.from('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/mfoAAAAASUVORK5CYII=', 'base64') });
  await expect(page.getByRole('status')).toHaveText('Аватар обновлён.');
  await expect(page.locator('.profile-avatar img')).toBeVisible();
  await expect(page.locator('.account img')).toBeVisible();
  expect(await page.locator('.profile-avatar img').evaluate((img) => (img as HTMLImageElement).naturalWidth)).toBeGreaterThan(0);
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  await page.screenshot({ path: testInfo.outputPath('profile-avatar.png'), fullPage: true });
});

test('server failure is shown with retry without a false empty feed', async ({ page }) => {
  const state = await mockApi(page); state.failFeed = true;
  await page.goto('/');
  await expect(page.getByRole('alert')).toContainText('Не удалось загрузить данные');
  await expect(page.getByText('Пока нет публикаций. Создайте первую тему.')).toHaveCount(0);
  state.failFeed = false;
  await page.getByRole('button', { name: 'Повторить', exact: true }).click();
  await expect(page.getByRole('button', { name: topic.title, exact: true })).toBeVisible();
});

test('catalog has its own limit and loads a second page without losing the first', async ({ page }) => {
  const state = await mockApi(page);
  await page.goto('/');
  await expect(page.getByRole('button', { name: topic.title, exact: true })).toBeVisible();
  await page.getByRole('link', { name: 'Все водоёмы' }).click();
  await expect(page.getByRole('heading', { name: 'Озеро 20', exact: true })).toBeVisible();
  expect(state.waterLimits).toContain('4'); expect(state.waterLimits).toContain('6'); expect(state.waterLimits).toContain('20');
  await page.getByRole('button', { name: 'Показать ещё', exact: true }).click();
  await expect(page.getByRole('heading', { name: 'Озеро 22', exact: true })).toBeVisible();
  await expect(page.getByRole('heading', { name: 'Озеро 20', exact: true })).toHaveCount(1);
});

test('modal traps focus, closes with Escape, and restores the trigger', async ({ page }) => {
  await mockApi(page); await page.goto('/profile');
  const trigger = page.getByRole('button', { name: 'Войти / Регистрация' });
  await trigger.click();
  const dialog = page.getByRole('dialog', { name: 'Вход', exact: true });
  await expect(dialog.getByLabel('Email')).toBeFocused();
  for (let i = 0; i < 8; i++) {
    await page.keyboard.press('Tab');
    expect(await page.evaluate(() => document.activeElement?.closest('dialog') !== null)).toBe(true);
  }
  await page.keyboard.press('Escape');
  await expect(dialog).toHaveCount(0); await expect(trigger).toBeFocused();
});

test('registration verifies email and opens an authenticated session', async ({ page }, testInfo) => {
  const state = await mockApi(page); await page.goto('/profile');
  await page.getByRole('button', { name: 'Войти / Регистрация' }).click();
  const dialog = page.getByRole('dialog');
  await dialog.getByRole('button', { name: 'Регистрация', exact: true }).click();
  await dialog.getByLabel('Email').fill('fisher@example.test');
  await dialog.getByLabel('Ник (3–32 символа)').fill(user.username);
  await dialog.getByLabel('Пароль (12–128 символов)').fill('valid-password-123');
  await dialog.getByLabel('Повторите пароль').fill('valid-password-123');
  await page.keyboard.press('Enter');
  await expect(dialog.getByRole('status')).toContainText('Мы отправили шестизначный код');
  expect(state.registered).toBe(true); expect(state.authenticated).toBe(false);
  await expect(dialog.getByLabel('Код подтверждения')).toBeFocused();
  const confirmButton = dialog.getByRole('button', { name: 'Подтвердить и войти' });
  const restartButton = dialog.getByRole('button', { name: 'Изменить данные или отправить код снова' });
  const confirmBox = await confirmButton.boundingBox();
  const restartBox = await restartButton.boundingBox();
  expect(confirmBox && restartBox && restartBox.y - confirmBox.y - confirmBox.height).toBeGreaterThanOrEqual(10);
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
  await page.screenshot({ path: testInfo.outputPath('verification-dialog.png') });
  await dialog.getByLabel('Код подтверждения').fill('123456');
  await page.keyboard.press('Enter');
  await expect(dialog).toHaveCount(0);
  await expect(page.getByRole('heading', { name: user.display_name, exact: true })).toBeVisible();
});

test('search works by keyboard at both viewport sizes and does not overflow', async ({ page }, testInfo) => {
  await mockApi(page); await page.goto('/');
  await expect(page.getByRole('button', { name: topic.title, exact: true })).toBeVisible();
  expect(await page.locator('.app').evaluate((element) => parseFloat(getComputedStyle(element).fontSize))).toBeGreaterThanOrEqual(15);
  if (testInfo.project.name === 'mobile') {
    expect(await page.locator('.bottom-nav a').first().evaluate((element) => parseFloat(getComputedStyle(element).fontSize))).toBeGreaterThanOrEqual(12);
  }
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
  await page.screenshot({ path: testInfo.outputPath('home-layout.png'), fullPage: true });
  const toggle = page.getByRole('button', { name: 'Открыть поиск' });
  if (await toggle.isVisible()) await toggle.click();
  const input = page.getByRole('textbox', { name: 'Поиск', exact: true });
  if (testInfo.project.name === 'mobile') await expect(input).toBeFocused();
  await input.fill('озеро'); await input.press('Enter');
  await expect(page).toHaveURL(/\/search\?q=/);
  await expect(page.getByRole('link', { name: 'Озеро 20 · Озеро' })).toBeVisible();
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
  await page.screenshot({ path: testInfo.outputPath('search-layout.png'), fullPage: true });
});

test('topic links survive navigation and all reply pages remain available', async ({ page }) => {
  const state = await mockApi(page); state.authenticated = true;
  await page.goto('/');
  await page.getByRole('button', { name: topic.title, exact: true }).click();
  await expect(page).toHaveURL('/forum/topics/' + topicId);
  await expect(page.getByText('Первый ответ', { exact: true })).toBeVisible();
  await page.getByRole('button', { name: 'Показать ещё', exact: true }).click();
  await expect(page.getByText('Третий ответ', { exact: true })).toBeVisible();
  await page.getByLabel('Ваш ответ', { exact: true }).fill('Новый опубликованный ответ');
  await page.getByRole('button', { name: 'Ответить', exact: true }).click();
  await expect(page.getByRole('status')).toContainText('Ответ опубликован.');
  await expect(page.getByText('Новый опубликованный ответ', { exact: true })).toBeVisible();
  await page.reload();
  await expect(page.getByRole('heading', { name: topic.title, exact: true })).toBeVisible();
});

test('water loading state is distinct from a confirmed 404', async ({ page }) => {
  await mockApi(page);
  let release!: () => void;
  const pending = new Promise<void>((resolve) => { release = resolve; });
  await page.route('**/api/v1/waters/missing', async (route) => {
    await pending;
    await route.fulfill({ status: 404, json: { error: { code: 'NOT_FOUND', message: 'not found', request_id: 'test' } } });
  });
  await page.goto('/waters/missing');
  await expect(page.getByRole('status')).toContainText('Загрузка водоёма');
  await expect(page.getByRole('heading', { name: 'Водоём не найден' })).toHaveCount(0);
  release();
  await expect(page.getByRole('heading', { name: 'Водоём не найден' })).toBeVisible();
});
