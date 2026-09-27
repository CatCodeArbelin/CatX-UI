import { useEffect, useState } from 'react';
import { Alert, Button, Card, Form, Input, InputNumber, Space, Switch, Table, Typography } from 'antd';
import { HttpUtil } from '@/utils';
import { useTranslation } from 'react-i18next';

type Credential = { id: number; clientId: number; enabled: boolean; expiresAt: number; lastUsed: number; createdAt: number };
type Grant = { id: number; subjectType: string; subjectId: number; hostId: number; enabled: boolean };

export default function PortalAdminPage() {
  const { t } = useTranslation();
  const [credentials, setCredentials] = useState<Credential[]>([]);
  const [grants, setGrants] = useState<Grant[]>([]);
  const [token, setToken] = useState('');
  const [error, setError] = useState('');
  const [enabled, setEnabled] = useState(false);
  async function load() {
    const [settings, c, g] = await Promise.all([
      HttpUtil.get<{ enabled: boolean }>('/panel/api/portal/settings', undefined, { silent: true }),
      HttpUtil.get<Credential[]>('/panel/api/portal/credentials', undefined, { silent: true }),
      HttpUtil.get<Grant[]>('/panel/api/portal/host-grants', undefined, { silent: true }),
    ]);
    if (settings.success) setEnabled(Boolean(settings.obj?.enabled));
    if (c.success) setCredentials(c.obj || []); else setError(c.msg);
    if (g.success) setGrants(g.obj || []); else setError(g.msg);
  }
  useEffect(() => { void load(); }, []);
  async function issue(clientId: number) {
    const result = await HttpUtil.post<{ token: string }>('/panel/api/portal/credentials', { clientId });
    if (result.success && result.obj) setToken(result.obj.token); else setError(result.msg);
    await load();
  }
  async function grant(values: { subjectType: string; subjectId: number; hostId: number }) {
    const result = await HttpUtil.post('/panel/api/portal/host-grants', values);
    if (!result.success) setError(result.msg);
    await load();
  }
  return <Space direction="vertical" style={{ width: '100%' }} size="large"><Card><Typography.Title level={2}>{t('fork.portal.adminTitle', 'Portal access')}</Typography.Title>{error && <Alert type="error" message={error} />}{token && <Alert type="warning" message={t('fork.portal.tokenOnce', 'Copy this token now. It will not be shown again.')} description={<Input.TextArea value={token} readOnly autoSize />} /> }<Space>{t('fork.portal.enabled', 'Self-service enabled')}<Switch checked={enabled} onChange={async (value) => { const result = await HttpUtil.post('/panel/api/portal/settings', { enabled: value }); if (result.success) setEnabled(value); else setError(result.msg); }} /></Space><Form layout="inline" onFinish={(values: { clientId: number }) => void issue(values.clientId)}><Form.Item name="clientId" label={t('fork.portal.clientId', 'Client ID')} rules={[{ required: true }]}><InputNumber min={1} /></Form.Item><Button htmlType="submit" type="primary">{t('fork.portal.issue', 'Issue token')}</Button></Form><Table rowKey="id" dataSource={credentials} columns={[{ title: t('fork.portal.clientId', 'Client ID'), dataIndex: 'clientId' }, { title: t('status', 'Status'), render: (_, row) => row.enabled ? t('enabled', 'Enabled') : t('disabled', 'Disabled') }, { title: t('fork.portal.lastUsed', 'Last used'), render: (_, row) => row.lastUsed ? new Date(row.lastUsed).toLocaleString() : '—' }, { title: t('actions', 'Actions'), render: (_, row) => <Space><Button onClick={() => void issue(row.clientId)}>{t('fork.portal.rotate', 'Rotate')}</Button><Button danger onClick={() => void HttpUtil.post(`/panel/api/portal/credentials/${row.clientId}/revoke`).then(load)}>{t('fork.portal.revoke', 'Revoke')}</Button></Space> }]} pagination={false} /></Card><Card><Typography.Title level={3}>{t('fork.portal.grants', 'Host visibility grants')}</Typography.Title><Form layout="inline" onFinish={(values: { subjectType: string; subjectId: number; hostId: number }) => void grant(values)}><Form.Item name="subjectType" initialValue="client" label={t('fork.portal.subjectType', 'Subject type')}><Input placeholder="client/group" /></Form.Item><Form.Item name="subjectId" label={t('fork.portal.subjectId', 'Subject ID')} rules={[{ required: true }]}><InputNumber min={1} /></Form.Item><Form.Item name="hostId" label={t('fork.portal.hostId', 'Host ID')} rules={[{ required: true }]}><InputNumber min={1} /></Form.Item><Button htmlType="submit" type="primary">{t('fork.portal.grant', 'Grant')}</Button></Form><Table rowKey="id" dataSource={grants} columns={[{ title: t('fork.portal.subject', 'Subject'), render: (_, row) => `${row.subjectType}:${row.subjectId}` }, { title: t('fork.portal.hostId', 'Host ID'), dataIndex: 'hostId' }, { title: t('actions', 'Actions'), render: (_, row) => <Button danger onClick={() => void HttpUtil.delete(`/panel/api/portal/host-grants/${row.id}`).then(load)}>{t('delete', 'Delete')}</Button> }]} pagination={false} /></Card></Space>;
}
