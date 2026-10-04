import { Alert, Space, Tag, Typography } from 'antd';
import type { ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import type { ForkFeature, ForkRuntimeState } from '@/lib/fork-feature';

const featureLabels: Record<ForkFeature, string> = {
  analytics: 'fork.activity.title',
  dns_intelligence: 'fork.common.labels.enableDns',
  policies: 'fork.policy.title',
  risk: 'fork.apiDocs.risk',
  audit: 'fork.audit.title',
  webhooks: 'fork.webhooks.title',
  self_service: 'fork.apiDocs.portal',
  fleet_updates: 'fork.fleetUpdate.title',
  sponsors: 'fork.sponsors.title',
  traffic_control: 'fork.apiDocs.traffic',
};

export default function CatxState({
  state,
  feature,
  title,
  description,
  action,
}: {
  state: ForkRuntimeState;
  feature?: ForkFeature;
  title?: ReactNode;
  description?: ReactNode;
  action?: ReactNode;
}) {
  const { t } = useTranslation();
  const featureLabel = feature ? t(featureLabels[feature]) : t('fork.common.name');
  const stateLabel = t(`fork.common.states.${state}`);
  const message =
    title ||
    (state === 'feature_off'
      ? t('fork.common.featureOffTitle', { feature: featureLabel })
      : stateLabel);
  const detail =
    description ||
    (state === 'feature_off'
      ? t('fork.common.featureOffDescription')
      : state === 'restart_required'
        ? t('fork.settings.restartRequiredDescription')
        : undefined);

  if (state === 'active') return <Tag color="green">{stateLabel}</Tag>;

  const details =
    detail || action ? (
      <Space direction="vertical" size={2}>
        {detail && <Typography.Text>{detail}</Typography.Text>}
        {action}
      </Space>
    ) : undefined;

  return (
    <Alert
      type={state === 'error' ? 'error' : state === 'degraded' ? 'warning' : 'info'}
      showIcon
      message={
        <Space size="small">
          <Tag>{stateLabel}</Tag>
          <Typography.Text>{message}</Typography.Text>
        </Space>
      }
      description={details}
    />
  );
}
