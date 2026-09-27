import { useCallback, useEffect, useState } from 'react';
import { Alert, Button, Card, Descriptions, List, Space, Tag, Typography, message } from 'antd';
import { useTranslation } from 'react-i18next';
import { HttpUtil } from '@/utils';

type RiskEvent = {
  id: number;
  kind: string;
  observedAt: number;
  scoreContribution: number;
  confidence: number;
  state: string;
  evidence?: string;
  sourceNode?: string;
  acknowledgedAt?: number;
};

type IPHistory = {
  id: number;
  ip: string;
  nodeGuid?: string;
  observedAt: number;
  country?: string;
  asn?: number;
  metadataState: string;
};

type RiskSummary = {
  enabled: boolean;
  clientEmail: string;
  score: number;
  confidence: number;
  state: string;
  evidence: number;
  calculatedAt: number;
  events: RiskEvent[];
  ipHistory: IPHistory[];
};

export default function RiskIntelligencePanel({ email }: { email: string }) {
  const { t } = useTranslation();
  const [summary, setSummary] = useState<RiskSummary | null>(null);
  const [loading, setLoading] = useState(false);
  const [messageApi, contextHolder] = message.useMessage();

  const load = useCallback(async () => {
    try {
      const result = await HttpUtil.get<{ success: boolean; obj: RiskSummary }>(
        `/panel/api/risk/clients/${encodeURIComponent(email)}`,
        undefined,
        { silent: true },
      );
      const raw = result?.obj;
      setSummary(raw?.success ? raw.obj : null);
    } finally {
      setLoading(false);
    }
  }, [email]);

  useEffect(() => {
    void load();
  }, [load]);

  const acknowledge = async (id: number) => {
    await HttpUtil.post(`/panel/api/risk/clients/${encodeURIComponent(email)}/events/${id}/ack`);
    messageApi.success(t('pages.clients.risk.acknowledged'));
    await load();
  };

  if ((!summary || !summary.enabled) && !loading) return null;
  if (!summary) return <Card loading size="small" />;

  return (
    <Card title={t('pages.clients.risk.title')} size="small" loading={loading}>
      {contextHolder}
      <Space direction="vertical" style={{ width: '100%' }}>
        <Descriptions size="small" column={2}>
          <Descriptions.Item label={t('pages.clients.risk.score')}>
            <Tag
              color={
                summary.score >= 70 && summary.confidence >= 0.75 && summary.evidence >= 2
                  ? 'orange'
                  : 'blue'
              }
            >
              {summary.score}/100
            </Tag>
          </Descriptions.Item>
          <Descriptions.Item label={t('pages.clients.risk.confidence')}>
            {Math.round(summary.confidence * 100)}%
          </Descriptions.Item>
          <Descriptions.Item label={t('pages.clients.risk.state')}>
            {t(`pages.clients.risk.states.${summary.state}`, summary.state)}
          </Descriptions.Item>
          <Descriptions.Item label={t('pages.clients.risk.evidence')}>
            {summary.evidence}
          </Descriptions.Item>
        </Descriptions>
        {summary.state === 'insufficient_evidence' && (
          <Alert type="info" showIcon message={t('pages.clients.risk.insufficient')} />
        )}
        <Typography.Text type="secondary">{t('pages.clients.risk.timeline')}</Typography.Text>
        <List
          size="small"
          dataSource={summary.events}
          locale={{ emptyText: t('pages.clients.risk.noEvents') }}
          renderItem={(event) => (
            <List.Item
              actions={
                event.acknowledgedAt
                  ? [<Tag key="ack">{t('pages.clients.risk.acknowledged')}</Tag>]
                  : [
                      <Button key="ack" size="small" onClick={() => void acknowledge(event.id)}>
                        {t('pages.clients.risk.acknowledge')}
                      </Button>,
                    ]
              }
            >
              <List.Item.Meta
                title={`${t(`pages.clients.risk.signals.${event.kind}`, event.kind)} +${event.scoreContribution}`}
                description={`${t('pages.clients.risk.confidence')}: ${Math.round(event.confidence * 100)}% · ${event.sourceNode || t('pages.clients.risk.unknownSource')} · ${event.evidence || t('pages.clients.risk.noEvidence')}`}
              />
            </List.Item>
          )}
        />
        <Typography.Text type="secondary">{t('pages.clients.risk.ipHistory')}</Typography.Text>
        <List
          size="small"
          dataSource={summary.ipHistory.slice(0, 12)}
          locale={{ emptyText: t('pages.clients.risk.noIpHistory') }}
          renderItem={(entry) => (
            <List.Item>
              <List.Item.Meta
                title={entry.ip}
                description={`${entry.country || t('pages.clients.risk.unknown')} · ASN ${entry.asn || t('pages.clients.risk.unknown')} · ${entry.nodeGuid || t('pages.clients.risk.unknownSource')}`}
              />
            </List.Item>
          )}
        />
      </Space>
    </Card>
  );
}
