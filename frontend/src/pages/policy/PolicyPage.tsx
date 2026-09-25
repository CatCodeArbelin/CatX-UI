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
          <Popconfirm title="Delete this policy?" onConfirm={() => void deletePolicy(row.id)}>
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
          title="Delete this assignment?"
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
        <Popconfirm title="Delete this override?" onConfirm={() => void deleteRecord(path, row.id)}>
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
          title="Delete this schedule?"
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
                <Typography.Title level={2}>Policy engine</Typography.Title>
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
            {!enabled && (
              <Alert
                type="info"
                showIcon
                message="Policies are disabled. The panel remains upstream-compatible and simulations are read-only no-ops."
              />
            )}
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
                      <Empty description="No policies yet." />
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
                              placeholder="Policy"
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
                            <Input placeholder="Target email or group" />
                          </Form.Item>
                          <Form.Item name="priority" initialValue={0}>
                            <InputNumber placeholder="Priority" />
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
                          locale={{ emptyText: <Empty description="No assignments." /> }}
                        />
                      </Space>
                    ),
                  },
                  {
                    key: 'overrides',
                    label: `Overrides (${overrides.length + temporary.length})`,
                    children: (
                      <Space direction="vertical" style={{ width: '100%' }}>
                        <Typography.Title level={4}>Explicit overrides</Typography.Title>
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
                              placeholder="Policy"
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
                            <Input placeholder="Target" />
                          </Form.Item>
                          <Form.Item name="scope" initialValue="domain">
                            <Select
                              options={['policy', 'domain', 'service', 'dns', 'quota', 'qos'].map(
                                (value) => ({ value, label: value }),
                              )}
                            />
                          </Form.Item>
                          <Form.Item name="value" initialValue="{}">
                            <Input placeholder="Value JSON" />
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
                          locale={{ emptyText: <Empty description="No explicit overrides." /> }}
                        />
                        <Typography.Title level={4}>Temporary overrides</Typography.Title>
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
                              placeholder="Policy"
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
                            <Input placeholder="Target" />
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
                            <InputNumber placeholder="Starts ms" />
                          </Form.Item>
                          <Form.Item name="expiresAt" rules={[{ required: true }]}>
                            <InputNumber placeholder="Expires ms" />
                          </Form.Item>
                          <Form.Item name="value" initialValue="{}">
                            <Input placeholder="Value JSON" />
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
                            emptyText: <Empty description="No active temporary overrides." />,
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
                              placeholder="Policy"
                              style={{ width: 160 }}
                              options={policies.map((p) => ({ value: p.id, label: p.name }))}
                            />
                          </Form.Item>
                          <Form.Item
                            name="timezone"
                            initialValue="UTC"
                            rules={[{ required: true }]}
                          >
                            <Input placeholder="IANA timezone" />
                          </Form.Item>
                          <Form.Item
                            name="weekdays"
                            initialValue="1,2,3,4,5"
                            rules={[{ required: true }]}
                          >
                            <Input placeholder="Weekdays 0–6" />
                          </Form.Item>
                          <Form.Item
                            name="startMinute"
                            initialValue={9 * 60}
                            rules={[{ required: true }]}
                          >
                            <InputNumber placeholder="Start minute" />
                          </Form.Item>
                          <Form.Item
                            name="endMinute"
                            initialValue={17 * 60}
                            rules={[{ required: true }]}
                          >
                            <InputNumber placeholder="End minute" />
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
                          locale={{ emptyText: <Empty description="No schedules." /> }}
                        />
                      </Space>
                    ),
                  },
                  {
                    key: 'simulator',
                    label: 'Simulator',
                    children: (
                      <Card size="small" title="Read-only simulation">
                        <Form
                          form={simForm}
                          layout="vertical"
                          onFinish={(values) => void simulate(values)}
                        >
                          <div className="policy-form-grid">
                            <Form.Item
                              name="clientEmail"
                              label="Client email"
                              rules={[{ required: true }]}
                            >
                              <Input placeholder="alice@example.test" />
                            </Form.Item>
                            <Form.Item name="groupName" label="Group">
                              <Input placeholder="staff" />
                            </Form.Item>
                            <Form.Item name="domain" label="Domain">
                              <Input placeholder="example.com" />
                            </Form.Item>
                            <Form.Item name="ip" label="IP / CIDR">
                              <Input placeholder="203.0.113.10" />
                            </Form.Item>
                            <Form.Item name="category" label="Category">
                              <Input placeholder="video/streaming" />
                            </Form.Item>
                            <Form.Item name="service" label="Service">
                              <Input placeholder="Example service" />
                            </Form.Item>
                          </div>
                          <Button
                            type="primary"
                            htmlType="submit"
                            icon={<SearchOutlined />}
                            loading={simLoading}
                          >
                            Simulate
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
          title={editing ? 'Edit policy' : 'New policy'}
          okText="Save"
          onCancel={() => setPolicyModal(false)}
          onOk={() => form.submit()}
          destroyOnClose
        >
          <Form form={form} layout="vertical" onFinish={(values) => void savePolicy(values)}>
            <Form.Item name="name" label="Name" rules={[{ required: true }]}>
              <Input />
            </Form.Item>
            <Form.Item name="description" label="Description">
              <Input.TextArea rows={2} />
            </Form.Item>
            <Space>
              <Form.Item name="priority" label="Priority">
                <InputNumber />
              </Form.Item>
              <Form.Item name="enabled" label="Enabled" valuePropName="checked">
                <Switch />
              </Form.Item>
            </Space>
            <Form.Item name="spec" label="Policy JSON" rules={[{ required: true }]}>
              <Input.TextArea rows={10} spellCheck={false} />
            </Form.Item>
            <Alert
              type="info"
              showIcon
              message="WP-4B capabilities"
              description='Use quarantine:true with an optional quarantineAllowlist. Use dns:{ managed:true, dnsOutboundTag:"dns-out" } for intercepted plaintext DNS. SafeSearch requires safeSearch:true and an explicit dnsOutboundTag whose resolver enforces SafeSearch. A bounded release uses temporary scope quarantine-release with value {"released":true}.'
            />
            <Form.Item name="categories" label="Known categories">
              <Select
                mode="multiple"
                options={POLICY_CATEGORIES}
                placeholder="Optional category targets"
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
    { title: 'Rule', dataIndex: 'ruleTag' },
    { title: 'Destination', dataIndex: 'destination' },
    { title: 'Outbound', dataIndex: 'outboundTag' },
    {
      title: 'Match',
      render: (_, row) => (row.matched ? <Tag color="green">Matched</Tag> : <Tag>Not matched</Tag>),
    },
    { title: 'Emission', render: (_, row) => (row.emitted ? 'Would emit' : row.reason || 'No-op') },
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
        {value.decision.quarantined && <Tag color="red">Quarantined</Tag>}
        {value.decision.quarantineReleased && <Tag color="orange">Bounded release active</Tag>}
        {value.decision.managedDns && <Tag color="blue">Managed DNS</Tag>}
        {value.decision.safeSearch && <Tag color="purple">SafeSearch capability</Tag>}
      </Space>
      {value.decision.dnsLimitations?.map((limitation) => (
        <Alert key={limitation} style={{ marginTop: 12 }} type="warning" message={limitation} />
      ))}
      <Typography.Title level={5}>Decision candidates</Typography.Title>
      <Table
        rowKey={(row) =>
          `${row.candidate.source}-${row.candidate.targetRef}-${row.candidate.priority}`
        }
        size="small"
        pagination={false}
        dataSource={value.decision.explanations}
        columns={decisionColumns}
      />
      <Typography.Title level={5}>Route preview</Typography.Title>
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
