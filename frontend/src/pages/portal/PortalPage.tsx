import { useCallback, useEffect, useMemo, useState } from 'react';
import { Alert, Button, Card, Descriptions, Input, Space, Table, Tabs, Typography } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { useTranslation } from 'react-i18next';
import { HttpUtil } from '@/utils';

type PortalClient = { id: number; email: string; group: string; enabled: boolean; totalGB: number; expiryTime: number; limitHwid: number };
type Device = { id: number; firstSeen: number; lastSeen: number; deviceName: string; deviceOs: string; osVersion: string; deviceModel: string };
type Host = { id: number; inboundId: number; remark: string; address: string; port: number; security: string; sni?: string; hostHeader?: string; path?: string; alpn?: string[]; fingerprint?: string };
type Traffic = { up: number; down: number; total: number; expiryTime: number };

function portalPath(path: string): string {
  const base = (window as Window & { X_UI_BASE_PATH?: string }).X_UI_BASE_PATH || '/';
  return `${base.replace(/\/+$/, '')}${path}`;
}

export default function PortalPage() {
  const { t } = useTranslation();
  const [token, setToken] = useState('');
  const [csrf, setCsrf] = useState('');
  const [client, setClient] = useState<PortalClient | null>(null);
  const [devices, setDevices] = useState<Device[]>([]);
  const [hosts, setHosts] = useState<Host[]>([]);
  const [traffic, setTraffic] = useState<Traffic | null>(null);
  const [error, setError] = useState('');

  const load = useCallback(async () => {
    const [me, deviceResult, hostResult, trafficResult, csrfResult] = await Promise.all([
      HttpUtil.get<PortalClient>(portalPath('/portal/me'), undefined, { silent: true }),
      HttpUtil.get<Device[]>(portalPath('/portal/devices'), undefined, { silent: true }),
      HttpUtil.get<Host[]>(portalPath('/portal/hosts'), undefined, { silent: true }),
      HttpUtil.get<Traffic>(portalPath('/portal/traffic'), undefined, { silent: true }),
      HttpUtil.get<string>(portalPath('/portal/csrf'), undefined, { silent: true }),
    ]);
    if (!me.success) { setClient(null); setError(me.msg || t('fork.portal.loginRequired', 'Enter your portal access token.')); return; }
    setClient(me.obj); setDevices(deviceResult.obj || []); setHosts(hostResult.obj || []); setTraffic(trafficResult.obj); setCsrf(csrfResult.obj || ''); setError('');
  }, [t]);

  useEffect(() => { void load(); }, [load]);

  async function login() {
    const result = await HttpUtil.post<{ csrfToken: string }>(portalPath('/portal/auth'), { token }, { silent: true });
    if (!result.success) { setError(result.msg || t('fork.portal.invalidToken', 'Invalid portal token.')); return; }
    setCsrf(result.obj?.csrfToken || ''); await load();
  }

  async function mutate(path: string, method: 'patch' | 'delete', body?: unknown) {
    const options = { headers: { 'X-CSRF-Token': csrf }, silent: true };
    const result = method === 'patch' ? await HttpUtil.put(portalPath(path), body, options) : await HttpUtil.delete(portalPath(path), options);
    if (!result.success) setError(result.msg || t('fork.portal.actionFailed', 'Action failed.')); else await load();
  }

  const columns = useMemo<ColumnsType<Device>>(() => [
    { title: t('fork.portal.device', 'Device'), render: (_, row) => row.deviceName || row.deviceModel || row.deviceOs || '—' },
    { title: t('fork.portal.lastSeen', 'Last seen'), render: (_, row) => row.lastSeen ? new Date(row.lastSeen).toLocaleString() : '—' },
    { title: t('fork.portal.actions', 'Actions'), render: (_, row) => <Space><Button onClick={() => { const name = window.prompt(t('fork.portal.renamePrompt', 'New device name'), row.deviceName); if (name != null) void mutate(`/portal/devices/${row.id}`, 'patch', { deviceName: name }); }}>{t('rename', 'Rename')}</Button><Button danger onClick={() => void mutate(`/portal/devices/${row.id}`, 'delete')}>{t('delete', 'Revoke')}</Button></Space> },
  ], [csrf, t]);

  if (!client) return <Card style={{ maxWidth: 520, margin: '8rem auto' }}><Space direction="vertical" style={{ width: '100%' }}><Typography.Title level={2}>{t('fork.portal.title', 'Client portal')}</Typography.Title><Typography.Paragraph>{t('fork.portal.tokenHint', 'Use the portal access token provided by your administrator.')}</Typography.Paragraph><Input.Password value={token} onChange={(e) => setToken(e.target.value)} onPressEnter={() => void login()} placeholder={t('fork.portal.token', 'Portal access token')} /><Button type="primary" onClick={() => void login()}>{t('fork.portal.signIn', 'Sign in')}</Button>{error && <Alert type="error" message={error} />}</Space></Card>;

  return <Card style={{ margin: 24 }}><Space direction="vertical" style={{ width: '100%' }} size="large"><div><Typography.Title level={2}>{t('fork.portal.title', 'Client portal')}</Typography.Title>{error && <Alert type="error" message={error} />}</div><Descriptions bordered items={[{ key: 'email', label: t('email', 'Email'), children: client.email }, { key: 'status', label: t('status', 'Status'), children: client.enabled ? t('enabled', 'Enabled') : t('disabled', 'Disabled') }, { key: 'quota', label: t('fork.portal.quota', 'Quota'), children: traffic ? `${traffic.up + traffic.down} / ${traffic.total || client.totalGB}` : '—' }, { key: 'expiry', label: t('fork.portal.expiry', 'Expiry'), children: client.expiryTime ? new Date(client.expiryTime).toLocaleString() : t('never', 'Never') }]} /><Tabs items={[{ key: 'devices', label: t('fork.portal.devices', 'Devices'), children: <Table rowKey="id" columns={columns} dataSource={devices} pagination={false} /> }, { key: 'hosts', label: t('fork.portal.hosts', 'Connection hosts'), children: <Table rowKey="id" columns={[{ title: t('fork.portal.host', 'Host'), dataIndex: 'address' }, { title: t('port', 'Port'), dataIndex: 'port' }, { title: t('security', 'Security'), dataIndex: 'security' }, { title: t('sni', 'SNI'), dataIndex: 'sni' }]} dataSource={hosts} pagination={false} /> }]} /></Space></Card>;
}
