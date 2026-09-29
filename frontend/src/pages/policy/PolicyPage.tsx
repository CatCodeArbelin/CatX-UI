import { useCallback, useEffect, useState } from 'react';
import {
  Alert,
  Button,
  Card,
  Empty,
  Form,
  Input,
  InputNumber,
  Modal,
  Popconfirm,
  Space,
  Spin,
  Select,
  Switch,
  Table,
  Tabs,
  Tag,
  Typography,
  message,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { DeleteOutlined, EditOutlined, PlusOutlined, SearchOutlined } from '@ant-design/icons';
import { HttpUtil } from '@/utils';
import ForkAdminPageShell from '@/components/fork/ForkAdminPageShell';
import FeatureOffState from '@/components/fork/FeatureOffState';
import { isKnownForkFeatureUnavailable } from '@/lib/fork-feature';
import { i18n } from '@/i18n/react';
import { useTranslation } from 'react-i18next';
import './PolicyPage.css';

type Policy = {
  id: number;
  name: string;
  description?: string;
  spec: string;
  priority: number;
  enabled: boolean;
};
type Assignment = {
  id: number;
  policyId: number;
  targetType: string;
  targetRef: string;
  priority: number;
  enabled: boolean;
};
type Override = Assignment & {
  scope: string;
  value: string;
  startsAt?: number;
  expiresAt?: number;
};
type Schedule = {
  id: number;
  policyId: number;
  timezone: string;
  weekdays: string;
  startMinute: number;
  endMinute: number;
  enabled: boolean;
  active: boolean;
  nextActiveAt: number;
};
type Decision = {
  policyName: string;
  action: string;
  explanation: string;
  quarantined?: boolean;
  quarantineReleased?: boolean;
  managedDns?: boolean;
  safeSearch?: boolean;
  dnsLimitations?: string[];
  winner: {
    source: string;
    targetType: string;
    targetRef: string;
    priority: number;
    scope?: string;
    expiresAt?: number;
  };
  explanations: {
    candidate: {
      source: string;
      targetRef: string;
      priority: number;
      scope?: string;
      expiresAt?: number;
    };
    won: boolean;
    reason: string;
  }[];
};
type RoutePreview = {
  ruleTag: string;
  destination: string;
  user: string;
  outboundTag?: string;
  order: number;
  emitted: boolean;
  matched: boolean;
  reason?: string;
};
type Simulation = {
  enabled: boolean;
  input: Record<string, unknown>;
  decision?: Decision;
  route: RoutePreview[];
  noOpReason?: string;
};

const parseSpec = (value: string): Record<string, unknown> => {
  let parsed: unknown;
  try {
    parsed = JSON.parse(value || '{}');
  } catch {
    throw new Error(i18n.t('fork.policy.errors.specSyntax'));
  }
  if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed))
    throw new Error(i18n.t('fork.policy.errors.specObject'));
  return parsed as Record<string, unknown>;
};
const POLICY_CATEGORIES = [
  'social',
  'video/streaming',
  'messaging',
  'gaming',
  'cloud/CDN',
  'search',
  'software/update',
  'advertising',
  'adult',
  'gambling',
].map((value) => ({ value, label: value }));

const formatMinute = (value: number) =>
  `${String(Math.floor(value / 60)).padStart(2, '0')}:${String(value % 60).padStart(2, '0')}`;

const targetTypeLabel = (value: string) =>
  i18n.t(`fork.policy.labels.${value}`, { defaultValue: value });

const scopeLabel = (value: string) =>
  i18n.t(`fork.policy.scopes.${value === 'quarantine-release' ? 'quarantineRelease' : value}`, {
    defaultValue: value,
  });

