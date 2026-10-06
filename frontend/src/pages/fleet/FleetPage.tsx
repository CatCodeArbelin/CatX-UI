import { Alert, Button, Card, Col, Empty, Row, Space, Statistic, Tag, Typography } from 'antd';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router';
import { useNodesQuery } from '@/api/queries/useNodesQuery';
import ForkAdminPageShell from '@/components/fork/ForkAdminPageShell';
import { ForkModuleTitle } from '@/components/fork/ForkModuleMaturity';

export default function FleetPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { nodes, loading, fetchError } = useNodesQuery();
  const nodeStatusLabel = (status?: string) =>
    status
      ? t(`fork.fleet.nodeStates.${status}`, { defaultValue: t('fork.common.unknown') })
      : t('fork.common.unknown');
  return (
    <ForkAdminPageShell pageClass="fleet-page">
      <Card
        size="small"
        loading={loading}
        title={<ForkModuleTitle moduleId="M09" title={t('fork.fleet.title')} />}
      >
        {fetchError && <Alert type="error" showIcon message={fetchError} />}
        {nodes.length ? (
          <Row gutter={[16, 16]}>
            {nodes.map((node) => (
              <Col xs={24} md={12} xl={8} key={node.id}>
                <Card size="small" title={node.name || node.remark}>
                  <Tag color={node.status === 'online' ? 'green' : 'default'}>
                    {nodeStatusLabel(node.status)}
                  </Tag>
                  <Row gutter={8} style={{ marginTop: 16 }}>
                    <Col span={8}>
                      <Statistic title={t('fork.fleet.clients')} value={node.clientCount} />
                    </Col>
                    <Col span={8}>
                      <Statistic title={t('fork.fleet.online')} value={node.onlineCount} />
                    </Col>
                    <Col span={8}>
                      <Statistic
                        title={t('fork.fleet.latency')}
                        value={node.latencyMs}
                        suffix="ms"
                      />
                    </Col>
                  </Row>
                  <Typography.Text type="secondary">
                    {t('fork.fleet.lastHeartbeat')}:{' '}
                    {node.lastHeartbeat ? new Date(node.lastHeartbeat).toLocaleString() : '—'}
                  </Typography.Text>
                </Card>
              </Col>
            ))}
          </Row>
        ) : (
          <Empty
            description={
              <Space direction="vertical" size="small">
                <Typography.Text>{t('fork.fleet.emptyTitle')}</Typography.Text>
                <Typography.Text type="secondary">
                  {t('fork.fleet.emptyDescription')}
                </Typography.Text>
                <Button type="primary" onClick={() => navigate('/nodes')}>
                  {t('fork.fleet.openNodes')}
                </Button>
              </Space>
            }
          />
        )}
      </Card>
    </ForkAdminPageShell>
  );
}
