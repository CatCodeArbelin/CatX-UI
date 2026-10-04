import { Alert, Button, Space, Typography } from 'antd';
import { SettingOutlined } from '@ant-design/icons';
import { useInRouterContext, useNavigate } from 'react-router';
import { useTranslation } from 'react-i18next';
import type { ForkFeature } from '@/lib/fork-feature';

function FeatureSettingsLink() {
  const navigate = useNavigate();
  const { t } = useTranslation();
  return (
    <Button
      type="link"
      icon={<SettingOutlined />}
      aria-label={t('fork.settings.title')}
      onClick={() => navigate('/settings#catx-features')}
    >
      {t('fork.settings.title')}
    </Button>
  );
}

export default function FeatureOffState({
  feature,
  messageKey,
}: {
  feature?: ForkFeature;
  messageKey?: string;
}) {
  const { t } = useTranslation();
  const inRouter = useInRouterContext();
  const labels: Record<ForkFeature, string> = {
    analytics: 'fork.activity.title',
    dns_intelligence: 'fork.common.labels.enableDns',
    policies: 'fork.policy.title',
    risk: 'fork.apiDocs.risk',
    audit: 'fork.audit.title',
    webhooks: 'fork.webhooks.title',
    self_service: 'fork.apiDocs.portal',
    fleet_updates: 'fork.fleetUpdate.title',
    sponsors: 'fork.sponsors.title',
  };
  const featureLabel = feature ? t(labels[feature]) : t('fork.common.featureOffTitle');
  return (
    <Alert
      type="info"
      showIcon
      message={
        messageKey ? t(messageKey) : t('fork.common.featureOffTitle', { feature: featureLabel })
      }
      description={
        <Space direction="vertical" size={2}>
          <Typography.Text>{t('fork.common.featureOffDescription')}</Typography.Text>
          {inRouter && <FeatureSettingsLink />}
        </Space>
      }
    />
  );
}
