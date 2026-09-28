import { useCallback, useEffect, useState } from 'react';
import {
  Alert,
  Button,
  Card,
  Checkbox,
  Form,
  Input,
  InputNumber,
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
    <Space direction="vertical" style={{ width: '100%' }} size="large">
      <Card title={t('fork.fleetUpdate.title', 'Fleet updates')}>
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
            <Input placeholder={t('fork.fleetUpdate.name', 'Campaign name')} />
          </Form.Item>
          <Form.Item name="channel">
            <Select
              options={[
                { value: 'stable', label: t('fork.fleetUpdate.stable', 'Stable') },
                { value: 'dev', label: t('fork.fleetUpdate.dev', 'Development') },
              ]}
            />
          </Form.Item>
          <Form.Item name="canaryCount">
            <InputNumber min={0} placeholder={t('fork.fleetUpdate.canary', 'Canary')} />
          </Form.Item>
          <Form.Item name="batchSize">
            <InputNumber min={1} placeholder={t('fork.fleetUpdate.batch', 'Batch')} />
          </Form.Item>
          <Form.Item name="maxParallel">
            <InputNumber min={1} placeholder={t('fork.fleetUpdate.parallel', 'Parallel')} />
          </Form.Item>
          <Form.Item name="dryRun" valuePropName="checked">
            <Checkbox>{t('fork.fleetUpdate.dryRun', 'Dry run (no mutation)')}</Checkbox>
          </Form.Item>
          <Form.Item name="confirmProduction" valuePropName="checked">
            <Checkbox>
              {t('fork.fleetUpdate.confirmProduction', 'Confirm production update')}
            </Checkbox>
          </Form.Item>
          <Button htmlType="submit" type="primary">
            {t('fork.fleetUpdate.plan', 'Plan dry run')}
          </Button>
        </Form>
      </Card>
      <Card title={t('fork.fleetUpdate.campaigns', 'Campaigns')}>
        <Table
          rowKey="id"
          dataSource={campaigns}
          columns={[
            { title: t('name', 'Name'), dataIndex: 'name' },
            { title: t('channel', 'Channel'), dataIndex: 'channel' },
            { title: t('status', 'Status'), render: (_, r) => <Tag>{r.state}</Tag> },
            {
              title: t('actions', 'Actions'),
              render: (_, r) => (
                <Space>
                  <Button onClick={() => void action(r.id, 'reconcile')}>
                    {t('fork.fleetUpdate.reconcile', 'Reconcile')}
                  </Button>
                  <Button onClick={() => void action(r.id, 'retry')}>
                    {t('fork.fleetUpdate.retry', 'Retry')}
                  </Button>
                  <Button danger onClick={() => void action(r.id, 'abort')}>
                    {t('fork.fleetUpdate.abort', 'Abort')}
                  </Button>
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
              { title: t('node', 'Node'), dataIndex: 'nodeName' },
              { title: t('status', 'Status'), render: (_, r) => <Tag>{r.state}</Tag> },
              { title: t('reason', 'Reason'), dataIndex: 'blockedReason' },
              { title: t('version', 'Version'), dataIndex: 'observedVersion' },
              { title: t('fork.fleetUpdate.runId', 'Run ID'), dataIndex: 'runId' },
              { title: t('fork.fleetUpdate.dispatch', 'Dispatch'), dataIndex: 'dispatchStatus' },
              { title: t('fork.fleetUpdate.result', 'Update result'), dataIndex: 'updateState' },
              {
                title: t('fork.fleetUpdate.rollback', 'Rollback'),
                render: (_, r) =>
                  r.rolledBack ? (r.rollbackHealthy ? 'healthy' : 'attempted') : '—',
              },
              {
                title: t('fork.fleetUpdate.soak', 'Soak'),
                render: (_, r) => (r.state === 'soaking' ? r.soakStartedAt || 'active' : r.state),
              },
              { title: t('fork.fleetUpdate.failure', 'Failure'), dataIndex: 'error' },
            ]}
            pagination={false}
          />
        </Card>
      )}
    </Space>
  );
}