export default function PolicyPage() {
  const { t } = useTranslation();
  const [policies, setPolicies] = useState<Policy[]>([]);
  const [assignments, setAssignments] = useState<Assignment[]>([]);
  const [overrides, setOverrides] = useState<Override[]>([]);
  const [temporary, setTemporary] = useState<Override[]>([]);
  const [schedules, setSchedules] = useState<Schedule[]>([]);
  const [enabled, setEnabled] = useState(true);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [featureOff, setFeatureOff] = useState(false);
  const [editing, setEditing] = useState<Policy | null>(null);
  const [policyModal, setPolicyModal] = useState(false);
  const [form] = Form.useForm();
  const [simForm] = Form.useForm();
  const [assignmentForm] = Form.useForm();
  const [overrideForm] = Form.useForm();
  const [temporaryForm] = Form.useForm();
  const [scheduleForm] = Form.useForm();
  const [simulation, setSimulation] = useState<Simulation | null>(null);
  const [simLoading, setSimLoading] = useState(false);

  const load = useCallback(
    async (showLoading = true) => {
      if (showLoading) setLoading(true);
      setError('');
      const [p, a, o, temporaryResponse, s] = await Promise.all([
        HttpUtil.get<{ enabled: boolean; items: Policy[] }>('/panel/api/policies', undefined, {
          silent: true,
        }),
        HttpUtil.get<{ items: Assignment[] }>('/panel/api/policies/assignments', undefined, {
          silent: true,
        }),
        HttpUtil.get<{ items: Override[] }>('/panel/api/policies/overrides', undefined, {
          silent: true,
        }),
        HttpUtil.get<{ items: Override[] }>('/panel/api/policies/temporary-overrides', undefined, {
          silent: true,
        }),
        HttpUtil.get<{ items: Schedule[] }>('/panel/api/policies/schedules', undefined, {
          silent: true,
        }),
      ]);
      if (
        [p, a, o, temporaryResponse, s].some((response) =>
          isKnownForkFeatureUnavailable(response, 'policies'),
        )
      ) {
        setFeatureOff(true);
        setLoading(false);
        return;
      }
      if (!p.success) setError(p.msg || t('fork.policy.loadFailed'));
      setEnabled(p.obj?.enabled !== false);
      setPolicies(p.obj?.items || []);
      setAssignments(a.obj?.items || []);
      setOverrides(o.obj?.items || []);
      setTemporary(temporaryResponse.obj?.items || []);
      setSchedules(s.obj?.items || []);
      setLoading(false);
    },
    [t],
  );
  useEffect(() => {
    const timer = window.setTimeout(() => void load(false), 0);
    return () => window.clearTimeout(timer);
  }, [load]);

  const openPolicy = (policy?: Policy) => {
    setEditing(policy || null);
    let policySpec: Record<string, unknown> = {};
    try {
      policySpec = parseSpec(policy?.spec || '{}');
    } catch {
      policySpec = {};
    }
    form.setFieldsValue({
      name: policy?.name || '',
      description: policy?.description || '',
      priority: policy?.priority || 0,
      enabled: policy?.enabled ?? true,
      spec: policy?.spec || '{\n  "action": "allow",\n  "destinations": ["example.com"]\n}',
      categories: Array.isArray(policySpec.categories) ? policySpec.categories : [],
    });
    setPolicyModal(true);
  };
  const savePolicy = async (values: {
    name: string;
    description?: string;
    priority?: number;
    enabled?: boolean;
    spec: string;
    categories?: string[];
  }) => {
    let spec: Record<string, unknown>;
    try {
      spec = parseSpec(values.spec);
    } catch (err) {
      message.error((err as Error).message);
      return;
    }
    const payloadSpec = { ...spec };
    if (values.categories?.length) payloadSpec.categories = values.categories;
    else delete payloadSpec.categories;
    const payload = { ...values, spec: payloadSpec };
    const result = editing
      ? await HttpUtil.put(`/panel/api/policies/${editing.id}`, payload)
      : await HttpUtil.post('/panel/api/policies', payload);
    if (!result.success) return;
    setPolicyModal(false);
    await load();
  };
  const deletePolicy = async (id: number) => {
    const result = await HttpUtil.delete(`/panel/api/policies/${id}`);
    if (result.success) await load();
  };
  const createRecord = async (
    path: string,
    values: Record<string, unknown>,
    formInstance: ReturnType<typeof Form.useForm>[0],
  ) => {
    const result = await HttpUtil.post(path, values);
    if (result.success) {
      formInstance.resetFields();
      await load();
    }
  };
  const deleteRecord = async (path: string, id: number) => {
    const result = await HttpUtil.delete(`${path}/${id}`);
    if (result.success) await load();
  };
  const simulate = async (values: Record<string, string>) => {
    setSimLoading(true);
    const params = Object.fromEntries(Object.entries(values).filter(([, value]) => value));
    const result = await HttpUtil.get<Simulation>('/panel/api/policies/simulate', params, {
      silent: true,
    });
    if (result.success && result.obj) setSimulation(result.obj);
    else message.error(result.msg || t('fork.policy.simulationFailed'));
    setSimLoading(false);
  };

  const policyColumns: ColumnsType<Policy> = [
    { title: t('fork.common.labels.name'), dataIndex: 'name' },
    { title: t('fork.policy.labels.priority'), dataIndex: 'priority' },
    {
      title: t('fork.common.labels.status'),
      render: (_, row) => (
        <Tag color={row.enabled ? 'green' : 'default'}>
          {row.enabled ? t('fork.common.labels.enabled') : t('fork.common.labels.disabled')}
        </Tag>
      ),
    },
    {
      title: t('fork.policy.labels.specification'),
      render: (_, row) => (
        <Typography.Text code ellipsis={{ tooltip: row.spec }}>
          {row.spec}
        </Typography.Text>
      ),
    },
    {
      title: t('fork.common.labels.actions'),
      render: (_, row) => (
        <Space>
          <Button
            aria-label={t('fork.policy.editAria', { name: row.name })}
            icon={<EditOutlined />}
            onClick={() => openPolicy(row)}
          />
          <Popconfirm
            title={t('fork.policy.deleteConfirm')}
            onConfirm={() => void deletePolicy(row.id)}
          >
            <Button
              danger
              aria-label={t('fork.policy.deleteAria', { name: row.name })}
              icon={<DeleteOutlined />}
            />
          </Popconfirm>
        </Space>
      ),
    },
  ];
  const assignmentColumns: ColumnsType<Assignment> = [
    {
      title: t('fork.policy.labels.policy'),
      render: (_, row) => policies.find((p) => p.id === row.policyId)?.name || `#${row.policyId}`,
    },
    {
      title: t('fork.policy.labels.target'),
      render: (_, row) => `${targetTypeLabel(row.targetType)}: ${row.targetRef}`,
    },
    { title: t('fork.policy.labels.priority'), dataIndex: 'priority' },
    {
      title: t('fork.common.labels.actions'),
      render: (_, row) => (
        <Popconfirm
          title={t('fork.policy.labels.deleteAssignment')}
          onConfirm={() => void deleteRecord('/panel/api/policies/assignments', row.id)}
        >
          <Button
            danger
            aria-label={t('fork.policy.labels.deleteAssignment')}
            icon={<DeleteOutlined />}
          />
        </Popconfirm>
      ),
    },
  ];
  const overrideColumns = (path: string): ColumnsType<Override> => [
    {
      title: t('fork.policy.labels.policy'),
      render: (_, row) => policies.find((p) => p.id === row.policyId)?.name || `#${row.policyId}`,
    },
    {
      title: t('fork.policy.labels.target'),
      render: (_, row) => `${targetTypeLabel(row.targetType)}: ${row.targetRef}`,
    },
    { title: t('fork.policy.labels.scope'), dataIndex: 'scope', render: scopeLabel },
    { title: t('fork.policy.labels.priority'), dataIndex: 'priority' },
    {
      title: t('fork.common.labels.actions'),
      render: (_, row) => (
        <Popconfirm
          title={t('fork.policy.labels.deleteOverride')}
          onConfirm={() => void deleteRecord(path, row.id)}
        >
          <Button
            danger
            aria-label={t('fork.policy.labels.deleteOverride')}
            icon={<DeleteOutlined />}
          />
        </Popconfirm>
      ),
    },
  ];
  const scheduleColumns: ColumnsType<Schedule> = [
    {
      title: t('fork.policy.labels.policy'),
      render: (_, row) => policies.find((p) => p.id === row.policyId)?.name || `#${row.policyId}`,
    },
    { title: t('fork.policy.labels.ianaTimezone'), dataIndex: 'timezone' },
    { title: t('fork.policy.labels.weekdays'), dataIndex: 'weekdays' },
    {
      title: t('fork.policy.labels.localWindow'),
      render: (_, row) => `${formatMinute(row.startMinute)}–${formatMinute(row.endMinute)}`,
    },
    {
      title: t('fork.common.labels.status'),
      render: (_, row) => (
        <Tag color={row.active ? 'green' : 'default'}>
          {row.active ? t('fork.policy.labels.active') : t('fork.policy.labels.inactive')}
        </Tag>
      ),
    },
    {
      title: t('fork.common.labels.actions'),
      render: (_, row) => (
        <Popconfirm
          title={t('fork.policy.labels.deleteSchedule')}
          onConfirm={() => void deleteRecord('/panel/api/policies/schedules', row.id)}
        >
          <Button
            danger
            aria-label={t('fork.policy.labels.deleteSchedule')}
            icon={<DeleteOutlined />}
          />
        </Popconfirm>
      ),
    },
  ];

  return (
    <ForkAdminPageShell pageClass="policy-page">
      <Card size="small">
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <div className="policy-heading">
            <div>
              <Typography.Title level={2}>{t('fork.policy.title')}</Typography.Title>
              <Typography.Paragraph type="secondary">
                {t('fork.policy.manageDescription')}
              </Typography.Paragraph>
            </div>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              disabled={!enabled}
              onClick={() => openPolicy()}
            >
              {t('fork.policy.newPolicy')}
            </Button>
          </div>
          {!featureOff && !enabled && (
            <FeatureOffState feature="policies" messageKey="fork.policy.labels.disabled" />
          )}
          {featureOff ? (
            <FeatureOffState feature="policies" messageKey="fork.policy.labels.disabled" />
          ) : (
            error && <Alert type="error" showIcon message={error} />
          )}
          {!featureOff && enabled && (
            <Spin spinning={loading}>
              <Tabs
                items={[
                  {
                    key: 'policies',
                    label: t('fork.policy.tabs.policies', { count: policies.length }),
                    children: policies.length ? (
                      <Table
                        rowKey="id"
                        columns={policyColumns}
                        dataSource={policies}
                        pagination={{ pageSize: 10 }}
                      />
                    ) : (
                      <Empty description={t('fork.policy.noPolicies')} />
                    ),
                  },
                  {
                    key: 'assignments',
                    label: t('fork.policy.tabs.assignments', { count: assignments.length }),
                    children: (
                      <Space direction="vertical" style={{ width: '100%' }}>
                        <Form
                          form={assignmentForm}
                          layout="inline"
                          onFinish={(values) =>
                            void createRecord(
                              '/panel/api/policies/assignments',
                              values,
                              assignmentForm,
                            )
                          }
                        >
                          <Form.Item name="policyId" rules={[{ required: true }]}>
                            <Select
                              placeholder={t('fork.policy.labels.policy')}
                              style={{ width: 180 }}
                              options={policies.map((p) => ({ value: p.id, label: p.name }))}
                            />
                          </Form.Item>
                          <Form.Item name="targetType" initialValue="client">
                            <Select
                              options={[
                                { value: 'client', label: t('fork.policy.labels.client') },
                                { value: 'group', label: t('fork.policy.labels.group') },
                              ]}
                            />
                          </Form.Item>
                          <Form.Item name="targetRef" rules={[{ required: true }]}>
                            <Input placeholder={t('fork.policy.labels.targetEmailGroup')} />
                          </Form.Item>
                          <Form.Item name="priority" initialValue={0}>
                            <InputNumber placeholder={t('fork.policy.labels.priority')} />
                          </Form.Item>
                          <Button type="primary" htmlType="submit">
                            {t('fork.policy.assign')}
                          </Button>
                        </Form>
                        <Table
                          rowKey="id"
                          columns={assignmentColumns}
                          dataSource={assignments}
                          pagination={{ pageSize: 10 }}
                          locale={{
                            emptyText: <Empty description={t('fork.policy.noAssignments')} />,
                          }}
                        />
                      </Space>
                    ),
                  },
                  {
                    key: 'overrides',
                    label: t('fork.policy.tabs.overrides', {
                      count: overrides.length + temporary.length,
                    }),
                    children: (
                      <Space direction="vertical" style={{ width: '100%' }}>
                        <Typography.Title level={4}>
                          {t('fork.policy.explicitOverrides')}
                        </Typography.Title>
                        <Form
                          form={overrideForm}
                          layout="inline"
                          onFinish={(values) => {
                            try {
                              void createRecord(
                                '/panel/api/policies/overrides',
                                { ...values, value: parseSpec(String(values.value || '{}')) },
                                overrideForm,
                              );
                            } catch (err) {
                              message.error((err as Error).message);
                            }
                          }}
                        >
                          <Form.Item name="policyId" rules={[{ required: true }]}>
                            <Select
                              placeholder={t('fork.policy.labels.policy')}
                              style={{ width: 160 }}
                              options={policies.map((p) => ({ value: p.id, label: p.name }))}
                            />
                          </Form.Item>
                          <Form.Item name="targetType" initialValue="client">
                            <Select
                              options={[
                                { value: 'client', label: t('fork.policy.labels.client') },
                                { value: 'group', label: t('fork.policy.labels.group') },
                              ]}
                            />
                          </Form.Item>
                          <Form.Item name="targetRef" rules={[{ required: true }]}>
                            <Input placeholder={t('fork.policy.labels.target')} />
                          </Form.Item>
                          <Form.Item name="scope" initialValue="domain">
                            <Select
                              options={['policy', 'domain', 'service', 'dns', 'quota', 'qos'].map(
                                (value) => ({
                                  value,
                                  label: t(`fork.policy.scopes.${value}`),
                                }),
                              )}
                            />
                          </Form.Item>
                          <Form.Item name="value" initialValue="{}">
                            <Input placeholder={t('fork.policy.labels.valueJson')} />
                          </Form.Item>
                          <Button type="primary" htmlType="submit">
                            {t('fork.policy.addOverride')}
                          </Button>
                        </Form>
                        <Table
                          rowKey="id"
                          columns={overrideColumns('/panel/api/policies/overrides')}
                          dataSource={overrides}
                          pagination={{ pageSize: 10 }}
                          locale={{
                            emptyText: <Empty description={t('fork.policy.noExplicitOverrides')} />,
                          }}
                        />
                        <Typography.Title level={4}>
                          {t('fork.policy.temporaryOverrides')}
                        </Typography.Title>
                        <Form
                          form={temporaryForm}
                          layout="inline"
                          onFinish={(values) => {
                            try {
                              void createRecord(
                                '/panel/api/policies/temporary-overrides',
                                { ...values, value: parseSpec(String(values.value || '{}')) },
                                temporaryForm,
                              );
                            } catch (err) {
                              message.error((err as Error).message);
                            }
                          }}
                        >
                          <Form.Item name="policyId" rules={[{ required: true }]}>
                            <Select
                              placeholder={t('fork.policy.labels.policy')}
                              style={{ width: 160 }}
                              options={policies.map((p) => ({ value: p.id, label: p.name }))}
                            />
                          </Form.Item>
                          <Form.Item name="targetType" initialValue="client">
                            <Select
                              options={[
                                { value: 'client', label: t('fork.policy.labels.client') },
                                { value: 'group', label: t('fork.policy.labels.group') },
                              ]}
                            />
                          </Form.Item>
                          <Form.Item name="targetRef" rules={[{ required: true }]}>
                            <Input placeholder={t('fork.policy.labels.target')} />
                          </Form.Item>
                          <Form.Item name="scope" initialValue="policy">
                            <Select
                              options={['policy', 'domain', 'service', 'quarantineRelease'].map(
                                (value) => ({
                                  value:
                                    value === 'quarantineRelease' ? 'quarantine-release' : value,
                                  label: t(`fork.policy.scopes.${value}`),
                                }),
                              )}
                            />
                          </Form.Item>
                          <Form.Item name="startsAt" rules={[{ required: true }]}>
                            <InputNumber placeholder={t('fork.policy.labels.startsMs')} />
                          </Form.Item>
                          <Form.Item name="expiresAt" rules={[{ required: true }]}>
                            <InputNumber placeholder={t('fork.policy.labels.expiresMs')} />
                          </Form.Item>
                          <Form.Item name="value" initialValue="{}">
                            <Input placeholder={t('fork.policy.labels.valueJson')} />
                          </Form.Item>
                          <Button type="primary" htmlType="submit">
                            {t('fork.policy.addTemporary')}
                          </Button>
                        </Form>
                        <Table
                          rowKey="id"
                          columns={overrideColumns('/panel/api/policies/temporary-overrides')}
                          dataSource={temporary}
                          pagination={{ pageSize: 10 }}
                          locale={{
                            emptyText: (
                              <Empty description={t('fork.policy.noTemporaryOverrides')} />
                            ),
                          }}
                        />
                      </Space>
                    ),
                  },
                  {
                    key: 'schedules',
                    label: t('fork.policy.tabs.schedules', { count: schedules.length }),
                    children: (
                      <Space direction="vertical" style={{ width: '100%' }}>
                        <Form
                          form={scheduleForm}
                          layout="inline"
                          onFinish={(values) =>
                            void createRecord('/panel/api/policies/schedules', values, scheduleForm)
                          }
                        >
                          <Form.Item name="policyId" rules={[{ required: true }]}>
                            <Select
                              placeholder={t('fork.policy.labels.policy')}
                              style={{ width: 160 }}
                              options={policies.map((p) => ({ value: p.id, label: p.name }))}
                            />
                          </Form.Item>
                          <Form.Item
                            name="timezone"
                            initialValue="UTC"
                            rules={[{ required: true }]}
                          >
                            <Input placeholder={t('fork.policy.labels.ianaTimezone')} />
                          </Form.Item>
                          <Form.Item
                            name="weekdays"
                            initialValue="1,2,3,4,5"
                            rules={[{ required: true }]}
                          >
                            <Input placeholder={t('fork.policy.labels.weekdays')} />
                          </Form.Item>
                          <Form.Item
                            name="startMinute"
                            initialValue={9 * 60}
                            rules={[{ required: true }]}
                          >
                            <InputNumber placeholder={t('fork.policy.labels.startMinute')} />
                          </Form.Item>
                          <Form.Item
                            name="endMinute"
                            initialValue={17 * 60}
                            rules={[{ required: true }]}
                          >
                            <InputNumber placeholder={t('fork.policy.labels.endMinute')} />
                          </Form.Item>
                          <Button type="primary" htmlType="submit">
                            {t('fork.policy.addSchedule')}
                          </Button>
                        </Form>
                        <Table
                          rowKey="id"
                          columns={scheduleColumns}
                          dataSource={schedules}
                          pagination={{ pageSize: 10 }}
                          locale={{
                            emptyText: <Empty description={t('fork.policy.noSchedules')} />,
                          }}
                        />
                      </Space>
                    ),
                  },
                  {
                    key: 'simulator',
                    label: t('fork.policy.simulator'),
                    children: (
                      <Card size="small" title={t('fork.policy.simulation')}>
                        <Form
                          form={simForm}
                          layout="vertical"
                          onFinish={(values) => void simulate(values)}
                        >
                          <div className="policy-form-grid">
                            <Form.Item
                              name="clientEmail"
                              label={t('fork.policy.labels.clientEmail')}
                              rules={[{ required: true }]}
                            >
                              <Input placeholder={t('fork.policy.clientPlaceholder')} />
                            </Form.Item>
                            <Form.Item name="groupName" label={t('fork.policy.labels.group')}>
                              <Input placeholder={t('fork.policy.groupPlaceholder')} />
                            </Form.Item>
                            <Form.Item name="domain" label={t('fork.policy.labels.domain')}>
                              <Input placeholder={t('fork.policy.domainPlaceholder')} />
                            </Form.Item>
                            <Form.Item name="ip" label={t('fork.policy.labels.ipCidr')}>
                              <Input placeholder={t('fork.policy.ipPlaceholder')} />
                            </Form.Item>
                            <Form.Item name="category" label={t('fork.policy.labels.category')}>
                              <Input placeholder={t('fork.policy.categoryPlaceholder')} />
                            </Form.Item>
                            <Form.Item name="service" label={t('fork.policy.labels.service')}>
                              <Input placeholder={t('fork.policy.labels.exampleService')} />
                            </Form.Item>
                          </div>
                          <Button
                            type="primary"
                            htmlType="submit"
                            icon={<SearchOutlined />}
                            loading={simLoading}
                          >
                            {t('fork.policy.simulate')}
                          </Button>
                        </Form>
                        {simulation && <SimulationView value={simulation} />}
                      </Card>
                    ),
                  },
                ]}
              />
            </Spin>
          )}
        </Space>
      </Card>
      <Modal
        open={policyModal}
        title={editing ? t('fork.policy.editPolicy') : t('fork.policy.newPolicy')}
        okText={t('fork.policy.save')}
        onCancel={() => setPolicyModal(false)}
        onOk={() => form.submit()}
        destroyOnClose
      >
        <Form form={form} layout="vertical" onFinish={(values) => void savePolicy(values)}>
          <Form.Item name="name" label={t('fork.policy.labels.name')} rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="description" label={t('fork.policy.labels.summary')}>
            <Input.TextArea rows={2} />
          </Form.Item>
          <Space>
            <Form.Item name="priority" label={t('fork.policy.labels.priority')}>
              <InputNumber />
            </Form.Item>
            <Form.Item
              name="enabled"
              label={t('fork.policy.labels.enabled')}
              valuePropName="checked"
            >
              <Switch />
            </Form.Item>
          </Space>
          <Form.Item
            name="spec"
            label={t('fork.policy.labels.policyJson')}
            rules={[{ required: true }]}
          >
            <Input.TextArea rows={10} spellCheck={false} />
          </Form.Item>
          <Alert
            type="info"
            showIcon
            message={t('fork.policy.labels.capabilities')}
            description={t('fork.policy.capabilitiesDescription')}
          />
          <Form.Item name="categories" label={t('fork.policy.labels.knownCategories')}>
            <Select
              mode="multiple"
              options={POLICY_CATEGORIES}
              placeholder={t('fork.policy.labels.optionalCategoryTargets')}
            />
          </Form.Item>
        </Form>
      </Modal>
    </ForkAdminPageShell>
  );
}

