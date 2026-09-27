import { useEffect, useState } from 'react';
import { Alert, Button, Card, Drawer, Empty, Input, Space, Table, Tag, Typography } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { HttpUtil } from '@/utils';
import { useTranslation } from 'react-i18next';

type AuditEvent = {
  id: string;
  createdAt: string;
  eventType: string;
  outcome: string;
  actorType: string;
  actorId?: string;
  actorName?: string;
  authMethod?: string;
  targetType?: string;
  targetRef?: string;
  nodeScope?: string;
  nodeId?: string;
  requestId: string;
  sourceIp?: string;
  statusCode?: number;
  metadata?: string;
};

export default function AuditPage() {
  const { t } = useTranslation();
  const [rows, setRows] = useState<AuditEvent[]>([]);
  const [eventType, setEventType] = useState('');
  const [selected, setSelected] = useState<AuditEvent | null>(null);
  const [error, setError] = useState('');
  const load = async () => {
    const result = await HttpUtil.get<{ items: AuditEvent[] }>(
      '/panel/api/fork/audit/events',
      {
        eventType: eventType || undefined,
        limit: 100,
      },
      { silent: true },
    );
    if (result.success && result.obj) setRows(result.obj.items);
    else setError(result.msg || t('fork.audit.error', 'Audit events could not be loaded.'));
  };
  useEffect(() => {
    void load();
  }, []);
  const columns: ColumnsType<AuditEvent> = [
    {
      title: 'Time',
      dataIndex: 'createdAt',
      render: (value: string) => new Date(value).toLocaleString(),
    },
    { title: 'Action', dataIndex: 'eventType' },
    {
      title: 'Outcome',
      dataIndex: 'outcome',
      render: (value: string) => <Tag color={value === 'success' ? 'green' : 'red'}>{value}</Tag>,
    },
    { title: 'Actor', render: (_, row) => row.actorName || row.actorType },
    { title: 'Target', render: (_, row) => row.targetType || '—' },
    { title: 'Node', render: (_, row) => row.nodeScope || 'local' },
    { title: 'Request', dataIndex: 'requestId', ellipsis: true },
    {
      title: t('fork.audit.details'),
      render: (_, row) => (
        <Button onClick={() => setSelected(row)}>{t('fork.audit.view', 'View')}</Button>
      ),
    },
  ];
  return (
    <Card>
      <Space direction="vertical" style={{ width: '100%' }} size="large">
        <div>
          <Typography.Title level={2}>{t('fork.audit.title')}</Typography.Title>
          <Typography.Text type="secondary">{t('fork.audit.description')}</Typography.Text>
        </div>
        <Space>
          <Input
            placeholder={t('fork.audit.eventType', 'Event type')}
            value={eventType}
            onChange={(e) => setEventType(e.target.value)}
            onPressEnter={() => void load()}
          />
          <Button type="primary" onClick={() => void load()}>
            {t('fork.audit.filter')}
          </Button>
        </Space>
        {error && <Alert type="error" message={error} />}
        {rows.length ? (
          <Table rowKey="id" columns={columns} dataSource={rows} pagination={{ pageSize: 25 }} />
        ) : (
          <Empty description={t('fork.audit.empty')} />
        )}
      </Space>
      <Drawer
        title={t('fork.audit.event', 'Audit event')}
        open={Boolean(selected)}
        onClose={() => setSelected(null)}
        width={520}
      >
        <pre style={{ whiteSpace: 'pre-wrap' }}>
          {selected ? JSON.stringify(selected, null, 2) : ''}
        </pre>
      </Drawer>
    </Card>
  );
}
