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
  try {
    const parsed: unknown = JSON.parse(value || '{}');
    if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed))
      throw new Error('Policy spec must be a JSON object.');
    return parsed as Record<string, unknown>;
  } catch {
    throw new Error('Policy spec must be valid JSON object syntax.');
  }
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

  const load = useCallback(async (showLoading = true) => {
    if (showLoading) setLoading(true);
    setError('');
    const [p, a, o, t, s] = await Promise.all([
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
    if (!p.success) setError(p.msg || 'Policies could not be loaded.');
    setEnabled(p.obj?.enabled !== false);
    setPolicies(p.obj?.items || []);
    setAssignments(a.obj?.items || []);
    setOverrides(o.obj?.items || []);
    setTemporary(t.obj?.items || []);
    setSchedules(s.obj?.items || []);
    setLoading(false);
  }, []);
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
    else message.error(result.msg || 'Simulation failed.');
    setSimLoading(false);
  };

  const policyColumns: ColumnsType<Policy> = [
    { title: 'Name', dataIndex: 'name' },
    { title: 'Priority', dataIndex: 'priority' },
    {
      title: 'State',
      render: (_, row) => (
        <Tag color={row.enabled ? 'green' : 'default'}>{row.enabled ? 'Enabled' : 'Disabled'}</Tag>
      ),
    },
    {
      title: 'Specification',
      render: (_, row) => (
        <Typography.Text code ellipsis={{ tooltip: row.spec }}>
          {row.spec}
        </Typography.Text>
      ),
    },
    {
      title: 'Actions',
      render: (_, row) => (
        <Space>
          <Button
            aria-label={`Edit ${row.name}`}
            icon={<EditOutlined />}
            onClick={() => openPolicy(row)}
          />
          <Popconfirm
            title={t('fork.policy.deleteConfirm')}
            onConfirm={() => void deletePolicy(row.id)}
          >
            <Button danger aria-label={`Delete ${row.name}`} icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ];
  const assignmentColumns: ColumnsType<Assignment> = [
    {
      title: 'Policy',
      render: (_, row) => policies.find((p) => p.id === row.policyId)?.name || `#${row.policyId}`,
    },
    { title: 'Target', render: (_, row) => `${row.targetType}: ${row.targetRef}` },
    { title: 'Priority', dataIndex: 'priority' },
    {
      title: 'Actions',
      render: (_, row) => (
        <Popconfirm
          title={t('fork.policy.labels.deleteAssignment')}
          onConfirm={() => void deleteRecord('/panel/api/policies/assignments', row.id)}
        >
          <Button danger icon={<DeleteOutlined />} />
        </Popconfirm>
      ),
    },
  ];
  const overrideColumns = (path: string): ColumnsType<Override> => [
    {
      title: 'Policy',
      render: (_, row) => policies.find((p) => p.id === row.policyId)?.name || `#${row.policyId}`,
    },
    { title: 'Target', render: (_, row) => `${row.targetType}: ${row.targetRef}` },
    { title: 'Scope', dataIndex: 'scope' },
    { title: 'Priority', dataIndex: 'priority' },
    {
      title: 'Actions',
      render: (_, row) => (
        <Popconfirm
          title={t('fork.policy.labels.deleteOverride')}
          onConfirm={() => void deleteRecord(path, row.id)}
        >
          <Button danger icon={<DeleteOutlined />} />
        </Popconfirm>
      ),
    },
  ];
  const scheduleColumns: ColumnsType<Schedule> = [
    {
      title: 'Policy',
      render: (_, row) => policies.find((p) => p.id === row.policyId)?.name || `#${row.policyId}`,
    },
    { title: 'Timezone', dataIndex: 'timezone' },
    { title: 'Weekdays', dataIndex: 'weekdays' },
    {
      title: 'Local window',
      render: (_, row) => `${formatMinute(row.startMinute)}–${formatMinute(row.endMinute)}`,
    },
    {
      title: 'State',
      render: (_, row) => (
        <Tag color={row.active ? 'green' : 'default'}>{row.active ? 'Active' : 'Inactive'}</Tag>
      ),
    },
    {
      title: 'Actions',
      render: (_, row) => (
        <Popconfirm
          title={t('fork.policy.labels.deleteSchedule')}
          onConfirm={() => void deleteRecord('/panel/api/policies/schedules', row.id)}
        >
          <Button danger icon={<DeleteOutlined />} />
        </Popconfirm>
      ),
    },
  ];

  return (
    <div className="policy-page">
      <div className="content-area">
        <Card>
          <Space direction="vertical" size="large" style={{ width: '100%' }}>
            <div className="policy-heading">
              <div>
                <Typography.Title level={2}>{t('fork.policy.title')}</Typography.Title>
                <Typography.Paragraph type="secondary">
                  Manage durable policies and inspect decisions without applying changes.
                </Typography.Paragraph>
              </div>
              <Button
                type="primary"
                icon={<PlusOutlined />}
                disabled={!enabled}
                onClick={() => openPolicy()}
              >
                New policy
              </Button>
            </div>
            {!enabled && <Alert type="info" showIcon message={t('fork.policy.labels.disabled')} />}
            {error && <Alert type="error" showIcon message={error} />}
            <Spin spinning={loading}>
              <Tabs
                items={[
                  {
                    key: 'policies',
                    label: `Policies (${policies.length})`,
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
                    label: `Assignments (${assignments.length})`,
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
                                { value: 'client', label: 'Client' },
                                { value: 'group', label: 'Group' },
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
                            Assign
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
                    label: `Overrides (${overrides.length + temporary.length})`,
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
                                { value: 'client', label: 'Client' },
                                { value: 'group', label: 'Group' },
                              ]}
                            />
                          </Form.Item>
                          <Form.Item name="targetRef" rules={[{ required: true }]}>
                            <Input placeholder={t('fork.policy.labels.target')} />
                          </Form.Item>
                          <Form.Item name="scope" initialValue="domain">
                            <Select
                              options={['policy', 'domain', 'service', 'dns', 'quota', 'qos'].map(
                                (value) => ({ value, label: value }),
                              )}
                            />
                          </Form.Item>
                          <Form.Item name="value" initialValue="{}">
                            <Input placeholder={t('fork.policy.labels.valueJson')} />
                          </Form.Item>
                          <Button type="primary" htmlType="submit">
                            Add override
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
                                { value: 'client', label: 'Client' },
                                { value: 'group', label: 'Group' },
                              ]}
                            />
                          </Form.Item>
                          <Form.Item name="targetRef" rules={[{ required: true }]}>
                            <Input placeholder={t('fork.policy.labels.target')} />
                          </Form.Item>
                          <Form.Item name="scope" initialValue="policy">
                            <Select
                              options={['policy', 'domain', 'service', 'quarantine-release'].map(
                                (value) => ({
                                  value,
                                  label: value,
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
                            Add temporary
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
                    label: `Schedules (${schedules.length})`,
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
                            Add schedule
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
                    label: t('fork.policy.labels.simulator'),
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
                              <Input placeholder={t('fork.policy.labels.clientPlaceholder')} />
                            </Form.Item>
                            <Form.Item name="groupName" label={t('fork.policy.labels.group')}>
                              <Input placeholder={t('fork.policy.labels.groupPlaceholder')} />
                            </Form.Item>
                            <Form.Item name="domain" label={t('fork.policy.labels.domain')}>
                              <Input placeholder={t('fork.policy.labels.domainPlaceholder')} />
                            </Form.Item>
                            <Form.Item name="ip" label={t('fork.policy.labels.ipCidr')}>
                              <Input placeholder={t('fork.policy.labels.ipPlaceholder')} />
                            </Form.Item>
                            <Form.Item name="category" label={t('fork.policy.labels.category')}>
                              <Input placeholder={t('fork.policy.labels.categoryPlaceholder')} />
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
                            {t('fork.policy.labels.simulate')}
                          </Button>
                        </Form>
                        {simulation && <SimulationView value={simulation} />}
                      </Card>
                    ),
                  },
                ]}
              />
            </Spin>
          </Space>
        </Card>
        <Modal
          open={policyModal}
          title={editing ? t('fork.policy.labels.editPolicy') : t('fork.policy.labels.newPolicy')}
          okText={t('fork.policy.labels.save')}
          onCancel={() => setPolicyModal(false)}
          onOk={() => form.submit()}
          destroyOnClose
        >
          <Form form={form} layout="vertical" onFinish={(values) => void savePolicy(values)}>
            <Form.Item
              name="name"
              label={t('fork.policy.labels.name')}
              rules={[{ required: true }]}
            >
              <Input />
            </Form.Item>
            <Form.Item name="description" label={t('fork.policy.labels.description')}>
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
              description={t('fork.policy.labels.capabilitiesDescription')}
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
      </div>
    </div>
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
        message={value.noOpReason || 'No effective policy.'}
      />
    );
  }
  const decisionColumns: ColumnsType<NonNullable<Simulation['decision']>['explanations'][number]> =
    [
      { title: 'Source', render: (_, row) => row.candidate.source },
      { title: 'Target', render: (_, row) => row.candidate.targetRef },
      { title: 'Priority', render: (_, row) => row.candidate.priority },
      {
        title: 'Result',
        render: (_, row) => (
          <Tag color={row.won ? 'green' : 'default'}>{row.won ? 'Winner' : row.reason}</Tag>
        ),
      },
    ];
  const routeColumns: ColumnsType<RoutePreview> = [
    { title: i18n.t('fork.policy.labels.rule'), dataIndex: 'ruleTag' },
    { title: i18n.t('fork.policy.labels.destination'), dataIndex: 'destination' },
    { title: i18n.t('fork.policy.labels.outbound'), dataIndex: 'outboundTag' },
    {
      title: i18n.t('fork.policy.labels.match'),
      render: (_, row) =>
        row.matched ? (
          <Tag color="green">{i18n.t('fork.policy.labels.matched')}</Tag>
        ) : (
          <Tag>{i18n.t('fork.policy.labels.notMatched')}</Tag>
        ),
    },
    {
      title: i18n.t('fork.policy.labels.emission'),
      render: (_, row) =>
        row.emitted
          ? i18n.t('fork.policy.labels.wouldEmit')
          : row.reason || i18n.t('fork.policy.labels.noOp'),
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
