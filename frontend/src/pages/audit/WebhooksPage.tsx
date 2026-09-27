import { useCallback, useEffect, useState } from 'react';
import { Alert, Button, Card, Form, Input, Modal, Space, Table, Tag, Typography } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { HttpUtil } from '@/utils';
import { useTranslation } from 'react-i18next';

type Endpoint = { id: string; name: string; url: string; eventTypes: string; enabled: boolean };
type Delivery = {
  id: string;
  eventId: string;
  endpointId: string;
  status: string;
  attempt: number;
  lastError?: string;
};

export default function WebhooksPage() {
  const { t } = useTranslation();
  const [endpoints, setEndpoints] = useState<Endpoint[]>([]);
  const [deliveries, setDeliveries] = useState<Delivery[]>([]);
  const [error, setError] = useState('');
  const [form] = Form.useForm();
  const load = useCallback(async () => {
    const [endpointResult, deliveryResult] = await Promise.all([
      HttpUtil.get<Endpoint[]>('/panel/api/fork/audit/webhooks', undefined, { silent: true }),
      HttpUtil.get<{ items: Delivery[] }>(
        '/panel/api/fork/audit/webhooks/deliveries',
        { limit: 50 },
        { silent: true },
      ),
    ]);
    if (endpointResult.success && endpointResult.obj) setEndpoints(endpointResult.obj);
    if (deliveryResult.success && deliveryResult.obj) setDeliveries(deliveryResult.obj.items);
    if (!endpointResult.success || !deliveryResult.success)
      setError(
        endpointResult.msg ||
          deliveryResult.msg ||
          t('fork.webhooks.error', 'Webhook data could not be loaded.'),
      );
  }, [t]);
  useEffect(() => {
    const timer = window.setTimeout(() => void load(), 0);
    return () => window.clearTimeout(timer);
  }, [load]);
  const create = async (value: {
    name: string;
    url: string;
    secret: string;
    eventTypes: string;
  }) => {
    const result = await HttpUtil.post('/panel/api/fork/audit/webhooks', {
      ...value,
      eventTypes: value.eventTypes
        .split(',')
        .map((item) => item.trim())
        .filter(Boolean),
    });
    if (result.success) {
      form.resetFields();
      void load();
    }
  };
  const deleteEndpoint = (id: string) =>
    Modal.confirm({
      title: t('fork.webhooks.confirmDelete', 'Delete webhook?'),
      okType: 'danger',
      onOk: async () => {
        await HttpUtil.delete(`/panel/api/fork/audit/webhooks/${id}`);
        void load();
      },
    });
  const replay = async (id: string) => {
    await HttpUtil.post(`/panel/api/fork/audit/webhooks/deliveries/${id}/replay`);
    void load();
  };
  const endpointColumns: ColumnsType<Endpoint> = [
    { title: 'Name', dataIndex: 'name' },
    { title: 'Destination', dataIndex: 'url' },
    {
      title: 'State',
      render: (_, row) => (
        <Tag color={row.enabled ? 'green' : 'default'}>{row.enabled ? 'Enabled' : 'Disabled'}</Tag>
      ),
    },
    {
      title: t('fork.webhooks.actions', 'Actions'),
      render: (_, row) => (
        <Button danger onClick={() => deleteEndpoint(row.id)}>
          {t('fork.webhooks.delete')}
        </Button>
      ),
    },
  ];
  const deliveryColumns: ColumnsType<Delivery> = [
    { title: 'Delivery', dataIndex: 'id' },
    { title: 'Status', dataIndex: 'status' },
    { title: 'Attempts', dataIndex: 'attempt' },
    { title: 'Error', dataIndex: 'lastError' },
    {
      title: t('fork.webhooks.replay'),
      render: (_, row) => (
        <Button onClick={() => void replay(row.id)}>{t('fork.webhooks.replay')}</Button>
      ),
    },
  ];
  return (
    <Space direction="vertical" style={{ width: '100%' }} size="large">
      <Card>
        <Typography.Title level={2}>{t('fork.webhooks.title')}</Typography.Title>
        <Typography.Text type="secondary">{t('fork.webhooks.summary')}</Typography.Text>
        <Form
          form={form}
          layout="vertical"
          onFinish={(value) => void create(value)}
          style={{ marginTop: 24 }}
        >
          <Space wrap align="start">
            <Form.Item
              name="name"
              label={t('fork.webhooks.name', 'Name')}
              rules={[{ required: true }]}
            >
              <Input />
            </Form.Item>
            <Form.Item
              name="url"
              label={t('fork.webhooks.url', 'Public HTTPS URL')}
              rules={[{ required: true, type: 'url' }]}
            >
              <Input style={{ width: 300 }} />
            </Form.Item>
            <Form.Item
              name="secret"
              label={t('fork.webhooks.secret', 'Secret')}
              rules={[{ required: true }]}
            >
              <Input.Password />
            </Form.Item>
            <Form.Item
              name="eventTypes"
              label={t('fork.webhooks.events', 'Events')}
              rules={[{ required: true }]}
            >
              <Input placeholder="* or auth.login,http.POST" />
            </Form.Item>
            <Form.Item>
              <Button type="primary" htmlType="submit">
                {t('fork.webhooks.create')}
              </Button>
            </Form.Item>
          </Space>
        </Form>
      </Card>
      {error && <Alert type="error" message={error} />}
      <Card title={t('fork.webhooks.destinations')}>
        <Table rowKey="id" columns={endpointColumns} dataSource={endpoints} pagination={false} />
      </Card>
      <Card title={t('fork.webhooks.deliveries')}>
        <Table
          rowKey="id"
          columns={deliveryColumns}
          dataSource={deliveries}
          pagination={{ pageSize: 25 }}
        />
      </Card>
    </Space>
  );
}
