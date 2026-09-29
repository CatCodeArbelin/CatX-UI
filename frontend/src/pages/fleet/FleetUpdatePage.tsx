import { useCallback, useEffect, useState } from 'react';
import {
  Alert,
  Button,
  Card,
  Checkbox,
  Form,
  Input,
  InputNumber,
  Popconfirm,
  Select,
  Space,
  Table,
  Tag,
} from 'antd';
import { useTranslation } from 'react-i18next';
import { HttpUtil } from '@/utils';

type Target = {
  id: number;
  nodeId: number;
  nodeName: string;
  state: string;
  blockedReason?: string;
  observedVersion?: string;
  runId?: string;
  dispatchStatus?: string;
  updateState?: string;
  rolledBack?: boolean;
  rollbackHealthy?: boolean;
  error?: string;
  soakStartedAt?: string;
};
type Campaign = {
  id: number;
  name: string;
  channel: string;
  releaseTag: string;
  state: string;
  dryRun: boolean;
  createdAt: string;
};
type Plan = { campaign: Campaign; targets: Target[] };

export default function FleetUpdatePage() {
  const { t } = useTranslation();
  const [campaigns, setCampaigns] = useState<Campaign[]>([]);
  const [selected, setSelected] = useState<Plan | null>(null);
  const [error, setError] = useState('');
  const load = useCallback(async () => {
    const result = await HttpUtil.get<Campaign[]>('/panel/api/fleet-updates/campaigns', undefined, {
      silent: true,
    });
    if (result.success) setCampaigns(result.obj || []);
    else setError(result.msg);
  }, []);
  useEffect(() => {
    void load();
  }, [load]);
  async function create(values: {
    name?: string;
    channel: string;
    canaryCount?: number;
    batchSize?: number;
    maxParallel?: number;
    healthTimeoutSecs?: number;
    soakSeconds?: number;
    dryRun?: boolean;
    confirmProduction?: boolean;
  }) {
    const result = await HttpUtil.post<Plan>('/panel/api/fleet-updates/campaigns', values);
    if (result.success && result.obj) {
      setSelected(result.obj);
      await load();
    } else setError(result.msg);
  }
  async function action(id: number, action: 'abort' | 'retry' | 'reconcile') {
    const result = await HttpUtil.post(`/panel/api/fleet-updates/campaigns/${id}/${action}`);
    if (!result.success) setError(result.msg);
    await load();
  }
  return (
    <div className="fleet-update-page">
      <div className="content-area">
        <Space direction="vertical" style={{ width: '100%' }} size="large">
          <Card title={t('fork.fleetUpdate.title')}>
            {error && <Alert type="error" message={error} />}
            <Form
              layout="inline"
              onFinish={(v) => void create(v)}
              initialValues={{
                channel: 'stable',
                canaryCount: 1,
                batchSize: 1,
                maxParallel: 1,
                healthTimeoutSecs: 300,
                soakSeconds: 0,
                dryRun: true,
                confirmProduction: false,
              }}
            >
              <Form.Item name="name">
                <Input
                  aria-label={t('fork.fleetUpdate.name')}
                  placeholder={t('fork.fleetUpdate.name')}
                />
              </Form.Item>
              <Form.Item name="channel">
                <Select
                  aria-label={t('fork.common.channel')}
                  options={[
                    { value: 'stable', label: t('fork.fleetUpdate.stable') },
                    { value: 'dev', label: t('fork.fleetUpdate.dev') },
                  ]}
                />
              </Form.Item>
              <Form.Item name="canaryCount">
                <InputNumber
                  aria-label={t('fork.fleetUpdate.canary')}
                  min={0}
                  placeholder={t('fork.fleetUpdate.canary')}
                />
              </Form.Item>
              <Form.Item name="batchSize">
                <InputNumber
                  aria-label={t('fork.fleetUpdate.batch')}
                  min={1}
                  placeholder={t('fork.fleetUpdate.batch')}
                />
              </Form.Item>
              <Form.Item name="maxParallel">
                <InputNumber
                  aria-label={t('fork.fleetUpdate.parallel')}
                  min={1}
                  placeholder={t('fork.fleetUpdate.parallel')}
                />
              </Form.Item>
              <Form.Item name="dryRun" valuePropName="checked">
                <Checkbox>{t('fork.fleetUpdate.dryRun')}</Checkbox>
              </Form.Item>
              <Form.Item name="confirmProduction" valuePropName="checked">
                <Checkbox>{t('fork.fleetUpdate.confirmProduction')}</Checkbox>
              </Form.Item>
              <Button htmlType="submit" type="primary">
                {t('fork.fleetUpdate.plan')}
              </Button>
            </Form>
          </Card>
          <Card title={t('fork.fleetUpdate.campaigns')}>
            <Table
              rowKey="id"
              dataSource={campaigns}
              columns={[
                { title: t('fork.common.name'), dataIndex: 'name' },
                { title: t('fork.common.channel'), dataIndex: 'channel' },
                { title: t('fork.common.status'), render: (_, r) => <Tag>{r.state}</Tag> },
                {
                  title: t('fork.common.actions'),
                  render: (_, r) => (
                    <Space>
                      <Button onClick={() => void action(r.id, 'reconcile')}>
                        {t('fork.fleetUpdate.reconcile')}
                      </Button>
                      <Button onClick={() => void action(r.id, 'retry')}>
                        {t('fork.fleetUpdate.retry')}
                      </Button>
                      <Popconfirm
                        title={`${t('fork.fleetUpdate.abort')}?`}
                        onConfirm={() => void action(r.id, 'abort')}
                      >
                        <Button danger>{t('fork.fleetUpdate.abort')}</Button>
                      </Popconfirm>
                    </Space>
                  ),
                },
              ]}
              pagination={false}
              onRow={(r) => ({
                onClick: async () => {
                  const v = await HttpUtil.get<Plan>(`/panel/api/fleet-updates/campaigns/${r.id}`);
                  if (v.success) setSelected(v.obj || null);
                },
              })}
            />
          </Card>
          {selected && (
            <Card title={`${selected.campaign.name} — ${selected.campaign.releaseTag}`}>
              <Table
                rowKey="id"
                dataSource={selected.targets}
                columns={[
                  { title: t('fork.common.node'), dataIndex: 'nodeName' },
                  { title: t('fork.common.status'), render: (_, r) => <Tag>{r.state}</Tag> },
                  { title: t('fork.common.reason'), dataIndex: 'blockedReason' },
                  { title: t('fork.common.version'), dataIndex: 'observedVersion' },
                  {
                    title: t('fork.fleetUpdate.runId'),
                    dataIndex: 'runId',
                    render: (value: string | undefined) =>
                      value ? (
                        <span className="catx-technical-value" dir="ltr">
                          {value}
                        </span>
                      ) : (
                        '—'
                      ),
                  },
                  {
                    title: t('fork.fleetUpdate.dispatch'),
                    dataIndex: 'dispatchStatus',
                  },
                  {
                    title: t('fork.fleetUpdate.result'),
                    dataIndex: 'updateState',
                  },
                  {
                    title: t('fork.fleetUpdate.rollback'),
                    render: (_, r) =>
                      r.rolledBack
                        ? r.rollbackHealthy
                          ? t('fork.common.labels.healthy')
                          : t('fork.common.labels.attempted')
                        : '—',
                  },
                  {
                    title: t('fork.fleetUpdate.soak'),
                    render: (_, r) =>
                      r.state === 'soaking'
                        ? r.soakStartedAt || t('fork.common.labels.active')
                        : r.state,
                  },
                  { title: t('fork.fleetUpdate.failure'), dataIndex: 'error' },
                ]}
                pagination={false}
              />
            </Card>
          )}
        </Space>
      </div>
    </div>
  );
}
