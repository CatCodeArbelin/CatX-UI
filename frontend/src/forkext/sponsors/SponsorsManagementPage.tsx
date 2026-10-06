import { useCallback, useEffect, useMemo, useState } from 'react';
import dayjs, { type Dayjs } from 'dayjs';
import {
  Alert,
  Button,
  Card,
  Checkbox,
  Col,
  DatePicker,
  Descriptions,
  Empty,
  Form,
  Input,
  InputNumber,
  Modal,
  Row,
  Select,
  Space,
  Switch,
  Table,
  Tabs,
  Tag,
  Typography,
  message,
} from 'antd';
import {
  DeleteOutlined,
  EditOutlined,
  EyeOutlined,
  PlusOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';

import { HttpUtil } from '@/utils';
import type { Sponsor } from '@/generated/types';
import { isKnownForkFeatureUnavailable } from '@/lib/fork-feature';
import FeatureOffState from '@/components/fork/FeatureOffState';
import ForkAdminPageShell from '@/components/fork/ForkAdminPageShell';
import SponsorCard from '@/components/sponsor/SponsorCard';

type Slot = 'dashboard' | 'sidebar' | 'page';
type ProviderMode = 'local' | 'remote';

type ManagedSponsor = {
  id: string;
  enabled: boolean;
  name: string;
  priority: number;
  slots: Slot[];
  startAt?: string;
  endAt?: string;
  destinationUrl: string;
  logoUrl?: string;
  title: Record<string, string>;
  text: Record<string, string>;
  createdAt: string;
  updatedAt: string;
};

type SponsorSettings = {
  enabled: boolean;
  configured: boolean;
  sourceUrl: string;
  contactUrl: string;
  providerMode: ProviderMode;
};

type SponsorStatus = {
  enabled: boolean;
  providerMode: ProviderMode;
  localSponsorCount: number;
  activeSponsorCount: number;
  remoteProviderConfigured: boolean;
  lastSuccessfulFetch?: string;
  lastError?: string;
  cacheState: string;
};

type SponsorFormValues = {
  id?: string;
  enabled: boolean;
  name: string;
  priority: number;
  slots: Slot[];
  window?: [Dayjs, Dayjs];
  destinationUrl: string;
  logoUrl?: string;
  titleEn: string;
  titleRu?: string;
  titleFa?: string;
  textEn?: string;
  textRu?: string;
  textFa?: string;
};

const JSON_HEADERS = { headers: { 'Content-Type': 'application/json' } } as const;
const slots: Slot[] = ['dashboard', 'sidebar', 'page'];
const slotLabelKey: Record<Slot, string> = {
  dashboard: 'pages.sponsors.placementDashboard',
  sidebar: 'pages.sponsors.placementSidebar',
  page: 'pages.sponsors.placementPage',
};

const providerStateLabel = (
  t: (key: string, options?: Record<string, unknown>) => string,
  value: string,
) => {
  const keys: Record<string, string> = {
    empty: 'pages.sponsors.providerStates.empty',
    local: 'pages.sponsors.providerStates.local',
    fresh: 'pages.sponsors.providerStates.fresh',
    error: 'pages.sponsors.providerStates.error',
    disabled: 'pages.sponsors.providerStates.disabled',
  };
  return t(keys[value] || 'fork.common.unknown');
};

function inputFromRecord(record: ManagedSponsor): SponsorFormValues {
  const window =
    record.startAt && record.endAt
      ? ([dayjs(record.startAt), dayjs(record.endAt)] as [Dayjs, Dayjs])
      : undefined;
  return {
    id: record.id,
    enabled: record.enabled,
    name: record.name,
    priority: record.priority,
    slots: record.slots,
    window,
    destinationUrl: record.destinationUrl,
    logoUrl: record.logoUrl,
    titleEn: record.title['en-US'] || record.title.en || record.name,
    titleRu: record.title['ru-RU'] || record.title.ru,
    titleFa: record.title['fa-IR'] || record.title.fa,
    textEn: record.text['en-US'] || record.text.en,
    textRu: record.text['ru-RU'] || record.text.ru,
    textFa: record.text['fa-IR'] || record.text.fa,
  };
}

export function sponsorRequestFromForm(values: SponsorFormValues) {
  const title = Object.fromEntries(
    [
      ['en-US', values.titleEn],
      ['ru-RU', values.titleRu],
      ['fa-IR', values.titleFa],
    ].filter(([, value]) => value?.trim()),
  );
  const text = Object.fromEntries(
    [
      ['en-US', values.textEn],
      ['ru-RU', values.textRu],
      ['fa-IR', values.textFa],
    ].filter(([, value]) => value?.trim()),
  );
  return {
    id: values.id,
    enabled: values.enabled,
    name: values.name,
    priority: values.priority,
    slots: values.slots,
    startAt: values.window?.[0]?.toISOString() || '',
    endAt: values.window?.[1]?.toISOString() || '',
    destinationUrl: values.destinationUrl,
    logoUrl: values.logoUrl || '',
    title,
    text,
  };
}

function previewValue(record: ManagedSponsor): Sponsor {
  return {
    id: record.id,
    name: record.name,
    enable: record.enabled,
    slots: record.slots,
    from: record.startAt,
    until: record.endAt || '9999-12-31T23:59:59Z',
    logo: record.logoUrl ? `/panel/api/fork/sponsors/logo/${record.id}` : undefined,
    title: record.title,
    text: record.text,
    link: record.destinationUrl,
  } as Sponsor;
}

export default function SponsorsManagementPage() {
  const { t } = useTranslation();
  const [form] = Form.useForm<SponsorFormValues>();
  const [providerForm] = Form.useForm<SponsorSettings>();
  const [messageApi, contextHolder] = message.useMessage();
  const [items, setItems] = useState<ManagedSponsor[]>([]);
  const [settings, setSettings] = useState<SponsorSettings | null>(null);
  const [status, setStatus] = useState<SponsorStatus | null>(null);
  const [selected, setSelected] = useState<ManagedSponsor | null>(null);
  const [editing, setEditing] = useState<ManagedSponsor | null>(null);
  const [modalOpen, setModalOpen] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [featureOff, setFeatureOff] = useState(false);
  const [error, setError] = useState('');

  const load = useCallback(async () => {
    setLoading(true);
    const [settingsResponse, statusResponse, listResponse] = await Promise.all([
      HttpUtil.get<SponsorSettings>('/panel/api/fork/sponsors/settings', undefined, {
        silent: true,
      }),
      HttpUtil.get<SponsorStatus>('/panel/api/fork/sponsors/status', undefined, { silent: true }),
      HttpUtil.get<ManagedSponsor[]>('/panel/api/fork/sponsors/manage', undefined, {
        silent: true,
      }),
    ]);
    if (
      [statusResponse, listResponse].some((response) =>
        isKnownForkFeatureUnavailable(response, 'sponsors'),
      )
    ) {
      setFeatureOff(true);
      setError('');
      setLoading(false);
      return;
    }
    const failed = [settingsResponse, statusResponse, listResponse].find(
      (response) => !response.success,
    );
    if (failed) {
      setError(t('pages.sponsors.loadFailed'));
    } else {
      setSettings(settingsResponse.obj || null);
      setStatus(statusResponse.obj || null);
      setItems(listResponse.obj || []);
      setError('');
      setFeatureOff(false);
    }
    setLoading(false);
  }, [t]);

  useEffect(() => {
    const timer = window.setTimeout(() => void load(), 0);
    return () => window.clearTimeout(timer);
  }, [load]);

  useEffect(() => {
    if (settings) providerForm.setFieldsValue(settings);
  }, [providerForm, settings]);

  const openCreate = () => {
    setEditing(null);
    form.resetFields();
    form.setFieldsValue({ enabled: true, priority: 0, slots: ['page'], titleEn: '' });
    setModalOpen(true);
  };

  const openEdit = (record: ManagedSponsor) => {
    setEditing(record);
    form.setFieldsValue(inputFromRecord(record));
    setModalOpen(true);
  };

  const saveSponsor = async (values: SponsorFormValues) => {
    setSaving(true);
    const path = editing
      ? `/panel/api/fork/sponsors/${encodeURIComponent(editing.id)}`
      : '/panel/api/fork/sponsors';
    const result = editing
      ? await HttpUtil.put<ManagedSponsor>(path, sponsorRequestFromForm(values), JSON_HEADERS)
      : await HttpUtil.post<ManagedSponsor>(path, sponsorRequestFromForm(values), JSON_HEADERS);
    if (!result.success) {
      messageApi.error(t('pages.sponsors.saveFailed'));
    } else {
      messageApi.success(t('pages.sponsors.saved'));
      setModalOpen(false);
      await load();
    }
    setSaving(false);
  };

  const deleteSponsor = (record: ManagedSponsor) => {
    Modal.confirm({
      title: t('pages.sponsors.deleteConfirm', { name: record.name }),
      okText: t('delete'),
      okButtonProps: { danger: true },
      cancelText: t('cancel'),
      onOk: async () => {
        const result = await HttpUtil.delete(
          `/panel/api/fork/sponsors/${encodeURIComponent(record.id)}`,
        );
        if (!result.success) throw new Error(t('pages.sponsors.deleteFailed'));
        messageApi.success(t('pages.sponsors.deleted'));
        if (selected?.id === record.id) setSelected(null);
        await load();
      },
    });
  };

  const toggleSponsor = async (record: ManagedSponsor, enabled: boolean) => {
    const result = await HttpUtil.put<ManagedSponsor>(
      `/panel/api/fork/sponsors/${encodeURIComponent(record.id)}`,
      { ...inputFromRecord(record), enabled },
      JSON_HEADERS,
    );
    if (!result.success) messageApi.error(t('pages.sponsors.saveFailed'));
    else await load();
  };

  const saveProvider = async (values: SponsorSettings) => {
    const result = await HttpUtil.put<SponsorSettings>(
      '/panel/api/fork/sponsors/settings',
      values,
      JSON_HEADERS,
    );
    if (!result.success) messageApi.error(t('pages.sponsors.saveFailed'));
    else {
      messageApi.success(t('pages.sponsors.saved'));
      await load();
    }
  };

  const preview = useMemo(() => (selected ? previewValue(selected) : null), [selected]);
  const remoteReadOnly = settings?.providerMode === 'remote';

  return (
    <ForkAdminPageShell pageClass="sponsors-management-page">
      {contextHolder}
      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <Card size="small">
          <Space direction="vertical" size="small" style={{ width: '100%' }}>
            <Space align="center" wrap>
              <Typography.Title level={2} style={{ margin: 0 }}>
                {t('pages.sponsors.managementTitle')}
              </Typography.Title>
              <Button icon={<ReloadOutlined />} onClick={() => void load()} loading={loading} />
            </Space>
            <Typography.Paragraph type="secondary">
              {t('pages.sponsors.managementIntro')}
            </Typography.Paragraph>
            {featureOff ? (
              <FeatureOffState feature="sponsors" />
            ) : (
              error && <Alert type="error" showIcon message={error} />
            )}
          </Space>
        </Card>

        {!featureOff && (
          <>
            <Card size="small" title={t('pages.sponsors.providerStatus')} loading={loading}>
              {status && (
                <Descriptions size="small" column={{ xs: 1, sm: 2, xl: 4 }}>
                  <Descriptions.Item label={t('pages.sponsors.providerMode')}>
                    <Tag color={status.providerMode === 'local' ? 'blue' : 'gold'}>
                      {status.providerMode === 'local'
                        ? t('pages.sponsors.providerLocal')
                        : t('pages.sponsors.providerRemote')}
                    </Tag>
                  </Descriptions.Item>
                  <Descriptions.Item label={t('pages.sponsors.localCount')}>
                    {status.localSponsorCount}
                  </Descriptions.Item>
                  <Descriptions.Item label={t('pages.sponsors.activeTotal')}>
                    {status.activeSponsorCount}
                  </Descriptions.Item>
                  <Descriptions.Item label={t('pages.sponsors.cacheState')}>
                    {providerStateLabel(t, status.cacheState)}
                  </Descriptions.Item>
                  {status.lastSuccessfulFetch && (
                    <Descriptions.Item label={t('pages.sponsors.lastFetch')}>
                      {new Date(status.lastSuccessfulFetch).toLocaleString()}
                    </Descriptions.Item>
                  )}
                  {status.lastError && (
                    <Descriptions.Item label={t('pages.sponsors.lastError')} span={2}>
                      <Typography.Text type="danger">{status.lastError}</Typography.Text>
                    </Descriptions.Item>
                  )}
                </Descriptions>
              )}
            </Card>

            <Card size="small" title={t('pages.sponsors.providerSettings')}>
              {settings && (
                <Form
                  form={providerForm}
                  layout="vertical"
                  initialValues={settings}
                  onFinish={(values) => void saveProvider(values)}
                >
                  <Row gutter={16}>
                    <Col xs={24} md={8}>
                      <Form.Item name="providerMode" label={t('pages.sponsors.providerMode')}>
                        <Select
                          options={[
                            { value: 'local', label: t('pages.sponsors.providerLocal') },
                            { value: 'remote', label: t('pages.sponsors.providerRemote') },
                          ]}
                        />
                      </Form.Item>
                    </Col>
                    <Col xs={24} md={8}>
                      <Form.Item name="sourceUrl" label={t('pages.sponsors.sourceUrl')}>
                        <Input type="url" placeholder="https://" />
                      </Form.Item>
                    </Col>
                    <Col xs={24} md={8}>
                      <Form.Item name="contactUrl" label={t('pages.sponsors.contactUrl')}>
                        <Input type="url" placeholder="https://" />
                      </Form.Item>
                    </Col>
                  </Row>
                  <Button type="primary" htmlType="submit">
                    {t('save')}
                  </Button>
                  {remoteReadOnly && (
                    <Alert
                      style={{ marginTop: 12 }}
                      type="info"
                      showIcon
                      message={t('pages.sponsors.remoteReadOnly')}
                    />
                  )}
                </Form>
              )}
            </Card>

            <Card
              size="small"
              title={t('pages.sponsors.listTitle')}
              extra={
                <Button
                  type="primary"
                  icon={<PlusOutlined />}
                  onClick={openCreate}
                  disabled={remoteReadOnly}
                >
                  {t('create')}
                </Button>
              }
            >
              <Table
                rowKey="id"
                loading={loading}
                dataSource={items}
                locale={{ emptyText: <Empty description={t('pages.sponsors.noItems')} /> }}
                columns={[
                  { title: t('pages.sponsors.name'), dataIndex: 'name' },
                  { title: t('pages.sponsors.priority'), dataIndex: 'priority', width: 100 },
                  {
                    title: t('pages.sponsors.slots'),
                    render: (_: unknown, row: ManagedSponsor) =>
                      row.slots.map((slot) => <Tag key={slot}>{t(slotLabelKey[slot])}</Tag>),
                  },
                  {
                    title: t('fork.common.status'),
                    render: (_: unknown, row: ManagedSponsor) => (
                      <Switch
                        checked={row.enabled}
                        disabled={remoteReadOnly}
                        onChange={(value) => void toggleSponsor(row, value)}
                      />
                    ),
                  },
                  {
                    title: t('fork.common.actions'),
                    render: (_: unknown, row: ManagedSponsor) => (
                      <Space>
                        <Button icon={<EyeOutlined />} onClick={() => setSelected(row)} />
                        <Button
                          icon={<EditOutlined />}
                          onClick={() => openEdit(row)}
                          disabled={remoteReadOnly}
                        />
                        <Button
                          danger
                          icon={<DeleteOutlined />}
                          onClick={() => deleteSponsor(row)}
                          disabled={remoteReadOnly}
                        />
                      </Space>
                    ),
                  },
                ]}
                pagination={{ pageSize: 10 }}
              />
            </Card>

            {preview && selected && (
              <Card size="small" title={t('pages.sponsors.preview')}>
                <Tabs
                  items={[
                    {
                      key: 'page',
                      label: t('pages.sponsors.placementPage'),
                      children: <SponsorCard sponsor={preview} variant="card" />,
                    },
                    {
                      key: 'sidebar',
                      label: t('pages.sponsors.placementSidebar'),
                      children: <SponsorCard sponsor={preview} variant="compact" />,
                    },
                    {
                      key: 'dashboard',
                      label: t('pages.sponsors.placementDashboard'),
                      children: <SponsorCard sponsor={preview} variant="banner" />,
                    },
                  ]}
                />
              </Card>
            )}
          </>
        )}
      </Space>

      <Modal
        open={modalOpen}
        title={editing ? t('edit') : t('pages.sponsors.createTitle')}
        okText={t('save')}
        cancelText={t('cancel')}
        confirmLoading={saving}
        onCancel={() => setModalOpen(false)}
        onOk={() => void form.submit()}
        destroyOnClose
        width={760}
      >
        <Form form={form} layout="vertical" onFinish={(values) => void saveSponsor(values)}>
          <Row gutter={16}>
            <Col xs={24} md={16}>
              <Form.Item
                name="name"
                label={t('pages.sponsors.name')}
                rules={[{ required: true, max: 160 }]}
              >
                <Input />
              </Form.Item>
            </Col>
            <Col xs={24} md={8}>
              <Form.Item
                name="priority"
                label={t('pages.sponsors.priority')}
                rules={[{ required: true }]}
              >
                <InputNumber style={{ width: '100%' }} min={-10000} max={10000} />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item
            name="slots"
            label={t('pages.sponsors.slots')}
            rules={[{ required: true, type: 'array', min: 1 }]}
          >
            <Checkbox.Group options={slots.map((slot) => ({ label: slot, value: slot }))} />
          </Form.Item>
          <Form.Item name="window" label={t('pages.sponsors.activePeriod')}>
            <DatePicker.RangePicker showTime style={{ width: '100%' }} />
          </Form.Item>
          <Row gutter={16}>
            <Col xs={24} md={12}>
              <Form.Item
                name="destinationUrl"
                label={t('pages.sponsors.destinationUrl')}
                rules={[{ required: true, type: 'url' }]}
              >
                <Input type="url" placeholder="https://" />
              </Form.Item>
            </Col>
            <Col xs={24} md={12}>
              <Form.Item
                name="logoUrl"
                label={t('pages.sponsors.logoUrl')}
                rules={[{ type: 'url' }]}
              >
                <Input type="url" placeholder="https://" />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item name="enabled" label={t('fork.common.status')} valuePropName="checked">
            <Switch checkedChildren={t('enabled')} unCheckedChildren={t('disabled')} />
          </Form.Item>
          <Typography.Title level={5}>{t('pages.sponsors.localization')}</Typography.Title>
          <Row gutter={16}>
            <Col xs={24} md={8}>
              <Form.Item name="titleEn" label="en-US" rules={[{ required: true, max: 500 }]}>
                <Input />
              </Form.Item>
              <Form.Item name="textEn" label={t('pages.sponsors.text')}>
                <Input.TextArea rows={3} maxLength={500} />
              </Form.Item>
            </Col>
            <Col xs={24} md={8}>
              <Form.Item name="titleRu" label="ru-RU">
                <Input />
              </Form.Item>
              <Form.Item name="textRu" label={t('pages.sponsors.text')}>
                <Input.TextArea rows={3} maxLength={500} />
              </Form.Item>
            </Col>
            <Col xs={24} md={8}>
              <Form.Item name="titleFa" label="fa-IR">
                <Input />
              </Form.Item>
              <Form.Item name="textFa" label={t('pages.sponsors.text')}>
                <Input.TextArea rows={3} maxLength={500} />
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Modal>
    </ForkAdminPageShell>
  );
}
