import { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Alert,
  Button,
  Card,
  Descriptions,
  Empty,
  Input,
  Modal,
  Space,
  Table,
  Tabs,
  Typography,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { useTranslation } from 'react-i18next';
import { HttpUtil } from '@/utils';
import FeatureOffState from '@/components/fork/FeatureOffState';
import { isKnownForkFeatureUnavailable } from '@/lib/fork-feature';

type PortalClient = {
  id: number;
  email: string;
  group: string;
  enabled: boolean;
  totalGB: number;
  expiryTime: number;
  limitHwid: number;
};
type Device = {
  id: number;
  firstSeen: number;
  lastSeen: number;
  deviceName: string;
  deviceOs: string;
  osVersion: string;
  deviceModel: string;
};
type Host = {
  id: number;
  inboundId: number;
  remark: string;
  address: string;
  port: number;
  security: string;
  sni?: string;
  hostHeader?: string;
  path?: string;
  alpn?: string[];
  fingerprint?: string;
};
type Traffic = { up: number; down: number; total: number; expiryTime: number };

// Portal data is always fetched through the dedicated session namespace.
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
  const [featureOff, setFeatureOff] = useState(false);

  const load = useCallback(async () => {
    const [me, deviceResult, hostResult, trafficResult, csrfResult] = await Promise.all([
      HttpUtil.get<PortalClient>(portalPath('/portal/me'), undefined, { silent: true }),
      HttpUtil.get<Device[]>(portalPath('/portal/devices'), undefined, { silent: true }),
      HttpUtil.get<Host[]>(portalPath('/portal/hosts'), undefined, { silent: true }),
      HttpUtil.get<Traffic>(portalPath('/portal/traffic'), undefined, { silent: true }),
      HttpUtil.get<string>(portalPath('/portal/csrf'), undefined, { silent: true }),
    ]);
    if (!me.success) {
      if (isKnownForkFeatureUnavailable(me, 'self_service')) {
        setFeatureOff(true);
        setError('');
        return;
      }
      setClient(null);
      setError(me.msg || t('fork.portal.loginRequired'));
      return;
    }
    setClient(me.obj);
    setDevices(deviceResult.obj || []);
    setHosts(hostResult.obj || []);
    setTraffic(trafficResult.obj);
    setCsrf(csrfResult.obj || '');
    setError('');
  }, [t]);

  useEffect(() => {
    const timer = window.setTimeout(() => void load(), 0);
    return () => window.clearTimeout(timer);
  }, [load]);

  async function login() {
    const result = await HttpUtil.post<{ csrfToken: string }>(
      portalPath('/portal/auth'),
      { token },
      { silent: true },
    );
    if (!result.success) {
      setError(result.msg || t('fork.portal.invalidToken'));
      return;
    }
    setCsrf(result.obj?.csrfToken || '');
    await load();
  }

  const mutate = useCallback(
    async (path: string, method: 'patch' | 'delete', body?: unknown) => {
      const options = { headers: { 'X-CSRF-Token': csrf }, silent: true };
      const result =
        method === 'patch'
          ? await HttpUtil.put(portalPath(path), body, options)
          : await HttpUtil.delete(portalPath(path), options);
      if (!result.success) setError(result.msg || t('fork.portal.actionFailed'));
      else await load();
    },
    [csrf, load, t],
  );

  const columns = useMemo<ColumnsType<Device>>(
    () => [
      {
        title: t('fork.portal.device'),
        render: (_, row) => row.deviceName || row.deviceModel || row.deviceOs || '—',
      },
      {
        title: t('fork.portal.lastSeen'),
        render: (_, row) => (row.lastSeen ? new Date(row.lastSeen).toLocaleString() : '—'),
      },
      {
        title: t('fork.portal.actions'),
        render: (_, row) => (
          <Space>
            <Button
              onClick={() => {
                const name = window.prompt(t('fork.portal.renamePrompt'), row.deviceName);
                if (name != null)
                  void mutate(`/portal/devices/${row.id}`, 'patch', { deviceName: name });
              }}
            >
              {t('fork.common.rename')}
            </Button>
            <Button
              danger
              onClick={() =>
                Modal.confirm({
                  title: `${t('fork.portal.revoke')}?`,
                  onOk: () => mutate(`/portal/devices/${row.id}`, 'delete'),
                })
              }
            >
              {t('fork.portal.revoke')}
            </Button>
          </Space>
        ),
      },
    ],
    [mutate, t],
  );

  if (featureOff)
    return (
      <div className="portal-page">
        <div className="content-area portal-login-shell">
          <Card>
            <FeatureOffState />
          </Card>
        </div>
      </div>
    );

  if (!client)
    return (
      <div className="portal-page">
        <div className="content-area portal-login-shell">
          <Card>
            <Space direction="vertical" style={{ width: '100%' }}>
              <Typography.Title level={2}>{t('fork.portal.title')}</Typography.Title>
              <Typography.Paragraph>{t('fork.portal.tokenHint')}</Typography.Paragraph>
              <Input.Password
                value={token}
                onChange={(e) => setToken(e.target.value)}
                onPressEnter={() => void login()}
                placeholder={t('fork.portal.token')}
              />
              <Button type="primary" onClick={() => void login()}>
                {t('fork.portal.signIn')}
              </Button>
              {error && <Alert type="error" message={error} />}
            </Space>
          </Card>
        </div>
      </div>
    );

  return (
    <div className="portal-page">
      <div className="content-area">
        <Card>
          <Space direction="vertical" style={{ width: '100%' }} size="large">
            <div>
              <Typography.Title level={2}>{t('fork.portal.title')}</Typography.Title>
              {error && <Alert type="error" message={error} />}
            </div>
            <Descriptions
              bordered
              items={[
                { key: 'email', label: t('fork.common.email'), children: client.email },
                {
                  key: 'status',
                  label: t('fork.common.status'),
                  children: client.enabled ? t('fork.common.enabled') : t('fork.common.disabled'),
                },
                {
                  key: 'quota',
                  label: t('fork.portal.quota'),
                  children: traffic
                    ? `${traffic.up + traffic.down} / ${traffic.total || client.totalGB}`
                    : '—',
                },
                {
                  key: 'expiry',
                  label: t('fork.portal.expiry'),
                  children: client.expiryTime
                    ? new Date(client.expiryTime).toLocaleString()
                    : t('fork.common.never'),
                },
              ]}
            />
            <Tabs
              items={[
                {
                  key: 'devices',
                  label: t('fork.portal.devices'),
                  children: (
                    <Table
                      rowKey="id"
                      columns={columns}
                      dataSource={devices}
                      pagination={false}
                      locale={{ emptyText: <Empty description={t('fork.common.noItems')} /> }}
                    />
                  ),
                },
                {
                  key: 'hosts',
                  label: t('fork.portal.hosts'),
                  children: (
                    <Table
                      rowKey="id"
                      columns={[
                        {
                          title: t('fork.portal.host'),
                          dataIndex: 'address',
                          render: (value: string) => (
                            <span className="catx-technical-value" dir="ltr">
                              {value}
                            </span>
                          ),
                        },
                        { title: t('fork.common.labels.port'), dataIndex: 'port' },
                        { title: t('fork.common.labels.security'), dataIndex: 'security' },
                        {
                          title: t('fork.common.labels.sni'),
                          dataIndex: 'sni',
                          render: (value: string) => (
                            <span className="catx-technical-value" dir="ltr">
                              {value}
                            </span>
                          ),
                        },
                      ]}
                      dataSource={hosts}
                      pagination={false}
                      locale={{ emptyText: <Empty description={t('fork.common.noItems')} /> }}
                    />
                  ),
                },
              ]}
            />
          </Space>
        </Card>
      </div>
    </div>
  );
}
