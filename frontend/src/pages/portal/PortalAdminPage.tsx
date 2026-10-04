import { useCallback, useEffect, useState } from 'react';
import { useSearchParams } from 'react-router';
import {
  Alert,
  Button,
  Card,
  Empty,
  Form,
  Input,
  Select,
  Space,
  Switch,
  Table,
  Typography,
} from 'antd';
import { HttpUtil } from '@/utils';
import ForkAdminPageShell from '@/components/fork/ForkAdminPageShell';
import FeatureOffState from '@/components/fork/FeatureOffState';
import { isKnownForkFeatureUnavailable } from '@/lib/fork-feature';
import { useTranslation } from 'react-i18next';

type Credential = {
  id: number;
  clientId: number;
  enabled: boolean;
  expiresAt: number;
  lastUsed: number;
  createdAt: number;
};
type Grant = {
  id: number;
  subjectType: string;
  subjectId: number;
  hostId: number;
  enabled: boolean;
};
type PortalOption = { id: number; label: string };
type PortalOptions = {
  clients: PortalOption[];
  groups: PortalOption[];
  hosts: PortalOption[];
};

export default function PortalAdminPage() {
  const { t } = useTranslation();
  const [searchParams] = useSearchParams();
  const clientIdParam = Number(searchParams.get('clientId'));
  const initialClientId =
    Number.isInteger(clientIdParam) && clientIdParam > 0 ? clientIdParam : undefined;
  const [credentials, setCredentials] = useState<Credential[]>([]);
  const [grants, setGrants] = useState<Grant[]>([]);
  const [token, setToken] = useState('');
  const [error, setError] = useState('');
  const [enabled, setEnabled] = useState(false);
  const [featureOff, setFeatureOff] = useState(false);
  const [options, setOptions] = useState<PortalOptions>({ clients: [], groups: [], hosts: [] });
  const [grantForm] = Form.useForm();
  const subjectType = Form.useWatch('subjectType', grantForm) || 'client';
  const load = useCallback(async () => {
    const [settings, optionResult, c, g] = await Promise.all([
      HttpUtil.get<{ enabled: boolean }>('/panel/api/portal/settings', undefined, { silent: true }),
      HttpUtil.get<PortalOptions>('/panel/api/portal/options', undefined, { silent: true }),
      HttpUtil.get<Credential[]>('/panel/api/portal/credentials', undefined, { silent: true }),
      HttpUtil.get<Grant[]>('/panel/api/portal/host-grants', undefined, { silent: true }),
    ]);
    if (
      [settings, optionResult, c, g].some((response) =>
        isKnownForkFeatureUnavailable(response, 'self_service'),
      )
    ) {
      setFeatureOff(true);
      setError('');
      return;
    }
    const failed = [settings, optionResult, c, g].find((response) => !response.success);
    if (failed) {
      setError(t('fork.portal.actionFailed'));
      return;
    }
    setEnabled(Boolean(settings.obj?.enabled));
    setOptions(optionResult.obj || { clients: [], groups: [], hosts: [] });
    setCredentials(c.obj || []);
    setGrants(g.obj || []);
  }, [t]);
  useEffect(() => {
    const timer = window.setTimeout(() => void load(), 0);
    return () => window.clearTimeout(timer);
  }, [load]);
  async function issue(clientId: number) {
    const result = await HttpUtil.post<{ token: string }>('/panel/api/portal/credentials', {
      clientId,
    });
    if (result.success && result.obj) setToken(result.obj.token);
    else setError(t('fork.portal.actionFailed'));
    await load();
  }
  async function grant(values: { subjectType: string; subjectId: number; hostId: number }) {
    const result = await HttpUtil.post('/panel/api/portal/host-grants', values);
    if (!result.success) setError(t('fork.portal.actionFailed'));
    await load();
  }
  const optionLabel = (items: PortalOption[], id: number) =>
    items.find((item) => item.id === id)?.label || t('fork.common.unknown');
  return (
    <ForkAdminPageShell pageClass="portal-admin-page">
      <Space direction="vertical" style={{ width: '100%' }} size="large">
        <Card size="small">
          <Typography.Title level={2}>{t('fork.portal.adminTitle')}</Typography.Title>
          {featureOff ? (
            <FeatureOffState feature="self_service" />
          ) : (
            error && <Alert type="error" message={error} />
          )}
          {token && (
            <Alert
              type="warning"
              message={t('fork.portal.tokenOnce')}
              description={<Input.TextArea value={token} readOnly autoSize />}
            />
          )}
          {!featureOff && (
            <Space>
              {t('fork.portal.enabled')}
              <Switch
                checked={enabled}
                onChange={async (value) => {
                  const result = await HttpUtil.post('/panel/api/portal/settings', {
                    enabled: value,
                  });
                  if (result.success) setEnabled(value);
                  else setError(t('fork.portal.actionFailed'));
                }}
              />
            </Space>
          )}
          {!featureOff && (
            <Form
              layout="inline"
              initialValues={{ clientId: initialClientId }}
              onFinish={(values: { clientId: number }) => void issue(values.clientId)}
            >
              <Form.Item
                name="clientId"
                label={t('fork.portal.clientId')}
                rules={[{ required: true }]}
              >
                <Select
                  showSearch
                  optionFilterProp="label"
                  options={options.clients.map((item) => ({ value: item.id, label: item.label }))}
                  style={{ minWidth: 240 }}
                />
              </Form.Item>
              <Button htmlType="submit" type="primary">
                {t('fork.portal.issue')}
              </Button>
            </Form>
          )}
          {!featureOff && (
            <Table
              rowKey="id"
              dataSource={credentials}
              columns={[
                {
                  title: t('fork.portal.clientId'),
                  render: (_, row) => optionLabel(options.clients, row.clientId),
                },
                {
                  title: t('fork.common.status'),
                  render: (_, row) =>
                    row.enabled ? t('fork.common.enabled') : t('fork.common.disabled'),
                },
                {
                  title: t('fork.portal.lastUsed'),
                  render: (_, row) =>
                    row.lastUsed ? new Date(row.lastUsed).toLocaleString() : '—',
                },
                {
                  title: t('fork.common.actions'),
                  render: (_, row) => (
                    <Space>
                      <Button onClick={() => void issue(row.clientId)}>
                        {t('fork.portal.rotate')}
                      </Button>
                      <Button
                        danger
                        onClick={() =>
                          void HttpUtil.post(
                            `/panel/api/portal/credentials/${row.clientId}/revoke`,
                          ).then(load)
                        }
                      >
                        {t('fork.portal.revoke')}
                      </Button>
                    </Space>
                  ),
                },
              ]}
              pagination={false}
              locale={{ emptyText: <Empty description={t('fork.common.noItems')} /> }}
            />
          )}
        </Card>
        {!featureOff && (
          <Card size="small">
            <Typography.Title level={3}>{t('fork.portal.grants')}</Typography.Title>
            <Form
              form={grantForm}
              layout="inline"
              initialValues={{ subjectType: 'client' }}
              onFinish={(values: { subjectType: string; subjectId: number; hostId: number }) =>
                void grant(values)
              }
            >
              <Form.Item
                name="subjectType"
                label={t('fork.portal.subjectType')}
                rules={[{ required: true }]}
              >
                <Select
                  options={[
                    { value: 'client', label: t('fork.policy.labels.client') },
                    { value: 'group', label: t('fork.policy.labels.group') },
                  ]}
                />
              </Form.Item>
              <Form.Item
                name="subjectId"
                label={t('fork.portal.subjectId')}
                rules={[{ required: true }]}
              >
                <Select
                  showSearch
                  optionFilterProp="label"
                  options={(subjectType === 'group' ? options.groups : options.clients).map(
                    (item) => ({ value: item.id, label: item.label }),
                  )}
                  style={{ minWidth: 220 }}
                />
              </Form.Item>
              <Form.Item name="hostId" label={t('fork.portal.hostId')} rules={[{ required: true }]}>
                <Select
                  showSearch
                  optionFilterProp="label"
                  options={options.hosts.map((item) => ({ value: item.id, label: item.label }))}
                  style={{ minWidth: 240 }}
                />
              </Form.Item>
              <Button htmlType="submit" type="primary">
                {t('fork.portal.grant')}
              </Button>
            </Form>
            <Table
              rowKey="id"
              dataSource={grants}
              columns={[
                {
                  title: t('fork.portal.subject'),
                  render: (_, row) =>
                    `${row.subjectType === 'group' ? t('fork.policy.labels.group') : t('fork.policy.labels.client')}: ${optionLabel(row.subjectType === 'group' ? options.groups : options.clients, row.subjectId)}`,
                },
                {
                  title: t('fork.portal.hostId'),
                  render: (_, row) => optionLabel(options.hosts, row.hostId),
                },
                {
                  title: t('fork.common.actions'),
                  render: (_, row) => (
                    <Button
                      danger
                      onClick={() =>
                        void HttpUtil.delete(`/panel/api/portal/host-grants/${row.id}`).then(load)
                      }
                    >
                      {t('fork.common.delete')}
                    </Button>
                  ),
                },
              ]}
              pagination={false}
              locale={{ emptyText: <Empty description={t('fork.common.noItems')} /> }}
            />
          </Card>
        )}
      </Space>
    </ForkAdminPageShell>
  );
}
