import { useCallback, useEffect, useState } from 'react';
import { Alert, Button, Card, Drawer, Empty, Input, Space, Table, Tag, Typography } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { HttpUtil } from '@/utils';
import FeatureOffState from '@/components/fork/FeatureOffState';
import { isKnownForkFeatureUnavailable } from '@/lib/fork-feature';
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
  const [featureOff, setFeatureOff] = useState(false);
  const load = useCallback(async () => {
    const result = await HttpUtil.get<{ items: AuditEvent[] }>(
      '/panel/api/fork/audit/events',
      {
        eventType: eventType || undefined,
        limit: 100,
      },
      { silent: true },
    );
    if (isKnownForkFeatureUnavailable(result, 'audit')) {
      setFeatureOff(true);
      setError('');
    } else if (result.success && result.obj) setRows(result.obj.items);
    else setError(result.msg || t('fork.audit.labels.error'));
  }, [eventType, t]);
  useEffect(() => {
    const timer = window.setTimeout(() => void load(), 0);
    return () => window.clearTimeout(timer);
  }, [load]);
  const columns: ColumnsType<AuditEvent> = [
    {
      title: t('fork.audit.labels.time'),
      dataIndex: 'createdAt',
      render: (value: string) => new Date(value).toLocaleString(),
    },
    { title: t('fork.audit.labels.action'), dataIndex: 'eventType' },
    {
      title: t('fork.audit.labels.outcome'),
      dataIndex: 'outcome',
      render: (value: string) => <Tag color={value === 'success' ? 'green' : 'red'}>{value}</Tag>,
    },
    { title: t('fork.audit.labels.actor'), render: (_, row) => row.actorName || row.actorType },
    { title: t('fork.audit.labels.target'), render: (_, row) => row.targetType || '—' },
    { title: 'Node', render: (_, row) => row.nodeScope || 'local' },
    { title: t('fork.audit.labels.request'), dataIndex: 'requestId', ellipsis: true },
    {
      title: t('fork.audit.details'),
      render: (_, row) => (
        <Button onClick={() => setSelected(row)}>{t('fork.audit.labels.view')}</Button>
      ),
    },
  ];
  return (
    <div className="audit-page">
      <div className="content-area">
        <Card>
          <Space direction="vertical" style={{ width: '100%' }} size="large">
            <div>
              <Typography.Title level={2}>{t('fork.audit.title')}</Typography.Title>
              <Typography.Text type="secondary">{t('fork.audit.summary')}</Typography.Text>
            </div>
            <Space>
              <Input
                placeholder={t('fork.audit.labels.eventType')}
                value={eventType}
                onChange={(e) => setEventType(e.target.value)}
                onPressEnter={() => void load()}
              />
              <Button type="primary" onClick={() => void load()}>
                {t('fork.audit.filter')}
              </Button>
            </Space>
            {featureOff ? <FeatureOffState /> : error && <Alert type="error" message={error} />}
            {!featureOff && rows.length ? (
              <Table
                rowKey="id"
                columns={columns}
                dataSource={rows}
                pagination={{ pageSize: 25 }}
              />
            ) : !featureOff ? (
              <Empty description={t('fork.audit.empty')} />
            ) : null}
          </Space>
          <Drawer
            title={t('fork.audit.labels.event')}
            open={Boolean(selected)}
            onClose={() => setSelected(null)}
            width={520}
          >
            <pre style={{ whiteSpace: 'pre-wrap' }}>
              {selected ? JSON.stringify(selected, null, 2) : ''}
            </pre>
          </Drawer>
        </Card>
      </div>
    </div>
  );
}
