import { Alert, Card, Col, Empty, Row, Statistic, Tag, Typography } from 'antd';
import { useTranslation } from 'react-i18next';
import { useNodesQuery } from '@/api/queries/useNodesQuery';

export default function FleetPage() {
  const { t } = useTranslation();
  const { nodes, loading, fetchError } = useNodesQuery();
  return (
    <div className="fleet-page">
      <div className="content-area">
        <Card loading={loading} title={t('fork.fleet.title')}>
          {fetchError && <Alert type="error" showIcon message={fetchError} />}
          {nodes.length ? (
            <Row gutter={[16, 16]}>
              {nodes.map((node) => (
                <Col xs={24} md={12} xl={8} key={node.id}>
                  <Card title={node.name || node.remark}>
                    <Tag color={node.status === 'online' ? 'green' : 'default'}>
                      {node.status || t('fork.common.unknown')}
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
            <Empty description={t('fork.common.noItems')} />
          )}
        </Card>
      </div>
    </div>
  );
}
