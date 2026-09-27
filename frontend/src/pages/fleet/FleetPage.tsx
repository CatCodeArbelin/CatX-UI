import { Card, Col, Row, Statistic, Tag, Typography } from 'antd';
import { useTranslation } from 'react-i18next';
import { useNodesQuery } from '@/api/queries/useNodesQuery';

export default function FleetPage() {
  const { t } = useTranslation();
  const { nodes, loading, fetchError } = useNodesQuery();
  return <Card loading={loading}><Typography.Title level={2}>{t('fork.fleet.title', 'Fleet')}</Typography.Title>{fetchError && <Typography.Text type="danger">{fetchError}</Typography.Text>}<Row gutter={[16, 16]}>{nodes.map((node) => <Col xs={24} md={12} xl={8} key={node.id}><Card title={node.name || node.remark}><Tag color={node.status === 'online' ? 'green' : 'default'}>{node.status || t('unknown', 'Unknown')}</Tag><Row gutter={8} style={{ marginTop: 16 }}><Col span={8}><Statistic title={t('fork.fleet.clients', 'Clients')} value={node.clientCount} /></Col><Col span={8}><Statistic title={t('fork.fleet.online', 'Online')} value={node.onlineCount} /></Col><Col span={8}><Statistic title={t('fork.fleet.latency', 'Latency')} value={node.latencyMs} suffix="ms" /></Col></Row><Typography.Text type="secondary">{t('fork.fleet.lastHeartbeat', 'Last heartbeat')}: {node.lastHeartbeat ? new Date(node.lastHeartbeat).toLocaleString() : '—'}</Typography.Text></Card></Col>)}</Row></Card>;
}