function SimulationView({ value }: { value: Simulation }) {
  if (!value.enabled) {
    return <Alert className="policy-result" type="info" message={value.noOpReason} />;
  }
  if (!value.decision) {
    return (
      <Alert
        className="policy-result"
        type="info"
        message={value.noOpReason || i18n.t('fork.policy.noEffectivePolicy')}
      />
    );
  }
  const decisionColumns: ColumnsType<NonNullable<Simulation['decision']>['explanations'][number]> =
    [
      { title: i18n.t('fork.policy.labels.source'), render: (_, row) => row.candidate.source },
      { title: i18n.t('fork.policy.labels.target'), render: (_, row) => row.candidate.targetRef },
      { title: i18n.t('fork.policy.labels.priority'), render: (_, row) => row.candidate.priority },
      {
        title: i18n.t('fork.policy.labels.result'),
        render: (_, row) => (
          <Tag color={row.won ? 'green' : 'default'}>
            {row.won ? i18n.t('fork.policy.labels.winner') : row.reason}
          </Tag>
        ),
      },
    ];
  const routeColumns: ColumnsType<RoutePreview> = [
    { title: i18n.t('fork.policy.rule'), dataIndex: 'ruleTag' },
    { title: i18n.t('fork.policy.destination'), dataIndex: 'destination' },
    { title: i18n.t('fork.policy.outbound'), dataIndex: 'outboundTag' },
    {
      title: i18n.t('fork.policy.match'),
      render: (_, row) =>
        row.matched ? (
          <Tag color="green">{i18n.t('fork.policy.labels.matched')}</Tag>
        ) : (
          <Tag>{i18n.t('fork.policy.labels.notMatched')}</Tag>
        ),
    },
    {
      title: i18n.t('fork.policy.emission'),
      render: (_, row) =>
        row.emitted ? i18n.t('fork.policy.wouldEmit') : row.reason || i18n.t('fork.policy.noOp'),
    },
  ];
  return (
    <div className="policy-result">
      <Alert
        type={
          value.decision.quarantined || value.decision.action === 'deny' ? 'warning' : 'success'
        }
        message={`${value.decision.action.toUpperCase()} · ${value.decision.policyName}`}
        description={value.decision.explanation}
      />
      <Space wrap style={{ marginTop: 12 }}>
        {value.decision.quarantined && (
          <Tag color="red">{i18n.t('fork.policy.labels.quarantined')}</Tag>
        )}
        {value.decision.quarantineReleased && (
          <Tag color="orange">{i18n.t('fork.policy.labels.boundedRelease')}</Tag>
        )}
        {value.decision.managedDns && (
          <Tag color="blue">{i18n.t('fork.policy.labels.managedDns')}</Tag>
        )}
        {value.decision.safeSearch && (
          <Tag color="purple">{i18n.t('fork.policy.labels.safeSearch')}</Tag>
        )}
      </Space>
      {value.decision.dnsLimitations?.map((limitation) => (
        <Alert key={limitation} style={{ marginTop: 12 }} type="warning" message={limitation} />
      ))}
      <Typography.Title level={5}>
        {i18n.t('fork.policy.labels.decisionCandidates')}
      </Typography.Title>
      <Table
        rowKey={(row) =>
          `${row.candidate.source}-${row.candidate.targetRef}-${row.candidate.priority}`
        }
        size="small"
        pagination={false}
        dataSource={value.decision.explanations}
        columns={decisionColumns}
      />
      <Typography.Title level={5}>{i18n.t('fork.policy.labels.routePreview')}</Typography.Title>
      <Table
        rowKey="ruleTag"
        size="small"
        pagination={false}
        dataSource={value.route}
        columns={routeColumns}
      />
    </div>
  );
}
