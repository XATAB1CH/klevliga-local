import { test, expect, type Page } from '@playwright/test';
import type { AuditEvent, ManagedUser, Role, SessionUser } from '../src/api/types';

const adminID = '11111111-1111-4111-8111-111111111111';
const targetID = '22222222-2222-4222-8222-222222222222';
const base = { avatar_url: null, region: null, bio: null, joined_at: '2026-11-01T12:00:00Z' };

async function mockAccess(page: Page, isAdmin = true) {
  const actor: SessionUser = { ...base, id: adminID, username: 'anton', display_name: 'Антон',
    roles: isAdmin ? ['user', 'admin'] : ['user'], permissions: isAdmin ? ['users.manage', 'audit.view', 'content.moderate', 'support.handle'] : [] };
  const target: ManagedUser = { ...base, id: targetID, username: 'angler', display_name: 'Рыболов', roles: ['user'], status: 'active' };
  const events: AuditEvent[] = [];
  const calls: { roles: Role[]; reason: string }[] = [];
  let adminReads = 0;
  await page.route('https://fonts.googleapis.com/**', (route) => route.abort());
  await page.route('https://fonts.gstatic.com/**', (route) => route.abort());
  await page.route('**/api/v1/**', async (route) => {
    const request = route.request();
    const url = new URL(request.url());
    const path = url.pathname.replace('/api/v1', '');
    if (path === '/auth/me') return route.fulfill({ json: actor });
    if (path.startsWith('/admin/')) adminReads++;
    if (path === '/admin/users') return route.fulfill({ json: { items: [target] } });
    if (path === `/admin/users/${targetID}`) return route.fulfill({ json: target });
    if (path === `/admin/users/${targetID}/roles`) {
      const body = request.postDataJSON() as { roles: Role[]; reason: string };
      calls.push(body);
      events.unshift({ id: 'event-' + calls.length, actor_type: 'user', actor_id: adminID,
        action: 'user.roles_changed', target_type: 'user', target_id: targetID, reason: body.reason,
        request_id: 'test-request', changes: { roles: { before: [...target.roles], after: body.roles } }, created_at: '2026-11-07T15:00:00Z' });
      target.roles = body.roles;
      return route.fulfill({ status: 204 });
    }
    if (path === `/admin/users/${targetID}/status`) {
      target.status = request.postDataJSON().status;
      return route.fulfill({ status: 204 });
    }
    if (path === '/admin/audit-events') return route.fulfill({ json: { items: events, next_cursor: null } });
    if (path === '/forum/categories') return route.fulfill({ json: [] });
    return route.fulfill({ json: { items: [], next_cursor: null } });
  });
  return { calls, actor, target, getAdminReads: () => adminReads };
}

test('administrator manages roles and reads the audit on desktop and mobile', async ({ page }, testInfo) => {
  const state = await mockAccess(page);
  await page.goto('/profile');
  await page.getByRole('link', { name: /Управление сообществом/ }).click();
  await expect(page.getByRole('heading', { name: 'Управление сообществом' })).toBeVisible();
  await page.getByLabel('Ник или имя').fill('angler');
  await page.getByRole('region', { name: 'Участники' }).getByRole('button', { name: 'Найти', exact: true }).click();
  await page.getByRole('button', { name: /Рыболов @angler/ }).click();
  await page.getByLabel('Поддержка', { exact: true }).check();
  await page.getByLabel('Модератор', { exact: true }).check();
  await expect(page.getByLabel('Пользователь', { exact: true })).toBeDisabled();
  await page.getByLabel('Причина изменения').fill('Помогает участникам сообщества');
  await page.getByRole('button', { name: 'Сохранить роли' }).focus();
  await page.keyboard.press('Enter');
  await expect(page.getByLabel('Поддержка', { exact: true })).toBeChecked();
  await expect(page.locator('.audit-list').getByText('Изменены роли', { exact: true })).toBeVisible();
  await expect(page.getByRole('status')).toContainText('Изменение сохранено в журнале.');
  expect(state.calls).toEqual([{ roles: ['user', 'support', 'moderator'], reason: 'Помогает участникам сообщества' }]);
  await page.getByLabel('Причина изменения').fill('Проверка блокировки аккаунта');
  await page.getByRole('button', { name: 'Заблокировать аккаунт', exact: true }).click();
  await expect(page.getByRole('button', { name: 'Разблокировать аккаунт' })).toBeVisible();
  await page.getByLabel('Причина изменения').fill('Возвращаем доступ участнику');
  await page.getByRole('button', { name: 'Разблокировать аккаунт' }).click();
  await expect(page.getByRole('button', { name: 'Заблокировать аккаунт', exact: true })).toBeVisible();
  await expect(page.locator('body')).toHaveJSProperty('scrollWidth', await page.locator('body').evaluate((body) => body.clientWidth));
  await page.evaluate(() => window.scrollTo({ top: 0, behavior: 'instant' }));
  await page.screenshot({ path: testInfo.outputPath('community-access.png'), fullPage: true });
});

test('ordinary user cannot enter community management or load its data', async ({ page }) => {
  const state = await mockAccess(page, false);
  await page.goto('/admin');
  await expect(page.getByRole('heading', { name: 'Доступ ограничен' })).toBeVisible();
  expect(state.getAdminReads()).toBe(0);
  await page.goto('/profile');
  await expect(page.getByRole('link', { name: /Управление сообществом/ })).toHaveCount(0);
});

test('last administrator error is understandable and preserves selected roles', async ({ page }) => {
  await mockAccess(page);
  await page.route(`**/api/v1/admin/users/${targetID}/roles`, (route) => route.fulfill({ status: 409,
    json: { error: { code: 'LAST_ADMIN_REQUIRED', message: 'at least one active administrator must remain', request_id: 'test' } } }));
  await page.goto('/admin');
  await page.getByLabel('Ник или имя').fill('angler');
  await page.getByRole('region', { name: 'Участники' }).getByRole('button', { name: 'Найти', exact: true }).click();
  await page.getByRole('button', { name: /Рыболов @angler/ }).click();
  await page.getByLabel('Администратор', { exact: true }).check();
  await page.getByLabel('Причина изменения').fill('Проверка ошибки защиты');
  await page.getByRole('button', { name: 'Сохранить роли' }).click();
  await expect(page.getByRole('alert')).toContainText('последнего активного администратора');
  await expect(page.getByLabel('Администратор', { exact: true })).toBeChecked();
});
